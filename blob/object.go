package blob

import (
	"context"
	"io"
)

// ObjectUploader is an interface for uploading objects to a storage bucket.
type ObjectUploader interface {
	// Upload uploads a file to the storage bucket using the provided object key and its data.
	Upload(ctx context.Context, key string, data io.Reader) error
}

// ObjectRemover is an interface for removing objects from a storage bucket.
type ObjectRemover interface {
	// Remove deletes a file from the storage bucket using the provided object key.
	Remove(ctx context.Context, key string) error
}
