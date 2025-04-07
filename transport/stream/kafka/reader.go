package kafka

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samber/lo"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

// -- Error(s) --

var (
	// ErrAlreadyRegistered is returned when a handler is already registered for a topic.
	ErrAlreadyRegistered = errors.New("reader handler already registered")
	// ErrEOF is returned when no records are fetched from Kafka.
	ErrEOF = errors.New("no records fetched")
	// ErrNoHandlerFound is returned when no handler is found for a topic.
	ErrNoHandlerFound = errors.New("no handler found for topic")
	// ErrReaderManagerClosed is returned when the reader manager is closed.
	ErrReaderManagerClosed = errors.New("reader manager is closed")
	// ErrReaderManagerAlreadyStarted is returned when the reader manager is already started.
	ErrReaderManagerAlreadyStarted = errors.New("reader manager already started")
)

// -- Reader Manager --

// A ReaderManager is a component that manages the reading of records from Apache Kafka topics.
type ReaderManager interface {
	// Register registers a handler for a specific topic. The handler will be invoked everytime a record is fetched
	// from the Apache Kafka topic.
	//
	// Use [ReaderRegisterOption] to configure additional mechanisms (like consumer groups).
	Register(name string, handler ReaderHandlerFunc, opts ...ReaderRegisterOption) error
	// MustRegister registers a handler for a specific topic. The handler will be invoked everytime a record is fetched
	// from the Apache Kafka topic.
	//
	// Use [ReaderRegisterOption] to configure additional mechanisms (like consumer groups).
	//
	// This routine will panic if any error occurs.
	MustRegister(name string, handler ReaderHandlerFunc, opts ...ReaderRegisterOption)
	// Start starts the reader manager. It begins polling records from Kafka and invoking the registered handlers.
	//
	// The reader manager will run until it is closed ([ReaderManager.Close]).
	Start() error
	// Close closes the reader manager. It stops polling records from Kafka and  waits for all in-flight handlers
	// to complete for a graceful shutdown.
	Close(ctx context.Context) error
}

// --- Channel ---

// ChannelReaderManager is a concrete implementation of [ReaderManager] using Go channels.
//
// This component uses a buffered channel to centrally manage the flow of polled Kafka records with a fixed
// pool size due channel buffering (similar to semaphores), ensuring a predictable performance and resource allocation.
//
// Moreover, a polling job will be allocated and executed for each consumer group registered; in any case, this
// component will also allocate and execute an additional polling job for the default group
// (defined by [WithReaderManagerGroupID]).
// This helps default group is provisioned to reduce duplicate processing of records across multiple
// system nodes (aka. cluster). If no group is defined, the global poller will run with no consumer group,
// leading to work competition (aka. race conditions) between system nodes.
//
// Finally, this component will commit all uncommitted offsets after all polled records have been processed (or failed to
// process). This is done to ensure that all records are marked as processed and avoid any permanent data-loss as
// Apache Kafka topic use append-log storage; marking individual messages is not possible. Marking record offsets individually
// can cause data-loss as one record offset is greater than other offsets from the polled batch, and thus,
// marking records (indirectly) with a lower offset number as processed as well.
//
// It is the responsibility of the user to handle the errors returned by the handler function. It is recommended
// to use a retry mechanism or a dead-letter queue (DLQ) to handle these errors to fully guarantee no data-loss.
type ChannelReaderManager struct {
	options readerManagerOptions
	client  *kgo.Client

	topicHandlerMap     map[string]ReaderHandlerFunc
	topicGroupClientMap map[string]*kgo.Client
	messageWorkerChanel chan *kgo.Record
	inFlightProcs       sync.WaitGroup
	alreadyStarted      atomic.Bool
	isClosed            atomic.Bool

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

// Register registers a handler for a specific topic. The handler will be invoked everytime a record is fetched
// from the Apache Kafka topic.
//
// Use [ReaderRegisterOption] to configure additional mechanisms (like consumer groups).
//
// This routine must be called before [ReaderManager.Start] is called. If this routine is called after
// [ReaderManager.Start] is called, it will return [ErrReaderManagerAlreadyStarted] as it does not allow consumer
// registrations at runtime.
func (c *ChannelReaderManager) Register(name string, handler ReaderHandlerFunc, opts ...ReaderRegisterOption) error {
	if c.alreadyStarted.Load() {
		return ErrReaderManagerAlreadyStarted
	} else if c.isClosed.Load() {
		return ErrReaderManagerClosed
	}

	options := readerRegisterOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	_, ok := c.topicHandlerMap[name]
	if ok {
		return ErrAlreadyRegistered
	}

	c.topicHandlerMap[name] = handler
	if options.group.IsZero() {
		c.client.AddConsumeTopics(name)
		return nil
	}

	var err error
	groupClient, ok := c.topicGroupClientMap[name]
	if ok {
		groupClient.AddConsumeTopics(name)
		return nil
	}

	groupClient, err = kgo.NewClient(
		append(c.options.baseOpts, kgo.ConsumerGroup(options.group.String()))...,
	)
	if err != nil {
		return err
	}
	c.topicGroupClientMap[name] = groupClient
	groupClient.AddConsumeTopics(name)
	return nil
}

// MustRegister registers a handler for a specific topic. The handler will be invoked everytime a record is fetched
// from the Apache Kafka topic.
//
// Use [ReaderRegisterOption] to configure additional mechanisms (like consumer groups).
//
// This routine must be called before [ReaderManager.Start] is called. If this routine is called after
// [ReaderManager.Start] is called, it will return [ErrReaderManagerAlreadyStarted] as it does not allow consumer
// registrations at runtime.
//
// This routine will panic if any error occurs.
func (c *ChannelReaderManager) MustRegister(name string, handler ReaderHandlerFunc, opts ...ReaderRegisterOption) {
	if err := c.Register(name, handler, opts...); err != nil {
		panic(err)
	}
}

// Start starts the reader manager. It begins polling records from Kafka and invoking the registered handlers.
func (c *ChannelReaderManager) Start() error {
	if c.alreadyStarted.Load() {
		return ErrReaderManagerAlreadyStarted
	} else if c.isClosed.Load() {
		return ErrReaderManagerClosed
	}

	c.ctxBase, c.ctxCancelFunc = context.WithCancel(context.Background())

	// set defaults
	c.options.pollBatchSize = lo.CoalesceOrEmpty(c.options.pollBatchSize, 100)
	c.options.pollInterval = lo.CoalesceOrEmpty(c.options.pollInterval, 500*time.Millisecond)
	c.options.workerPoolSize = lo.CoalesceOrEmpty(c.options.workerPoolSize, c.options.pollBatchSize/2)
	c.options.handlerTimeout = lo.CoalesceOrEmpty(c.options.handlerTimeout, 30*time.Second)

	// bootstrap worker pool
	c.messageWorkerChanel = make(chan *kgo.Record, c.options.workerPoolSize)
	go c.startWorkerProc()

	c.alreadyStarted.Store(true)
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
				c.options.errorHandler(c.ctxBase, ErrEOF)
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

// Close closes the reader manager. It stops polling records from Kafka and waits for all in-flight handlers
// to complete for a graceful shutdown.
func (c *ChannelReaderManager) Close(ctx context.Context) error {
	if c.isClosed.Load() {
		return ErrReaderManagerClosed
	}
	c.isClosed.Store(true)
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

// --- Registrar ---

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

// --- Manager ---

type readerManagerOptions struct {
	baseOpts []kgo.Opt

	groupID        string
	workerPoolSize int
	pollBatchSize  int
	pollInterval   time.Duration
	handlerTimeout time.Duration
	errorHandler   func(context.Context, error)
}

// ReaderManagerOption represents an option for configuring the [ReaderManager].
type ReaderManagerOption func(*readerManagerOptions)

// WithReaderManagerClientOpts sets the base options ([][kgo.Opt]) of the underlying clients used by
// [ReaderManager].
func WithReaderManagerClientOpts(opts ...kgo.Opt) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		if len(opts) == 0 {
			o.baseOpts = make([]kgo.Opt, 0, len(o.baseOpts))
		}
		o.baseOpts = append(o.baseOpts, opts...)
	}
}

// WithReaderManagerGroupID sets the global consumer group ID of a [ReaderManager] instance.
func WithReaderManagerGroupID(id string) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.groupID = id
	}
}

// WithReaderManagerPoolSize sets the size of the worker pool used by [ReaderManager].
func WithReaderManagerPoolSize(size int) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.workerPoolSize = size
	}
}

// WithReaderManagerPollBatchSize sets the batch size for polling records from Kafka.
func WithReaderManagerPollBatchSize(size int) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.pollBatchSize = size
	}
}

// WithReaderManagerPollInterval sets the interval for polling records from Kafka.
func WithReaderManagerPollInterval(interval time.Duration) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.pollInterval = interval
	}
}

// WithReaderManagerHandlerTimeout sets the timeout for the handler function.
func WithReaderManagerHandlerTimeout(timeout time.Duration) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.handlerTimeout = timeout
	}
}

// WithReaderManagerErrorHandler sets the error handler function for the [ReaderManager].
func WithReaderManagerErrorHandler(handler func(context.Context, error)) ReaderManagerOption {
	return func(o *readerManagerOptions) {
		o.errorHandler = handler
	}
}
