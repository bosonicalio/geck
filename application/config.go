package application

import (
	"github.com/hadroncorp/geck/environment"
	"github.com/hadroncorp/geck/version"
)

// Config is a configuration structure for applications.
// Contains general properties for a basic application.
type Config struct {
	Name        string                  `env:"APP_NAME"`
	Version     version.Version         `env:"APP_VERSION" envDefault:"v0.1.0-alpha"`
	Environment environment.Environment `env:"APP_ENVIRONMENT" envDefault:"local"`
}
