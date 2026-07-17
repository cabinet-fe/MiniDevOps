package pkg

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type backupTestUser struct {
	ID           uint `gorm:"primaryKey"`
	Username     string
	PasswordHash string
	Role         string
	IsActive     bool
}

func (backupTestUser) TableName() string { return "users" }

type backupTestProject struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	CreatedBy uint
}

func (backupTestProject) TableName() string { return "projects" }

type backupTestEnvironment struct {
	ID        uint `gorm:"primaryKey"`
	ProjectID uint
	Name      string
	Branch    string
}

func (backupTestEnvironment) TableName() string { return "environments" }

type backupTestBuild struct {
	ID            uint `gorm:"primaryKey"`
	ProjectID     uint
	EnvironmentID uint
	BuildNumber   int
	Status        string
	CurrentStage  string
}

func (backupTestBuild) TableName() string { return "builds" }

type backupTestAuditLog struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint
	Action       string
	ResourceType string
	ResourceID   uint
}

func (backupTestAuditLog) TableName() string { return "audit_logs" }

type backupTestNotification struct {
	ID      uint `gorm:"primaryKey"`
	UserID  uint
	Type    string
	Title   string
	Message string
}

func (backupTestNotification) TableName() string { return "notifications" }

func TestPrepareSlimSQLiteBackup_PrunesTables(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "db.sqlite")

	db, err := gorm.Open(sqlite.Open(src), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(
		&backupTestUser{},
		&backupTestProject{},
		&backupTestEnvironment{},
		&backupTestBuild{},
		&backupTestAuditLog{},
		&backupTestNotification{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	u := &backupTestUser{Username: "u1", PasswordHash: "x", Role: "admin", IsActive: true}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("user: %v", err)
	}
	p := &backupTestProject{Name: "p1", CreatedBy: u.ID}
	if err := db.Create(p).Error; err != nil {
		t.Fatalf("project: %v", err)
	}
	env := &backupTestEnvironment{ProjectID: p.ID, Name: "dev", Branch: "main"}
	if err := db.Create(env).Error; err != nil {
		t.Fatalf("env: %v", err)
	}
	if err := db.Create(&backupTestBuild{
		ProjectID: p.ID, EnvironmentID: env.ID, BuildNumber: 1, Status: "success", CurrentStage: "done",
	}).Error; err != nil {
		t.Fatalf("build: %v", err)
	}
	if err := db.Create(&backupTestAuditLog{UserID: u.ID, Action: "create", ResourceType: "project", ResourceID: p.ID}).Error; err != nil {
		t.Fatalf("audit: %v", err)
	}
	if err := db.Create(&backupTestNotification{UserID: u.ID, Type: "build", Title: "t", Message: "m"}).Error; err != nil {
		t.Fatalf("notification: %v", err)
	}

	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	out, cleanup, err := PrepareSlimSQLiteBackup(src)
	if err != nil {
		t.Fatalf("PrepareSlimSQLiteBackup: %v", err)
	}
	defer cleanup()

	db2, err := gorm.Open(sqlite.Open(out), &gorm.Config{})
	if err != nil {
		t.Fatalf("open slim: %v", err)
	}

	var users, builds, audits, notifs int64
	for _, tc := range []struct {
		name  string
		model interface{}
		cnt   *int64
		want  int64
	}{
		{"users", &backupTestUser{}, &users, 1},
		{"builds", &backupTestBuild{}, &builds, 0},
		{"audit_logs", &backupTestAuditLog{}, &audits, 0},
		{"notifications", &backupTestNotification{}, &notifs, 0},
	} {
		if err := db2.Model(tc.model).Count(tc.cnt).Error; err != nil {
			t.Fatalf("count %s: %v", tc.name, err)
		}
		if *tc.cnt != tc.want {
			t.Fatalf("%s: got count %d want %d", tc.name, *tc.cnt, tc.want)
		}
	}
}

func TestPrepareSlimSQLiteBackup_MissingFile(t *testing.T) {
	t.Parallel()
	_, _, err := PrepareSlimSQLiteBackup(filepath.Join(t.TempDir(), "nope.sqlite"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
