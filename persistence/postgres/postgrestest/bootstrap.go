package postgrestest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/caarlos0/env/v11"

	"github.com/hadroncorp/geck/persistence/postgres"
	"github.com/hadroncorp/geck/persistence/sqltest"
)

// StartContainer starts a container and runs migrations.
func StartContainer(ctx context.Context, t *testing.T, c *Container, dir string) (*sql.DB, error) {
	t.Helper()
	if err := c.Instance.Start(ctx); err != nil {
		return nil, err
	}
	t.Setenv("SQL_CONNECTION_STRING", c.ConnectionString)
	psqlConfig, err := env.ParseAs[postgres.DBConfig]()
	if err != nil {
		return nil, err
	}
	psqlDB, err := postgres.NewPooledDB(psqlConfig)
	if err != nil {
		return nil, err
	}
	return psqlDB, sqltest.RunGooseMigrations(psqlDB, "postgres", dir)
}
