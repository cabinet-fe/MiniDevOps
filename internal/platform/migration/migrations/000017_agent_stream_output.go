package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000017_agent_stream_output", upAgentStreamOutput)
}

func upAgentStreamOutput(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	agent := &aiAgentStreamOutputMigrationModel{}
	if !db.Migrator().HasColumn(agent, "stream_output") {
		if err := db.Migrator().AddColumn(agent, "StreamOutput"); err != nil {
			return err
		}
	}
	return nil
}

type aiAgentStreamOutputMigrationModel struct {
	ID           uint `gorm:"primaryKey"`
	StreamOutput bool `gorm:"not null;default:false"`
}

func (aiAgentStreamOutputMigrationModel) TableName() string { return "ai_agents" }
