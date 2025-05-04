package postgresfx

import (
	"github.com/caarlos0/env/v11"
	"go.uber.org/fx"

	"github.com/tesserical/geck/persistence/postgres"
)

// Module is the `uber/fx` module for `geck` Persistence API Postgres integrations.
var Module = fx.Module("geck/persistence/postgres",
	fx.Provide(
		env.ParseAs[postgres.DBConfig],
		postgres.NewPooledDB,
	),
)
