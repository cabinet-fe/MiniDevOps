package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_PathResolution(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a config.yaml in that directory
	configContent := `
server:
  port: 8080
database:
  path: "./data/db.sqlite"
build:
  workspace_dir: "./data/workspaces"
  artifact_dir: "./data/artifacts"
  log_dir: "./data/logs"
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load the config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check if paths are absolute and resolved relative to tmpDir
	expectedBaseDir := tmpDir

	checkPath := func(got, expectedName string) {
		absGot, _ := filepath.Abs(got)
		expected := filepath.Join(expectedBaseDir, "data", expectedName)
		if absGot != expected {
			t.Errorf("Path mismatch: got %v, expected %v", absGot, expected)
		}
	}

	checkPath(cfg.Database.Path, "db.sqlite")
	checkPath(cfg.Build.WorkspaceDir, "workspaces")
	checkPath(cfg.Build.ArtifactDir, "artifacts")
	checkPath(cfg.Build.LogDir, "logs")
}
