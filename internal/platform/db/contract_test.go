//go:build contract

package db_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bedrock/internal/cicd/model"
	"bedrock/internal/cicd/repository"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Three-driver migration + thin CI/CD CRUD contract (AGENTS.md).
//
//	go test ./internal/platform/db/... -tags=contract
//
// Postgres/MySQL skip with a clear message when DSN env is unset:
//
//	BEDROCK_CONTRACT_POSTGRES_DSN
//	BEDROCK_CONTRACT_MYSQL_DSN

func TestContract_MigrationsAndCICDTables(t *testing.T) {
	for _, driver := range []string{"sqlite", "postgres", "mysql"} {
		t.Run(driver, func(t *testing.T) {
			gdb := openDriver(t, driver)
			if err := migration.Up(context.Background(), gdb, migration.Driver(db.NormalizeDriver(driver))); err != nil {
				t.Fatalf("migration.Up: %v", err)
			}
			for _, table := range []string{
				"repositories", "build_jobs", "build_runs", "credentials",
				"deploy_targets", "build_deploy_attempts", "schema_migrations",
			} {
				if !gdb.Migrator().HasTable(table) {
					t.Fatalf("missing table %s on %s", table, driver)
				}
			}

			suffix := fmt.Sprintf("%s-%d", driver, time.Now().UnixNano())
			credRepo := repository.NewCredentialRepository(gdb)
			repoRepo := repository.NewRepositoryRepository(gdb)
			cred := &model.Credential{Name: "db-contract-" + suffix, Type: "token", SecretCipher: "x", CreatedBy: 99}
			if err := credRepo.Create(cred); err != nil {
				t.Fatalf("create credential: %v", err)
			}
			repo := &model.Repository{
				Name: "db-contract-repo-" + suffix, RepoURL: "https://example.com/" + suffix + ".git",
				AuthType: "none", CreatedBy: 99,
			}
			if err := repoRepo.Create(repo); err != nil {
				t.Fatalf("create repository: %v", err)
			}
			_ = repoRepo.Delete(repo.ID)
			_ = credRepo.Delete(cred.ID)
		})
	}
}

func openDriver(t *testing.T, driver string) *gorm.DB {
	t.Helper()
	switch db.NormalizeDriver(driver) {
	case "sqlite":
		gdb, err := db.Open(&config.DatabaseConfig{
			Driver: "sqlite",
			Path:   filepath.Join(t.TempDir(), "contract.sqlite"),
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			sqlDB, _ := gdb.DB()
			if sqlDB != nil {
				_ = sqlDB.Close()
			}
		})
		return gdb
	case "postgres":
		dsn := os.Getenv("BEDROCK_CONTRACT_POSTGRES_DSN")
		if dsn == "" {
			t.Skip("BEDROCK_CONTRACT_POSTGRES_DSN not set; skipping postgres contract test")
		}
		gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatalf("open postgres: %v", err)
		}
		sqlDB, err := gdb.DB()
		if err != nil {
			t.Fatal(err)
		}
		if err := sqlDB.Ping(); err != nil {
			t.Skipf("postgres unreachable: %v", err)
		}
		t.Cleanup(func() { _ = sqlDB.Close() })
		return gdb
	case "mysql":
		dsn := os.Getenv("BEDROCK_CONTRACT_MYSQL_DSN")
		if dsn == "" {
			t.Skip("BEDROCK_CONTRACT_MYSQL_DSN not set; skipping mysql contract test")
		}
		gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			t.Fatalf("open mysql: %v", err)
		}
		sqlDB, err := gdb.DB()
		if err != nil {
			t.Fatal(err)
		}
		if err := sqlDB.Ping(); err != nil {
			t.Skipf("mysql unreachable: %v", err)
		}
		t.Cleanup(func() { _ = sqlDB.Close() })
		return gdb
	default:
		t.Fatalf("unknown driver %s", driver)
		return nil
	}
}
