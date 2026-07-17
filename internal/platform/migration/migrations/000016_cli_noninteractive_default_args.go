package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000016_cli_noninteractive_default_args", upCLINonInteractiveDefaultArgs)
}

// upCLINonInteractiveDefaultArgs sets headless subcommands for Codex, OpenCode,
// and Reasonix so AgentRun invokes non-interactive CLIs that stream stdout.
func upCLINonInteractiveDefaultArgs(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	updates := map[string]string{
		"codex":    "exec",
		"opencode": "run",
		"reasonix": "run",
	}
	for key, args := range updates {
		if err := db.Model(&cliRuntimeDefinitionMigrationModel{}).
			Where("`key` = ? AND (default_args = '' OR default_args IS NULL)", key).
			Update("default_args", args).Error; err != nil {
			return err
		}
	}
	return nil
}
