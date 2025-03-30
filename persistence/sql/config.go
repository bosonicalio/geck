package sql

import "time"

// DBConfig is a structure used by factory routines generating sql.DB instances
// to define pooling/client general-purpose settings.
//
// Embed this structure to a driver-specific DBConfig to extend these properties.
type DBConfig struct {
	// ConnectionString is the connection string used to connect to the database.
	ConnectionString string `env:"SQL_CONNECTION_STRING,unset"`
	// InitConnectTimeout is the maximum time to wait for a connection to be established.
	InitConnectTimeout time.Duration `env:"SQL_INIT_CONNECT_TIMEOUT" envDefault:"15s"`
	// MaxConnLifetime is the maximum amount of time a connection may be reused.
	MaxConnLifetime time.Duration `env:"SQL_MAX_CONN_LIFETIME" envDefault:"15s"`
	// MaxConnIdleTime is the maximum amount of time a connection may be idle.
	MaxConnIdleTime time.Duration `env:"SQL_MAX_CONN_IDLE_TIME" envDefault:"60s"`
	// MaxConnections is the maximum number of open connections to the database.
	MaxConnections int64 `env:"SQL_MAX_CONNS" envDefault:"4"`
	// MinConnections is the minimum number of open connections to the database.
	MinConnections int64 `env:"SQL_MIN_CONNS"`
}
