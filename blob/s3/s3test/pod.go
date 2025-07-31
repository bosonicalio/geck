package s3test

import (
	"context"
	"io"
	"io/fs"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/samber/lo"

	"github.com/bosonicalio/geck/blob"
	"github.com/bosonicalio/geck/cloud/aws/awstest"
	"github.com/bosonicalio/geck/testutil"
)

// Pod is a test component for running a LocalStack docker instance with S3 service.
type Pod struct {
	awsPod awstest.Pod
	client *s3.Client
}

// compile-time assertions
var _ testutil.Pod = (*Pod)(nil)

// NewPod creates a new LocalStack container configured to run the S3 service.
//
// It allows for optional configuration such as bucket name, seed data from a filesystem,
// and seed bytes to be uploaded to the S3 service upon initialization.
func NewPod(ctx context.Context, opts ...PodOption) (Pod, error) {
	podConfig := &podOptions{}
	for _, opt := range opts {
		opt(podConfig)
	}
	// Ensure only the S3 service is started
	podConfig.baseOpts = append(podConfig.baseOpts, awstest.WithPodServices("s3"))
	awsPod, err := awstest.NewPod(ctx, podConfig.baseOpts...)
	if err != nil {
		return Pod{}, err
	}
	defer func() {
		// Ensure the pod is closed if an error occurs during initialization
		if err != nil {
			_ = awsPod.Close()
		}
	}()
	client := s3.NewFromConfig(awsPod.Config(), func(options *s3.Options) {
		options.UsePathStyle = true
	})

	if podConfig.bucketName != "" {
		_, errCreate := client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: lo.EmptyableToPtr(podConfig.bucketName),
		})
		if errCreate != nil {
			return Pod{}, errCreate
		}
	}

	uploader := seedUploader{}
	if len(podConfig.baseOpts) > 0 || podConfig.seedFs != nil {
		uploader.bucketName = podConfig.bucketName
		uploader.client = client
	}

	if podConfig.seedFs != nil {
		if errSeed := blob.UploadAllFromFS(ctx, uploader, podConfig.seedFs); errSeed != nil {
			return Pod{}, errSeed
		}
	}

	if len(podConfig.seedBytes) > 0 {
		for i := range podConfig.seedBytes {
			errSeed := blob.UploadAll(ctx, uploader, blob.WithBatchUploadItemBytes(podConfig.seedBytes[i].key, podConfig.seedBytes[i].data))
			if errSeed != nil {
				return Pod{}, errSeed
			}
		}
	}

	return Pod{
		awsPod: awsPod,
		client: client,
	}, nil
}

// Client returns the S3 client for interacting with the LocalStack S3 service.
func (p Pod) Client() *s3.Client {
	return p.client
}

// Close terminates the LocalStack container and cleans up resources.
func (p Pod) Close() error {
	return p.awsPod.Close()
}

// -- Seed Uploader --

type seedUploader struct {
	bucketName string
	client     *s3.Client
}

var _ blob.ObjectUploader = (*seedUploader)(nil)

func (s seedUploader) Upload(ctx context.Context, key string, data io.Reader) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: lo.EmptyableToPtr(s.bucketName),
		Key:    lo.EmptyableToPtr(key),
		Body:   data,
	})
	return err
}

// -- Options --

type seedItem struct {
	key  string
	data []byte
}

type podOptions struct {
	baseOpts   []awstest.PodOption
	bucketName string
	seedFs     fs.FS
	seedBytes  []seedItem
}

// PodOption is a functional option type for configuring the S3 test pod.
type PodOption func(*podOptions)

// WithPodBaseOptions adds a set of basic options ([awstest.PodOption]) to the pod configuration.
func WithPodBaseOptions(opts ...awstest.PodOption) PodOption {
	return func(rootOpts *podOptions) {
		if rootOpts.baseOpts == nil {
			rootOpts.baseOpts = make([]awstest.PodOption, 0, len(opts))
		}
		rootOpts.baseOpts = append(rootOpts.baseOpts, opts...)
	}
}

// WithPodBucketName sets the name of the S3 bucket to be created in the LocalStack instance.
func WithPodBucketName(name string) PodOption {
	return func(opts *podOptions) {
		opts.bucketName = name
	}
}

// WithPodSeedFS sets the filesystem containing seed data to be used by the S3 pod.
func WithPodSeedFS(seedFs fs.FS) PodOption {
	return func(opts *podOptions) {
		opts.seedFs = seedFs
	}
}

// WithPodSeedBytes adds a key-value pair of seed data to be uploaded to the S3 pod.
func WithPodSeedBytes(key string, data []byte) PodOption {
	return func(opts *podOptions) {
		if key == "" || data == nil {
			return
		}
		if opts.seedBytes == nil {
			opts.seedBytes = make([]seedItem, 0, 1)
		}
		opts.seedBytes = append(opts.seedBytes, seedItem{
			key:  key,
			data: data,
		})
	}
}
