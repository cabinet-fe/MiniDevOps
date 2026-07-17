package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000018_agent_persistent_workspace", upAgentPersistentWorkspace)
}

func upAgentPersistentWorkspace(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	agent := &aiAgentArtifactColumnsMigrationModel{}
	for _, column := range []string{"artifact_format", "max_artifacts"} {
		if db.Migrator().HasColumn(agent, column) {
			if err := db.Migrator().DropColumn(agent, column); err != nil {
				return err
			}
		}
	}
	// Keep ai_agents.output_dir: each agent still has one fixed relative output directory.

	run := &agentRunArtifactColumnMigrationModel{}
	if db.Migrator().HasColumn(run, "artifact_path") {
		if err := db.Migrator().DropColumn(run, "artifact_path"); err != nil {
			return err
		}
	}
	return nil
}

type aiAgentArtifactColumnsMigrationModel struct {
	ID             uint   `gorm:"primaryKey"`
	OutputDir      string `gorm:"size:200;not null;default:output"`
	ArtifactFormat string `gorm:"size:20;not null;default:gzip"`
	MaxArtifacts   int    `gorm:"not null;default:10"`
}

func (aiAgentArtifactColumnsMigrationModel) TableName() string { return "ai_agents" }

type agentRunArtifactColumnMigrationModel struct {
	ID           uint   `gorm:"primaryKey"`
	WorkDir      string `gorm:"size:500"`
	ArtifactPath string `gorm:"size:500"`
}

func (agentRunArtifactColumnMigrationModel) TableName() string { return "agent_runs" }
