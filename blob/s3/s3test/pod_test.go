package s3test_test

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tesserical/geck/blob/s3/s3test"
	"github.com/tesserical/geck/cloud/aws/awstest"
)

func TestNewPod(t *testing.T) {
	bucketName := "test-bucket"
	pod, err := s3test.NewPod(t.Context(),
		s3test.WithPodBaseOptions(
			awstest.WithPodImageTag("4.6"),
		),
		s3test.WithPodBucketName(bucketName),
		s3test.WithPodSeedBytes("test-key-0", []byte("test-data-0")),
		s3test.WithPodSeedFS(os.DirFS("testdata")),
	)
	require.NoError(t, err)
	require.NotNil(t, pod)
	require.NotNil(t, pod.Client())
	defer func() {
		assert.NoError(t, pod.Close())
	}()
	res, err := pod.Client().HeadBucket(t.Context(), &s3.HeadBucketInput{
		Bucket:              lo.EmptyableToPtr(bucketName),
		ExpectedBucketOwner: lo.EmptyableToPtr("000000000000"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, lo.FromPtr(res.BucketRegion))
	resObj, err := pod.Client().HeadObject(t.Context(), &s3.HeadObjectInput{
		Bucket: lo.EmptyableToPtr(bucketName),
		Key:    lo.EmptyableToPtr("test-key-0"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, lo.FromPtr(resObj.ETag))
	resObj, err = pod.Client().HeadObject(t.Context(), &s3.HeadObjectInput{
		Bucket: lo.EmptyableToPtr(bucketName),
		Key:    lo.EmptyableToPtr("some-text.txt"),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, lo.FromPtr(resObj.ETag))
}
