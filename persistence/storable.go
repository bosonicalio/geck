package persistence

// Storable is an interface indicating the implementing type can be stored into a persistence system.
type Storable interface {
	// IsNew checks if the type was just created.
	IsNew() bool
}

// NoopStorable is a no-op implementation of the [Storable] interface.
type NoopStorable struct {
	IsNewValue bool
}

// compile-time assertion
var _ Storable = (*NoopStorable)(nil)

// IsNew checks if the type was just created.
func (n NoopStorable) IsNew() bool {
	return n.IsNewValue
}
