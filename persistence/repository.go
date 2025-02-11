package persistence

import "context"

// ReadRepository component used to retrieve T instances from a storage system.
type ReadRepository[K comparable, T Storable] interface {
	GetByID(ctx context.Context, key K) (*T, error)
}

// WriteRepository component used to write T instances into a storage system.
type WriteRepository[K comparable, T Storable] interface {
	Save(ctx context.Context, entity T) error
	DeleteByID(ctx context.Context, key K) error
	Delete(ctx context.Context, entity T) error
}

// WriteBatchRepository component used to write several T instances into a storage system.
type WriteBatchRepository[K comparable, T Storable] interface {
	SaveAll(ctx context.Context, entities []T) error
	DeleteAll(ctx context.Context, entities []T) error
	DeleteAllByIDs(ctx context.Context, keys []K) error
}
