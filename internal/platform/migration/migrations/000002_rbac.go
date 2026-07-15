package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000002_rbac", upRBAC)
}

func upRBAC(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	models := []interface{}{
		&roleMigrationModel{},
		&rolePermissionMigrationModel{},
		&rbacResourceMigrationModel{},
		&menuMetadataMigrationModel{},
		&userRoleMigrationModel{},
		&dictionaryMigrationModel{},
		&dictItemMigrationModel{},
		&operationLogMigrationModel{},
	}
	for _, m := range models {
		if db.Migrator().HasTable(m) {
			continue
		}
		if err := db.Migrator().CreateTable(m); err != nil {
			return err
		}
	}
	return nil
}

type roleMigrationModel struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;uniqueIndex;not null"`
	Code        string    `gorm:"size:100;uniqueIndex;not null"`
	Description string    `gorm:"size:500"`
	CreatedAt   time.Time `gorm:""`
	UpdatedAt   time.Time `gorm:""`
}

func (roleMigrationModel) TableName() string { return "roles" }

type rolePermissionMigrationModel struct {
	ID         uint   `gorm:"primaryKey"`
	RoleID     uint   `gorm:"index;not null"`
	Permission string `gorm:"size:200;not null;index"`
}

func (rolePermissionMigrationModel) TableName() string { return "role_permissions" }

type rbacResourceMigrationModel struct {
	ID        uint      `gorm:"primaryKey"`
	Path      string    `gorm:"size:200;uniqueIndex;not null"`
	Type      string    `gorm:"size:20;not null;index"` // menu|page|action|card
	ParentID  *uint     `gorm:"index"`
	Enabled   bool      `gorm:"not null;default:true"`
	SortKey   int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:""`
	UpdatedAt time.Time `gorm:""`
}

func (rbacResourceMigrationModel) TableName() string { return "rbac_resources" }

type menuMetadataMigrationModel struct {
	ID         uint   `gorm:"primaryKey"`
	ResourceID uint   `gorm:"uniqueIndex;not null"`
	Title      string `gorm:"size:100;not null"`
	Route      string `gorm:"size:200"`
	IconBase64 string `gorm:"type:text"`
	IconMime   string `gorm:"size:64"`
}

func (menuMetadataMigrationModel) TableName() string { return "menu_metadata" }

type userRoleMigrationModel struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

func (userRoleMigrationModel) TableName() string { return "user_roles" }

type dictionaryMigrationModel struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Code        string    `gorm:"size:100;uniqueIndex;not null"`
	Description string    `gorm:"size:500"`
	CreatedAt   time.Time `gorm:""`
	UpdatedAt   time.Time `gorm:""`
}

func (dictionaryMigrationModel) TableName() string { return "dictionaries" }

type dictItemMigrationModel struct {
	ID           uint      `gorm:"primaryKey"`
	DictionaryID uint      `gorm:"index;not null"`
	Label        string    `gorm:"size:200;not null"`
	Value        string    `gorm:"size:200;not null"`
	SortOrder    int       `gorm:"not null;default:0"`
	Enabled      bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:""`
	UpdatedAt    time.Time `gorm:""`
}

func (dictItemMigrationModel) TableName() string { return "dict_items" }

type operationLogMigrationModel struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"index"`
	Username     string    `gorm:"size:50"`
	Action       string    `gorm:"size:50;not null;index"`
	ResourceType string    `gorm:"size:50"`
	ResourceID   string    `gorm:"size:64"`
	Details      string    `gorm:"type:text"`
	IPAddress    string    `gorm:"size:45"`
	CreatedAt    time.Time `gorm:"index"`
}

func (operationLogMigrationModel) TableName() string { return "operation_logs" }
