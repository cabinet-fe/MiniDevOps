//go:build contract

package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"bedrock/internal/ai/model"
	"bedrock/internal/ai/repository"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestContract_AI_CRUD(t *testing.T) {
	for _, driver := range []string{"sqlite", "postgres", "mysql"} {
		t.Run(driver, func(t *testing.T) {
			gdb := openContractDB(t, driver)
			if err := migration.Up(context.Background(), gdb, migration.Driver(db.NormalizeDriver(driver))); err != nil {
				t.Fatalf("migration.Up(%s): %v", driver, err)
			}
			if err := migration.Up(context.Background(), gdb, migration.Driver(db.NormalizeDriver(driver))); err != nil {
				t.Fatalf("idempotent migration.Up(%s): %v", driver, err)
			}
			for _, column := range []string{"artifact_format", "max_artifacts"} {
				if gdb.Migrator().HasColumn("ai_agents", column) {
					t.Fatalf("ai_agents.%s still exists on %s", column, driver)
				}
			}
			if !gdb.Migrator().HasColumn("ai_agents", "output_dir") {
				t.Fatalf("ai_agents.output_dir missing on %s", driver)
			}
			if gdb.Migrator().HasColumn("agent_runs", "artifact_path") {
				t.Fatalf("agent_runs.artifact_path still exists on %s", driver)
			}
			if !gdb.Migrator().HasColumn("agent_runs", "work_dir") {
				t.Fatalf("agent_runs.work_dir missing on %s", driver)
			}
			repo := repository.NewAIRepository(gdb)

			clis, err := repo.ListCLIs()
			if err != nil || len(clis) != 4 {
				t.Fatalf("seeded CLIs: %v len=%d", err, len(clis))
			}

			agent := &model.AiAgent{
				Name: "a", CliKey: "claude_code", Enabled: true, TimeoutSec: 60,
				SkillIDsJSON: "[]", BuildJobIDsJSON: "[]",
				OutputDir: "output",
			}
			if err := repo.CreateAgent(agent); err != nil {
				t.Fatal(err)
			}
			trig := &model.AgentTrigger{AgentID: agent.ID, Type: model.TriggerManual, Enabled: true}
			if err := repo.CreateTrigger(trig); err != nil {
				t.Fatal(err)
			}
			run := &model.AgentRun{AgentID: agent.ID, TriggerType: model.TriggerManual, Status: model.JobQueued}
			if err := repo.CreateRun(run); err != nil {
				t.Fatal(err)
			}
			skill := &model.SkillPackage{
				Name: "s", Visibility: model.SkillPrivate, StorageObjectID: 1,
				PackageDigest: "abc", SizeBytes: 1, CreatedBy: 1, UpdatedBy: 1,
			}
			if err := repo.CreateSkill(skill); err != nil {
				t.Fatal(err)
			}
			pat := &model.PersonalAccessToken{
				UserID: 1, Name: "p", TokenPrefix: "br_pat_xxxx", TokenHash: "hash-" + driver,
				ScopesJSON: `["skills:read"]`,
			}
			if err := repo.CreatePAT(pat); err != nil {
				t.Fatal(err)
			}
			if _, err := repo.FindPATByHash(pat.TokenHash); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func openContractDB(t *testing.T, driver string) *gorm.DB {
	t.Helper()
	switch db.NormalizeDriver(driver) {
	case "sqlite":
		gdb, err := db.Open(&config.DatabaseConfig{Driver: "sqlite", Path: filepath.Join(t.TempDir(), "c.sqlite")})
		if err != nil {
			t.Fatal(err)
		}
		return gdb
	case "postgres":
		dsn := os.Getenv("BEDROCK_CONTRACT_POSTGRES_DSN")
		if dsn == "" {
			t.Skip("BEDROCK_CONTRACT_POSTGRES_DSN not set")
		}
		gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal(err)
		}
		return gdb
	case "mysql":
		dsn := os.Getenv("BEDROCK_CONTRACT_MYSQL_DSN")
		if dsn == "" {
			t.Skip("BEDROCK_CONTRACT_MYSQL_DSN not set")
		}
		gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatal(err)
		}
		return gdb
	default:
		t.Fatalf("unknown driver %s", driver)
		return nil
	}
}
