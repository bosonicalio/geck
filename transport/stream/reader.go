package stream

import (
	"context"
	"log/slog"
)

type HandlerFunc func(ctx context.Context, message Message) error

type ReaderInterceptorFunc func(next HandlerFunc) HandlerFunc

func NewLogMessageInterceptor(logger *slog.Logger) ReaderInterceptorFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, message Message) error {
			logger.Info("received message", slog.String("key", message.Key))
			return next(ctx, message)
		}
	}
}

func NewRecoverInterceptor(logger *slog.Logger) ReaderInterceptorFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, message Message) error {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic recovered", slog.Any("panic", r))
				}
			}()
			return next(ctx, message)
		}
	}
}

type ReaderManager interface {
	Register(name string, handler HandlerFunc)
}

type Controller interface {
	RegisterReaders()
}
