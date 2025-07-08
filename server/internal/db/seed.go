package db

import (
	"minidevops/server/internal/models"
	"minidevops/server/internal/utils"

	"gorm.io/gorm"
)

// SeedData 初始化种子数据
func SeedData(db *gorm.DB) error {
	// 创建默认权限
	if err := seedPermissions(db); err != nil {
		return err
	}

	// 创建默认角色
	if err := seedRoles(db); err != nil {
		return err
	}

	// 创建默认用户
	if err := seedUsers(db); err != nil {
		return err
	}

	// 创建默认系统配置
	if err := seedSystemConfig(db); err != nil {
		return err
	}

	return nil
}

// seedPermissions 创建默认权限
func seedPermissions(db *gorm.DB) error {
	permissions := []models.Permission{
		{Name: "系统管理", Type: models.PermissionTypeMenu, Code: "system", Sort: 1},
		{Name: "用户管理", Type: models.PermissionTypeMenu, Code: "user", Sort: 2},
		{Name: "角色管理", Type: models.PermissionTypeMenu, Code: "role", Sort: 3},
		{Name: "权限管理", Type: models.PermissionTypeMenu, Code: "permission", Sort: 4},
		{Name: "仓库管理", Type: models.PermissionTypeMenu, Code: "repository", Sort: 5},
		{Name: "任务管理", Type: models.PermissionTypeMenu, Code: "task", Sort: 6},
		{Name: "远程服务器", Type: models.PermissionTypeMenu, Code: "remote", Sort: 7},
		{Name: "系统配置", Type: models.PermissionTypeMenu, Code: "config", Sort: 8},
	}

	for _, permission := range permissions {
		var existingPermission models.Permission
		if err := db.Where("code = ?", permission.Code).First(&existingPermission).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&permission).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// seedRoles 创建默认角色
func seedRoles(db *gorm.DB) error {
	var adminRole models.Role
	if err := db.Where("code = ?", "admin").First(&adminRole).Error; err == gorm.ErrRecordNotFound {
		// 获取所有权限
		var permissions []models.Permission
		db.Find(&permissions)

		adminRole = models.Role{
			Name:        "管理员",
			Code:        "admin",
			Description: "系统管理员，拥有所有权限",
			DataScope:   "1",
			Permissions: permissions,
		}

		if err := db.Create(&adminRole).Error; err != nil {
			return err
		}
	}

	return nil
}

// seedUsers 创建默认用户
func seedUsers(db *gorm.DB) error {
	var adminUser models.User
	if err := db.Where("username = ?", "admin").First(&adminUser).Error; err == gorm.ErrRecordNotFound {
		// 获取管理员角色
		var adminRole models.Role
		if err := db.Where("code = ?", "admin").First(&adminRole).Error; err != nil {
			return err
		}

		hashedPassword, err := utils.HashPassword("admin123")
		if err != nil {
			return err
		}

		adminUser = models.User{
			Username: "admin",
			Password: hashedPassword,
			Name:     "系统管理员",
			Email:    "admin@example.com",
			Roles:    []models.Role{adminRole},
		}

		if err := db.Create(&adminUser).Error; err != nil {
			return err
		}
	}

	return nil
}

// seedSystemConfig 创建默认系统配置
func seedSystemConfig(db *gorm.DB) error {
	configs := []models.SystemConfig{
		{
			Key:         models.ConfigKeyMountPath,
			Value:       "~/dev-ops",
			Description: "任务代码挂载路径",
		},
	}

	for _, config := range configs {
		var existingConfig models.SystemConfig
		if err := db.Where("key = ?", config.Key).First(&existingConfig).Error; err == gorm.ErrRecordNotFound {
			if err := db.Create(&config).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
