package postgrestest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	geckpostgres "github.com/bosonicalio/geck/persistence/postgres"
	"github.com/bosonicalio/geck/persistence/sqltest"
	"github.com/bosonicalio/geck/testutil"
)

// Pod is a test component for running a Postgres docker instance.
type Pod struct {
	baseCtx   context.Context
	container *postgres.PostgresContainer
	client    *sql.DB
}

// compile-time assertions
var _ testutil.Pod = (*Pod)(nil)

// NewPod creates a new Postgres test container instance.
//
// It accepts options to configure the image tag, database name, and migrations filesystem.
//
// If migrations are provided, they will be executed after the container is started.
//
// If seed data is provided, it will be executed after the migrations using transactions to ensure operation atomicity.
func NewPod(ctx context.Context, opts ...PodOption) (Pod, error) {
	podConfig := newPodOptions()
	for _, opt := range opts {
		opt(podConfig)
	}

	container, err := postgres.Run(ctx, fmt.Sprintf("postgres:%s", podConfig.imageTag),
		postgres.WithDatabase(podConfig.databaseName),
		postgres.WithUsername("some_user"),
		postgres.WithPassword("some_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return Pod{}, err
	}
	defer func() {
		// Ensure the container is terminated if an error occurs
		if err != nil && container != nil {
			_ = container.Terminate(ctx)
		}
	}()

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		return Pod{}, err
	}
	client, err := geckpostgres.NewConnectionPool(ctx, connString)
	if err != nil {
		return Pod{}, err
	}

	if podConfig.migrationsFs != nil {
		if errRunMigrations := sqltest.RunMigrations(ctx, "postgres", client, podConfig.migrationsFs, ""); errRunMigrations != nil {
			return Pod{}, errRunMigrations
		}
	}

	if podConfig.seedFs != nil {
		if errRunSeeds := sqltest.RunSeeds(ctx, client, podConfig.seedFs); errRunSeeds != nil {
			return Pod{}, errRunSeeds
		}
	}

	return Pod{
		baseCtx:   ctx,
		container: container,
		client:    client,
	}, nil
}

// Client returns the SQL client for the Postgres container.
func (p Pod) Client() *sql.DB {
	return p.client
}

// Close terminates the Postgres container and closes the SQL client connection.
func (p Pod) Close() error {
	if p.container == nil && p.client == nil {
		return nil
	}

	errs := make([]error, 0)
	if p.client != nil {
		if err := p.client.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if p.container != nil {
		if err := p.container.Terminate(context.Background()); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// -- Options --

type podOptions struct {
	imageTag     string
	databaseName string
	migrationsFs fs.FS
	seedFs       fs.FS
}

type PodOption func(*podOptions)

func newPodOptions() *podOptions {
	return &podOptions{
		imageTag:     "alpine",
		databaseName: "testdb",
	}
}

// WithPodImageTag sets the image tag for the Postgres container.
func WithPodImageTag(imageTag string) PodOption {
	return func(o *podOptions) {
		o.imageTag = imageTag
	}
}

// WithPodDatabaseName sets the database name for the Postgres container.
func WithPodDatabaseName(databaseName string) PodOption {
	return func(o *podOptions) {
		o.databaseName = databaseName
	}
}

// WithPodMigrationsFS sets the filesystem for migrations.
func WithPodMigrationsFS(fs fs.FS) PodOption {
	return func(o *podOptions) {
		o.migrationsFs = fs
	}
}

// WithPodSeedFS sets the filesystem for seed data.
func WithPodSeedFS(fs fs.FS) PodOption {
	return func(o *podOptions) {
		o.seedFs = fs
	}
}
