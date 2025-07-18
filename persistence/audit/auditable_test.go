package audit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tesserical/geck/persistence/audit"
	"github.com/tesserical/geck/security/identity"
)

func TestNew(t *testing.T) {
	// Set up a default auditable
	auditable := audit.New(t.Context())
	assert.Equal(t, "unknown", auditable.CreateBy)
	assert.NotEmpty(t, auditable.CreateTime)
	assert.Equal(t, "unknown", auditable.LastUpdateBy)
	assert.NotEmpty(t, auditable.LastUpdateTime)
	assert.Equal(t, int64(0), auditable.Version)
	assert.False(t, auditable.IsDeleted)

	// Set up a default auditable with custom location
	loc := time.FixedZone("UTC-8", -8*60*60)
	auditable = audit.New(t.Context(),
		audit.WithLocation(loc),
	)
	assert.Equal(t, "unknown", auditable.CreateBy)
	assert.NotEmpty(t, auditable.CreateTime)
	assert.Equal(t, loc, auditable.CreateTime.Location())
	assert.Equal(t, "unknown", auditable.LastUpdateBy)
	assert.NotEmpty(t, auditable.LastUpdateTime)
	assert.Equal(t, loc, auditable.LastUpdateTime.Location())
	assert.Equal(t, int64(0), auditable.Version)
	assert.False(t, auditable.IsDeleted)

	// Set up a customized auditable
	principal := identity.NewBasicPrincipal("test_user")
	ctx := identity.WithPrincipal(t.Context(), principal)
	auditable = audit.New(ctx,
		audit.WithCreateTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		audit.WithVersion(65),
	)
	assert.Equal(t, "test_user", auditable.CreateBy)
	assert.NotEmpty(t, auditable.CreateTime)
	assert.Equal(t, "test_user", auditable.LastUpdateBy)
	assert.NotEmpty(t, auditable.LastUpdateTime)
	assert.Equal(t, int64(65), auditable.Version)
	assert.False(t, auditable.IsDeleted)

	// Soft delete the auditable
	pastVersion := auditable.Version
	principal = identity.NewBasicPrincipal("test_user_soft_delete")
	ctx = identity.WithPrincipal(t.Context(), principal)
	audit.SoftDelete(ctx, &auditable)
	assert.True(t, auditable.IsDeleted)
	assert.True(t, auditable.LastUpdateTime.After(auditable.CreateTime))
	assert.Greater(t, auditable.Version, pastVersion)
	assert.Equal(t, "test_user", auditable.CreateBy)
	assert.Equal(t, "test_user_soft_delete", auditable.LastUpdateBy)
}
