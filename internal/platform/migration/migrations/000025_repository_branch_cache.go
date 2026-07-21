package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000025_repository_branch_cache", upRepositoryBranchCache)
}

func upRepositoryBranchCache(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	repo := &repositoryBranchCacheMigrationModel{}
	if !db.Migrator().HasColumn(repo, "branches_json") {
		if err := db.Migrator().AddColumn(repo, "BranchesJSON"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(repo, "branches_synced_at") {
		if err := db.Migrator().AddColumn(repo, "BranchesSyncedAt"); err != nil {
			return err
		}
	}
	return nil
}

type repositoryBranchCacheMigrationModel struct {
	ID               uint       `gorm:"primaryKey"`
	BranchesJSON     string     `gorm:"type:text"`
	BranchesSyncedAt *time.Time `gorm:"index"`
}

func (repositoryBranchCacheMigrationModel) TableName() string { return "repositories" }
