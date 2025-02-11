package postgresfx

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"

	"github.com/hadroncorp/geck/persistence/driver/postgres"
)

// Module is the `uber/fx` module for `geck` Persistence API Postgres integrations.
var Module = fx.Module("geck/persistence/driver/postgres",
	fx.Provide(
		env.ParseAs[postgres.DBConfig],
		postgres.NewPooledDB,
	),
)
