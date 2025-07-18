package http

import (
	"context"

	"github.com/labstack/echo/v4"

	"github.com/tesserical/geck/persistence"
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
//
// It uses the persistence API's transaction factories to create a transaction for each registered transaction
// factory.
// If no transaction factories are available, it returns a middleware that simply calls the next handler without
// wrapping it in a transaction.
//
// This middleware is useful for ensuring that all database operations within a request are executed within a transaction,
// allowing for rollback in case of errors.
//
// It is pretty important to add transactional factories to the registry before using this middleware. This can be done
// by using the `persistence.RegisterTxFactory` function in the application initialization phase.
//
// This middleware will look up all registered transaction factories and chain them together so even with many
// database backends.
//
// If you want to use a custom transaction factory, you can use the [WithTxFactory] option to set it.
// If a custom transaction factory is provided, it will be used directly, bypassing the global registry of
// transaction factories.
func Transactional(opts ...TransactionalOption) echo.MiddlewareFunc {
	config := &transactionalOptions{}
	for _, opt := range opts {
		opt(config)
	}
	if config.txFactory != nil {
		// If a custom transaction factory is provided, use it directly.
		return transactional(config.txFactory)
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return persistence.ExecInTxAll(c.Request().Context(), func(ctx context.Context) error {
				c.SetRequest(c.Request().WithContext(ctx))
				return next(c)
			})
		}
	}

	// This is the original implementation that uses the global registry of transaction factories, chaining them
	// with echo middleware functions.

	//factories := persistence.GetTxFactories()
	//if len(factories) == 0 {
	//	return func(next echo.HandlerFunc) echo.HandlerFunc {
	//		return next
	//	}
	//}
	//mws := make([]echo.MiddlewareFunc, 0, len(factories))
	//for i := range factories {
	//	mws = append(mws, transactional(factories[i]))
	//}
	//return chain(mws...)
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
	txFactory persistence.TxFactory
}

// TransactionalOption is a functional option type for configuring the Transactional middleware.
type TransactionalOption func(*transactionalOptions)

// WithTxFactory allows to set a custom transaction factory for the Transactional middleware.
//
// This will cause the middleware to use the provided factory instead of looking up the global registry of
// transaction factories.
func WithTxFactory(factory persistence.TxFactory) TransactionalOption {
	return func(opts *transactionalOptions) {
		opts.txFactory = factory
	}
}
