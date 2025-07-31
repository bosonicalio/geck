package application_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bosonicalio/geck/application"
	"github.com/bosonicalio/geck/environment"
	"github.com/bosonicalio/geck/version"
)

func TestNew(t *testing.T) {
	app, err := application.New(
		application.WithName("TestApp"),
		application.WithVersion(version.MustParse("v10.5.150-beta.1")),
		application.WithEnvironment(environment.Staging),
		application.WithInstanceID("test-instance-id"),
	)
	assert.NoError(t, err)
	assert.Equal(t, "TestApp", app.Name)
	assert.Equal(t, "v10.5.150-beta.1", app.Version.String())
	assert.Equal(t, environment.Staging.String(), app.Environment.String())
	assert.Equal(t, "test-instance-id", app.InstanceID)

	// Test the default values
	app, err = application.New(
		application.WithName("TestApp2"),
		application.WithVersion(version.MustParse("v11.5.150-beta.1")),
		application.WithEnvironment(environment.Production),
	)
	assert.NoError(t, err)
	assert.Equal(t, "TestApp2", app.Name)
	assert.Equal(t, "v11.5.150-beta.1", app.Version.String())
	assert.Equal(t, environment.Production.String(), app.Environment.String())
	assert.NotEmpty(t, app.InstanceID)
	instanceID, err := uuid.Parse(app.InstanceID)
	assert.NoError(t, err)
	assert.NotNil(t, instanceID)
}
