package http

import (
	"context"
	"errors"

	"github.com/labstack/echo/v4"

	"github.com/bosonicalio/geck/persistence"
)

// - Utilities -

// Chain is a helper function that chains multiple echo.MiddlewareFunc into one.
// It processes the request in the order the middlewares are passed.
func Chain(middlewares ...echo.MiddlewareFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		// Apply the middlewares in reverse order. The last middleware
		// will be the innermost one, wrapping the final handler.
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// - Persistence -

// Transactional is an Echo middleware that wraps the request in a database transaction.
// If the request handler returns an error, the transaction will
// be rolled back. If the request handler completes successfully, the transaction will be committed.
//
// If a [persistence.TxManager] is provided via the [WithTxManager] option, the request handler will be executed
// within a set of transactions managed by the provided manager. This allows for executing multiple database transactions
// in a single request, which can be useful for complex operations that require multiple steps to be
// executed atomically.
//
// If you want to use a specific transaction factory, you can enable it by passing the [WithTxFactory] option to set it.
func Transactional(opts ...TransactionalOption) echo.MiddlewareFunc {
	config := &transactionalOptions{}
	for _, opt := range opts {
		opt(config)
	}
	if config.txFactory != nil {
		// If a custom transaction factory is provided, use it directly.
		return transactional(config.txFactory)
	} else if config.txManager == nil {
		panic(errors.New("geck.http: Transactional middleware requires a transaction manager or factory"))
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return config.txManager.Execute(c.Request().Context(), func(ctx context.Context) error {
				// Set the context with the transaction for the next handler.
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			})
		}
	}
}

func transactional(factory persistence.TxFactory) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return persistence.ExecInTx(c.Request().Context(), factory, func(ctx context.Context) error {
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			})
		}
	}
}

// -- Options --

type transactionalOptions struct {
	txManager *persistence.TxManager
	txFactory persistence.TxFactory
}

// TransactionalOption is a functional option type for configuring the Transactional middleware.
type TransactionalOption func(*transactionalOptions)

// WithTxManager allows to set a custom transaction manager for the Transactional middleware.
// This will cause the middleware to run HTTP requests within transactions managed by the provided manager.
func WithTxManager(manager *persistence.TxManager) TransactionalOption {
	return func(opts *transactionalOptions) {
		opts.txManager = manager
	}
}

// WithTxFactory allows to set a custom transaction factory for the Transactional middleware.
//
// This will cause the middleware to use the provided factory instead of using a [persistence.TxManager].
func WithTxFactory(factory persistence.TxFactory) TransactionalOption {
	return func(opts *transactionalOptions) {
		opts.txFactory = factory
	}
}
