package sqlfx

import (
	"database/sql"
	"time"

	"github.com/doug-martin/goqu/v9"
	"go.uber.org/fx"

	gecksql "github.com/hadroncorp/geck/persistence/sql"
)

type databaseInterceptorsDeps struct {
	fx.In
	Database     *sql.DB
	Interceptors []gecksql.DBInterceptor `group:"db_interceptors_sql"`
}

// InterceptorModule is a `uber/fx` module providing [gecksql.DBInterceptor] instances
// so driver-specific submodules can integrate additional behavior (e.g. observability, transaction contexts)
// into their concrete implementations of [gecksql.DB].
var InterceptorModule = fx.Module("geck/persistence/sql/interceptors",
	fx.Provide(
		func(deps databaseInterceptorsDeps) gecksql.DB {
			opts := make([]gecksql.DatabaseOption, 0, len(deps.Interceptors))
			for _, interceptor := range deps.Interceptors {
				opts = append(opts, gecksql.WithInterceptor(interceptor))
			}
			return gecksql.NewDB(deps.Database, opts...)
		},
	),
)

// ObservabilityModule is a `uber/fx` module providing [gecksql.DBInterceptor] instances
// so driver-specific submodules can integrate observability tools (i.e. logging, metrics, tracing)
// into their concrete implementations of [gecksql.DB].
//
// Requires to be declared along [InterceptorModule].
var ObservabilityModule = fx.Module("geck/persistence/sql/interceptors/observability",
	fx.Provide(
		AsDBInterceptor(gecksql.NewDatabaseLogger),
	),
)

// GoquModule is the `uber/fx` module of the [goqu] package.
var GoquModule = fx.Module("geck/persistence/sql/goqu",
	fx.Invoke(
		func() {
			goqu.SetTimeLocation(time.UTC)
		},
	),
)
