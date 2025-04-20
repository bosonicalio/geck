package kafka

import (
	"github.com/twmb/franz-go/pkg/kgo"
)

// SkipFunc is a function that determines whether to skip a middleware.
type SkipFunc func(msg *kgo.Record) bool

// -- Options --

type InterceptorOptions struct {
	Skip SkipFunc
}

// InterceptorOption is a function that modifies the options for the interceptor.
type InterceptorOption func(*InterceptorOptions)

// WithSkipInterceptor is an option to skip the interceptor for certain messages.
func WithSkipInterceptor(skip SkipFunc) InterceptorOption {
	return func(o *InterceptorOptions) {
		o.Skip = skip
	}
}

// -- Reader --

// ReaderInterceptor is a routine to be executed before/after (depending on the implementation) for
// Apache Kafka readers. These routines can be chained to achieve additional behaviors.
type ReaderInterceptor func(next ReaderHandlerFunc) ReaderHandlerFunc
