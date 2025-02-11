package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewEchoServer allocates a new [echo.Echo] instance with default configurations.
func NewEchoServer(config ServerConfig) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = NewErrorHandler(config)
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	return e
}

// StartServer starts `e` ([echo.Echo]) server, listening on [ServerConfig.Address].
func StartServer(e *echo.Echo, config ServerConfig) error {
	return e.Start(config.Address)
}
