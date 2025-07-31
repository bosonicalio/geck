package http

import (
	"github.com/labstack/echo/v4"

	"github.com/bosonicalio/geck/application"
)

// Controller is a transport component exposing a set of endpoints through the HTTP protocol
// to give access to external system clients.
type Controller interface {
	// SetEndpoints takes `e` and sets one or many endpoints.
	SetEndpoints(e *echo.Echo)
	// SetVersionedEndpoints takes `g` and sets one or many endpoints.
	//
	// The `g` argument MUST have the major version of the system as path prefix.
	SetVersionedEndpoints(g *echo.Group)
}

// RegisterServerEndpoints registers `ctrls` (slice of [Controller]) into `e` ([echo.Echo]).
//
// Uses [application.Application.Version].Major() value as path prefix for versioned endpoints.
func RegisterServerEndpoints(e *echo.Echo, config application.Application, ctrls []Controller) string {
	pathPrefix := "/" + config.Version.Major()
	versionedGroup := e.Group(pathPrefix)
	for _, controller := range ctrls {
		controller.SetEndpoints(e)
		controller.SetVersionedEndpoints(versionedGroup)
	}
	return pathPrefix
}
