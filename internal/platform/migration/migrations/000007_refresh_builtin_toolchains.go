package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000007_refresh_builtin_toolchains", upRefreshBuiltinToolchains)
}

// upRefreshBuiltinToolchains replaces the original P2 placeholder commands on
// existing installs without altering custom toolchains.
func upRefreshBuiltinToolchains(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver
	for _, builtin := range builtinToolchains() {
		updates := map[string]interface{}{
			"executable":         builtin.Executable,
			"description":        builtin.Description,
			"detect_command":     builtin.DetectCommand,
			"install_template":   builtin.InstallTemplate,
			"upgrade_template":   builtin.UpgradeTemplate,
			"uninstall_template": builtin.UninstallTemplate,
			"versions_command":   builtin.VersionsCommand,
			"switch_template":    builtin.SwitchTemplate,
		}
		if err := db.Model(&toolchainDefinitionMigrationModel{}).
			Where("name = ? AND kind = ?", builtin.Name, "builtin").
			Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}
