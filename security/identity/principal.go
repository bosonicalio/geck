package identity

import "github.com/samber/lo"

// A Principal represents an entity in a system that can be authenticated. It could be a user,
// a system process, or a service.
type Principal interface {
	// ID identifier of the [Principal].
	ID() string
	// Authorities permissions or roles granted to a [Principal]. It specifies what a [Principal] is
	// allowed to do within the system.
	Authorities() []string
	// HasAuthority checks if the [Principal] has a specific authority.
	HasAuthority(authority string) bool
	// HasAllAuthorities checks if the [Principal] has all the specified authorities.
	//
	// Returns false if no authorities are provided.
	HasAllAuthorities(authorities ...string) bool
	// HasAnyAuthorities checks if the [Principal] has any of the specified authorities.
	//
	// Returns false if no authorities are provided.
	HasAnyAuthorities(authorities ...string) bool
}

// -- Basic --

// BasicPrincipal is a simple implementation of the [Principal] interface.
type BasicPrincipal struct {
	id             string
	authoritiesSet map[string]struct{}
}

// compile-time assertion
var _ Principal = (*BasicPrincipal)(nil)

// NewBasicPrincipal creates a new [BasicPrincipal] with the given ID and authorities.
func NewBasicPrincipal(id string, authorities ...string) BasicPrincipal {
	return BasicPrincipal{
		id:             id,
		authoritiesSet: lo.Keyify(authorities),
	}
}

// ID identifier of the [Principal].
func (p BasicPrincipal) ID() string {
	return p.id
}

// Authorities permissions or roles granted to a [Principal]. It specifies what a [Principal] is
// allowed to do within the system.
func (p BasicPrincipal) Authorities() []string {
	return lo.Keys(p.authoritiesSet)
}

// HasAuthority checks if the [Principal] has a specific authority.
func (p BasicPrincipal) HasAuthority(authority string) bool {
	_, ok := p.authoritiesSet[authority]
	return ok
}

// HasAllAuthorities checks if the [Principal] has all the specified authorities.
//
// Returns false if no authorities are provided.
func (p BasicPrincipal) HasAllAuthorities(authorities ...string) bool {
	if len(authorities) == 0 {
		return false
	}
	for _, authority := range authorities {
		if !p.HasAuthority(authority) {
			return false
		}
	}
	return true
}

// HasAnyAuthorities checks if the [Principal] has any of the specified authorities.
//
// Returns false if no authorities are provided.
func (p BasicPrincipal) HasAnyAuthorities(authorities ...string) bool {
	if len(authorities) == 0 {
		return false
	}
	for _, authority := range authorities {
		if p.HasAuthority(authority) {
			return true
		}
	}
	return false
}
