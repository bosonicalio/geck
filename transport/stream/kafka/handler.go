package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

// ReaderHandlerFunc is a function that processes a Kafka record ([kgo.Record]).
type ReaderHandlerFunc func(ctx context.Context, msg *kgo.Record) error
