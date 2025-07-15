package testutil

import (
	"io"
)

// Pod is a utility component providing users with a way to run and manage a hermetic environment for testing purposes.
//
// It is designed to be used in testing scenarios where a controlled environment is required, such as for integration tests.
// Moreover, works as an aggregate for a range of infrastructure and their properly configured clients.
type Pod interface {
	io.Closer
}
