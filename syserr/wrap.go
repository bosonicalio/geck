package syserr

// Unwrapper is construct holding a set of errors.
//
// Its main usage is by casting this from an error interface to retrieve the slice of errors joined with [errors.Join]
// (or any structure implementing [errors.Unwrap]).
type Unwrapper interface {
	// Unwrap retrieves the slice of errors.
	Unwrap() []error
}
