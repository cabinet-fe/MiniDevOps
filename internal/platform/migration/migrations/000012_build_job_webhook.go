package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000012_build_job_webhook", upBuildJobWebhook)
}

func upBuildJobWebhook(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	job := &buildJobWebhookMigrationModel{}
	if !db.Migrator().HasColumn(job, "webhook_secret") {
		if err := db.Migrator().AddColumn(job, "WebhookSecret"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(job, "webhook_type") {
		if err := db.Migrator().AddColumn(job, "WebhookType"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(job, "webhook_ref_path") {
		if err := db.Migrator().AddColumn(job, "WebhookRefPath"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(job, "webhook_commit_path") {
		if err := db.Migrator().AddColumn(job, "WebhookCommitPath"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(job, "webhook_message_path") {
		if err := db.Migrator().AddColumn(job, "WebhookMessagePath"); err != nil {
			return err
		}
	}

	repo := &repositoryWebhookCleanupMigrationModel{}
	for _, col := range []string{
		"default_branch",
		"webhook_secret",
		"webhook_type",
		"webhook_ref_path",
		"webhook_commit_path",
		"webhook_message_path",
	} {
		if db.Migrator().HasColumn(repo, col) {
			if err := db.Migrator().DropColumn(repo, col); err != nil {
				return err
			}
		}
	}

	if db.Migrator().HasTable(&webhookDeliveryLegacyMigrationModel{}) {
		if err := db.Migrator().DropTable(&webhookDeliveryLegacyMigrationModel{}); err != nil {
			return err
		}
	}
	if !db.Migrator().HasTable(&webhookDeliveryJobMigrationModel{}) {
		if err := db.Migrator().CreateTable(&webhookDeliveryJobMigrationModel{}); err != nil {
			return err
		}
	}
	return nil
}

type buildJobWebhookMigrationModel struct {
	ID                 uint   `gorm:"primaryKey"`
	WebhookSecret      string `gorm:"size:64"`
	WebhookType        string `gorm:"size:20;default:auto"`
	WebhookRefPath     string `gorm:"size:300"`
	WebhookCommitPath  string `gorm:"size:300"`
	WebhookMessagePath string `gorm:"size:300"`
}

func (buildJobWebhookMigrationModel) TableName() string { return "build_jobs" }

type repositoryWebhookCleanupMigrationModel struct {
	ID uint `gorm:"primaryKey"`
}

func (repositoryWebhookCleanupMigrationModel) TableName() string { return "repositories" }

type webhookDeliveryLegacyMigrationModel struct {
	ID           uint   `gorm:"primaryKey"`
	RepositoryID uint   `gorm:"uniqueIndex:idx_wh_delivery;not null"`
	DeliveryKey  string `gorm:"size:200;uniqueIndex:idx_wh_delivery;not null"`
}

func (webhookDeliveryLegacyMigrationModel) TableName() string { return "webhook_deliveries" }

type webhookDeliveryJobMigrationModel struct {
	ID          uint      `gorm:"primaryKey"`
	BuildJobID  uint      `gorm:"uniqueIndex:idx_wh_delivery;not null"`
	DeliveryKey string    `gorm:"size:200;uniqueIndex:idx_wh_delivery;not null"`
	CreatedAt   time.Time `gorm:""`
}

func (webhookDeliveryJobMigrationModel) TableName() string { return "webhook_deliveries" }
