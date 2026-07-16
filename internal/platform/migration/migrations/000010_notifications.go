package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000010_notifications", upNotifications)
}

func upNotifications(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver
	if db.Migrator().HasTable(&notificationMigrationModel{}) {
		return nil
	}
	return db.Migrator().CreateTable(&notificationMigrationModel{})
}

// notificationMigrationModel is the in-app notification inbox (DESIGN §12).
type notificationMigrationModel struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	Type       string    `gorm:"size:50;not null"`
	Title      string    `gorm:"size:200;not null"`
	Message    string    `gorm:"size:500"`
	BuildRunID *uint     `gorm:"index"`
	AgentRunID *uint     `gorm:"index"`
	IsRead     bool      `gorm:"not null;default:false;index"`
	CreatedAt  time.Time `gorm:"index"`
}

func (notificationMigrationModel) TableName() string { return "notifications" }
