package interceptor

import (
	"context"
	"errors"

	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/bosonicalio/geck/transport/stream/kafka"
)

// -- Dead Letter Queue --

// UseDeadLetter is a [kafka.ReaderInterceptor] that sends messages to a dead letter queue (DLQ) if the handler returns
// an error.
//
// Moreover, this routine adds a header to the message with the original topic name, so that the consumer can
// identify the topic from which the message originated. If the topic is not set, it will be set to the
// default topic name with a `-dlq` suffix.
func UseDeadLetter(client *kgo.Client, topic string, opts ...kafka.InterceptorOption) kafka.ReaderInterceptor {
	ops := &kafka.InterceptorOptions{}
	for _, opt := range opts {
		opt(ops)
	}
	if ops.Skip == nil {
		ops.Skip = func(msg *kgo.Record) bool {
			return false
		}
	}
	return func(next kafka.ReaderHandlerFunc) kafka.ReaderHandlerFunc {
		return func(ctx context.Context, msg *kgo.Record) error {
			if ops.Skip(msg) {
				return next(ctx, msg)
			}

			err := next(ctx, msg)
			if err == nil {
				return nil
			}

			if msg.Headers == nil {
				msg.Headers = make([]kgo.RecordHeader, 0, 1)
			}
			msg.Headers = append(msg.Headers, kgo.RecordHeader{
				Key:   "Original-Topic",
				Value: []byte(msg.Topic),
			})
			msg.Topic = lo.CoalesceOrEmpty(topic, msg.Topic+"-dlq")
			if errProduce := client.ProduceSync(ctx, msg).FirstErr(); errProduce != nil {
				return errors.Join(err, errProduce)
			}
			return err
		}
	}
}
