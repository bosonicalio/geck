package paging

import (
	"context"

	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/persistence/criteria"
)

// Repository component used to retrieve several T instances from a storage system.
type Repository[T persistence.Storable] interface {
	FindAll(ctx context.Context, opts ...criteria.Option) (*Page[T], error)
}
