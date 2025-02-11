package identifier

// Factory is a builder component that generates identifiers.
type Factory interface {
	// NewID generates a new identifier.
	NewID() (string, error)
}
