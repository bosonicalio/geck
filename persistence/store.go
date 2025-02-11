package persistence

// Storable is an interface indicating the implementing type can be stored into a persistence system.
type Storable interface {
	// IsNew checks if the type was just created.
	IsNew() bool
}
