package sqlfx

import (
	"go.uber.org/fx"

	gecksql "github.com/tesserical/geck/persistence/sql"
)

// AsDB annotates `t` as a [gecksql.DB] implementation.
//
// This annotation only works for `uber/fx` providers.
func AsDB(t any) any {
	return fx.Annotate(
		t,
		fx.As(new(gecksql.DB)),
	)
}

// AsDBInterceptor annotates `t` as a [gecksql.DBInterceptor] implementation.
//
// This annotation only works for `uber/fx` providers.
//
// In addition, it adds `t` into the SQL database interceptor group, meaning the dependency framework will
// aggregate all components annotated by this routine to later offer them to other components in form of a slice
// for its usage.
func AsDBInterceptor(t any) any {
	return fx.Annotate(
		t,
		fx.As(new(gecksql.DBInterceptor)),
		fx.ResultTags(`group:"db_interceptors_sql"`),
	)
}
