package sql

import (
	"time"

	"github.com/hadroncorp/geck/persistence/audit"
)

// Auditable is a version of audit.Auditable for SQL databases.
type Auditable struct {
	CreateTime     time.Time `db:"create_time"`
	CreateBy       string    `db:"create_by"`
	LastUpdateTime time.Time `db:"last_update_time"`
	LastUpdateBy   string    `db:"last_update_by"`
	Version        int64     `db:"version"`
	IsActive       bool      `db:"is_active"`
}

// NewAuditable allocates a new [Auditable] instance based on `v` ([audit.Auditable]).
func NewAuditable(v audit.Auditable) Auditable {
	return Auditable{
		CreateTime:     v.CreateTime,
		CreateBy:       v.CreateBy,
		LastUpdateTime: v.LastUpdateTime,
		LastUpdateBy:   v.LastUpdateBy,
		Version:        v.Version,
		IsActive:       v.IsActive,
	}
}

// ToAudit converts [Auditable] into an [audit.Auditable].
func (a Auditable) ToAudit() audit.Auditable {
	return audit.Auditable{
		CreateTime:     a.CreateTime.UTC(),
		CreateBy:       a.CreateBy,
		LastUpdateTime: a.LastUpdateTime.UTC(),
		LastUpdateBy:   a.LastUpdateBy,
		Version:        a.Version,
		IsActive:       a.IsActive,
	}
}
