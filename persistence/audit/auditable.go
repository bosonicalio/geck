package audit

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/tesserical/geck/persistence"
	"github.com/tesserical/geck/security/identity"
)

const _defaultPrincipalUsername = "unknown"

// Auditable is a structure provisioning metadata for persistence operations.
//
// Embed this structure into your entities/aggregates to enhance and control write operations.
// Call [New] routine to create an instance with default values.
//
// Implements [persistence.Storable] interface.
type Auditable struct {
	CreateTime     time.Time
	CreateBy       string
	LastUpdateTime time.Time
	LastUpdateBy   string
	Version        int64
	IsDeleted      bool

	loc *time.Location
}

// compile-time assertions
var _ persistence.Storable = (*Auditable)(nil)

// New allocates a new [Auditable] instance using default values.
//
// This routine takes `ctx` argument to retrieve the [identity.Principal].
// If no principal is found, an `unknown` value will be placed instead.
// The principal is used to set the `CreateBy` and `LastUpdateBy` fields.
//
// Use [AuditableOption] routines to customize how the instance is created.
func New(ctx context.Context, opts ...AuditableOption) Auditable {
	auditable := &Auditable{}
	for _, opt := range opts {
		opt(auditable)
	}

	principal, _ := identity.GetPrincipal(ctx)
	var username string
	if principal != nil {
		username = principal.ID()
	}
	username = lo.CoalesceOrEmpty(username, _defaultPrincipalUsername)
	now := time.Now().In(lo.CoalesceOrEmpty(auditable.loc, time.UTC))

	auditable.CreateBy = username
	auditable.CreateTime = lo.CoalesceOrEmpty(auditable.CreateTime, now)
	auditable.LastUpdateBy = username
	auditable.LastUpdateTime = lo.CoalesceOrEmpty(auditable.LastUpdateTime, auditable.CreateTime)
	auditable.Version = lo.CoalesceOrEmpty(auditable.Version)
	return *auditable
}

// IsNew checks if the type was just created.
func (a Auditable) IsNew() bool {
	return a.Version == 0
}

// Touch increases the version, updates last update fields, both time and by.
//
// This routine takes `ctx` argument to retrieve the [identity.Principal] instance performing
// the operation. If no principal is found, an `unknown` value will be placed instead.
func Touch(ctx context.Context, auditable *Auditable) {
	auditable.Version++
	auditable.LastUpdateTime = time.Now().In(auditable.LastUpdateTime.Location())
	var username string
	principal, _ := identity.GetPrincipal(ctx)
	if principal != nil {
		username = principal.ID()
	}
	auditable.LastUpdateBy = lo.CoalesceOrEmpty(username, _defaultPrincipalUsername)
}

// SoftDelete Marks `auditable` as deleted. It also increases the version, updates last update
// fields, both time and by.
//
// This routine takes `ctx` argument to retrieve the [identity.Principal] instance performing
// the operation. If no principal is found, an `unknown` value will be placed instead.
func SoftDelete(ctx context.Context, auditable *Auditable) {
	Touch(ctx, auditable)
	auditable.IsDeleted = true
}

// -- Options --

// AuditableOption routine used to set non-required options to [Auditable]-related routines.
type AuditableOption func(auditable *Auditable)

// WithCreateTime sets the creation time for [Auditable].
func WithCreateTime(t time.Time) AuditableOption {
	return func(o *Auditable) {
		o.CreateTime = t
	}
}

// WithUpdateTime sets the last update time for [Auditable].
func WithUpdateTime(t time.Time) AuditableOption {
	return func(o *Auditable) {
		o.LastUpdateTime = t
	}
}

// WithLocation sets the location for [Auditable] default timestamps.
//
// This value will be ignored if any of the `createTime` or `updateTime` options are set.
func WithLocation(loc *time.Location) AuditableOption {
	return func(o *Auditable) {
		o.loc = loc
	}
}

// WithVersion sets the initial version for [Auditable].
func WithVersion(version int64) AuditableOption {
	return func(o *Auditable) {
		o.Version = version
	}
}
