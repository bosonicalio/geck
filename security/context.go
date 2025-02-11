package security

import (
	"context"
)

type principalContextKey struct{}

// GetPrincipal retrieves a [Principal] instance from `ctx`.
func GetPrincipal(ctx context.Context) (Principal, error) {
	val, ok := ctx.Value(principalContextKey{}).(Principal)
	if !ok {
		return nil, ErrPrincipalNotFound
	}
	return val, nil
}
