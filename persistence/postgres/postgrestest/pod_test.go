package postgrestest_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tesserical/geck/persistence/postgres/postgrestest"
)

func TestNewPod(t *testing.T) {
	pod, err := postgrestest.NewPod(t.Context(),
		postgrestest.WithPodImageTag("16-alpine"),
		postgrestest.WithPodDatabaseName("mytestdb"),
		postgrestest.WithPodMigrationsFS(os.DirFS("testdata/migration")),
		postgrestest.WithPodSeedFS(os.DirFS("testdata/seed")),
	)
	require.NoError(t, err)
	assert.NotNil(t, pod.Client())
	defer func() {
		errClose := pod.Close()
		assert.NoError(t, errClose)
	}()
	err = pod.Client().Ping()
	assert.NoError(t, err)
}
