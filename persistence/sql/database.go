package sql

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence"
)

// DB represents a SQL database client based on stdlib [sql.DB].
//
// The intention of this interface is to avoid persistence components to depend on a concrete implementation
// ([sql.DB]). By not depending, testing/mocking is easier and also, it allows consumers to implement patterns
// like chain-of-responsibility to adhere additional functionality without affecting the final concrete
// implementation (e.g. logging, tracing, transaction management).
type DB interface {
	// Begin starts a transaction. The default isolation level is dependent on
	// the driver.
	//
	// Begin uses [context.Background] internally; to specify the context, use
	// [DB.BeginTx].
	Begin() (*sql.Tx, error)
	// BeginTx starts a transaction.
	//
	// The provided context is used until the transaction is committed or rolled back.
	// If the context is canceled, the sql package will roll back
	// the transaction. [sql.Tx.Commit] will return an error if the context provided to
	// BeginTx is canceled.
	//
	// The provided [sql.TxOptions] is optional and may be nil if defaults should be used.
	// If a non-default isolation level is used that the driver doesn't support,
	// an error will be returned.
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	// QueryContext executes a query that returns rows, typically a SELECT.
	// The args are for any placeholder parameters in the query.
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	// QueryRowContext executes a query that is expected to return at most one row.
	// QueryRowContext always returns a non-nil value. Errors are deferred until
	// [sql.Row]'s Scan method is called.
	// If the query selects no rows, the [*sql.Row.Scan] will return [sql.ErrNoRows].
	// Otherwise, [*sql.Row.Scan] scans the first selected row and discards
	// the rest.
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	// ExecContext executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	// PrepareContext creates a prepared statement for later queries or executions.
	// Multiple queries or executions may be run concurrently from the
	// returned statement.
	// The caller must call the statement's [*sql.Stmt.Close] method
	// when the statement is no longer needed.
	//
	// The provided context is used for the preparation of the statement, not for the
	// execution of the statement.
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// ---> Interceptors <---

// DBInterceptor is a component extending [DB] to adhere new functionality to a [DB] through a
// chain of responsibility.
type DBInterceptor interface {
	DB
	// SetNext sets the parent instance to call after/before DBInterceptor operations.
	SetNext(db DB)
}

// -- OPTIONS --

type databaseOptions struct {
	Interceptors []DBInterceptor
}

// DatabaseOption is a routine used to set up optional configurations of [DB].
type DatabaseOption func(*databaseOptions)

// WithInterceptor appends a [DBInterceptor] to a chain of the same type, creating an execution stack
// when any of the operations defined in [DB] executes.
func WithInterceptor(interceptor DBInterceptor) DatabaseOption {
	return func(do *databaseOptions) {
		if do.Interceptors == nil {
			do.Interceptors = make([]DBInterceptor, 0, 1)
		}
		do.Interceptors = append(do.Interceptors, interceptor)
	}
}

// NewDB allocates a new [DB] out of a [sql.DB]. In addition, this routine also accepts [DatabaseOption]
// derivatives to set up optional configurations (e.g. [WithInterceptor]).
func NewDB(parent *sql.DB, opts ...DatabaseOption) DB {
	options := databaseOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	if len(options.Interceptors) == 0 {
		return parent
	}

	var db DB = parent
	for _, interceptor := range options.Interceptors {
		interceptor.SetNext(db)
		db = interceptor
	}
	return db
}

// -- LOGGER --

// DatabaseLogger is an interceptor component adhering logging capabilities to an existing [DB].
type DatabaseLogger struct {
	next     DB
	logger   *slog.Logger
	logLevel slog.Level
}

type databaseLoggerOptions struct {
	logLevel slog.Level
}

// DatabaseLoggerOption is a routine used to set up [DatabaseLogger] optional configuration.
type DatabaseLoggerOption func(*databaseLoggerOptions)

// WithLogLevel sets the base log level to use for a [DatabaseLogger].
func WithLogLevel(lvl slog.Level) DatabaseLoggerOption {
	return func(o *databaseLoggerOptions) {
		o.logLevel = lvl
	}
}

// compile-time assertion
var _ DBInterceptor = (*DatabaseLogger)(nil)

// NewDatabaseLogger allocates a new [DatabaseLogger].
func NewDatabaseLogger(logger *slog.Logger, opts ...DatabaseLoggerOption) *DatabaseLogger {
	options := databaseLoggerOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return &DatabaseLogger{
		logger:   logger,
		logLevel: lo.Min([]slog.Level{options.logLevel, slog.LevelDebug}),
	}
}

func (d *DatabaseLogger) SetNext(db DB) {
	d.next = db
}

func (d *DatabaseLogger) Begin() (tx *sql.Tx, err error) {
	start := time.Now()
	tx, err = d.next.Begin()
	if err != nil {
		d.logger.Error("failed to start transaction",
			slog.String("err", err.Error()),
			slog.String("took", time.Since(start).String()),
		)
		return
	}
	d.logger.Log(context.Background(), d.logLevel, "started transaction",
		slog.String("took", time.Since(start).String()),
	)
	return
}

func (d *DatabaseLogger) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, err error) {
	start := time.Now()
	var optLogAttributes slog.Attr
	if opts != nil {
		optLogAttributes = slog.Group("tx_options",
			slog.Bool("read_only", opts.ReadOnly),
		)
	}

	tx, err = d.next.BeginTx(ctx, opts)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to start transaction",
			slog.String("err", err.Error()),
			slog.String("took", time.Since(start).String()),
			optLogAttributes,
		)
		return
	}
	d.logger.Log(ctx, d.logLevel, "started transaction",
		slog.String("took", time.Since(start).String()),
		optLogAttributes)
	return
}

func (d *DatabaseLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
	start := time.Now()
	rows, err = d.next.QueryContext(ctx, query, args...)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to query rows",
			slog.String("err", err.Error()),
			slog.String("query", query),
			slog.Int("total_args", len(args)),
			slog.String("took", time.Since(start).String()),
		)
		return
	}
	d.logger.Log(ctx, d.logLevel, "performed query",
		slog.String("query", query),
		slog.Int("total_args", len(args)),
		slog.String("took", time.Since(start).String()),
	)
	return
}

func (d *DatabaseLogger) QueryRowContext(ctx context.Context, query string, args ...interface{}) (row *sql.Row) {
	start := time.Now()
	row = d.next.QueryRowContext(ctx, query, args...)
	if row != nil && row.Err() != nil {
		d.logger.ErrorContext(ctx, "failed to query row",
			slog.String("err", row.Err().Error()),
			slog.String("query", query),
			slog.Int("total_args", len(args)),
			slog.String("took", time.Since(start).String()),
		)
		return
	}
	d.logger.Log(ctx, d.logLevel, "performed query row",
		slog.String("query", query),
		slog.Int("total_args", len(args)),
		slog.String("took", time.Since(start).String()),
	)
	return
}

func (d *DatabaseLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (res sql.Result, err error) {
	start := time.Now()
	res, err = d.next.ExecContext(ctx, query, args...)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to exec query",
			slog.String("err", err.Error()),
			slog.String("query", query),
			slog.Int("total_args", len(args)),
			slog.String("took", time.Since(start).String()),
		)
		return
	}
	d.logger.Log(ctx, d.logLevel, "performed exec",
		slog.String("query", query),
		slog.Int("total_args", len(args)),
		slog.String("took", time.Since(start).String()),
	)
	return
}

func (d *DatabaseLogger) PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error) {
	start := time.Now()
	stmt, err = d.next.PrepareContext(ctx, query)
	if err != nil {
		d.logger.ErrorContext(ctx, "failed to prepare statement",
			slog.String("err", err.Error()),
			slog.String("query", query),
			slog.String("took", time.Since(start).String()),
		)
		return
	}
	d.logger.Log(ctx, d.logLevel, "prepared statement",
		slog.String("query", query),
		slog.String("took", time.Since(start).String()),
	)
	return
}

// -- Transaction Propagator --

// DatabaseTxPropagator is an interceptor component adhering transaction propagation
// to all operations of an existing [DB], using transaction contexts.
type DatabaseTxPropagator struct {
	next         DB
	txOpts       *sql.TxOptions
	autoCreateTx bool
}

type databaseTxPropagatorOptions struct {
	txOpts     *sql.TxOptions
	autoCreate bool
}

// DatabaseTxPropagatorOption is a routine used to set up [DatabaseTxPropagator] optional configuration.
type DatabaseTxPropagatorOption func(*databaseTxPropagatorOptions)

// WithTxOptions sets transaction options for a [DatabaseTxPropagator].
func WithTxOptions(opts *sql.TxOptions) DatabaseTxPropagatorOption {
	return func(o *databaseTxPropagatorOptions) {
		o.txOpts = opts
	}
}

// WithAutoCreateTx enables the automatic creation of a transaction if one is not found in the context.
func WithAutoCreateTx() DatabaseTxPropagatorOption {
	return func(o *databaseTxPropagatorOptions) {
		o.autoCreate = true
	}
}

// compile-time assertion
var _ DBInterceptor = (*DatabaseTxPropagator)(nil)

// NewDatabaseTxPropagator allocates a new [DatabaseTxPropagator].
func NewDatabaseTxPropagator(opts ...DatabaseTxPropagatorOption) *DatabaseTxPropagator {
	options := databaseTxPropagatorOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return &DatabaseTxPropagator{}
}

// getTxCtx retrieves a transaction context from the provided context. If the transaction context is not found
// and the auto-create transaction option is enabled, a new transaction context is created.
func (d *DatabaseTxPropagator) getTxCtx(ctx context.Context) (context.Context, error) {
	_, found := persistence.FromTxContext[Transaction](ctx)
	if found || !d.autoCreateTx {
		return ctx, nil
	}
	tx, err := d.next.BeginTx(ctx, d.txOpts)
	if err != nil {
		return ctx, err
	}
	ctxTx := persistence.NewTxContext[Transaction](ctx, Transaction{Parent: tx})
	return ctxTx, nil
}

func (d *DatabaseTxPropagator) SetNext(db DB) {
	d.next = db
}

func (d *DatabaseTxPropagator) Begin() (*sql.Tx, error) {
	return d.next.Begin()
}

func (d *DatabaseTxPropagator) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}
	tx, found := persistence.FromTxContext[Transaction](ctxTx)
	if found {
		return tx.Parent, nil
	}
	return d.next.BeginTx(ctxTx, opts)
}

func (d *DatabaseTxPropagator) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}

	tx, found := persistence.FromTxContext[Transaction](ctxTx)
	if !found {
		return d.next.QueryContext(ctxTx, query, args...)
	}
	return tx.Parent.QueryContext(ctxTx, query, args...)
}

func (d *DatabaseTxPropagator) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil
	}
	tx, found := persistence.FromTxContext[Transaction](ctxTx)
	if !found {
		return d.next.QueryRowContext(ctxTx, query, args...)
	}
	return tx.Parent.QueryRowContext(ctxTx, query, args...)
}

func (d *DatabaseTxPropagator) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}
	tx, found := persistence.FromTxContext[Transaction](ctxTx)
	if !found {
		return d.next.ExecContext(ctxTx, query, args...)
	}
	return tx.Parent.ExecContext(ctxTx, query, args...)
}

func (d *DatabaseTxPropagator) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}
	tx, found := persistence.FromTxContext[Transaction](ctxTx)
	if !found {
		return d.next.PrepareContext(ctxTx, query)
	}
	return tx.Parent.PrepareContext(ctxTx, query)
}
