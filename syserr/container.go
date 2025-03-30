package syserr

// Container is a sentinel structure holding a slice of errors.
type Container interface {
	// Unwrap retrieves the slice of errors.
	Unwrap() []error
}
