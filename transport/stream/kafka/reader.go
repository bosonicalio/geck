package kafka

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"golang.org/x/sync/semaphore"

	"github.com/hadroncorp/geck/transport/stream"
)

// ReaderManager is a Kafka [stream.ReaderManager] that reads messages from Kafka topics.
type ReaderManager struct {
	client          *kgo.Client
	topicHandlerMap map[string]stream.HandlerFunc

	options         readerManagerOptions
	workerSemaphore *semaphore.Weighted
	inFlightProcs   sync.WaitGroup
}

var (
	// compile-time assertion
	_ stream.ReaderManager = (*ReaderManager)(nil)
)

// NewReaderManager creates a new instance of [ReaderManager].
func NewReaderManager(client *kgo.Client, opts ...ReaderManagerOption) *ReaderManager {
	options := readerManagerOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	options.pollBatchSize = lo.CoalesceOrEmpty(options.pollBatchSize, 100)
	options.pollInterval = lo.CoalesceOrEmpty(options.pollInterval, 500*time.Millisecond)
	options.workerPoolSize = lo.CoalesceOrEmpty(options.workerPoolSize, options.pollBatchSize/2)
	options.handlerTimeout = lo.CoalesceOrEmpty(options.handlerTimeout, 30*time.Second)

	return &ReaderManager{
		client:          client,
		topicHandlerMap: make(map[string]stream.HandlerFunc),
		options:         options,
		workerSemaphore: semaphore.NewWeighted(int64(options.workerPoolSize)),
	}
}

func (r *ReaderManager) Register(topic string, handler stream.HandlerFunc) {
	r.topicHandlerMap[topic] = handler
	r.client.AddConsumeTopics(topic)
}

func (r *ReaderManager) Start(ctx context.Context) error {
	var err error
	successMu := sync.Mutex{}
	for {
		select {
		case <-ctx.Done():
			return err
		default:
		}

		fetches := r.client.PollRecords(ctx, r.options.pollBatchSize)
		if fetches.IsClientClosed() {
			return fetches.Err()
		} else if fetches.Err() != nil {
			if kerr.IsRetriable(fetches.Err()) {
				r.options.logger.WarnContext(ctx, "retriable error", slog.String("error", fetches.Err().Error()))
				time.Sleep(r.options.pollInterval)
				continue
			}
			r.options.logger.ErrorContext(ctx, "non-retriable error", fetches.Err())
			return fetches.Err()
		}

		fetches.EachError(func(topic string, partition int32, err error) {
			r.options.logger.ErrorContext(ctx, "error while fetching records",
				slog.String("topic", topic),
				slog.Int("partition", int(partition)),
				slog.String("error", err.Error()),
			)
		})

		if fetches.Empty() {
			r.options.logger.DebugContext(ctx, "no records fetched")
			time.Sleep(r.options.pollInterval)
			continue
		}

		numRecords := fetches.NumRecords()
		iter := fetches.RecordIter()
		successRecords := make([]*kgo.Record, 0, numRecords)
		r.inFlightProcs.Add(numRecords)
		for !iter.Done() {
			record := iter.Next()
			handler, ok := r.topicHandlerMap[record.Topic]
			if !ok {
				r.options.logger.WarnContext(ctx, "no handler found for topic", slog.String("topic", record.Topic))
				r.inFlightProcs.Done()
				continue
			}

			if err = r.workerSemaphore.Acquire(ctx, 1); err != nil {
				r.options.logger.ErrorContext(ctx, "failed to acquire worker semaphore", slog.String("error", err.Error()))
				r.inFlightProcs.Done()
				continue
			}
			go func() {
				defer r.workerSemaphore.Release(1)
				defer r.inFlightProcs.Done()
				scopedCtx, cancelFunc := context.WithTimeout(ctx, r.options.handlerTimeout)
				defer cancelFunc()
				err = handler(scopedCtx, stream.Message{
					Key:      string(record.Key),
					Metadata: unmarshalHeaders(record),
					Data:     record.Value,
				})
				if err != nil {
					r.options.logger.ErrorContext(ctx, "failed to process record", slog.String("error", err.Error()))
					return
				}
				successMu.Lock()
				successRecords = append(successRecords, record)
				successMu.Unlock()
			}()
		}
		r.inFlightProcs.Wait()

		if r.options.commitAll {
			err = r.client.CommitUncommittedOffsets(ctx)
		} else {
			err = r.client.CommitRecords(ctx, successRecords...)
		}
		if err != nil {
			r.options.logger.ErrorContext(ctx, "failed to commit records", slog.String("error", err.Error()))
			continue
		}

		r.options.logger.DebugContext(ctx, "records committed", slog.Int("num_records", len(successRecords)))
	}
}

func (r *ReaderManager) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	r.inFlightProcs.Wait()
	r.client.Close()
	return nil
}

// -- Option(s) --

type readerManagerOptions struct {
	logger         *slog.Logger
	commitAll      bool
	workerPoolSize int
	pollBatchSize  int
	pollInterval   time.Duration
	handlerTimeout time.Duration
}

type ReaderManagerOption func(*readerManagerOptions)

func WithReaderCommitAll() ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.commitAll = true
	}
}

func WithReaderWorkerPoolSize(size int) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.workerPoolSize = size
	}
}

func WithReaderPollBatchSize(size int) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.pollBatchSize = size
	}
}

func WithReaderPollInterval(interval time.Duration) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.pollInterval = interval
	}
}

func WithReaderHandlerTimeout(timeout time.Duration) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.handlerTimeout = timeout
	}
}

func WithReaderLogger(logger *slog.Logger) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.logger = logger
	}
}
