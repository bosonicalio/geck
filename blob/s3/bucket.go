package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/samber/lo"

	"github.com/tesserical/geck/blob"
)

// Bucket is the Amazon Simple Storage Service (S3) implementation of [blob.Bucket].
type Bucket struct {
	name     string
	client   *s3.Client
	uploader *manager.Uploader
}

var (
	// compile-time assertions
	_ blob.Bucket = (*Bucket)(nil)
)

// NewBucket creates a new S3 bucket instance with the provided name, client, and uploader.
func NewBucket(name string, client *s3.Client) Bucket {
	return Bucket{
		name:     name,
		client:   client,
		uploader: manager.NewUploader(client),
	}
}

func (b Bucket) Upload(ctx context.Context, key string, data io.Reader) error {
	_, err := b.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: lo.EmptyableToPtr(b.name),
		Key:    lo.EmptyableToPtr(key),
		Body:   data,
	})
	return err
}

func (b Bucket) Remove(ctx context.Context, key string) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: lo.EmptyableToPtr(b.name),
		Key:    lo.EmptyableToPtr(key),
	})
	return err
}
