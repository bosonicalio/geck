package application

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/tesserical/geck/environment"
	"github.com/tesserical/geck/version"
)

// Application is a configuration structure for applications, containing basic metadata for an application.
type Application struct {
	Name        string
	Version     version.Version
	Environment environment.Environment
	InstanceID  string
}

// compile-time assertion
var _ fmt.Stringer = (*Application)(nil)

// New allocates an [Application].
//
// If [Application.InstanceID] is not set, it generates a new UUID v7 instance ID.
func New(opts ...Option) (Application, error) {
	app := Application{}
	for _, opt := range opts {
		opt(&app)
	}
	if app.InstanceID != "" {
		return app, nil
	}
	id, err := uuid.NewV7()
	if err != nil {
		return Application{}, err
	}
	app.InstanceID = id.String()
	return app, nil
}

// InstanceName returns the instance name of the application.
func (c Application) InstanceName() string {
	return c.Name + "-" + c.InstanceID
}

// FullInstanceName returns the full instance name of the application, including the environment.
func (c Application) FullInstanceName() string {
	return c.Name + "-" + c.Environment.String() + "-" + c.InstanceID
}

// String returns a string representation of the application configuration.
func (c Application) String() string {
	return c.Name + " (" + c.Version.String() + ") [" + c.Environment.String() + "] " + c.InstanceID
}

// -- Options --

type Option func(*Application)

// WithName sets the name of the application.
func WithName(name string) Option {
	return func(app *Application) {
		app.Name = name
	}
}

// WithVersion sets the version of the application.
func WithVersion(v version.Version) Option {
	return func(app *Application) {
		app.Version = v
	}
}

// WithEnvironment sets the environment of the application.
func WithEnvironment(env environment.Environment) Option {
	return func(app *Application) {
		app.Environment = env
	}
}

// WithInstanceID sets the instance ID of the application.
func WithInstanceID(id string) Option {
	return func(app *Application) {
		app.InstanceID = id
	}
}
