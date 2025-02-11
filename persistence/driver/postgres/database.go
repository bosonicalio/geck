package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	gecksql "github.com/hadroncorp/geck/persistence/sql"
)

type DBConfig struct {
	gecksql.DBConfig
	MaxConnLifetimeJitter time.Duration `env:"PSQL_MAX_CONN_LIFETIME_JITTER"`
	HealthCheckInterval   time.Duration `env:"PSQL_HEALTHCHECK_INTERVAL" envDefault:"5s"`
}

// NewPooledDB allocates a [sql.DB] instance.
//
// It uses a custom pooling mechanism provided by the package `jackc/pgx`
// specially tuned for Postgres.
func NewPooledDB(config DBConfig) (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.InitConnectTimeout)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(config.ConnectionString)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConnLifetime = config.MaxConnLifetimeJitter
	poolConfig.MaxConnLifetimeJitter = config.MaxConnLifetimeJitter
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.MaxConns = int32(config.MaxConnections)
	poolConfig.MinConns = int32(config.MinConnections)
	poolConfig.HealthCheckPeriod = config.HealthCheckInterval

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDBFromPool(pool), nil
}
