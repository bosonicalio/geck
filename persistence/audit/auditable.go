package audit

import (
	"context"
	"time"

	"github.com/samber/lo"

	"github.com/hadroncorp/geck/persistence"
	"github.com/hadroncorp/geck/security"
)

// -- AUDITABLE OPTIONS --

type auditableOptions struct {
	location *time.Location
}

// AuditableOption routine used to set non-required options to [Auditable]-related routines.
type AuditableOption func(*auditableOptions)

// WithLocation sets the location for [Auditable] timestamps.
func WithLocation(loc *time.Location) AuditableOption {
	return func(o *auditableOptions) {
		o.location = loc
	}
}

// -- AUDITABLE --

// Auditable is a structure provisioning metadata for persistence operations.
//
// Embed this structure into your entities/aggregates to enhance and control write operations.
// Call [NewAuditable] routine to create an instance with default values.
//
// Implements [persistence.Storable] interface.
type Auditable struct {
	CreateTime     time.Time
	CreateBy       string
	LastUpdateTime time.Time
	LastUpdateBy   string
	Version        int64
	IsActive       bool
}

const _defaultPrincipalUsername = "unknown"

// compile-time assertions
var _ persistence.Storable = (*Auditable)(nil)

// NewAuditable allocates a new [Auditable] instance using default values.
//
// This routine takes `ctx` argument to retrieve the [security.Principal] instance performing
// the operation. If no principal is found, an `unknown` value will be placed instead.
//
// Use [AuditableOption] routines to customize how the instance is created.
func NewAuditable(ctx context.Context, opts ...AuditableOption) Auditable {
	options := auditableOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	now := time.Now().In(lo.CoalesceOrEmpty(options.location, time.UTC))
	principal, _ := security.GetPrincipal(ctx)
	var username string
	if principal != nil {
		username = principal.ID()
	}
	username = lo.CoalesceOrEmpty(username, _defaultPrincipalUsername)
	return Auditable{
		CreateTime:     now,
		CreateBy:       username,
		LastUpdateTime: now,
		LastUpdateBy:   username,
		Version:        0,
		IsActive:       true,
	}
}

// UpdateAuditable increases the version, updates last update fields, both time and by.
//
// This routine takes `ctx` argument to retrieve the [security.Principal] instance performing
// the operation. If no principal is found, an `unknown` value will be placed instead.
func UpdateAuditable(ctx context.Context, auditable *Auditable) {
	auditable.Version++
	auditable.LastUpdateTime = time.Now().UTC()
	var username string
	principal, _ := security.GetPrincipal(ctx)
	if principal != nil {
		username = principal.ID()
	}
	auditable.LastUpdateBy = lo.CoalesceOrEmpty(username, _defaultPrincipalUsername)
}

// IsNew checks if the type was just created.
func (a Auditable) IsNew() bool {
	return a.Version == 0
}

// ToView converts the current [Auditable] instance to an [AuditableView].
func (a Auditable) ToView() AuditableView {
	return AuditableView{
		CreateTime:           a.CreateTime.Format(time.RFC3339),
		CreateTimeMillis:     a.CreateTime.UnixMilli(),
		CreateBy:             a.CreateBy,
		LastUpdateTime:       a.LastUpdateTime.Format(time.RFC3339),
		LastUpdateTimeMillis: a.LastUpdateTime.UnixMilli(),
		LastUpdateBy:         a.LastUpdateBy,
		Version:              a.Version,
		IsActive:             a.IsActive,
	}
}
