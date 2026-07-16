package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"bedrock/internal/ai/model"
	"bedrock/internal/ai/repository"
)

type AuditWriter interface {
	Write(userID uint, username, action, resourceType, resourceID, details, ip string) error
}

type CLIService struct {
	repo  *repository.AIRepository
	audit AuditWriter

	jobs    chan uint
	stop    chan struct{}
	wg      sync.WaitGroup
	startMu sync.Mutex
	started bool
}

func NewCLIService(repo *repository.AIRepository, audit ...AuditWriter) *CLIService {
	svc := &CLIService{
		repo: repo,
		jobs: make(chan uint, 128),
		stop: make(chan struct{}),
	}
	if len(audit) > 0 {
		svc.audit = audit[0]
	}
	return svc
}

func (s *CLIService) Start() {
	s.startMu.Lock()
	defer s.startMu.Unlock()
	if s.started {
		return
	}
	s.started = true
	s.wg.Add(1)
	go s.worker()
}

func (s *CLIService) Shutdown() {
	s.startMu.Lock()
	if !s.started {
		s.startMu.Unlock()
		return
	}
	s.started = false
	close(s.stop)
	s.startMu.Unlock()
	s.wg.Wait()
}

func (s *CLIService) RecoverOnStartup() error {
	if _, err := s.repo.MarkRunningInstallJobsInterrupted(); err != nil {
		return err
	}
	queued, err := s.repo.ListInstallJobsByStatuses(model.JobQueued)
	if err != nil {
		return err
	}
	for _, job := range queued {
		if err := s.submit(job.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *CLIService) ListCLIs() ([]model.CliRuntimeDefinition, error) {
	items, err := s.repo.ListCLIs()
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].RiskNotice = model.RiskNoticeSameUID
	}
	return items, nil
}

type DetectResult struct {
	Detected bool   `json:"detected"`
	Output   string `json:"output"`
	Path     string `json:"path"`
	Version  string `json:"version"`
	Healthy  bool   `json:"healthy"`
	RiskNotice string `json:"risk_notice"`
}

func (s *CLIService) Detect(key string) (*DetectResult, error) {
	cli, err := s.repo.FindCLIByKey(key)
	if err != nil {
		return nil, err
	}
	cmd := strings.TrimSpace(cli.DetectCommand)
	if cmd == "" {
		cmd = "command -v " + cli.BinaryName
	}
	output, runErr := executeShell(context.Background(), cmd)
	result := &DetectResult{
		Detected:   runErr == nil,
		Output:     strings.TrimSpace(output),
		RiskNotice: model.RiskNoticeSameUID,
	}
	if path, lookErr := exec.LookPath(cli.BinaryName); lookErr == nil {
		result.Path = path
		cli.InstalledPath = path
	}
	if runErr == nil {
		cli.InstallStatus = "installed"
		cli.Healthy = true
		result.Healthy = true
		result.Version = firstLine(output)
		cli.InstalledVersion = result.Version
	} else {
		cli.InstallStatus = "missing"
		cli.Healthy = false
		result.Healthy = false
		if result.Output == "" {
			result.Output = runErr.Error()
		}
	}
	_ = s.repo.UpdateCLI(cli)
	return result, nil
}

type JobInput struct {
	Version string `json:"version"`
}

func (s *CLIService) Enqueue(key, operation string, input JobInput, createdBy uint) (*model.CliInstallJob, error) {
	cli, err := s.repo.FindCLIByKey(key)
	if err != nil {
		return nil, err
	}
	template := templateFor(cli, operation)
	if template == "" {
		return nil, errors.New("该 CLI 未配置此操作命令")
	}
	job := &model.CliInstallJob{
		CliKey: key, Operation: operation, RequestedVersion: strings.TrimSpace(input.Version),
		Status: model.JobQueued, CreatedBy: createdBy,
	}
	if err := s.repo.CreateInstallJob(job); err != nil {
		return nil, err
	}
	s.auditJob(job, "cli_job_enqueued")
	if err := s.submit(job.ID); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *CLIService) ListSources(cliKey string) ([]model.CliInstallSource, error) {
	return s.repo.ListSources(cliKey)
}

type SourceInput struct {
	CliKey   string `json:"cli_key"`
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

func (s *CLIService) CreateSource(input SourceInput) (*model.CliInstallSource, error) {
	if strings.TrimSpace(input.CliKey) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.BaseURL) == "" {
		return nil, errors.New("cli_key、名称和地址不能为空")
	}
	if _, err := s.repo.FindCLIByKey(input.CliKey); err != nil {
		return nil, errors.New("CLI 不存在")
	}
	item := &model.CliInstallSource{
		CliKey: input.CliKey, Name: input.Name, BaseURL: input.BaseURL,
		Priority: input.Priority, Enabled: input.Enabled,
	}
	if err := s.repo.CreateSource(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *CLIService) UpdateSource(id uint, input SourceInput) (*model.CliInstallSource, error) {
	item, err := s.repo.FindSource(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Name) != "" {
		item.Name = input.Name
	}
	if strings.TrimSpace(input.BaseURL) != "" {
		item.BaseURL = input.BaseURL
	}
	item.Priority = input.Priority
	item.Enabled = input.Enabled
	if err := s.repo.UpdateSource(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *CLIService) DeleteSource(id uint) error {
	return s.repo.DeleteSource(id)
}

func (s *CLIService) ListJobs(page, pageSize int, cliKey, status string) ([]model.CliInstallJob, int64, error) {
	return s.repo.ListInstallJobs(page, pageSize, cliKey, status)
}

func (s *CLIService) GetJob(id uint) (*model.CliInstallJob, error) {
	return s.repo.FindInstallJob(id)
}

func (s *CLIService) JobLogs(id uint) (string, error) {
	job, err := s.repo.FindInstallJob(id)
	if err != nil {
		return "", err
	}
	return job.LogText, nil
}

func (s *CLIService) Retry(id uint, createdBy uint) (*model.CliInstallJob, error) {
	old, err := s.repo.FindInstallJob(id)
	if err != nil {
		return nil, err
	}
	if old.Status == model.JobQueued || old.Status == model.JobRunning {
		return nil, errors.New("任务仍在执行")
	}
	return s.Enqueue(old.CliKey, old.Operation, JobInput{Version: old.RequestedVersion}, createdBy)
}

func (s *CLIService) submit(id uint) error {
	s.startMu.Lock()
	started := s.started
	s.startMu.Unlock()
	if !started {
		return errors.New("CLI 任务调度器未启动")
	}
	select {
	case s.jobs <- id:
		return nil
	default:
		go func() { s.jobs <- id }()
		return nil
	}
}

func (s *CLIService) worker() {
	defer s.wg.Done()
	for {
		select {
		case <-s.stop:
			return
		case id := <-s.jobs:
			s.ExecuteJob(context.Background(), id)
		}
	}
}

func (s *CLIService) ExecuteJob(ctx context.Context, id uint) {
	job, err := s.repo.FindInstallJob(id)
	if err != nil || job.Status != model.JobQueued {
		return
	}
	now := time.Now().UTC()
	job.Status, job.StartedAt = model.JobRunning, &now
	job.LogText += fmt.Sprintf("%s job started: %s (same UID as Bedrock; no sandbox)\n", now.Format(time.RFC3339), job.Operation)
	if err := s.repo.UpdateInstallJob(job); err != nil {
		return
	}
	cli, err := s.repo.FindCLIByKey(job.CliKey)
	if err != nil {
		s.failJob(job, err)
		return
	}
	template := templateFor(cli, job.Operation)
	if template == "" {
		s.failJob(job, errors.New("未配置命令模板"))
		return
	}
	if needsSource(job.Operation, template) {
		sources, err := s.repo.ListEnabledSources(job.CliKey)
		if err != nil {
			s.failJob(job, err)
			return
		}
		if len(sources) == 0 {
			s.failJob(job, errors.New("没有可用安装源"))
			return
		}
		for i, source := range sources {
			command := renderCLICommand(template, job.RequestedVersion, source.BaseURL)
			job.CommandSnapshot = command
			job.LogText += fmt.Sprintf("trying source %q (priority %d)\n", source.Name, source.Priority)
			output, runErr := executeShell(ctx, command)
			job.LogText += output
			if runErr == nil {
				sid := source.ID
				job.SourceID = &sid
				job.LogText += fmt.Sprintf("source %q succeeded\n", source.Name)
				if i > 0 {
					job.LogText += "multi-source fallback succeeded after earlier failures\n"
				}
				s.succeedJob(job, cli)
				return
			}
			job.LogText += fmt.Sprintf("source %q failed: %v\n", source.Name, runErr)
		}
		s.failJob(job, errors.New("所有安装源均失败"))
		return
	}
	command := renderCLICommand(template, job.RequestedVersion, "")
	job.CommandSnapshot = command
	output, runErr := executeShell(ctx, command)
	job.LogText += output
	if runErr != nil {
		s.failJob(job, runErr)
		return
	}
	s.succeedJob(job, cli)
}

func (s *CLIService) succeedJob(job *model.CliInstallJob, cli *model.CliRuntimeDefinition) {
	finished := time.Now().UTC()
	job.Status = model.JobSuccess
	job.FinishedAt = &finished
	_ = s.repo.UpdateInstallJob(job)
	cli.InstallStatus = "installed"
	if path, err := exec.LookPath(cli.BinaryName); err == nil {
		cli.InstalledPath = path
	}
	cli.Healthy = true
	_ = s.repo.UpdateCLI(cli)
	s.auditJob(job, "cli_job_success")
}

func (s *CLIService) failJob(job *model.CliInstallJob, err error) {
	finished := time.Now().UTC()
	job.Status = model.JobFailed
	job.FinishedAt = &finished
	job.ErrorMessage = err.Error()
	job.LogText += "error: " + err.Error() + "\n"
	_ = s.repo.UpdateInstallJob(job)
	s.auditJob(job, "cli_job_failed")
}

func (s *CLIService) auditJob(job *model.CliInstallJob, action string) {
	if s.audit == nil {
		return
	}
	_ = s.audit.Write(job.CreatedBy, "", action, "cli_install_job", fmt.Sprintf("%d", job.ID),
		fmt.Sprintf("cli=%s op=%s status=%s", job.CliKey, job.Operation, job.Status), "")
}

func templateFor(cli *model.CliRuntimeDefinition, operation string) string {
	switch operation {
	case "install":
		return cli.InstallTemplate
	case "upgrade":
		return cli.UpgradeTemplate
	case "uninstall":
		return cli.UninstallTemplate
	default:
		return ""
	}
}

func needsSource(operation, template string) bool {
	return (operation == "install" || operation == "upgrade") && strings.Contains(template, "{{base_url}}")
}

func renderCLICommand(template, version, baseURL string) string {
	out := strings.ReplaceAll(template, "{{version}}", shellQuote(version))
	out = strings.ReplaceAll(out, "{{base_url}}", shellQuote(baseURL))
	return out
}

func shellQuote(s string) string {
	if s == "" {
		return ""
	}
	return strings.ReplaceAll(s, `'`, `'\''`)
}

func executeShell(ctx context.Context, command string) (string, error) {
	if runtime.GOOS == "windows" {
		cmd := exec.CommandContext(ctx, "cmd", "/C", command)
		var buf bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		return buf.String(), err
	}
	cmd := exec.CommandContext(ctx, "bash", "-lc", command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

// ResolveBinary returns absolute path for a CLI binary if installed.
func ResolveBinary(cli *model.CliRuntimeDefinition) (string, error) {
	if cli.InstalledPath != "" {
		if _, err := os.Stat(cli.InstalledPath); err == nil {
			return cli.InstalledPath, nil
		}
	}
	return exec.LookPath(cli.BinaryName)
}

// BuildRuntimeEnv injects API base / env templates without overwriting CLI login state files.
func BuildRuntimeEnv(cli *model.CliRuntimeDefinition, apiBase string, extra map[string]string) []string {
	env := os.Environ()
	if apiBase != "" && cli.APIBaseEnv != "" {
		env = append(env, cli.APIBaseEnv+"="+apiBase)
	}
	for k, v := range extra {
		if strings.TrimSpace(k) == "" {
			continue
		}
		env = append(env, k+"="+v)
	}
	// Ensure PATH still finds the binary.
	if cli.InstalledPath != "" {
		dir := filepath.Dir(cli.InstalledPath)
		env = append(env, "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	return env
}
