package migrations

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000020_cli_npm_install_sources", upCLINPMInstallSources)
}

// upCLINPMInstallSources switches all AI CLIs to npm install/upgrade/uninstall
// templates and replaces non-npm install sources with npm registry mirrors.
// Configured sources are passed as npm --registry; no sources means the default registry.
func upCLINPMInstallSources(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	updates := []struct {
		key   string
		pkg   string
		label string
	}{
		{"claude_code", "@anthropic-ai/claude-code", "claude_code"},
		{"opencode", "opencode-ai", "opencode"},
		{"reasonix", "reasonix", "reasonix"},
		{"codex", "@openai/codex", "codex"},
	}
	for _, u := range updates {
		if err := db.Model(&cliRuntimeDefinitionMigrationModel{}).
			Where("`key` = ?", u.key).
			Updates(map[string]any{
				"install_template":   npmCLIInstallTemplate(u.pkg, u.label),
				"upgrade_template":   npmCLIUpgradeTemplate(u.pkg, u.label),
				"uninstall_template": npmCLIUninstallTemplate(u.pkg),
			}).Error; err != nil {
			return err
		}
	}

	now := time.Now().UTC()
	legacyNames := []string{"Official", "GitHub releases", "Mirror"}
	for _, cliKey := range []string{"opencode", "reasonix"} {
		if err := db.Where("cli_key = ? AND name IN ?", cliKey, legacyNames).
			Delete(&cliInstallSourceMigrationModel{}).Error; err != nil {
			return err
		}
	}

	sources := []cliInstallSourceMigrationModel{
		{CliKey: "opencode", Name: "npm registry", BaseURL: "https://registry.npmjs.org", Priority: 10, Enabled: true, CreatedAt: now},
		{CliKey: "opencode", Name: "npm mirror", BaseURL: "https://registry.npmmirror.com", Priority: 20, Enabled: true, CreatedAt: now},
		{CliKey: "reasonix", Name: "npm registry", BaseURL: "https://registry.npmjs.org", Priority: 10, Enabled: true, CreatedAt: now},
		{CliKey: "reasonix", Name: "npm mirror", BaseURL: "https://registry.npmmirror.com", Priority: 20, Enabled: true, CreatedAt: now},
	}
	for _, source := range sources {
		var existing cliInstallSourceMigrationModel
		err := db.Where("cli_key = ? AND name = ?", source.CliKey, source.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&source).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		existing.BaseURL = source.BaseURL
		existing.Priority = source.Priority
		existing.Enabled = source.Enabled
		if err := db.Save(&existing).Error; err != nil {
			return err
		}
	}
	return nil
}
