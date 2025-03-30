package identity

// A Principal represents an entity in a system that can be authenticated. It could be a user,
// a system process, or a service.
type Principal interface {
	// ID identifier of the [Principal].
	ID() string
	// Authorities permissions or roles granted to a [Principal]. It specifies what a [Principal] is
	// allowed to do within the system.
	Authorities() []string
}

// -- Basic --

// BasicPrincipal is a simple implementation of the [Principal] interface.
type BasicPrincipal struct {
	id          string
	authorities []string
}

// compile-time assertion
var _ Principal = (*BasicPrincipal)(nil)

// NewBasicPrincipal creates a new [BasicPrincipal] with the given ID and authorities.
func NewBasicPrincipal(id string, authorities ...string) BasicPrincipal {
	return BasicPrincipal{
		id:          id,
		authorities: authorities,
	}
}

// ID identifier of the [Principal].
func (p BasicPrincipal) ID() string {
	return p.id
}

// Authorities permissions or roles granted to a [Principal]. It specifies what a [Principal] is
// allowed to do within the system.
func (p BasicPrincipal) Authorities() []string {
	return p.authorities
}
