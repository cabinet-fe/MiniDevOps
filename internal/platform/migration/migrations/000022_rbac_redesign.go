package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000022_rbac_redesign", upRBACRedesign)
}

// upRBACRedesign destructively rebuilds RBAC: menu_groups, rbac_resources
// (code/full_code/group_id/merged menu fields), roles.type, clears role_permissions.
// Fresh-install only — custom role grants are not preserved.
func upRBACRedesign(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	return db.Transaction(func(tx *gorm.DB) error {
		// Clear grants before dropping resource identity.
		if tx.Migrator().HasTable(&rolePermissionMigrationModel{}) {
			if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).
				Delete(&rolePermissionMigrationModel{}).Error; err != nil {
				return err
			}
		}

		if tx.Migrator().HasTable(&menuMetadataMigrationModel{}) {
			if err := tx.Migrator().DropTable(&menuMetadataMigrationModel{}); err != nil {
				return err
			}
		}
		if tx.Migrator().HasTable(&rbacResourceMigrationModel{}) {
			if err := tx.Migrator().DropTable(&rbacResourceMigrationModel{}); err != nil {
				return err
			}
		}

		if !tx.Migrator().HasTable(&menuGroupMigrationModel{}) {
			if err := tx.Migrator().CreateTable(&menuGroupMigrationModel{}); err != nil {
				return err
			}
		}

		if err := tx.Migrator().CreateTable(&rbacResourceV2MigrationModel{}); err != nil {
			return err
		}

		role := &roleTypeMigrationModel{}
		if tx.Migrator().HasTable(role) && !tx.Migrator().HasColumn(role, "Type") {
			if err := tx.Migrator().AddColumn(role, "Type"); err != nil {
				return err
			}
			if err := tx.Exec(`UPDATE roles SET type = ? WHERE type IS NULL OR type = ''`, "custom").Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// roleTypeMigrationModel only exists to add roles.type.
type roleTypeMigrationModel struct {
	ID   uint   `gorm:"primaryKey"`
	Type string `gorm:"size:20;not null;default:custom"`
}

func (roleTypeMigrationModel) TableName() string { return "roles" }

type menuGroupMigrationModel struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100;not null"`
	Code        string    `gorm:"size:100;uniqueIndex;not null"`
	RoutePrefix string    `gorm:"size:200"`
	SortKey     int       `gorm:"not null;default:0"`
	Enabled     bool      `gorm:"not null;default:true"`
	CreatedAt   time.Time `gorm:""`
	UpdatedAt   time.Time `gorm:""`
}

func (menuGroupMigrationModel) TableName() string { return "menu_groups" }

// rbacResourceV2MigrationModel is the rebuilt resource table (no path / menu_metadata).
type rbacResourceV2MigrationModel struct {
	ID             uint      `gorm:"primaryKey"`
	Code           string    `gorm:"size:100;not null;index"`
	FullCode       string    `gorm:"size:200;uniqueIndex;not null"`
	Type           string    `gorm:"size:20;not null;index"` // menu|action|card
	GroupID        *uint     `gorm:"index"`
	ParentID       *uint     `gorm:"index"`
	SuperAdminOnly bool      `gorm:"not null;default:false"`
	Hidden         bool      `gorm:"not null;default:false"`
	Enabled        bool      `gorm:"not null;default:true"`
	SortKey        int       `gorm:"not null;default:0"`
	Title          string    `gorm:"size:100"`
	Route          string    `gorm:"size:200"`
	IconBase64     string    `gorm:"type:text"`
	IconMime       string    `gorm:"size:64"`
	CreatedAt      time.Time `gorm:""`
	UpdatedAt      time.Time `gorm:""`
}

func (rbacResourceV2MigrationModel) TableName() string { return "rbac_resources" }
