package sql

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/samber/lo"

	"github.com/tesserical/geck/persistence"
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

// - Interceptors -

// -- Logger --

// DBLogger is an interceptor component adhering logging capabilities to an existing [DB].
type DBLogger struct {
	next     DB
	logger   *slog.Logger
	logLevel slog.Level
}

// compile-time assertion
var _ DB = (*DBLogger)(nil)

// NewDBLogger allocates a new [DBLogger].
func NewDBLogger(parent DB, logger *slog.Logger, opts ...DBLoggerOption) DBLogger {
	options := dbLoggerOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return DBLogger{
		next:     parent,
		logger:   logger,
		logLevel: lo.Min([]slog.Level{options.logLevel, slog.LevelDebug}),
	}
}

func (d DBLogger) Begin() (tx *sql.Tx, err error) {
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

func (d DBLogger) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx *sql.Tx, err error) {
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

func (d DBLogger) QueryContext(ctx context.Context, query string, args ...interface{}) (rows *sql.Rows, err error) {
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

func (d DBLogger) QueryRowContext(ctx context.Context, query string, args ...interface{}) (row *sql.Row) {
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

func (d DBLogger) ExecContext(ctx context.Context, query string, args ...interface{}) (res sql.Result, err error) {
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

func (d DBLogger) PrepareContext(ctx context.Context, query string) (stmt *sql.Stmt, err error) {
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

// --- Options ---

type dbLoggerOptions struct {
	logLevel slog.Level
}

// DBLoggerOption is a routine used to set up [DBLogger] optional configuration.
type DBLoggerOption func(*dbLoggerOptions)

// WithLogLevel sets the base log level to use for a [DBLogger].
func WithLogLevel(lvl slog.Level) DBLoggerOption {
	return func(o *dbLoggerOptions) {
		o.logLevel = lvl
	}
}

// -- Transaction Propagator --

// TxExecutor is a type alias for the [persistence.TxExecutor] used in the context of a [DBTxPropagator].
const TxExecutor persistence.TxExecutor = "sql"

// DBTxPropagator is an interceptor component adhering transaction propagation
// to all operations of an existing [DB], using transaction contexts.
type DBTxPropagator struct {
	next         DB
	txOpts       *sql.TxOptions
	autoCreateTx bool
}

// compile-time assertion
var _ DB = (*DBTxPropagator)(nil)

// NewDBTxPropagator allocates a new [DBTxPropagator].
func NewDBTxPropagator(parent DB, opts ...DBTxPropagatorOption) DBTxPropagator {
	options := dbTxPropagatorOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return DBTxPropagator{
		next:         parent,
		txOpts:       options.txOpts,
		autoCreateTx: options.autoCreate,
	}
}

// getTxCtx retrieves a transaction context from the provided context. If the transaction context is not found
// and the auto-create transaction option is enabled, a new transaction context is created.
func (d DBTxPropagator) getTxCtx(ctx context.Context) (context.Context, error) {
	_, found := persistence.FromTxContext(ctx, TxExecutor)
	if found || !d.autoCreateTx {
		return ctx, nil
	}
	tx, err := d.next.BeginTx(ctx, d.txOpts)
	if err != nil {
		return ctx, err
	}
	ctxTx := persistence.WithTxContext(ctx, TxExecutor, Transaction{Parent: tx})
	return ctxTx, nil
}

func (d DBTxPropagator) Begin() (*sql.Tx, error) {
	return d.next.Begin()
}

func (d DBTxPropagator) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}

	txIface, found := persistence.FromTxContext(ctxTx, TxExecutor)
	if !found {
		return d.next.BeginTx(ctxTx, opts)
	}

	tx, ok := txIface.(Transaction)
	if !ok {
		return nil, persistence.ErrInvalidTxContext
	}
	return tx.Parent, nil
}

func (d DBTxPropagator) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}

	txIface, found := persistence.FromTxContext(ctxTx, TxExecutor)
	if !found {
		return d.next.QueryContext(ctxTx, query, args...)
	}

	tx, ok := txIface.(Transaction)
	if !ok {
		return nil, persistence.ErrInvalidTxContext
	}
	return tx.Parent.QueryContext(ctxTx, query, args...)
}

func (d DBTxPropagator) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil
	}
	txIface, found := persistence.FromTxContext(ctxTx, TxExecutor)
	if !found {
		return d.next.QueryRowContext(ctxTx, query, args...)
	}
	tx, ok := txIface.(Transaction)
	if !ok {
		panic(persistence.ErrInvalidTxContext)
	}
	return tx.Parent.QueryRowContext(ctxTx, query, args...)
}

func (d DBTxPropagator) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}
	txIface, found := persistence.FromTxContext(ctxTx, TxExecutor)
	if !found {
		return d.next.ExecContext(ctxTx, query, args...)
	}

	tx, ok := txIface.(Transaction)
	if !ok {
		return nil, persistence.ErrInvalidTxContext
	}
	return tx.Parent.ExecContext(ctxTx, query, args...)
}

func (d DBTxPropagator) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	ctxTx, err := d.getTxCtx(ctx)
	if err != nil {
		return nil, err
	}
	txIface, found := persistence.FromTxContext(ctxTx, TxExecutor)
	if !found {
		return d.next.PrepareContext(ctxTx, query)
	}

	tx, ok := txIface.(Transaction)
	if !ok {
		return nil, persistence.ErrInvalidTxContext
	}
	return tx.Parent.PrepareContext(ctxTx, query)
}

// --- Options ---

type dbTxPropagatorOptions struct {
	txOpts     *sql.TxOptions
	autoCreate bool
}

// DBTxPropagatorOption is a routine used to set up [DBTxPropagator] optional configuration.
type DBTxPropagatorOption func(*dbTxPropagatorOptions)

// WithTxOptions sets transaction options for a [DBTxPropagator].
func WithTxOptions(opts *sql.TxOptions) DBTxPropagatorOption {
	return func(o *dbTxPropagatorOptions) {
		o.txOpts = opts
	}
}

// WithAutoCreateTx enables the automatic creation of a transaction if one is not found in the context.
func WithAutoCreateTx(v bool) DBTxPropagatorOption {
	return func(o *dbTxPropagatorOptions) {
		o.autoCreate = v
	}
}
