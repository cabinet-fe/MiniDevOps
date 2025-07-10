package db

import (
	"server/internal/models"

	"gorm.io/gorm"
)

// RunMigrations 运行数据库迁移
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Repository{},
		&models.Task{},
		&models.RemoteServer{},
		&models.SystemConfig{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.TaskRemote{},
	)
}
