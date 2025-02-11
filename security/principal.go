package security

// A Principal represents an entity in a system that can be authenticated. It could be a user,
// a system process, or a service.
type Principal interface {
	// ID identifier of the [Principal].
	ID() string
	// Authorities permissions or roles granted to a [Principal]. It specifies what a [Principal] is
	// allowed to do within the system.
	Authorities() []string
}
