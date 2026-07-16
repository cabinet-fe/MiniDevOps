package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000007_refresh_builtin_toolchains", upRefreshBuiltinToolchains)
}

// upRefreshBuiltinToolchains was a one-shot refresh for the retired toolchain
// tables. After 000011 those tables are dropped; keep this migration as a
// no-op so already-applied databases remain valid in schema_migrations.
func upRefreshBuiltinToolchains(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = db
	_ = driver
	return nil
}
