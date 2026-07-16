package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000013_drop_build_job_branch_policy", upDropBuildJobBranchPolicy)
}

func upDropBuildJobBranchPolicy(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	job := &buildJobBranchPolicyDropModel{}
	if db.Migrator().HasColumn(job, "branch_policy") {
		if err := db.Migrator().DropColumn(job, "BranchPolicy"); err != nil {
			return err
		}
	}
	return nil
}

type buildJobBranchPolicyDropModel struct {
	ID uint `gorm:"primaryKey"`
}

func (buildJobBranchPolicyDropModel) TableName() string { return "build_jobs" }
