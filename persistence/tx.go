package persistence

import "context"

// Transaction is a database transaction interface that provides methods for
// managing transaction lifecycle. Implementations should handle the underlying
// database-specific transaction operations while maintaining consistent behavior.
// Use with [NewTxContext] and [FromTxContext] to propagate transactions through context.
type Transaction interface {
	// Commit commits the transaction.
	Commit(ctx context.Context) error
	// Rollback rolls back the transaction.
	Rollback(ctx context.Context) error
}

type txCtxKey struct{}

// NewTxContext creates a new context with the given tx ([Transaction]).
func NewTxContext[T Transaction](parent context.Context, tx T) context.Context {
	return context.WithValue(parent, txCtxKey{}, tx)
}

// FromTxContext retrieves the tx ([Transaction]) from the given context.
func FromTxContext[T Transaction](ctx context.Context) (T, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(T)
	return tx, ok
}
