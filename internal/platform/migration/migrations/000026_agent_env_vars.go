package migrations

import (
	"context"

	"gorm.io/gorm"

	"bedrock/internal/platform/migration"
)

func init() {
	migration.Register("000026_agent_env_vars", upAgentEnvVars)
}

func upAgentEnvVars(ctx context.Context, db *gorm.DB, driver migration.Driver) error {
	_ = ctx
	_ = driver

	agent := &agentEnvVarsMigrationModel{}
	if !db.Migrator().HasColumn(agent, "env_vars_cipher") {
		if err := db.Migrator().AddColumn(agent, "EnvVarsCipher"); err != nil {
			return err
		}
	}
	return nil
}

type agentEnvVarsMigrationModel struct {
	ID            uint   `gorm:"primaryKey"`
	EnvVarsCipher string `gorm:"type:text"`
}

func (agentEnvVarsMigrationModel) TableName() string { return "ai_agents" }
