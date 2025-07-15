//go:build integration

package s3_test

import (
	"bytes"
	"io"
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	gecks3 "github.com/tesserical/geck/blob/s3"
	"github.com/tesserical/geck/blob/s3/s3test"
	"github.com/tesserical/geck/cloud/aws/awstest"
)

func TestBucket_Upload(t *testing.T) {
	// arrange
	bucketName := strconv.FormatUint(rand.Uint64(), 10)
	pod, err := s3test.NewPod(t.Context(),
		s3test.WithPodBucketName(bucketName),
		s3test.WithPodBaseOptions(awstest.WithPodImageTag("4.6")),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, pod.Close())
	}()

	tests := []struct {
		name   string
		inKey  string
		in     io.Reader
		expErr error
	}{
		{
			name:   "Should upload file successfully",
			inKey:  "test-key",
			in:     bytes.NewReader([]byte("hello world")),
			expErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(scopedT *testing.T) {
			// arrange
			bucket := gecks3.NewBucket(bucketName, pod.Client())
			// act
			err = bucket.Upload(scopedT.Context(), tt.inKey, tt.in)
			// assert
			assert.ErrorIs(scopedT, err, tt.expErr)
		})
	}
}

func TestBucket_Remove(t *testing.T) {
	// arrange
	bucketName := strconv.FormatUint(rand.Uint64(), 10)
	pod, err := s3test.NewPod(t.Context(),
		s3test.WithPodBucketName(bucketName),
		s3test.WithPodBaseOptions(awstest.WithPodImageTag("4.6")),
		s3test.WithPodSeedBytes("test-key", []byte("hello world")),
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, pod.Close())
	}()

	tests := []struct {
		name   string
		inKey  string
		expErr bool
	}{
		{
			name:   "Should remove file successfully",
			inKey:  "test-key",
			expErr: false,
		},
		{
			name:   "Should not return error when trying to remove non-existent file",
			inKey:  "test-key-2",
			expErr: false, // idempotent operation, should not return an error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(scopedT *testing.T) {
			// arrange
			bucket := gecks3.NewBucket(bucketName, pod.Client())
			// act
			err = bucket.Remove(scopedT.Context(), tt.inKey)
			// assert
			assert.Equal(scopedT, tt.expErr, err != nil)
		})
	}
}
