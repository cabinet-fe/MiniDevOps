package model

import (
	"fmt"
	"os"
	"path/filepath"

	"buildflow/internal/config"
	"golang.org/x/crypto/bcrypt"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() (*gorm.DB, error) {
	if config.C == nil {
		return nil, fmt.Errorf("config not loaded: call config.Load before InitDB")
	}

	dbPath := config.C.Database.Path
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting underlying db: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	if err := db.AutoMigrate(
		&User{},
		&Server{},
		&Project{},
		&Environment{},
		&EnvVar{},
		&VarGroup{},
		&VarGroupItem{},
		&EnvironmentVarGroup{},
		&Build{},
		&Notification{},
		&AuditLog{},
		&Dictionary{},
		&DictItem{},
	); err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	// Drop legacy group_name column from projects table if it still exists
	if db.Migrator().HasColumn(&Project{}, "group_name") {
		_ = db.Migrator().DropColumn(&Project{}, "group_name")
	}

	// Seed default project_tags dictionary
	var dictCount int64
	db.Model(&Dictionary{}).Where("code = ?", "project_tags").Count(&dictCount)
	if dictCount == 0 {
		db.Create(&Dictionary{
			Name:        "项目标签",
			Code:        "project_tags",
			Description: "项目可选标签列表",
		})
	}

	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("counting users: %w", err)
	}
	if count == 0 {
		adminCfg := config.C.Admin
		if adminCfg.Username == "" || adminCfg.Password == "" {
			return nil, fmt.Errorf("admin username and password must be set in config when no users exist")
		}
		displayName := adminCfg.DisplayName
		if displayName == "" {
			displayName = "Administrator"
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(adminCfg.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hashing admin password: %w", err)
		}
		admin := User{
			Username:     adminCfg.Username,
			PasswordHash: string(hash),
			DisplayName:  displayName,
			Role:         "admin",
			IsActive:     true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return nil, fmt.Errorf("creating admin user: %w", err)
		}
	}

	return db, nil
}
