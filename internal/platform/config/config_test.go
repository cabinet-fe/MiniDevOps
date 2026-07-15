package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_sqliteDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	content := `
server:
  port: 8080
database:
  driver: sqlite
  path: "./data/db.sqlite"
jwt:
  secret: "test-secret"
encryption:
  key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
admin:
  username: admin
  password: admin123
build:
  workspace_dir: "./data/workspaces"
  artifact_dir: "./data/artifacts"
  log_dir: "./data/logs"
  cache_dir: "./data/caches"
`
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Database.Driver != "sqlite" {
		t.Fatalf("driver=%q", cfg.Database.Driver)
	}
	if cfg.Database.Path != filepath.Join(tmpDir, "data", "db.sqlite") {
		t.Fatalf("path=%q", cfg.Database.Path)
	}
}

func TestLoad_rejectsBadDriver(t *testing.T) {
	tmpDir := t.TempDir()
	content := `
server:
  port: 8080
database:
  driver: oracle
  path: "./data/db.sqlite"
jwt:
  secret: "test-secret"
encryption:
  key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
`
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLoad_postgresRequiresHost(t *testing.T) {
	tmpDir := t.TempDir()
	content := `
database:
  driver: postgres
  name: bedrock
  user: bedrock
jwt:
  secret: "test-secret"
encryption:
  key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
`
	path := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}
