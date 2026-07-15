package seed

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"bedrock/internal/platform/config"
)

// UserRow is the minimal row used for super-admin seeding (avoids auth→platform cycle).
type UserRow struct {
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

func (UserRow) TableName() string { return "users" }

// EnsureSuperAdmin creates the built-in super-admin from config when no users exist.
// The seeded user is marked IsSuperAdmin and must not be deletable by application logic.
func EnsureSuperAdmin(db *gorm.DB, admin config.AdminConfig) error {
	var count int64
	if err := db.Model(&UserRow{}).Count(&count).Error; err != nil {
		return fmt.Errorf("counting users: %w", err)
	}
	if count > 0 {
		return nil
	}
	if admin.Username == "" || admin.Password == "" {
		return fmt.Errorf("admin.username and admin.password must be set when no users exist")
	}
	displayName := admin.DisplayName
	if displayName == "" {
		displayName = "Administrator"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}
	now := time.Now().UTC()
	row := UserRow{
		Username:     admin.Username,
		PasswordHash: string(hash),
		DisplayName:  displayName,
		IsActive:     true,
		IsSuperAdmin: true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := db.Create(&row).Error; err != nil {
		return fmt.Errorf("creating super-admin: %w", err)
	}
	return nil
}
