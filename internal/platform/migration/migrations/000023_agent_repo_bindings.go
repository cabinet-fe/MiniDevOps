package migrations

import (
	"context"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000023_agent_repo_bindings", upAgentRepoBindings)
}

func upAgentRepoBindings(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	agent := &aiAgentDropBuildJobsMigrationModel{}
	if db.Migrator().HasColumn(agent, "build_job_ids_json") {
		if err := db.Migrator().DropColumn(agent, "BuildJobIDsJSON"); err != nil {
			return err
		}
	}

	binding := &aiAgentRepoBindingMigrationModel{}
	if !db.Migrator().HasTable(binding) {
		if err := db.Migrator().CreateTable(binding); err != nil {
			return err
		}
	}
	return nil
}

type aiAgentDropBuildJobsMigrationModel struct {
	ID              uint   `gorm:"primaryKey"`
	BuildJobIDsJSON string `gorm:"type:text"`
}

func (aiAgentDropBuildJobsMigrationModel) TableName() string { return "ai_agents" }

type aiAgentRepoBindingMigrationModel struct {
	ID           uint      `gorm:"primaryKey"`
	AgentID      uint      `gorm:"not null;uniqueIndex:uidx_agent_repo;index"`
	RepositoryID uint      `gorm:"not null;uniqueIndex:uidx_agent_repo;index"`
	Branch       string    `gorm:"size:200;not null;default:main"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (aiAgentRepoBindingMigrationModel) TableName() string { return "ai_agent_repo_bindings" }
