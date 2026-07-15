package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000004_webhook_deliveries", upWebhookDeliveries)
}

func upWebhookDeliveries(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver
	m := &webhookDeliveryMigrationModel{}
	if db.Migrator().HasTable(m) {
		return nil
	}
	return db.Migrator().CreateTable(m)
}

type webhookDeliveryMigrationModel struct {
	ID           uint      `gorm:"primaryKey"`
	RepositoryID uint      `gorm:"uniqueIndex:idx_wh_delivery;not null"`
	DeliveryKey  string    `gorm:"size:200;uniqueIndex:idx_wh_delivery;not null"`
	CreatedAt    time.Time `gorm:""`
}

func (webhookDeliveryMigrationModel) TableName() string { return "webhook_deliveries" }
