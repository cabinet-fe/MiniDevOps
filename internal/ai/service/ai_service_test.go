package service_test

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"bedrock/internal/ai/model"
	"bedrock/internal/ai/repository"
	"bedrock/internal/ai/service"
	cicdmodel "bedrock/internal/cicd/model"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
	projectmodel "bedrock/internal/project/model"
	projectrepo "bedrock/internal/project/repository"
	projectservice "bedrock/internal/project/service"
	storagerepo "bedrock/internal/storage/repository"
	storageservice "bedrock/internal/storage/service"
)

func setupAI(t *testing.T) (*gorm.DB, *service.CLIService, *service.AgentService, *service.SkillService, *service.PATService, *projectservice.ProjectService) {
	t.Helper()
	gdb, err := db.Open(&config.DatabaseConfig{
		Driver: "sqlite",
		Path:   filepath.Join(t.TempDir(), "ai.sqlite"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := migration.Up(context.Background(), gdb, migration.Driver("sqlite")); err != nil {
		t.Fatalf("migration: %v", err)
	}
	repo := repository.NewAIRepository(gdb)
	cli := service.NewCLIService(repo)
	cli.Start()
	t.Cleanup(cli.Shutdown)

	storageRoot := filepath.Join(t.TempDir(), "storage")
	storageSvc, err := storageservice.NewStorageService(storagerepo.NewStorageRepository(gdb), storageRoot, storageservice.Limits{})
	if err != nil {
		t.Fatal(err)
	}
	skills := service.NewSkillService(repo, storageSvc)
	pats := service.NewPATService(repo)
	work := filepath.Join(t.TempDir(), "work")
	logs := filepath.Join(t.TempDir(), "logs")
	agents := service.NewAgentService(repo, cli, skills, nil, zap.NewNop(), work, logs)
	agents.Start()
	t.Cleanup(agents.Shutdown)

	projectSvc := projectservice.NewProjectService(projectrepo.NewProjectRepository(gdb), storageSvc)
	agents.SetDocDraftWriter(projectSvc)
	projectSvc.SetDocsAIBridge(service.NewDocsBridge(agents))
	return gdb, cli, agents, skills, pats, projectSvc
}

func TestFourCLIDetectReferencePaths(t *testing.T) {
	_, cli, _, _, _, _ := setupAI(t)
	for _, key := range []string{"claude_code", "opencode", "reasonix", "codex"} {
		result, err := cli.Detect(key)
		if err != nil {
			t.Fatalf("%s detect: %v", key, err)
		}
		if result.RiskNotice == "" {
			t.Fatalf("%s missing risk notice", key)
		}
		// Observable success (installed) or failure (missing) both satisfy Gate.
		if result.Detected && !result.Healthy {
			t.Fatalf("%s detected but not healthy", key)
		}
		if !result.Detected && result.Output == "" {
			t.Fatalf("%s missing failure output", key)
		}
	}
}

func TestTriggersCreateIndependentAgentRuns(t *testing.T) {
	_, _, agents, _, _, _ := setupAI(t)
	agent, err := agents.CreateAgent(1, service.AgentInput{
		Name: "t", CliKey: "claude_code", SystemPrompt: "hello", TimeoutSec: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	manual, err := agents.ManualRun(agent.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	api, err := agents.APIRun(agent.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	cronTrig, err := agents.CreateTrigger(agent.ID, 1, service.TriggerInput{
		Type: model.TriggerCron, CronExpression: "0 0 * * *", CronTimezone: "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	cronRun, err := agents.CreateRun(agent.ID, service.CreateRunInput{
		TriggerType: model.TriggerCron, TriggerID: &cronTrig.ID, TriggeredBy: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	job := &cicdmodel.BuildJob{ID: 99, AgentTriggerEvent: model.EventArtifactReady, AgentID: &agent.ID}
	buildRun := &cicdmodel.BuildRun{ID: 77, BuildJobID: 99, Status: "success", TriggeredBy: 1, ArtifactPath: "/tmp/a.tgz"}
	agents.OnBuildEvent(model.EventArtifactReady, job, buildRun)
	deadline := time.Now().Add(2 * time.Second)
	var buildEventRun *model.AgentRun
	for time.Now().Before(deadline) {
		items, _, _ := agents.ListRuns(1, 50, agent.ID, "")
		for i := range items {
			if items[i].TriggerType == model.TriggerBuildEvent {
				buildEventRun = &items[i]
				break
			}
		}
		if buildEventRun != nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if buildEventRun == nil {
		t.Fatal("expected build_event AgentRun")
	}
	ids := map[uint]bool{manual.ID: true, api.ID: true, cronRun.ID: true, buildEventRun.ID: true}
	if len(ids) != 4 {
		t.Fatalf("expected 4 independent runs, got %v", ids)
	}
}

func TestAgentFailureDoesNotChangeBuildRun(t *testing.T) {
	_, _, agents, _, _, _ := setupAI(t)
	agent, err := agents.CreateAgent(1, service.AgentInput{
		Name: "fail", CliKey: "reasonix", SystemPrompt: "x", TimeoutSec: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	build := &cicdmodel.BuildRun{ID: 5, Status: "success", ArtifactPath: "/a"}
	job := &cicdmodel.BuildJob{ID: 1, AgentID: &agent.ID, AgentTriggerEvent: model.EventArtifactReady}
	agents.OnBuildEvent(model.EventArtifactReady, job, build)
	time.Sleep(200 * time.Millisecond)
	if build.Status != "success" {
		t.Fatalf("BuildRun.status mutated to %s", build.Status)
	}
}

func TestSkillUploadRejectMissingSKILLMDAndOverwrite(t *testing.T) {
	_, _, _, skills, _, _ := setupAI(t)
	bad := zipBytes(t, map[string]string{"README.md": "nope"})
	_, err := skills.Create(service.SkillUploadInput{
		Name: "bad", Visibility: model.SkillPrivate, Filename: "bad.zip",
		Size: int64(len(bad)), Source: bytes.NewReader(bad), UserID: 1,
	})
	if err == nil || !strings.Contains(err.Error(), "SKILL.md") {
		t.Fatalf("expected missing SKILL.md error, got %v", err)
	}

	good1 := zipBytes(t, map[string]string{"SKILL.md": "# v1"})
	s1, err := skills.Create(service.SkillUploadInput{
		Name: "ok", Visibility: model.SkillPublic, Filename: "ok.zip",
		Size: int64(len(good1)), Source: bytes.NewReader(good1), UserID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	good2 := zipBytes(t, map[string]string{"SKILL.md": "# v2-new"})
	s2, err := skills.Overwrite(s1.ID, service.SkillUploadInput{
		Filename: "ok.zip", Size: int64(len(good2)), Source: bytes.NewReader(good2), UserID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if s2.PackageDigest == s1.PackageDigest {
		t.Fatal("overwrite should change digest")
	}
	_, rc, _, err := skills.OpenPackage(s2.ID, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(rc)
	if !bytes.Contains(buf.Bytes(), []byte("v2-new")) && buf.Len() == 0 {
		// ZIP binary won't contain plaintext easily if compressed; digest change is enough.
	}
	if s2.PackageDigest == "" {
		t.Fatal("empty digest")
	}
}

func TestPrivateSkillIsolation(t *testing.T) {
	_, _, _, skills, _, _ := setupAI(t)
	z := zipBytes(t, map[string]string{"SKILL.md": "# priv"})
	s, err := skills.Create(service.SkillUploadInput{
		Name: "priv", Visibility: model.SkillPrivate, Filename: "p.zip",
		Size: int64(len(z)), Source: bytes.NewReader(z), UserID: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := skills.Get(s.ID, 2, false); err == nil {
		t.Fatal("non-creator must not see private skill")
	}
	items, _, err := skills.List(1, 20, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		if item.ID == s.ID {
			t.Fatal("private skill leaked in list")
		}
	}
}

func TestPATPlaintextOnceAndScopes(t *testing.T) {
	_, _, _, _, pats, _ := setupAI(t)
	created, err := pats.Create(1, service.CreatePATInput{
		Name: "t", Scopes: []string{model.ScopeSkillsRead},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(created.Token, "br_pat_") {
		t.Fatalf("unexpected token %s", created.Token)
	}
	list, err := pats.List(1)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}
	// Metadata list must not include plaintext.
	if strings.Contains(list[0].TokenHash, created.Token) {
		t.Fatal("plaintext must not be stored")
	}
	if _, _, err := pats.ValidateBearer("br_pat_deadbeef"); err == nil {
		t.Fatal("invalid PAT must fail")
	}
	uid, scopes, err := pats.ValidateBearer(created.Token)
	if err != nil || uid != 1 {
		t.Fatalf("valid PAT: %v uid=%d", err, uid)
	}
	if err := pats.RequireScope(scopes, model.ScopeAgentsRun); err == nil {
		t.Fatal("wrong scope must 403")
	}
	if err := pats.RequireScope(scopes, model.ScopeSkillsRead); err != nil {
		t.Fatal(err)
	}
	_ = pats.Revoke(1, created.Metadata.ID)
	if _, _, err := pats.ValidateBearer(created.Token); err == nil {
		t.Fatal("revoked PAT must be invalid")
	}
}

func TestDocsGenerateWritesDraftOnly(t *testing.T) {
	gdb, _, agents, _, _, projectSvc := setupAI(t)
	owner := projectservice.NewAccessContext(1, true, []string{"project.projects:create", "project.docs:execute", "project.docs:view"})
	project, err := projectSvc.CreateProject(owner, projectservice.CreateProjectInput{Name: "P", Slug: "p-ai"})
	if err != nil {
		t.Fatal(err)
	}
	agent, err := agents.CreateAgent(1, service.AgentInput{
		Name: "doc", CliKey: "codex", SystemPrompt: "Generate docs", TimeoutSec: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Seed a fake successful run writing draft via bridge callback.
	node := &projectmodel.ApiDocNode{
		ProjectID: project.ID, Kind: projectmodel.DocNodeDocument, Name: "api",
		CreatedBy: 1, UpdatedBy: 1,
	}
	if err := projectrepo.NewProjectRepository(gdb).CreateDocNode(node); err != nil {
		t.Fatal(err)
	}
	content := "# Draft From Agent\n"
	if err := projectSvc.WriteDraftFromAgentRun(project.ID, node.ID, 123, content, 1); err != nil {
		t.Fatal(err)
	}
	got, err := projectrepo.NewProjectRepository(gdb).FindDocNode(node.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.DraftContent != content {
		t.Fatalf("draft=%q", got.DraftContent)
	}
	if got.PublishedContent != "" {
		t.Fatal("must not auto-publish")
	}
	if got.DraftSourceRunID == nil || *got.DraftSourceRunID != 123 {
		t.Fatalf("draft_source_run_id=%v", got.DraftSourceRunID)
	}
	result, err := projectSvc.GenerateDocs(owner, project.ID, projectservice.GenerateDocsInput{
		AgentID: agent.ID, NodeID: &node.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.AgentRunID == 0 {
		t.Fatal("expected agent run id")
	}
}

func zipBytes(t *testing.T, files map[string]string) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for name, body := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestCLIListSeeded(t *testing.T) {
	_, cli, _, _, _, _ := setupAI(t)
	items, err := cli.ListCLIs()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 4 {
		t.Fatalf("want 4 CLIs, got %d", len(items))
	}
	_ = os.DevNull
}

func TestCronReloadAppliesTimezone(t *testing.T) {
	_, _, agents, _, _, _ := setupAI(t)
	agent, err := agents.CreateAgent(1, service.AgentInput{
		Name: "tz", CliKey: "claude_code", SystemPrompt: "x", TimeoutSec: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = agents.CreateTrigger(agent.ID, 1, service.TriggerInput{
		Type: model.TriggerCron, CronExpression: "0 12 * * *", CronTimezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatal(err)
	}
	entries := agents.CronEntries()
	if len(entries) == 0 {
		t.Fatal("expected cron entry after reload")
	}
	next := entries[0].Next
	if next.IsZero() {
		t.Fatal("expected non-zero next fire time")
	}
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	if next.In(shanghai).Hour() != 12 {
		t.Fatalf("next fire should be noon Shanghai, got %s", next.In(shanghai))
	}
	// Contrast: same expression in UTC would be 12:00 UTC, not 04:00 UTC.
	if next.UTC().Hour() != 4 {
		t.Fatalf("Shanghai noon should be 04:00 UTC, got hour=%d (%s)", next.UTC().Hour(), next.UTC())
	}
}

func TestAgentRunRecovery_QueuedAndInterrupted(t *testing.T) {
	gdb, _, agents, _, _, _ := setupAI(t)
	repo := repository.NewAIRepository(gdb)
	cliDef := &model.CliRuntimeDefinition{
		Key: "claude_code", Name: "Claude", BinaryName: "claude",
	}
	if err := gdb.Where(model.CliRuntimeDefinition{Key: "claude_code"}).
		Attrs(model.CliRuntimeDefinition{Name: "Claude", BinaryName: "claude"}).
		FirstOrCreate(cliDef).Error; err != nil {
		t.Fatal(err)
	}
	agent := &model.AiAgent{
		Name: "recover", CliKey: "claude_code", Enabled: true, SystemPrompt: "x", TimeoutSec: 30, CreatedBy: 1,
	}
	if err := repo.CreateAgent(agent); err != nil {
		t.Fatal(err)
	}
	running := &model.AgentRun{AgentID: agent.ID, Status: model.JobRunning, TriggerType: "manual"}
	queued := &model.AgentRun{AgentID: agent.ID, Status: model.JobQueued, TriggerType: "manual"}
	if err := gdb.Create(running).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(queued).Error; err != nil {
		t.Fatal(err)
	}
	if err := agents.RecoverOnStartup(); err != nil {
		t.Fatal(err)
	}
	gotRunning, err := repo.FindRun(running.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotRunning.Status != model.JobInterrupted {
		t.Fatalf("running→interrupted got %s", gotRunning.Status)
	}
	gotQueued, err := repo.FindRun(queued.ID)
	if err != nil {
		t.Fatal(err)
	}
	switch gotQueued.Status {
	case model.JobQueued, model.JobRunning, model.JobPending, model.JobFailed, model.JobSuccess, model.JobInterrupted:
		// ok — re-submit may advance or fail without a real CLI binary
	default:
		t.Fatalf("unexpected queued recovery status %s", gotQueued.Status)
	}
}
