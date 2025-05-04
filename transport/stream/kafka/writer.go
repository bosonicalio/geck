package kafka

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/tesserical/geck/transport/stream"
)

// - Asynchronous Writer -

// AsyncWriter is a Kafka [stream.Writer] that writes messages asynchronously.
//
// Compared with the [SyncWriter] option, this writer is more efficient for high-throughput scenarios
// where the application does not need to wait for the result of the write operation.
type AsyncWriter struct {
	client *kgo.Client
	opts   asyncStreamWriterOptions
}

var (
	// compile-time assertions
	_ stream.Writer = (*AsyncWriter)(nil)
)

// NewAsyncWriter creates a new instance of [AsyncWriter].
func NewAsyncWriter(client *kgo.Client, opts ...AsyncWriterOption) AsyncWriter {
	o := asyncStreamWriterOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	return AsyncWriter{
		client: client,
		opts:   o,
	}
}

func (s AsyncWriter) Write(ctx context.Context, name string, message stream.Message) error {
	s.client.Produce(ctx, &kgo.Record{
		Key:     []byte(message.Key),
		Value:   message.Data,
		Headers: marshalHeaders(message.Header),
		Topic:   name,
	}, s.opts.handler)
	return nil
}

func (s AsyncWriter) WriteBatch(ctx context.Context, name string, messages []stream.Message) (int, error) {
	for _, m := range messages {
		s.client.Produce(ctx, &kgo.Record{
			Key:     []byte(m.Key),
			Value:   m.Data,
			Headers: marshalHeaders(m.Header),
			Topic:   name,
		}, s.opts.handler)
	}
	return 0, nil
}

// -- Handler(s) --

// AsyncWriterHandler is a function that intercepts the result of a produced message.
type AsyncWriterHandler func(record *kgo.Record, err error)

// NewAsyncWriterLogHandler creates a new instance of [AsyncWriterHandler] that logs the result of a
// produced message.
func NewAsyncWriterLogHandler(logger *slog.Logger) AsyncWriterHandler {
	return func(r *kgo.Record, err error) {
		if err != nil {
			logger.ErrorContext(r.Context, "stream.writer.kafka: failed to produce message",
				slog.String("err", err.Error()),
				slog.String("topic", r.Topic),
				slog.String("key", string(r.Key)),
				slog.Int("data_size", len(r.Value)),
			)
			return
		}
		logger.DebugContext(r.Context, "stream.writer.kafka: message produced",
			slog.String("topic", r.Topic),
			slog.String("key", string(r.Key)),
			slog.Int("data_size", len(r.Value)),
		)
	}
}

// -- Option(s) --

type asyncStreamWriterOptions struct {
	handler AsyncWriterHandler
}

// AsyncWriterOption is a function that modifies the behavior of an [AsyncWriter].
type AsyncWriterOption func(*asyncStreamWriterOptions)

// WithAsyncWriterHandler sets the handler for an [AsyncWriter].
func WithAsyncWriterHandler(handler AsyncWriterHandler) AsyncWriterOption {
	return func(o *asyncStreamWriterOptions) {
		o.handler = handler
	}
}

// - Synchronous Writer -

// SyncWriter is a Kafka stream writer that writes messages synchronously.
//
// Compared with the [AsyncWriter] option, this writer is more suitable for scenarios where the application
// needs to wait for the result of the write operation.
type SyncWriter struct {
	client *kgo.Client
}

var (
	// compile-time assertions
	_ stream.Writer = (*SyncWriter)(nil)
)

// NewSyncWriter creates a new instance of [SyncWriter].
func NewSyncWriter(client *kgo.Client) SyncWriter {
	return SyncWriter{client: client}
}

func (s SyncWriter) Write(ctx context.Context, name string, message stream.Message) error {
	return s.client.ProduceSync(ctx, &kgo.Record{
		Key:     []byte(message.Key),
		Value:   message.Data,
		Headers: marshalHeaders(message.Header),
		Topic:   name,
		Context: ctx,
	}).FirstErr()
}

func (s SyncWriter) WriteBatch(ctx context.Context, name string, messages []stream.Message) (int, error) {
	buf := make([]*kgo.Record, 0, len(messages))
	for _, m := range messages {
		buf = append(buf, &kgo.Record{
			Key:     []byte(m.Key),
			Value:   m.Data,
			Headers: marshalHeaders(m.Header),
			Topic:   name,
			Context: ctx,
		})
	}

	err := s.client.ProduceSync(ctx, buf...).FirstErr()
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}

// - Transactional Writer -

type TransactionalWriter struct {
	client *kgo.Client
}

var (
	// compile-time assertions
	_ stream.Writer = (*TransactionalWriter)(nil)
)

// NewTransactionalWriter creates a new instance of [TransactionalWriter].
func NewTransactionalWriter(client *kgo.Client) TransactionalWriter {
	return TransactionalWriter{client: client}
}

// writeInTransaction handles a complete transaction with multiple writes.
func (s TransactionalWriter) writeInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := s.client.BeginTransaction(); err != nil {
		return err
	}

	if err := fn(ctx); err != nil {
		return err
	}

	if err := s.client.Flush(ctx); err != nil {
		return err
	}

	err := s.client.EndTransaction(ctx, kgo.TryCommit)
	if err == nil {
		return nil
	} else if errors.Is(err, kerr.OperationNotAttempted) {
		return err
	}

	if err = s.client.AbortBufferedRecords(ctx); err != nil {
		return err
	}
	return s.client.EndTransaction(ctx, kgo.TryAbort)
}

func (s TransactionalWriter) Write(ctx context.Context, name string, message stream.Message) error {
	return s.writeInTransaction(ctx, func(ctx context.Context) error {
		return s.client.ProduceSync(ctx, &kgo.Record{
			Key:       []byte(message.Key),
			Value:     message.Data,
			Headers:   marshalHeaders(message.Header),
			Timestamp: time.Time{},
			Topic:     name,
			Context:   ctx,
		}).FirstErr()
	})
}

func (s TransactionalWriter) WriteBatch(ctx context.Context, name string, messages []stream.Message) (int, error) {
	writeCount := 0
	err := s.writeInTransaction(ctx, func(ctx context.Context) error {
		buf := make([]*kgo.Record, 0, len(messages))
		for _, m := range messages {
			buf = append(buf, &kgo.Record{
				Key:       []byte(m.Key),
				Value:     m.Data,
				Headers:   marshalHeaders(m.Header),
				Timestamp: time.Time{},
				Topic:     name,
				Context:   ctx,
			})
		}

		err := s.client.ProduceSync(ctx, buf...).FirstErr()
		if err != nil {
			return err
		}
		writeCount = len(buf)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return writeCount, nil
}
