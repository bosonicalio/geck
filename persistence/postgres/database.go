package postgres

import (
	"context"
	"database/sql"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// NewConnectionPool creates a new connection pool for Postgres using the provided connection string.
//
// The connection string uses standard Postgres format:
// `postgres://user:password@host:port/database?sslmode=disable`.
//
// It returns a [sql.DB] instance that can be used to interact with the database.
//
// The function accepts optional configuration options to customize the connection pool.
func NewConnectionPool(ctx context.Context, connString string, opts ...ConnectionPoolOption) (*sql.DB, error) {
	config, err := newConnectionPoolConfig(connString)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(config)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDBFromPool(pool), nil
}

// -- Options --

// ConnectionPoolOption is a function type that modifies the pgxpool.Config for a Postgres connection pool.
type ConnectionPoolOption func(config *pgxpool.Config)

func newConnectionPoolConfig(connString string) (*pgxpool.Config, error) {
	pgxConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}
	pgxConfig.MaxConns = max(int32(runtime.NumCPU())*2, 4)
	pgxConfig.MinConns = min(int32(runtime.NumCPU())/4, 2)
	pgxConfig.MaxConnLifetime = time.Hour
	pgxConfig.MaxConnIdleTime = 30 * time.Minute
	pgxConfig.HealthCheckPeriod = 1 * time.Minute
	return pgxConfig, nil
}

// WithMaxConnections sets the maximum number of connections in the pool.
func WithMaxConnections(max int32) ConnectionPoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConns = max
	}
}

// WithMinConnections sets the minimum number of connections in the pool.
func WithMinConnections(min int32) ConnectionPoolOption {
	return func(config *pgxpool.Config) {
		config.MinConns = min
	}
}

// WithMaxConnLifetimeJitter sets the maximum lifetime jitter for connections in the pool.
func WithMaxConnLifetimeJitter(d time.Duration) ConnectionPoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConnLifetimeJitter = d
	}
}

// WithMaxConnLifetime sets the maximum lifetime of a connection in the pool.
func WithMaxConnLifetime(d time.Duration) ConnectionPoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConnLifetime = d
	}
}

// WithMaxConnIdleTime sets the maximum idle time of a connection in the pool.
func WithMaxConnIdleTime(d time.Duration) ConnectionPoolOption {
	return func(config *pgxpool.Config) {
		config.MaxConnIdleTime = d
	}
}

// WithMinIdleConnections sets the minimum number of idle connections in the pool.
func WithMinIdleConnections(min int32) ConnectionPoolOption {
	return func(config *pgxpool.Config) {
		config.MinConns = min
	}
}

// WithHealthCheckPeriod sets the interval for health checks on connections in the pool.
func WithHealthCheckPeriod(d time.Duration) ConnectionPoolOption {
	return func(config *pgxpool.Config) {
		config.HealthCheckPeriod = d
	}
}
