package awstest_test

import (
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bosonicalio/geck/cloud/aws/awstest"
)

func TestNewPod(t *testing.T) {
	pod, err := awstest.NewPod(t.Context(),
		awstest.WithPodImageTag("4.6.0"),
		awstest.WithPodServices("s3"),
	)
	require.NoError(t, err)
	require.NotNil(t, pod)
	baseEndpoint := lo.FromPtr(pod.Config().BaseEndpoint)
	assert.NotEmpty(t, baseEndpoint)
}
