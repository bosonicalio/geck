package sql

import "time"

// DBConfig is a structure used by factory routines generating sql.DB instances
// to define pooling/client general-purpose settings.
//
// Embed this structure to a driver-specific DBConfig to extend these properties.
type DBConfig struct {
	ConnectionString   string        `env:"SQL_CONNECTION_STRING,unset"`
	InitConnectTimeout time.Duration `env:"SQL_INIT_CONNECT_TIMEOUT" envDefault:"15s"`
	MaxConnLifetime    time.Duration `env:"SQL_MAX_CONN_LIFETIME" envDefault:"15s"`
	MaxConnIdleTime    time.Duration `env:"SQL_MAX_CONN_IDLE_TIME" envDefault:"60s"`
	MaxConnections     int64         `env:"SQL_MAX_CONNS" envDefault:"4"`
	MinConnections     int64         `env:"SQL_MIN_CONNS"`
}
