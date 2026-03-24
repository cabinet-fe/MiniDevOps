package deployer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalDeployer_copiesAndOverwrites(t *testing.T) {
	ctx := context.Background()
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	sub := filepath.Join(src, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "b.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	d := &LocalDeployer{}
	if err := d.Deploy(ctx, DeployOptions{SourceDir: src, RemotePath: dst}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil || string(got) != "new" {
		t.Fatalf("a.txt: %s %v", got, err)
	}

	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := d.Deploy(ctx, DeployOptions{SourceDir: src, RemotePath: dst}); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil || string(got) != "v2" {
		t.Fatalf("overwrite: %s %v", got, err)
	}
}

func TestLocalDeployer_keepsResidualFiles(t *testing.T) {
	ctx := context.Background()
	src := t.TempDir()
	dst := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "only.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(dst, "legacy.dat")
	if err := os.WriteFile(legacy, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	d := &LocalDeployer{}
	if err := d.Deploy(ctx, DeployOptions{SourceDir: src, RemotePath: dst}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacy); err != nil {
		t.Fatalf("residual removed: %v", err)
	}
}

func TestLocalDeployer_requiresAbsolutePath(t *testing.T) {
	ctx := context.Background()
	src := t.TempDir()
	d := &LocalDeployer{}
	err := d.Deploy(ctx, DeployOptions{SourceDir: src, RemotePath: "relative/path"})
	if err == nil {
		t.Fatal("expected error for non-absolute path")
	}
}
