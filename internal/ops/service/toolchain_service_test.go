package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bedrock/internal/ops/model"
	"bedrock/internal/ops/repository"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
	systemrepository "bedrock/internal/system/repository"
	systemservice "bedrock/internal/system/service"

	"gorm.io/gorm"
)

func TestInstallFallsBackToSecondSourceAndKeepsLogs(t *testing.T) {
	gdb := newOpsDatabase(t)
	repo := repository.NewOpsRepository(gdb)
	audit := systemservice.NewAuditService(systemrepository.NewOperationLogRepository(gdb))
	svc := NewToolchainService(repo, audit)
	svc.Start()
	t.Cleanup(svc.Shutdown)

	toolchain, err := svc.CreateCustom(ToolchainInput{
		Name:            "fallback-test",
		Executable:      "sh",
		InstallTemplate: `if [ "{{source_url}}" = "second" ]; then echo second-ok; else echo first-failed; exit 1; fi`,
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateSource(SourceInput{Name: "first-source", BaseURL: "first", Priority: -20, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateSource(SourceInput{Name: "second-source", BaseURL: "second", Priority: -10, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	job, err := svc.Enqueue(toolchain.ID, "install", JobInput{}, 1)
	if err != nil {
		t.Fatal(err)
	}
	finished := awaitJob(t, svc, job.ID)
	if finished.Status != model.JobSuccess {
		t.Fatalf("job status = %s, logs:\n%s", finished.Status, finished.LogText)
	}
	if !strings.Contains(finished.LogText, `source "first-source" failed`) ||
		!strings.Contains(finished.LogText, `source "second-source" succeeded`) {
		t.Fatalf("fallback is not observable in logs:\n%s", finished.LogText)
	}
	audits, total, err := audit.List(systemrepository.OperationLogFilters{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if total < 3 {
		t.Fatalf("audit events = %d, want at least enqueue/start/completion", total)
	}
	wantResourceID := fmt.Sprint(job.ID)
	var completionFound bool
	for _, event := range audits {
		if event.Action == "toolchain_job_completed" && event.ResourceID == wantResourceID {
			completionFound = strings.Contains(event.Details, "source=second-source") &&
				strings.Contains(event.Details, "source_fallback=true")
			break
		}
	}
	if !completionFound {
		t.Fatalf("completion audit does not record fallback: %#v", audits)
	}
}

func TestRecoverInterruptsRunningJobAndRetryCreatesNewJob(t *testing.T) {
	repo := newOpsRepository(t)
	svc := NewToolchainService(repo)
	svc.Start()
	t.Cleanup(svc.Shutdown)

	toolchain, err := svc.CreateCustom(ToolchainInput{
		Name: "retry-test", Executable: "sh", UninstallTemplate: "true",
	}, 1)
	if err != nil {
		t.Fatal(err)
	}
	running := &model.ToolchainInstallJob{
		ToolchainID: toolchain.ID, Operation: "uninstall", Status: model.JobRunning, LogText: "kept log\n",
	}
	if err := repo.CreateJob(running); err != nil {
		t.Fatal(err)
	}
	if err := svc.RecoverOnStartup(); err != nil {
		t.Fatal(err)
	}
	interrupted, err := svc.GetJob(running.ID)
	if err != nil {
		t.Fatal(err)
	}
	if interrupted.Status != model.JobInterrupted || interrupted.LogText != "kept log\n" {
		t.Fatalf("restart must retain logs and interrupt: %#v", interrupted)
	}
	retry, err := svc.Retry(running.ID, 2)
	if err != nil {
		t.Fatal(err)
	}
	if retry.ID == running.ID {
		t.Fatalf("retry must create a new job: %#v", retry)
	}
}

func TestKillSelfAndDangerousProcessAreRejected(t *testing.T) {
	svc := NewProcessService()
	if _, err := svc.KillProcess(int32(os.Getpid())); !errors.Is(err, ErrKillSelf) {
		t.Fatalf("self kill error = %v", err)
	}
	for _, item := range []struct {
		pid  int32
		name string
	}{{1, "anything"}, {999, "systemd"}, {999, "Bedrock"}} {
		if !IsDangerousProcess(item.pid, item.name) {
			t.Fatalf("expected protected process: %#v", item)
		}
	}
}

func TestToolchainLifecycleExecutesEachOperation(t *testing.T) {
	repo := newOpsRepository(t)
	svc := NewToolchainService(repo)
	svc.Start()
	t.Cleanup(svc.Shutdown)

	marker := filepath.Join(t.TempDir(), "toolchain-lifecycle.log")
	command := func(operation string) string {
		return fmt.Sprintf("printf '%s\\n' >> %q", operation, marker)
	}
	toolchain, err := svc.CreateCustom(ToolchainInput{
		Name:              "lifecycle-test",
		Executable:        "sh",
		DetectCommand:     command("detect"),
		InstallTemplate:   command("install"),
		UpgradeTemplate:   command("upgrade"),
		UninstallTemplate: command("uninstall"),
		SwitchTemplate:    command("switch"),
	}, 1)
	if err != nil {
		t.Fatal(err)
	}

	detect, err := svc.Detect(toolchain.ID)
	if err != nil || !detect.Detected {
		t.Fatalf("detect = %#v, %v", detect, err)
	}
	for _, operation := range []string{"install", "upgrade", "uninstall", "switch"} {
		job, err := svc.Enqueue(toolchain.ID, operation, JobInput{Version: "1.2.3"}, 1)
		if err != nil {
			t.Fatalf("enqueue %s: %v", operation, err)
		}
		if finished := awaitJob(t, svc, job.ID); finished.Status != model.JobSuccess {
			t.Fatalf("%s status = %s, logs:\n%s", operation, finished.Status, finished.LogText)
		}
	}
	content, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"detect", "install", "upgrade", "uninstall", "switch"} {
		if !strings.Contains(string(content), operation+"\n") {
			t.Fatalf("lifecycle marker is missing %q: %s", operation, content)
		}
	}
	updated, err := repo.FindToolchain(toolchain.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.DefaultVersion != "1.2.3" {
		t.Fatalf("switch did not update default version: %q", updated.DefaultVersion)
	}
}

func TestSeededGoToolchainLifecycleExecutesWithStubbedManagers(t *testing.T) {
	repo := newOpsRepository(t)
	svc := NewToolchainService(repo)
	svc.Start()
	t.Cleanup(svc.Shutdown)

	stubDir := t.TempDir()
	logPath := filepath.Join(stubDir, "toolchain-lifecycle.log")
	writeStubExecutable(t, filepath.Join(stubDir, "go"), `#!/bin/sh
printf 'go %s\n' "$*" >> "$TOOLCHAIN_LIFECYCLE_LOG"
printf 'go version go1.23.4 stub/amd64\n'
`)
	writeStubExecutable(t, filepath.Join(stubDir, "asdf"), `#!/bin/sh
printf 'asdf %s\n' "$*" >> "$TOOLCHAIN_LIFECYCLE_LOG"
`)
	t.Setenv("TOOLCHAIN_LIFECYCLE_LOG", logPath)
	t.Setenv("PATH", stubDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	items, err := repo.ListToolchains()
	if err != nil {
		t.Fatal(err)
	}
	var goToolchain *model.ToolchainDefinition
	for i := range items {
		if items[i].Name == "Go" {
			goToolchain = &items[i]
			break
		}
	}
	if goToolchain == nil || goToolchain.Kind != model.ToolchainBuiltin {
		t.Fatalf("seeded Go toolchain = %#v, want builtin definition", goToolchain)
	}

	detected, err := svc.Detect(goToolchain.ID)
	if err != nil || !detected.Detected {
		t.Fatalf("detect Go = %#v, %v", detected, err)
	}
	for _, operation := range []string{"install", "upgrade", "uninstall", "switch"} {
		job, err := svc.Enqueue(goToolchain.ID, operation, JobInput{Version: "1.23.4"}, 1)
		if err != nil {
			t.Fatalf("enqueue Go %s: %v", operation, err)
		}
		if finished := awaitJob(t, svc, job.ID); finished.Status != model.JobSuccess {
			t.Fatalf("Go %s status = %s, logs:\n%s", operation, finished.Status, finished.LogText)
		}
	}

	logContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	logs := string(logContent)
	for _, want := range []string{
		"go version",
		"asdf plugin add golang https://github.com/asdf-community/asdf-golang.git",
		"asdf install golang 1.23.4",
		"asdf plugin update golang",
		"asdf uninstall golang 1.23.4",
		"asdf set -u golang 1.23.4",
	} {
		if !strings.Contains(logs, want+"\n") {
			t.Fatalf("seeded Go lifecycle did not invoke %q:\n%s", want, logs)
		}
	}
	updated, err := repo.FindToolchain(goToolchain.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.DefaultVersion != "1.23.4" {
		t.Fatalf("switch did not update seeded Go default version: %q", updated.DefaultVersion)
	}
}

func TestSeededGoToolchainUsesRealLifecycleTemplates(t *testing.T) {
	repo := newOpsRepository(t)
	items, err := repo.ListToolchains()
	if err != nil {
		t.Fatal(err)
	}
	var goToolchain *model.ToolchainDefinition
	for i := range items {
		if items[i].Name == "Go" {
			goToolchain = &items[i]
			break
		}
	}
	if goToolchain == nil {
		t.Fatal("seeded Go toolchain is missing")
	}
	for _, operation := range []string{"install", "upgrade", "uninstall", "switch"} {
		template, err := commandTemplate(goToolchain, operation)
		if err != nil {
			t.Fatalf("Go %s template: %v", operation, err)
		}
		if !strings.Contains(template, "asdf") || strings.Contains(template, "go{{version}} version") {
			t.Fatalf("Go %s template is a placeholder: %q", operation, template)
		}
	}
}

func TestRedactSpacedCLISecrets(t *testing.T) {
	input := `tool --token token-secret -t short-token --password "password secret" --api-key api-secret --secret secret-value --access-token access-secret --auth auth-secret --credential credential-secret --header "Authorization: Bearer header-secret"`
	redacted := redact(input)
	for _, secret := range []string{
		"token-secret",
		"short-token",
		"password secret",
		"api-secret",
		"secret-value",
		"access-secret",
		"auth-secret",
		"credential-secret",
		"header-secret",
	} {
		if strings.Contains(redacted, secret) {
			t.Fatalf("redacted command leaks %q: %s", secret, redacted)
		}
	}
	for _, want := range []string{
		"--token [REDACTED]",
		"-t [REDACTED]",
		"--password [REDACTED]",
		`--header "Authorization: Bearer [REDACTED]"`,
	} {
		if !strings.Contains(redacted, want) {
			t.Fatalf("redaction missing %q: %s", want, redacted)
		}
	}
}

func TestFailedJobPersistsRedactedSpacedCLISecrets(t *testing.T) {
	repo := newOpsRepository(t)
	svc := NewToolchainService(repo)
	job := &model.ToolchainInstallJob{
		ToolchainID:     1,
		Operation:       "install",
		Status:          model.JobRunning,
		CommandSnapshot: "tool --token snapshot-secret",
		LogText:         "tool --password log-secret",
	}
	if err := repo.CreateJob(job); err != nil {
		t.Fatal(err)
	}

	svc.fail(job, errors.New("tool --api-key error-secret --access-token access-secret --secret secret-value"), "", false)

	persisted, err := repo.FindJob(job.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{persisted.CommandSnapshot, persisted.LogText, persisted.ErrorMessage} {
		for _, secret := range []string{"snapshot-secret", "log-secret", "error-secret", "access-secret", "secret-value"} {
			if strings.Contains(value, secret) {
				t.Fatalf("persisted failed job leaks %q: %q", secret, value)
			}
		}
	}
}

func awaitJob(t *testing.T, svc *ToolchainService, id uint) *model.ToolchainInstallJob {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		job, err := svc.GetJob(id)
		if err != nil {
			t.Fatal(err)
		}
		if job.Status != model.JobQueued && job.Status != model.JobRunning {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %d did not finish", id)
	return nil
}

func writeStubExecutable(t *testing.T, path, script string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
}

func newOpsRepository(t *testing.T) *repository.OpsRepository {
	t.Helper()
	return repository.NewOpsRepository(newOpsDatabase(t))
}

func newOpsDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := db.Open(&config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "ops.sqlite"),
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
	return gdb
}
