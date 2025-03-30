package kafka

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kgo"
)

// ClientConfig is a structure used by factory routines generating Kafka clients.
type ClientConfig struct {
	SeedBrokers []string `env:"KAFKA_SEED_BROKERS" envSeparator:"," envDefault:"localhost:9092"`
	ClientID    string   `env:"KAFKA_CLIENT_ID"`
}

// NewClient creates a new Kafka client using [kgo] package.
func NewClient(config ClientConfig, opts ...kgo.Opt) (*kgo.Client, error) {
	clientID := lo.CoalesceOrEmpty(config.ClientID, fmt.Sprintf("geck-kafka-client-%s", lo.RandomString(6, lo.LettersCharset)))
	opts = append(opts,
		kgo.SeedBrokers(config.SeedBrokers...),
		kgo.ClientID(clientID),
		kgo.ProducerBatchCompression(kgo.Lz4Compression()),
	)
	return kgo.NewClient(opts...)
}

// - Transactional -

// NewTxClient creates a new Kafka transactional producer client using [kgo] package.
func NewTxClient(config ClientConfig, opts ...TxClientOption) (*kgo.Client, error) {
	options := txClientOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	if options.baseOpts == nil {
		options.baseOpts = make([]kgo.Opt, 0, 2)
	}

	options.baseOpts = append(options.baseOpts, kgo.RequiredAcks(kgo.AllISRAcks()))
	options.transactionalID = lo.CoalesceOrEmpty(options.transactionalID,
		fmt.Sprintf("geck-producer-%s", lo.RandomString(8, lo.LettersCharset)))
	options.baseOpts = append(options.baseOpts, kgo.TransactionalID(options.transactionalID))
	return NewClient(config, options.baseOpts...)
}

// -- Option(s) --

type txClientOptions struct {
	transactionalID string

	baseOpts []kgo.Opt
}

// TxClientOption is a functional option for configuring Kafka clients.
type TxClientOption func(*txClientOptions)

// WithClientTxID sets the transactional ID for the Kafka client.
func WithClientTxID(id string) TxClientOption {
	return func(o *txClientOptions) {
		o.transactionalID = id
	}
}

// WithTxClientBaseOpts sets the base options for the Kafka client.
func WithTxClientBaseOpts(opts ...kgo.Opt) TxClientOption {
	return func(o *txClientOptions) {
		o.baseOpts = opts
	}
}

// - Reader -

// NewReaderClient creates a new Kafka reader client using [kgo] package.
func NewReaderClient(config ClientConfig, opts ...ReaderClientOption) (*kgo.Client, error) {
	options := readerClientOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	options.consumerGroup = lo.CoalesceOrEmpty(options.consumerGroup,
		fmt.Sprintf("geck-consumer-%s", lo.RandomString(6, lo.LettersCharset)))

	if options.baseOpts == nil {
		options.baseOpts = make([]kgo.Opt, 0, 2)
	}
	options.baseOpts = append(options.baseOpts, kgo.ConsumerGroup(options.consumerGroup))
	if options.readCommittedOnly {
		options.baseOpts = append(options.baseOpts, kgo.FetchIsolationLevel(kgo.ReadCommitted()))
	}
	return NewClient(config, options.baseOpts...)
}

// -- Option(s) --

type readerClientOptions struct {
	consumerGroup     string
	readCommittedOnly bool
	baseOpts          []kgo.Opt
}

// ReaderClientOption is a functional option for configuring Kafka clients.
type ReaderClientOption func(*readerClientOptions)

// WithClientConsumerGroup sets the consumer group for the Kafka client.
func WithClientConsumerGroup(name string) ReaderClientOption {
	return func(o *readerClientOptions) {
		o.consumerGroup = name
	}
}

// WithClientReadCommittedOnly sets the Kafka client to read committed messages only.
//
// Enable this option if the topic to be read is configured with transactional semantics.
func WithClientReadCommittedOnly() ReaderClientOption {
	return func(o *readerClientOptions) {
		o.readCommittedOnly = true
	}
}

// WithReaderClientBaseOpts sets the base options for the Kafka client.
func WithReaderClientBaseOpts(opts ...kgo.Opt) ReaderClientOption {
	return func(o *readerClientOptions) {
		o.baseOpts = opts
	}
}
