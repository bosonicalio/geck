package postgrestest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// -- Container --

// Container represents a postgres container for testing.
type Container struct {
	Instance         testcontainers.Container
	ConnectionString string
}

// NewContainer creates and starts a Postgres container with configurations for testing scenarios.
func NewContainer(ctx context.Context, t *testing.T, opts ...ContainerOption) (*Container, error) {
	t.Helper() // Marks this function as a test helper

	options := containerOptions{
		connString: "postgres://user:password@localhost:5432/dbname",
	}
	for _, opt := range opts {
		opt(&options)
	}

	config, err := pgx.ParseConfig(options.connString)
	if err != nil {
		return nil, err
	}

	container, err := postgres.Run(ctx,
		fmt.Sprintf("postgres:%s", lo.CoalesceOrEmpty(options.imageTag, "alpine")),
		postgres.WithDatabase(config.Database),
		postgres.WithUsername(config.User),
		postgres.WithPassword(config.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	connString, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	return &Container{
		Instance:         container,
		ConnectionString: connString,
	}, nil
}

// --- Option(s) ---

type containerOptions struct {
	connString string
	imageTag   string
}

// ContainerOption represents an option for the container.
type ContainerOption func(*containerOptions)

// WithContainerConnString sets the connection string for the container.
//
// Port will be ignored as containers will be bind to a dynamic port to avoid collisions between
// exposed ports.
func WithContainerConnString(connString string) ContainerOption {
	return func(o *containerOptions) {
		o.connString = connString
	}
}

// WithContainerImageTag sets the image tag for the container.
func WithContainerImageTag(imageTag string) ContainerOption {
	return func(o *containerOptions) {
		o.imageTag = imageTag
	}
}

// -- Test Runners --

// WithTestDatabase runs a test with a provisioned database
func WithTestDatabase(ctx context.Context, t *testing.T, test func(db *sql.DB)) {
	t.Helper()

	container, err := NewContainer(ctx, t)
	if err != nil {
		t.Fatalf("Failed to create container: %v", err)
	}
	// Setup database
	db, err := sql.Open("postgres", container.ConnectionString)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Run the test with the database
	test(db)

	// Cleanup
	_ = db.Close()
	_ = container.Instance.Terminate(context.Background())
}
