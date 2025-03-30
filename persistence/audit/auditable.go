package audit

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/security/identity"
)

// --> Auditable <--

const _defaultPrincipalUsername = "unknown"

// Auditable is a structure provisioning metadata for persistence operations.
//
// Embed this structure into your entities/aggregates to enhance and control write operations.
// Call [NewWithDefaults] routine to create an instance with default values.
//
// Implements [persistence.Storable] interface.
type Auditable struct {
	createTime     time.Time
	createBy       string
	lastUpdateTime time.Time
	lastUpdateBy   string
	version        uint64
	isDeleted      bool
}

// compile-time assertions
var _ persistence.Storable = (*Auditable)(nil)

// NewWithDefaults allocates a new [Auditable] instance using default values.
//
// This routine takes `ctx` argument to retrieve the [security.Principal] instance performing
// the operation. If no principal is found, an `unknown` value will be placed instead.
//
// Use [AuditableOption] routines to customize how the instance is created.
func NewWithDefaults(ctx context.Context, opts ...AuditableOption) Auditable {
	options := auditableOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	now := time.Now().In(lo.CoalesceOrEmpty(options.location, time.UTC))
	principal, _ := identity.GetPrincipal(ctx)
	var username string
	if principal != nil {
		username = principal.ID()
	}
	username = lo.CoalesceOrEmpty(username, _defaultPrincipalUsername)
	return Auditable{
		createTime:     now,
		createBy:       username,
		lastUpdateTime: now,
		lastUpdateBy:   username,
		version:        0,
		isDeleted:      false,
	}
}

// Update increases the version, updates last update fields, both time and by.
//
// This routine takes `ctx` argument to retrieve the [security.Principal] instance performing
// the operation. If no principal is found, an `unknown` value will be placed instead.
func Update(ctx context.Context, auditable *Auditable) {
	auditable.version++
	auditable.lastUpdateTime = time.Now().UTC()
	var username string
	principal, _ := identity.GetPrincipal(ctx)
	if principal != nil {
		username = principal.ID()
	}
	auditable.lastUpdateBy = lo.CoalesceOrEmpty(username, _defaultPrincipalUsername)
}

// Delete Marks `auditable` as deleted. It also increases the version, updates last update
// fields, both time and by.
//
// This routine takes `ctx` argument to retrieve the [security.Principal] instance performing
// the operation. If no principal is found, an `unknown` value will be placed instead.
func Delete(ctx context.Context, auditable *Auditable) {
	auditable.isDeleted = true
	Update(ctx, auditable)
}

func (a Auditable) CreateTime() time.Time {
	return a.createTime
}

func (a Auditable) CreateBy() string {
	return a.createBy
}

func (a Auditable) LastUpdateTime() time.Time {
	return a.lastUpdateTime
}

func (a Auditable) LastUpdateBy() string {
	return a.lastUpdateBy
}

func (a Auditable) Version() uint64 {
	return a.version
}

func (a Auditable) IsDeleted() bool {
	return a.isDeleted
}

// IsNew checks if the type was just created.
func (a Auditable) IsNew() bool {
	return a.version == 0
}

// ToView converts the current [Auditable] instance to a [View].
func (a Auditable) ToView() View {
	return View{
		CreateTime:           a.createTime.Format(time.RFC3339),
		CreateTimeMillis:     a.createTime.UnixMilli(),
		CreateBy:             a.createBy,
		LastUpdateTime:       a.lastUpdateTime.Format(time.RFC3339),
		LastUpdateTimeMillis: a.lastUpdateTime.UnixMilli(),
		LastUpdateBy:         a.lastUpdateBy,
		Version:              a.version,
		IsDeleted:            a.isDeleted,
	}
}

type auditableOptions struct {
	location *time.Location
}

// -- Options --

// AuditableOption routine used to set non-required options to [Auditable]-related routines.
type AuditableOption func(*auditableOptions)

// WithLocation sets the location for [Auditable] timestamps.
func WithLocation(loc *time.Location) AuditableOption {
	return func(o *auditableOptions) {
		o.location = loc
	}
}

// -- New --

// NewArgs is a structure used to provision arguments for [New] routine.
type NewArgs struct {
	CreateTime     time.Time
	CreateBy       string
	LastUpdateTime time.Time
	LastUpdateBy   string
	Version        uint64
	IsDeleted      bool
}

// New allocates a new [Auditable].
func New(args NewArgs) Auditable {
	return Auditable{
		createTime:     args.CreateTime,
		createBy:       args.CreateBy,
		lastUpdateTime: args.LastUpdateTime,
		lastUpdateBy:   args.LastUpdateBy,
		version:        args.Version,
		isDeleted:      args.IsDeleted,
	}
}
