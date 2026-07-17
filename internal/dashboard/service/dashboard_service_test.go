package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bedrock/internal/dashboard/model"
	"bedrock/internal/dashboard/repository"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
)

func TestSystemStatusReportsHostDiskAndDirectorySizes(t *testing.T) {
	root := t.TempDir()
	workspace := filepath.Join(root, "workspaces")
	artifacts := filepath.Join(root, "artifacts")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(artifacts, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := []byte("bedrock-dashboard-disk")
	if err := os.WriteFile(filepath.Join(workspace, "a.bin"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(artifacts, "b.bin"), append(payload, payload...), 0o644); err != nil {
		t.Fatal(err)
	}

	repo := newDashboardRepository(t)
	svc := NewDashboardService(repo, "test", time.Now(), []string{workspace, artifacts})
	status, err := svc.SystemStatus()
	if err != nil {
		t.Fatal(err)
	}
	if status.DiskTotalBytes == 0 {
		t.Fatalf("expected host disk total, got %#v", status)
	}
	if status.DiskUsedBytes == 0 && status.DiskUsagePercent == 0 && status.DiskFreeBytes == 0 {
		t.Fatalf("expected host disk usage sample, got %#v", status)
	}
	if len(status.Directories) != 2 {
		t.Fatalf("directories = %#v", status.Directories)
	}
	if status.Directories[0].UsedBytes != uint64(len(payload)) {
		t.Fatalf("workspace used = %d, want %d", status.Directories[0].UsedBytes, len(payload))
	}
	if status.Directories[1].UsedBytes != uint64(len(payload)*2) {
		t.Fatalf("artifacts used = %d, want %d", status.Directories[1].UsedBytes, len(payload)*2)
	}
}

func TestDirectoryUsedBytesMissingRootIsZero(t *testing.T) {
	got, err := directoryUsedBytes(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if got != 0 {
		t.Fatalf("got %d, want 0", got)
	}
}

func TestLayoutFiltersStaleBuildCardAndRejectsUnauthorizedAddition(t *testing.T) {
	repo := newDashboardRepository(t)
	svc := NewDashboardService(repo, "test", time.Now(), []string{"."})
	permissions := []string{"dashboard.system_info:view", "dashboard.system_status:view"}

	if err := repo.CreateLayout(&model.Layout{
		UserID:    42,
		CardsJSON: `[{"id":"build_summary","visible":true,"order":0},{"id":"system_info","visible":true,"order":1}]`,
	}); err != nil {
		t.Fatal(err)
	}
	layout, err := svc.GetLayout(42, false, permissions)
	if err != nil {
		t.Fatal(err)
	}
	for _, card := range layout.Cards {
		if card.ID == CardBuildSummary {
			t.Fatalf("stale build card must be filtered: %#v", layout.Cards)
		}
	}
	_, err = svc.PutLayout(42, false, permissions, []model.CardLayout{
		{ID: CardBuildSummary, Visible: true, Order: 0},
	})
	if !errors.Is(err, ErrUnauthorizedCard) {
		t.Fatalf("expected ErrUnauthorizedCard, got %v", err)
	}
}

func TestLayoutPersistsAfterPut(t *testing.T) {
	repo := newDashboardRepository(t)
	svc := NewDashboardService(repo, "test", time.Now(), []string{"."})
	permissions := []string{
		"cicd.build_runs:view", "dashboard.system_info:view", "dashboard.system_status:view",
	}
	want := []model.CardLayout{
		{ID: CardSystemStatus, Visible: false, Order: 0},
		{ID: CardBuildSummary, Visible: true, Order: 1},
		{ID: CardSystemInfo, Visible: true, Order: 2},
	}
	if _, err := svc.PutLayout(7, false, permissions, want); err != nil {
		t.Fatal(err)
	}
	// A fresh service models a restart or a new request lifecycle.
	got, err := NewDashboardService(repo, "test", time.Now(), []string{"."}).GetLayout(7, false, permissions)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Cards) != len(want) {
		t.Fatalf("got %d cards, want %d: %#v", len(got.Cards), len(want), got.Cards)
	}
	for index, card := range got.Cards {
		if card != want[index] {
			t.Fatalf("card %d = %#v, want %#v", index, card, want[index])
		}
	}
}

func newDashboardRepository(t *testing.T) *repository.DashboardRepository {
	t.Helper()
	gdb, err := db.Open(&config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "dashboard.sqlite"),
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
	if err := migration.Up(context.Background(), gdb, "sqlite"); err != nil {
		t.Fatal(err)
	}
	return repository.NewDashboardRepository(gdb)
}
