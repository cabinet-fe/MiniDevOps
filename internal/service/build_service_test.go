package service

import (
	"os"
	"testing"

	"buildflow/internal/model"
	"buildflow/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newBuildServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.Project{}, &model.Environment{}, &model.Build{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func TestTriggerBuildUsesEnvironmentBranchAndInitialStage(t *testing.T) {
	db := newBuildServiceTestDB(t)
	envRepo := repository.NewEnvironmentRepository(db)
	buildRepo := repository.NewBuildRepository(db)

	env := &model.Environment{
		ProjectID: 1,
		Name:      "dev",
		Branch:    "feature/login-fix",
	}
	if err := envRepo.Create(env); err != nil {
		t.Fatalf("create env: %v", err)
	}

	svc := NewBuildService(buildRepo, nil, envRepo, nil)

	build, err := svc.TriggerBuild(1, env.ID, 7, "manual", "", "", "")
	if err != nil {
		t.Fatalf("trigger build: %v", err)
	}

	if build.Branch != env.Branch {
		t.Fatalf("expected branch %q, got %q", env.Branch, build.Branch)
	}
	if build.Status != "pending" {
		t.Fatalf("expected status pending, got %q", build.Status)
	}
	if build.CurrentStage != "pending" {
		t.Fatalf("expected current stage pending, got %q", build.CurrentStage)
	}
}

func TestGetBuildDetailFallsBackToEnvironmentBranchAndInferredStage(t *testing.T) {
	db := newBuildServiceTestDB(t)
	projectRepo := repository.NewProjectRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	buildRepo := repository.NewBuildRepository(db)

	project := &model.Project{
		Name:         "demo",
		RepoURL:      "https://example.com/repo.git",
		RepoAuthType: "none",
	}
	if err := projectRepo.Create(project); err != nil {
		t.Fatalf("create project: %v", err)
	}

	env := &model.Environment{
		ProjectID: project.ID,
		Name:      "prod",
		Branch:    "release",
	}
	if err := envRepo.Create(env); err != nil {
		t.Fatalf("create env: %v", err)
	}

	logFile, err := os.CreateTemp(t.TempDir(), "build-*.log")
	if err != nil {
		t.Fatalf("create temp log: %v", err)
	}
	defer logFile.Close()

	if _, err := logFile.WriteString("=== Stage: Cloning ===\n=== Stage: Building ===\nERROR: Build failed with exit status 1\n"); err != nil {
		t.Fatalf("write temp log: %v", err)
	}

	build := &model.Build{
		ProjectID:     project.ID,
		EnvironmentID: env.ID,
		BuildNumber:   1,
		Status:        "failed",
		CurrentStage:  "pending",
		LogPath:       logFile.Name(),
		ErrorMessage:  "构建失败: exit status 1",
	}
	if err := buildRepo.Create(build); err != nil {
		t.Fatalf("create build: %v", err)
	}

	svc := NewBuildService(buildRepo, projectRepo, envRepo, nil)
	detail, err := svc.GetBuildDetail(build.ID)
	if err != nil {
		t.Fatalf("get build detail: %v", err)
	}

	if detail.Branch != env.Branch {
		t.Fatalf("expected branch %q, got %q", env.Branch, detail.Branch)
	}
	if detail.CurrentStage != "building" {
		t.Fatalf("expected inferred current stage building, got %q", detail.CurrentStage)
	}
}
