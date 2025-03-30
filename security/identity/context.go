package identity

import (
	"context"
	"errors"
)

type principalContextKey struct{}

var (
	// ErrPrincipalNotFound is returned when a principal is not found.
	ErrPrincipalNotFound = errors.New("principal not found")
)

// GetPrincipal retrieves a [Principal] instance from `ctx`.
func GetPrincipal(ctx context.Context) (Principal, error) {
	val, ok := ctx.Value(principalContextKey{}).(Principal)
	if !ok {
		return nil, ErrPrincipalNotFound
	}
	return val, nil
}

// WithPrincipal sets a [Principal] instance in `ctx`.
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}
