package audit

import "time"

// AuditableView is a version of [Auditable] with primitive-only fields.
//
// Moreover, this structure adds new fields such as CreateTimeMillis to
// guarantee compatibility with external systems using time in unix epoch format.
//
// Finally, this structure is meant to be exposed to external systems (hence the name `view`).
type AuditableView struct {
	CreateTime           string `json:"create_time"`
	CreateTimeMillis     int64  `json:"create_time_millis"`
	CreateBy             string `json:"create_by"`
	LastUpdateTime       string `json:"last_update_time"`
	LastUpdateTimeMillis int64  `json:"last_update_time_millis"`
	LastUpdateBy         string `json:"last_update_by"`
	Version              uint64 `json:"version"`
	IsDeleted            bool   `json:"is_deleted"`
}

// NewAuditableFromView allocates a new [Auditable] from an [AuditableView].
func NewAuditableFromView(v AuditableView) Auditable {
	createTime, _ := time.Parse(time.RFC3339, v.CreateTime)
	updateTime, _ := time.Parse(time.RFC3339, v.LastUpdateTime)
	return Auditable{
		createTime:     createTime,
		createBy:       v.CreateBy,
		lastUpdateTime: updateTime,
		lastUpdateBy:   v.LastUpdateBy,
		version:        v.Version,
		isDeleted:      v.IsDeleted,
	}
}
