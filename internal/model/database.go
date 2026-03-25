package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"buildflow/internal/config"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
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
		&Credential{},
		&Project{},
		&Environment{},
		&Distribution{},
		&EnvVar{},
		&VarGroup{},
		&VarGroupItem{},
		&EnvironmentVarGroup{},
		&Build{},
		&BuildDistribution{},
		&Notification{},
		&AuditLog{},
		&Dictionary{},
		&DictItem{},
	); err != nil {
		return nil, fmt.Errorf("auto migrating: %w", err)
	}

	legacyEnvCols := []string{"deploy_server_id", "deploy_path", "deploy_method", "post_deploy_script"}
	for _, col := range legacyEnvCols {
		if db.Migrator().HasColumn(&Environment{}, col) {
			_ = db.Migrator().DropColumn(&Environment{}, col)
		}
	}

	if err := migrateProjectRepoCredentials(db); err != nil {
		return nil, fmt.Errorf("migrate project credentials: %w", err)
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

func migrateProjectRepoCredentials(db *gorm.DB) error {
	var projects []Project
	if err := db.
		Where("repo_auth_type <> ? AND repo_auth_type <> ? AND repo_password <> '' AND credential_id IS NULL", "none", "credential").
		Find(&projects).Error; err != nil {
		return err
	}
	if len(projects) == 0 {
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for i := range projects {
			project := &projects[i]
			if strings.TrimSpace(project.RepoPassword) == "" {
				continue
			}

			credentialType := "password"
			if strings.TrimSpace(project.RepoAuthType) == "token" {
				credentialType = "token"
			}

			baseName := strings.TrimSpace(project.Name) + "-仓库凭证"
			if strings.TrimSpace(project.Name) == "" {
				baseName = "仓库凭证"
			}
			name, err := nextCredentialName(tx, baseName, project.CreatedBy)
			if err != nil {
				return err
			}

			credential := &Credential{
				Name:        name,
				Type:        credentialType,
				Username:    project.RepoUsername,
				Password:    project.RepoPassword,
				Description: "由历史项目仓库认证自动迁移",
				CreatedBy:   project.CreatedBy,
			}
			if err := tx.Create(credential).Error; err != nil {
				return err
			}

			if err := tx.Model(&Project{}).Where("id = ?", project.ID).Updates(map[string]interface{}{
				"credential_id":  credential.ID,
				"repo_auth_type": "credential",
				"repo_username":  "",
				"repo_password":  "",
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func nextCredentialName(tx *gorm.DB, baseName string, createdBy uint) (string, error) {
	name := baseName
	for i := 0; ; i++ {
		var count int64
		if err := tx.Model(&Credential{}).Where("name = ? AND created_by = ?", name, createdBy).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return name, nil
		}
		name = fmt.Sprintf("%s-%d", baseName, i+1)
	}
}
