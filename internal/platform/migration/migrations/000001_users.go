package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000001_users", upUsers)
}

func upUsers(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver
	if db.Migrator().HasTable(&userMigrationModel{}) {
		return nil
	}
	return db.Migrator().CreateTable(&userMigrationModel{})
}

// userMigrationModel defines the Wave 1 users table (compatible with JWT login).
type userMigrationModel struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"size:50;uniqueIndex;not null"`
	PasswordHash string    `gorm:"size:255;not null"`
	DisplayName  string    `gorm:"size:100"`
	Email        string    `gorm:"size:200"`
	Avatar       string    `gorm:"size:500"`
	IsActive     bool      `gorm:"not null;default:true"`
	IsSuperAdmin bool      `gorm:"not null;default:false"`
	CreatedAt    time.Time `gorm:""`
	UpdatedAt    time.Time `gorm:""`
}

func (userMigrationModel) TableName() string { return "users" }
