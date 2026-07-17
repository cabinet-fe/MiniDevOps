package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000015_agent_workspace_fields", upAgentWorkspaceFields)
}

func upAgentWorkspaceFields(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	agent := &aiAgentWorkspaceMigrationModel{}
	if db.Migrator().HasColumn(agent, "repository_id") {
		if err := db.Migrator().DropColumn(agent, "RepositoryID"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(agent, "build_job_ids_json") {
		if err := db.Migrator().AddColumn(agent, "BuildJobIDsJSON"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(agent, "output_dir") {
		if err := db.Migrator().AddColumn(agent, "OutputDir"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(agent, "artifact_format") {
		if err := db.Migrator().AddColumn(agent, "ArtifactFormat"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(agent, "max_artifacts") {
		if err := db.Migrator().AddColumn(agent, "MaxArtifacts"); err != nil {
			return err
		}
	}

	run := &agentRunWorkspaceMigrationModel{}
	if !db.Migrator().HasColumn(run, "work_dir") {
		if err := db.Migrator().AddColumn(run, "WorkDir"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(run, "artifact_path") {
		if err := db.Migrator().AddColumn(run, "ArtifactPath"); err != nil {
			return err
		}
	}

	// Backfill defaults for existing rows (GORM defaults apply to new inserts only).
	if err := db.Model(agent).Where("output_dir = '' OR output_dir IS NULL").Update("output_dir", "output").Error; err != nil {
		return err
	}
	if err := db.Model(agent).Where("artifact_format = '' OR artifact_format IS NULL").Update("artifact_format", "gzip").Error; err != nil {
		return err
	}
	if err := db.Model(agent).Where("max_artifacts = 0 OR max_artifacts IS NULL").Update("max_artifacts", 10).Error; err != nil {
		return err
	}
	if err := db.Model(agent).Where("build_job_ids_json = '' OR build_job_ids_json IS NULL").Update("build_job_ids_json", "[]").Error; err != nil {
		return err
	}
	return nil
}

type aiAgentWorkspaceMigrationModel struct {
	ID               uint   `gorm:"primaryKey"`
	RepositoryID     *uint  `gorm:"index"`
	BuildJobIDsJSON  string `gorm:"type:text"`
	OutputDir        string `gorm:"size:200;not null;default:output"`
	ArtifactFormat   string `gorm:"size:20;not null;default:gzip"`
	MaxArtifacts     int    `gorm:"not null;default:10"`
}

func (aiAgentWorkspaceMigrationModel) TableName() string { return "ai_agents" }

type agentRunWorkspaceMigrationModel struct {
	ID           uint   `gorm:"primaryKey"`
	WorkDir      string `gorm:"size:500"`
	ArtifactPath string `gorm:"size:500"`
}

func (agentRunWorkspaceMigrationModel) TableName() string { return "agent_runs" }
