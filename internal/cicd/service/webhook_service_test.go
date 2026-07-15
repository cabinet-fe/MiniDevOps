package service_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"bedrock/internal/cicd/repository"
	"bedrock/internal/cicd/service"
	"gorm.io/gorm"
)

func TestWebhook_SignatureFailRejects(t *testing.T) {
	_, repoSvc, _, _, runSvc, gdb := setupCICD(t)

	repo, err := repoSvc.Create(1, service.CreateRepositoryInput{
		Name: "wh-sig", RepoURL: "https://example.com/a.git", AuthType: "none",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	secret := revealedWebhookSecret(t, repoSvc, repo.ID)

	wh := newWebhookSvc(gdb, runSvc)
	body := []byte(`{"ref":"refs/heads/main","after":"abc"}`)
	headers := map[string]string{
		"X-GitHub-Event":      "push",
		"X-Hub-Signature-256": "sha256=deadbeef",
		"X-GitHub-Delivery":   "del-1",
	}
	_, err = wh.Receive(repo.ID, secret, headers, body, 0)
	if err == nil || !service.IsUnauthorized(err) {
		t.Fatalf("want unauthorized, got %v", err)
	}
}

func TestWebhook_GenericSecretAndDedup(t *testing.T) {
	_, repoSvc, _, jobSvc, runSvc, gdb := setupCICD(t)

	repo, err := repoSvc.Create(1, service.CreateRepositoryInput{
		Name: "wh-gen", RepoURL: "https://example.com/b.git", AuthType: "none",
		WebhookType: "generic",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	secret := revealedWebhookSecret(t, repoSvc, repo.ID)

	_, err = jobSvc.Create(1, service.CreateBuildJobInput{
		RepositoryID:   repo.ID,
		Name:           "job1",
		Branch:         "main",
		BranchPolicy:   "fixed",
		TriggerWebhook: boolPtr(true),
		BuildScript:    "echo ok",
	})
	if err != nil {
		t.Fatal(err)
	}

	wh := newWebhookSvc(gdb, runSvc)
	body := []byte(`{"ref":"refs/heads/main","after":"abc123","message":"hi"}`)
	headers := map[string]string{"X-GitHub-Delivery": "same-del"}

	r1, err := wh.Receive(repo.ID, secret, headers, body, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !r1.Accepted || r1.Triggered != 1 {
		t.Fatalf("first: %+v", r1)
	}

	r2, err := wh.Receive(repo.ID, secret, headers, body, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !r2.Duplicate || r2.Triggered != 0 {
		t.Fatalf("dedup: %+v", r2)
	}
}

func TestWebhook_LogsNoSecret(t *testing.T) {
	secret := "super-secret-value-xyz"
	msg := service.RedactSecret("invalid secret super-secret-value-xyz in path", secret)
	if strings.Contains(msg, secret) {
		t.Fatalf("secret leaked: %s", msg)
	}
	if !strings.Contains(msg, "***") {
		t.Fatalf("expected redaction: %s", msg)
	}
}

func TestWebhook_BranchMatching_FixedMissDoesNotEnqueue(t *testing.T) {
	_, repoSvc, _, jobSvc, runSvc, gdb := setupCICD(t)

	repo, err := repoSvc.Create(1, service.CreateRepositoryInput{
		Name: "wh-branch-miss", RepoURL: "https://example.com/miss.git", AuthType: "none",
		WebhookType: "generic",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	secret := revealedWebhookSecret(t, repoSvc, repo.ID)

	_, err = jobSvc.Create(1, service.CreateBuildJobInput{
		RepositoryID:   repo.ID,
		Name:           "fixed-main",
		Branch:         "main",
		BranchPolicy:   "fixed",
		TriggerWebhook: boolPtr(true),
		BuildScript:    "echo ok",
	})
	if err != nil {
		t.Fatal(err)
	}

	wh := newWebhookSvc(gdb, runSvc)
	res, err := wh.Receive(repo.ID, secret, map[string]string{"X-GitHub-Delivery": "branch-miss-1"},
		[]byte(`{"ref":"refs/heads/develop","after":"abc"}`), 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.Triggered != 0 || len(res.RunIDs) != 0 {
		t.Fatalf("fixed-branch miss should not enqueue: %+v", res)
	}
}

func TestWebhook_BranchMatching_MultiJobOnlyMatching(t *testing.T) {
	_, repoSvc, _, jobSvc, runSvc, gdb := setupCICD(t)

	repo, err := repoSvc.Create(1, service.CreateRepositoryInput{
		Name: "wh-branch-multi", RepoURL: "https://example.com/multi.git", AuthType: "none",
		WebhookType: "generic",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	secret := revealedWebhookSecret(t, repoSvc, repo.ID)

	mainJob, err := jobSvc.Create(1, service.CreateBuildJobInput{
		RepositoryID: repo.ID, Name: "job-main", Branch: "main", BranchPolicy: "fixed",
		TriggerWebhook: boolPtr(true), BuildScript: "echo main",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = jobSvc.Create(1, service.CreateBuildJobInput{
		RepositoryID: repo.ID, Name: "job-develop", Branch: "develop", BranchPolicy: "fixed",
		TriggerWebhook: boolPtr(true), BuildScript: "echo develop",
	})
	if err != nil {
		t.Fatal(err)
	}

	wh := newWebhookSvc(gdb, runSvc)
	res, err := wh.Receive(repo.ID, secret, map[string]string{"X-GitHub-Delivery": "branch-multi-1"},
		[]byte(`{"ref":"refs/heads/main","after":"deadbeef"}`), 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.Triggered != 1 || len(res.JobIDs) != 1 {
		t.Fatalf("want only matching job: %+v", res)
	}
	if res.JobIDs[0] != mainJob.ID {
		t.Fatalf("job_ids=%v want %d", res.JobIDs, mainJob.ID)
	}
}

func TestWebhook_BranchMatching_ParamAcceptsAnyBranch(t *testing.T) {
	_, repoSvc, _, jobSvc, runSvc, gdb := setupCICD(t)

	repo, err := repoSvc.Create(1, service.CreateRepositoryInput{
		Name: "wh-branch-param", RepoURL: "https://example.com/param.git", AuthType: "none",
		WebhookType: "generic",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	secret := revealedWebhookSecret(t, repoSvc, repo.ID)

	paramJob, err := jobSvc.Create(1, service.CreateBuildJobInput{
		RepositoryID: repo.ID, Name: "job-param", Branch: "main", BranchPolicy: "param",
		TriggerWebhook: boolPtr(true), BuildScript: "echo param",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = jobSvc.Create(1, service.CreateBuildJobInput{
		RepositoryID: repo.ID, Name: "job-fixed-main", Branch: "main", BranchPolicy: "fixed",
		TriggerWebhook: boolPtr(true), BuildScript: "echo fixed",
	})
	if err != nil {
		t.Fatal(err)
	}

	wh := newWebhookSvc(gdb, runSvc)
	res, err := wh.Receive(repo.ID, secret, map[string]string{"X-GitHub-Delivery": "branch-param-1"},
		[]byte(`{"ref":"refs/heads/feature/x","after":"cafebabe"}`), 0)
	if err != nil {
		t.Fatal(err)
	}
	// param = all-branch policy; fixed main must not match feature/x
	if res.Triggered != 1 || len(res.JobIDs) != 1 || res.JobIDs[0] != paramJob.ID {
		t.Fatalf("param policy should enqueue only param job: %+v", res)
	}
	if res.Branch != "feature/x" {
		t.Fatalf("branch=%q", res.Branch)
	}
	run, err := runSvc.Get(res.RunIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	if run.Branch != "feature/x" {
		t.Fatalf("enqueued run branch=%q want feature/x", run.Branch)
	}
}

func TestWebhook_ValidGitHubSignature(t *testing.T) {
	_, repoSvc, _, jobSvc, runSvc, gdb := setupCICD(t)

	repo, err := repoSvc.Create(1, service.CreateRepositoryInput{
		Name: "wh-ok", RepoURL: "https://example.com/c.git", AuthType: "none",
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	secret := revealedWebhookSecret(t, repoSvc, repo.ID)
	_, err = jobSvc.Create(1, service.CreateBuildJobInput{
		RepositoryID: repo.ID, Name: "j", Branch: "main", TriggerWebhook: boolPtr(true), BuildScript: "echo",
	})
	if err != nil {
		t.Fatal(err)
	}

	body := []byte(`{"ref":"refs/heads/main","after":"deadbeef","head_commit":{"id":"deadbeef","message":"m"}}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	wh := newWebhookSvc(gdb, runSvc)
	res, err := wh.Receive(repo.ID, secret, map[string]string{
		"X-GitHub-Event":      "push",
		"X-Hub-Signature-256": sig,
		"X-GitHub-Delivery":   "del-ok",
	}, body, 0)
	if err != nil {
		t.Fatal(err)
	}
	if res.Triggered != 1 {
		t.Fatalf("triggered=%d", res.Triggered)
	}
}

func revealedWebhookSecret(t *testing.T, repoSvc *service.RepositoryService, id uint) string {
	t.Helper()
	repo, err := repoSvc.Get(id, true)
	if err != nil {
		t.Fatal(err)
	}
	if repo.WebhookSecret == "" {
		rotated, err := repoSvc.RotateWebhookSecret(id)
		if err != nil {
			t.Fatal(err)
		}
		return rotated.WebhookSecret
	}
	return repo.WebhookSecret
}

func newWebhookSvc(gdb *gorm.DB, runSvc *service.BuildRunService) *service.WebhookService {
	return service.NewWebhookService(
		repository.NewRepositoryRepository(gdb),
		repository.NewBuildJobRepository(gdb),
		repository.NewWebhookDeliveryRepository(gdb),
		runSvc,
	)
}
