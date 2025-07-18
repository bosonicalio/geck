package persistence

import (
	"context"
	"errors"
)

// Transaction is a database transaction interface that provides methods for
// managing transaction lifecycle. Implementations should handle the underlying
// database-specific transaction operations while maintaining consistent behavior.
// Use with [WithTxContext] and [FromTxContext] to propagate transactions through context.
type Transaction interface {
	// Commit commits the transaction.
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error
}

// -- Context Management --

// TxExecutor is a type that represents the executor of a transaction.
//
// It is used as a key in the context to store and retrieve the [Transaction] instance.
type TxExecutor string

var (
	// ErrInvalidTxContext is returned when a transaction context is invalid or does not contain a transaction.
	ErrInvalidTxContext = errors.New("geck.persistence: invalid transaction context")
)

// WithTxContext creates a new context with the given [Transaction].
func WithTxContext(parent context.Context, executor TxExecutor, tx Transaction) context.Context {
	return context.WithValue(parent, executor, tx)
}

// FromTxContext retrieves the [Transaction] from the given context.
func FromTxContext(ctx context.Context, executor TxExecutor) (Transaction, bool) {
	tx, ok := ctx.Value(executor).(Transaction)
	return tx, ok
}
