package engine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"buildflow/internal/model"
	"buildflow/internal/repository"
	"buildflow/internal/ws"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newPipelineTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&model.Build{}, &model.Notification{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	return db
}

func TestFailBuildKeepsCurrentStage(t *testing.T) {
	db := newPipelineTestDB(t)
	buildRepo := repository.NewBuildRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	startedAt := time.Now().Add(-2 * time.Second)
	build := &model.Build{
		ProjectID:     1,
		EnvironmentID: 1,
		BuildNumber:   1,
		Status:        "building",
		CurrentStage:  "building",
		TriggeredBy:   1,
		StartedAt:     &startedAt,
	}
	if err := buildRepo.Create(build); err != nil {
		t.Fatalf("create build: %v", err)
	}

	pipeline := &Pipeline{
		buildRepo: buildRepo,
		notifRepo: notifRepo,
		hub:       ws.NewHub(),
	}

	pipeline.failBuild(build, "构建失败: exit status 1")

	saved, err := buildRepo.FindByID(build.ID)
	if err != nil {
		t.Fatalf("find build: %v", err)
	}

	if saved.Status != "failed" {
		t.Fatalf("expected status failed, got %q", saved.Status)
	}
	if saved.CurrentStage != "building" {
		t.Fatalf("expected current stage building, got %q", saved.CurrentStage)
	}
	if saved.ErrorMessage != "构建失败: exit status 1" {
		t.Fatalf("unexpected error message %q", saved.ErrorMessage)
	}
	if saved.DurationMs <= 0 {
		t.Fatalf("expected positive duration, got %d", saved.DurationMs)
	}
}

func TestFailBuildNoOpWhenAlreadySuccess(t *testing.T) {
	db := newPipelineTestDB(t)
	buildRepo := repository.NewBuildRepository(db)
	notifRepo := repository.NewNotificationRepository(db)

	build := &model.Build{
		ProjectID:     1,
		EnvironmentID: 1,
		BuildNumber:   1,
		Status:        "success",
		CurrentStage:  "success",
		TriggeredBy:   1,
	}
	if err := buildRepo.Create(build); err != nil {
		t.Fatalf("create build: %v", err)
	}

	pipeline := &Pipeline{
		buildRepo: buildRepo,
		notifRepo: notifRepo,
		hub:       ws.NewHub(),
	}

	pipeline.failBuild(build, "should not apply")

	saved, err := buildRepo.FindByID(build.ID)
	if err != nil {
		t.Fatalf("find build: %v", err)
	}
	if saved.Status != "success" {
		t.Fatalf("expected status to stay success, got %q", saved.Status)
	}
}

func TestCreateArtifactArchiveSupportsZipAndGzip(t *testing.T) {
	sourceDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sourceDir, "index.html"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	gzipPath := filepath.Join(t.TempDir(), "artifact.tar.gz")
	if err := createArtifactArchive(gzipPath, sourceDir, "gzip"); err != nil {
		t.Fatalf("create gzip archive: %v", err)
	}
	if err := assertTarGzContainsFile(gzipPath, "index.html"); err != nil {
		t.Fatalf("verify gzip archive: %v", err)
	}

	zipPath := filepath.Join(t.TempDir(), "artifact.zip")
	if err := createArtifactArchive(zipPath, sourceDir, "zip"); err != nil {
		t.Fatalf("create zip archive: %v", err)
	}
	if err := assertZipContainsFile(zipPath, "index.html"); err != nil {
		t.Fatalf("verify zip archive: %v", err)
	}
}

func assertTarGzContainsFile(path, name string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return os.ErrNotExist
		}
		if err != nil {
			return err
		}
		if header.Name == name {
			return nil
		}
	}
}

func assertZipContainsFile(path, name string) error {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		if file.Name == name {
			return nil
		}
	}
	return os.ErrNotExist
}
