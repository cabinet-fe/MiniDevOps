package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000014_drop_cli_install_jobs", upDropCLIInstallJobs)
}

func upDropCLIInstallJobs(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	job := &cliInstallJobDropModel{}
	if db.Migrator().HasTable(job) {
		if err := db.Migrator().DropTable(job); err != nil {
			return err
		}
	}
	return nil
}

type cliInstallJobDropModel struct {
	ID uint `gorm:"primaryKey"`
}

func (cliInstallJobDropModel) TableName() string { return "cli_install_jobs" }
