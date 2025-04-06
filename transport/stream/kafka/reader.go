package kafka

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

// -- Error(s) --

var (
	// ErrAlreadyRegistered is returned when a handler is already registered for a topic.
	ErrAlreadyRegistered = errors.New("reader handler already registered")
	// ErrNoRecords is returned when no records are fetched from Kafka.
	ErrNoRecords = errors.New("no records fetched")
	// ErrNoHandlerFound is returned when no handler is found for a topic.
	ErrNoHandlerFound = errors.New("no handler found for topic")
	// ErrReaderManagerClosed is returned when the reader manager is closed.
	ErrReaderManagerClosed = errors.New("reader manager is closed")
)

// A ReaderManager is a component that manages the registration and lifecycle of
// Kafka reader handlers. It is responsible for starting the reader, polling
// records from Kafka, and invoking the registered handlers for each record.
type ReaderManager interface {
	// Register registers a handler for a specific topic. The handler is invoked
	// for each record fetched from Kafka. The handler function should return an
	// error if the processing of the record fails. The handler is invoked in a
	// separate goroutine, and the reader manager will handle the error
	// according to the configured error handler.
	Register(name string, handler ReaderHandlerFunc, opts ...ReaderRegisterOption)
	// Start starts the reader manager. It begins polling records from Kafka and
	// invoking the registered handlers. The reader manager will run until it is closed ([ReaderManager.Close]).
	Start() error
	// Close closes the reader manager. It stops polling records from Kafka and
	// waits for all in-flight handlers to complete for a graceful shutdown.
	Close(ctx context.Context) error
}

// -- Channel --

// ChannelReaderManager is a concrete implementation of [ReaderManager] using Go channels.
//
// This component is responsible for managing the registration and lifecycle of
// Kafka reader handlers. It is responsible for starting the reader, polling
// records from Kafka, and invoking the registered handlers for each record.
// It uses a worker pool to process records concurrently and allows for
// custom error handling and configuration options.
// It is designed to be used in a streaming architecture where multiple
// handlers can be registered for different topics, and each handler can
// process records independently.
type ChannelReaderManager struct {
	options readerManagerOptions
	client  *kgo.Client

	topicHandlerMap     map[string]ReaderHandlerFunc
	topicGroupClientMap map[string]*kgo.Client
	messageWorkerChanel chan *kgo.Record
	inFlightProcs       sync.WaitGroup

	ctxBase       context.Context
	ctxCancelFunc context.CancelFunc
}

// compile-time assertion
var _ ReaderManager = (*ChannelReaderManager)(nil)

// NewChannelReaderManager creates a new instance of [ChannelReaderManager].
func NewChannelReaderManager(opts ...ReaderManagerOption) (*ChannelReaderManager, error) {
	options := readerManagerOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	client, err := kgo.NewClient(options.baseOpts...)
	if err != nil {
		return nil, err
	}
	return &ChannelReaderManager{
		options:             options,
		client:              client,
		topicHandlerMap:     make(map[string]ReaderHandlerFunc),
		topicGroupClientMap: make(map[string]*kgo.Client),
	}, nil
}

func (c *ChannelReaderManager) Register(name string, handler ReaderHandlerFunc, opts ...ReaderRegisterOption) {
	options := readerRegisterOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	_, ok := c.topicHandlerMap[name]
	if ok {
		panic(ErrAlreadyRegistered)
	}

	c.topicHandlerMap[name] = handler
	if options.group.IsZero() {
		c.client.AddConsumeTopics(name)
		return
	}

	var err error
	groupClient, ok := c.topicGroupClientMap[name]
	if ok {
		groupClient.AddConsumeTopics(name)
		return
	}

	groupClient, err = kgo.NewClient(
		append(c.options.baseOpts, kgo.ConsumerGroup(options.group.String()))...,
	)
	if err != nil {
		panic(err)
	}
	c.topicGroupClientMap[name] = groupClient
	groupClient.AddConsumeTopics(name)
}

func (c *ChannelReaderManager) Start() error {
	c.ctxBase, c.ctxCancelFunc = context.WithCancel(context.Background())

	// set defaults
	c.options.pollBatchSize = lo.CoalesceOrEmpty(c.options.pollBatchSize, 100)
	c.options.pollInterval = lo.CoalesceOrEmpty(c.options.pollInterval, 500*time.Millisecond)
	c.options.workerPoolSize = lo.CoalesceOrEmpty(c.options.workerPoolSize, c.options.pollBatchSize/2)
	c.options.handlerTimeout = lo.CoalesceOrEmpty(c.options.handlerTimeout, 30*time.Second)

	// bootstrap worker pool
	c.messageWorkerChanel = make(chan *kgo.Record, c.options.workerPoolSize)
	go c.startWorkerProc()

	errs := make([]error, 0, len(c.topicGroupClientMap)+1)
	errsMu := sync.Mutex{}
	go func() {
		if err := c.startPoller(c.client); err != nil {
			errsMu.Lock()
			errs = append(errs, err)
			errsMu.Unlock()
		}
	}()
	for _, groupClient := range c.topicGroupClientMap {
		go func() {
			if err := c.startPoller(groupClient); err != nil {
				errsMu.Lock()
				errs = append(errs, err)
				errsMu.Unlock()
			}
		}()
	}
	return errors.Join(errs...)
}

func (c *ChannelReaderManager) startPoller(client *kgo.Client) error {
	for {
		select {
		case <-c.ctxBase.Done():
			return c.ctxBase.Err()
		default:
		}

		fetches := client.PollRecords(c.ctxBase, c.options.pollBatchSize)
		err := fetches.Err()
		if fetches.IsClientClosed() || errors.Is(err, context.Canceled) {
			return ErrReaderManagerClosed
		} else if kerr.IsRetriable(err) {
			if c.options.errorHandler != nil {
				c.options.errorHandler(c.ctxBase, err)
			}
			time.Sleep(c.options.pollInterval)
			continue
		} else if err != nil {
			return err
		}

		if fetches.Empty() {
			if c.options.errorHandler != nil {
				c.options.errorHandler(c.ctxBase, ErrNoRecords)
			}
			time.Sleep(c.options.pollInterval)
			continue
		}

		numRecords := fetches.NumRecords()
		iter := fetches.RecordIter()
		c.inFlightProcs.Add(numRecords)
		for !iter.Done() {
			record := iter.Next()
			// send to worker channel
			c.messageWorkerChanel <- record
		}
		c.inFlightProcs.Wait()
		err = client.CommitUncommittedOffsets(c.ctxBase)
		if err != nil && c.options.errorHandler != nil {
			c.options.errorHandler(c.ctxBase, err)
		}
	}
}

func (c *ChannelReaderManager) startWorkerProc() {
	for message := range c.messageWorkerChanel {
		err := c.processRecord(message)
		if err != nil && c.options.errorHandler != nil {
			c.options.errorHandler(c.ctxBase, err)
		}
	}
}

func (c *ChannelReaderManager) processRecord(record *kgo.Record) error {
	defer c.inFlightProcs.Done()
	handlerFunc, ok := c.topicHandlerMap[record.Topic]
	if !ok {
		return ErrNoHandlerFound
	}
	scopedCtx, cancelFunc := context.WithTimeout(c.ctxBase, c.options.handlerTimeout)
	defer cancelFunc()
	return handlerFunc(scopedCtx, record)
}

func (c *ChannelReaderManager) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	c.ctxCancelFunc()
	c.inFlightProcs.Wait()
	close(c.messageWorkerChanel)
	c.client.Close()
	for _, groupClient := range c.topicGroupClientMap {
		groupClient.Close()
	}
	return nil
}

// -- Option(s) --

type readerRegisterOptions struct {
	group ConsumerGroup
}

// ReaderRegisterOption represents an option for registering a reader handler.
type ReaderRegisterOption func(*readerRegisterOptions)

// WithReaderGroup sets the consumer group for the reader handler.
func WithReaderGroup(group ConsumerGroup) ReaderRegisterOption {
	return func(o *readerRegisterOptions) {
		o.group = group
	}
}

type readerManagerOptions struct {
	baseOpts []kgo.Opt

	workerPoolSize int
	pollBatchSize  int
	pollInterval   time.Duration
	handlerTimeout time.Duration
	errorHandler   func(context.Context, error)
}

type ReaderManagerOption func(*readerManagerOptions)

// WithReaderClientOpts sets the base options ([][kgo.Opt]) of the underlying clients used by
// [ReaderManager].
func WithReaderClientOpts(opts ...kgo.Opt) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		if len(opts) == 0 {
			o.baseOpts = make([]kgo.Opt, 0, len(o.baseOpts))
		}
		o.baseOpts = append(o.baseOpts, opts...)
	}
}

// WithReaderPoolSize sets the size of the worker pool used by [ReaderManager].
func WithReaderPoolSize(size int) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.workerPoolSize = size
	}
}

// WithReaderPollBatchSize sets the batch size for polling records from Kafka.
func WithReaderPollBatchSize(size int) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.pollBatchSize = size
	}
}

// WithReaderPollInterval sets the interval for polling records from Kafka.
func WithReaderPollInterval(interval time.Duration) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.pollInterval = interval
	}
}

// WithReaderHandlerTimeout sets the timeout for the handler function.
func WithReaderHandlerTimeout(timeout time.Duration) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.handlerTimeout = timeout
	}
}

// WithReaderErrorHandler sets the error handler function for the [ReaderManager].
func WithReaderErrorHandler(handler func(context.Context, error)) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.errorHandler = handler
	}
}
