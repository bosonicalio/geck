package persistence

import "context"

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type txCtxKey struct{}

func NewTxContext[T Transaction](parent context.Context, tx T) context.Context {
	return context.WithValue(parent, txCtxKey{}, tx)
}

func FromTxContext[T Transaction](ctx context.Context) (T, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(T)
	return tx, ok
}
