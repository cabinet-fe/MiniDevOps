package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000024_agent_workspace_status", upAgentWorkspaceStatus)
}

func upAgentWorkspaceStatus(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	agent := &aiAgentWorkspaceStatusMigrationModel{}
	if !db.Migrator().HasColumn(agent, "workspace_status") {
		if err := db.Migrator().AddColumn(agent, "WorkspaceStatus"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasColumn(agent, "workspace_error") {
		if err := db.Migrator().AddColumn(agent, "WorkspaceError"); err != nil {
			return err
		}
	}
	return nil
}

type aiAgentWorkspaceStatusMigrationModel struct {
	ID              uint   `gorm:"primaryKey"`
	WorkspaceStatus string `gorm:"size:20;not null;default:ready"`
	WorkspaceError  string `gorm:"type:text"`
}

func (aiAgentWorkspaceStatusMigrationModel) TableName() string { return "ai_agents" }
