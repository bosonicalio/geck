package application

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"

	"github.com/hadroncorp/geck/environment"
	"github.com/hadroncorp/geck/version"
)

// Config is a configuration structure for applications.
// Contains general properties for a basic application.
type Config struct {
	Name        string                  `env:"APP_NAME"`
	Version     version.Version         `env:"APP_VERSION" envDefault:"v0.1.0-alpha"`
	Environment environment.Environment `env:"APP_ENVIRONMENT" envDefault:"local"`
	InstanceID  string                  `env:"APP_INSTANCE_ID"`
}

// compile-time assertion
var _ fmt.Stringer = (*Config)(nil)

// New creates a new application configuration ([Config]).
//
// It uses the [env] package to parse environment variables and populate the configuration structure.
//
// If [Config.InstanceID] is not set, it generates a new UUID v7 instance ID.
func New() (Config, error) {
	config, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}

	if config.InstanceID != "" {
		return config, nil
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Config{}, err
	}
	config.InstanceID = id.String()
	return config, nil
}

// InstanceName returns the instance name of the application.
func (c Config) InstanceName() string {
	return c.Name + "-" + c.InstanceID
}

// FullInstanceName returns the full instance name of the application, including the environment.
func (c Config) FullInstanceName() string {
	return c.Name + "-" + c.Environment.String() + "-" + c.InstanceID
}

// String returns a string representation of the application configuration.
func (c Config) String() string {
	return c.Name + " (" + c.Version.String() + ") [" + c.Environment.String() + "] " + c.InstanceID
}
