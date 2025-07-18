package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewEchoServer allocates a new [echo.Echo] instance with default configurations.
func NewEchoServer(opts ...ServerOption) *echo.Echo {
	config := &serverOptions{
		errorResponseCodec: "json", // Default error response format
	}
	for _, opt := range opts {
		opt(config)
	}
	e := echo.New()
	e.HTTPErrorHandler = NewErrorHandler(config.errorResponseCodec)
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())
	return e
}

// -- Options --

type serverOptions struct {
	errorResponseCodec string
}

// ServerOption is a function that modifies the server behaviors.
type ServerOption func(options *serverOptions)

// WithServerErrorResponseCodec sets the codec for error responses.
func WithServerErrorResponseCodec(format string) ServerOption {
	return func(opts *serverOptions) {
		opts.errorResponseCodec = format
	}
}
