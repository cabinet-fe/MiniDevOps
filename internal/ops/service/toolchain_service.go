package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"bedrock/internal/ops/model"
	"bedrock/internal/ops/repository"
)

var (
	ErrBuiltinImmutable = errors.New("内置工具链不可删除")
	ErrMissingTemplate  = errors.New("该工具链未配置此操作命令")
	ErrInvalidOperation = errors.New("不支持的工具链操作")
	secretFlagName      = `(?:token|password|secret|api[_-]?(?:key|token)|access[_-]?token|auth(?:orization)?(?:[_-]?token)?|credential(?:s)?|client[_-]?secret|private[_-]?key)`
	secretAssignment    = regexp.MustCompile(`(?i)((?:^|[^a-z0-9])` + secretFlagName + `\s*=\s*)[^\s&'"]+`)
	secretSpacedFlag    = regexp.MustCompile(`(?i)((?:^|[^a-z0-9_-])(?:--` + secretFlagName + `|-t)\s+)(?:"[^"]*"|'[^']*'|[^\s&'"]+)`)
	secretBearer        = regexp.MustCompile(`(?i)(authorization\s*[:=]\s*bearer\s+)[^\s&'"]+`)
	secretURLUserInfo   = regexp.MustCompile(`(?i)(https?://[^/\s:@]+:)[^@/\s]+@`)
)

// AuditWriter is the small audit boundary used by asynchronous operations.
// Worker events do not have an HTTP request, so username and IP are empty.
type AuditWriter interface {
	Write(userID uint, username, action, resourceType, resourceID, details, ip string) error
}

type ToolchainInput struct {
	Name              string `json:"name"`
	Executable        string `json:"executable"`
	Description       string `json:"description"`
	DetectCommand     string `json:"detect_command"`
	InstallTemplate   string `json:"install_template"`
	UpgradeTemplate   string `json:"upgrade_template"`
	UninstallTemplate string `json:"uninstall_template"`
	VersionsCommand   string `json:"versions_command"`
	SwitchTemplate    string `json:"switch_template"`
	DefaultVersion    string `json:"default_version"`
}

type SourceInput struct {
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	Priority int    `json:"priority"`
	Enabled  bool   `json:"enabled"`
}

type JobInput struct {
	Version string `json:"version"`
}

type DetectResult struct {
	Detected bool   `json:"detected"`
	Output   string `json:"output"`
}

type ToolchainService struct {
	repo  *repository.OpsRepository
	audit AuditWriter

	jobs    chan uint
	stop    chan struct{}
	wg      sync.WaitGroup
	startMu sync.Mutex
	started bool
}

func NewToolchainService(repo *repository.OpsRepository, audit ...AuditWriter) *ToolchainService {
	service := &ToolchainService{
		repo: repo,
		jobs: make(chan uint, 128),
		stop: make(chan struct{}),
	}
	if len(audit) > 0 {
		service.audit = audit[0]
	}
	return service
}

func (s *ToolchainService) Start() {
	s.startMu.Lock()
	defer s.startMu.Unlock()
	if s.started {
		return
	}
	s.started = true
	s.wg.Add(1)
	go s.worker()
}

func (s *ToolchainService) Shutdown() {
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

// RecoverOnStartup implements DESIGN D18: active work is marked interrupted,
// while never-started queued jobs are submitted again.
func (s *ToolchainService) RecoverOnStartup() error {
	if _, err := s.repo.MarkRunningInterrupted(); err != nil {
		return err
	}
	queued, err := s.repo.ListJobsByStatuses(model.JobQueued)
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

func (s *ToolchainService) ListToolchains() ([]model.ToolchainDefinition, error) {
	return s.repo.ListToolchains()
}

func (s *ToolchainService) CreateCustom(input ToolchainInput, createdBy uint) (*model.ToolchainDefinition, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Executable) == "" {
		return nil, errors.New("名称和可执行文件不能为空")
	}
	item := &model.ToolchainDefinition{
		Name: input.Name, Kind: model.ToolchainCustom, Executable: input.Executable,
		Description: input.Description, DetectCommand: input.DetectCommand,
		InstallTemplate: input.InstallTemplate, UpgradeTemplate: input.UpgradeTemplate,
		UninstallTemplate: input.UninstallTemplate, VersionsCommand: input.VersionsCommand,
		SwitchTemplate: input.SwitchTemplate, DefaultVersion: input.DefaultVersion, CreatedBy: createdBy,
	}
	if err := s.repo.CreateToolchain(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ToolchainService) UpdateCustom(id uint, input ToolchainInput) (*model.ToolchainDefinition, error) {
	item, err := s.repo.FindToolchain(id)
	if err != nil {
		return nil, err
	}
	if item.Kind != model.ToolchainCustom {
		return nil, ErrBuiltinImmutable
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Executable) == "" {
		return nil, errors.New("名称和可执行文件不能为空")
	}
	item.Name, item.Executable, item.Description = input.Name, input.Executable, input.Description
	item.DetectCommand, item.InstallTemplate, item.UpgradeTemplate = input.DetectCommand, input.InstallTemplate, input.UpgradeTemplate
	item.UninstallTemplate, item.VersionsCommand, item.SwitchTemplate = input.UninstallTemplate, input.VersionsCommand, input.SwitchTemplate
	item.DefaultVersion = input.DefaultVersion
	if err := s.repo.UpdateToolchain(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ToolchainService) DeleteCustom(id uint) error {
	item, err := s.repo.FindToolchain(id)
	if err != nil {
		return err
	}
	if item.Kind != model.ToolchainCustom {
		return ErrBuiltinImmutable
	}
	return s.repo.DeleteToolchain(id)
}

func (s *ToolchainService) Detect(id uint) (*DetectResult, error) {
	item, err := s.repo.FindToolchain(id)
	if err != nil {
		return nil, err
	}
	command := item.DetectCommand
	if command == "" {
		command = item.Executable + " --version"
	}
	output, err := executeCommand(context.Background(), command)
	return &DetectResult{Detected: err == nil, Output: redact(output)}, nil
}

func (s *ToolchainService) Enqueue(id uint, operation string, input JobInput, createdBy uint) (*model.ToolchainInstallJob, error) {
	item, err := s.repo.FindToolchain(id)
	if err != nil {
		return nil, err
	}
	if _, err := commandTemplate(item, operation); err != nil {
		return nil, err
	}
	job := &model.ToolchainInstallJob{
		ToolchainID: id, Operation: operation, RequestedVersion: strings.TrimSpace(input.Version),
		Status: model.JobQueued, CreatedBy: createdBy,
	}
	if err := s.repo.CreateJob(job); err != nil {
		return nil, err
	}
	s.auditJob(job, "toolchain_job_enqueued", "", false)
	if err := s.submit(job.ID); err != nil {
		return nil, err
	}
	return job, nil
}

func (s *ToolchainService) Retry(id uint, createdBy uint) (*model.ToolchainInstallJob, error) {
	old, err := s.repo.FindJob(id)
	if err != nil {
		return nil, err
	}
	if old.Status == model.JobQueued || old.Status == model.JobRunning {
		return nil, errors.New("任务仍在执行")
	}
	return s.Enqueue(old.ToolchainID, old.Operation, JobInput{Version: old.RequestedVersion}, createdBy)
}

func (s *ToolchainService) ListSources() ([]model.InstallSource, error) {
	return s.repo.ListSources()
}

func (s *ToolchainService) CreateSource(input SourceInput) (*model.InstallSource, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.BaseURL) == "" {
		return nil, errors.New("名称和地址不能为空")
	}
	item := &model.InstallSource{Name: input.Name, BaseURL: input.BaseURL, Priority: input.Priority, Enabled: input.Enabled}
	if err := s.repo.CreateSource(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ToolchainService) UpdateSource(id uint, input SourceInput) (*model.InstallSource, error) {
	item, err := s.repo.FindSource(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.BaseURL) == "" {
		return nil, errors.New("名称和地址不能为空")
	}
	item.Name, item.BaseURL, item.Priority, item.Enabled = input.Name, input.BaseURL, input.Priority, input.Enabled
	if err := s.repo.UpdateSource(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *ToolchainService) DeleteSource(id uint) error {
	return s.repo.DeleteSource(id)
}

func (s *ToolchainService) PingSource(id uint) (bool, string, error) {
	source, err := s.repo.FindSource(id)
	if err != nil {
		return false, "", err
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodHead, source.BaseURL, nil)
	if err != nil {
		return false, "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return false, err.Error(), nil
	}
	defer response.Body.Close()
	return response.StatusCode >= 200 && response.StatusCode < 400, response.Status, nil
}

func (s *ToolchainService) ListJobs(page, pageSize int, status string) ([]model.ToolchainInstallJob, int64, error) {
	items, total, err := s.repo.ListJobs(page, pageSize, status)
	for i := range items {
		sanitizeJob(&items[i])
	}
	return items, total, err
}

func (s *ToolchainService) GetJob(id uint) (*model.ToolchainInstallJob, error) {
	job, err := s.repo.FindJob(id)
	if err != nil {
		return nil, err
	}
	sanitizeJob(job)
	return job, nil
}

func (s *ToolchainService) JobLogs(id uint) (string, error) {
	job, err := s.repo.FindJob(id)
	if err != nil {
		return "", err
	}
	return redact(job.LogText), nil
}

func (s *ToolchainService) submit(id uint) error {
	s.startMu.Lock()
	started := s.started
	s.startMu.Unlock()
	if !started {
		return errors.New("工具链任务调度器未启动")
	}
	select {
	case s.jobs <- id:
		return nil
	default:
		s.jobs <- id
		return nil
	}
}

func (s *ToolchainService) worker() {
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

// ExecuteJob is exported for deterministic integration tests and worker use.
func (s *ToolchainService) ExecuteJob(ctx context.Context, id uint) {
	job, err := s.repo.FindJob(id)
	if err != nil || job.Status != model.JobQueued {
		return
	}
	now := time.Now().UTC()
	job.Status, job.StartedAt = model.JobRunning, &now
	job.LogText += fmt.Sprintf("%s job started: %s\n", now.Format(time.RFC3339), job.Operation)
	if err := s.repo.UpdateJob(job); err != nil {
		return
	}
	s.auditJob(job, "toolchain_job_started", "", false)

	toolchain, err := s.repo.FindToolchain(job.ToolchainID)
	if err != nil {
		s.fail(job, err, "", false)
		return
	}
	template, err := commandTemplate(toolchain, job.Operation)
	if err != nil {
		s.fail(job, err, "", false)
		return
	}

	if needsSource(job.Operation, template) {
		sources, err := s.repo.ListEnabledSources()
		if err != nil {
			s.fail(job, err, "", false)
			return
		}
		if len(sources) == 0 {
			s.fail(job, errors.New("没有可用安装源"), "", false)
			return
		}
		for index, source := range sources {
			command := renderCommand(template, *toolchain, job.RequestedVersion, source.BaseURL)
			job.CommandSnapshot = redactSourceURL(command, source.BaseURL)
			job.LogText += fmt.Sprintf("trying source %q (priority %d)\n", source.Name, source.Priority)
			output, runErr := executeCommand(ctx, command)
			job.LogText += redactSourceURL(output, source.BaseURL)
			if runErr == nil {
				sourceID := source.ID
				job.SourceID = &sourceID
				job.LogText += fmt.Sprintf("source %q succeeded\n", source.Name)
				s.succeed(job, toolchain, source.Name, index > 0)
				return
			}
			job.LogText += fmt.Sprintf("source %q failed: %s\n", source.Name, redact(runErr.Error()))
		}
		s.fail(job, errors.New("所有安装源均失败"), "", len(sources) > 1)
		return
	}

	command := renderCommand(template, *toolchain, job.RequestedVersion, "")
	job.CommandSnapshot = redact(command)
	output, runErr := executeCommand(ctx, command)
	job.LogText += redact(output)
	if runErr != nil {
		s.fail(job, runErr, "", false)
		return
	}
	s.succeed(job, toolchain, "", false)
}

func (s *ToolchainService) succeed(
	job *model.ToolchainInstallJob,
	toolchain *model.ToolchainDefinition,
	sourceName string,
	sourceFallback bool,
) {
	now := time.Now().UTC()
	job.Status, job.FinishedAt, job.ErrorMessage = model.JobSuccess, &now, ""
	job.LogText += fmt.Sprintf("%s job succeeded\n", now.Format(time.RFC3339))
	job.CommandSnapshot, job.LogText = redact(job.CommandSnapshot), redact(job.LogText)
	if err := s.repo.UpdateJob(job); err != nil {
		return
	}
	if job.Operation == "switch" && job.RequestedVersion != "" {
		toolchain.DefaultVersion = job.RequestedVersion
		_ = s.repo.UpdateToolchain(toolchain)
	}
	s.auditJob(job, "toolchain_job_completed", sourceName, sourceFallback)
}

func (s *ToolchainService) fail(
	job *model.ToolchainInstallJob,
	err error,
	sourceName string,
	sourceFallback bool,
) {
	now := time.Now().UTC()
	job.Status, job.FinishedAt, job.ErrorMessage = model.JobFailed, &now, redact(err.Error())
	job.LogText += fmt.Sprintf("%s job failed: %s\n", now.Format(time.RFC3339), job.ErrorMessage)
	job.CommandSnapshot, job.LogText = redact(job.CommandSnapshot), redact(job.LogText)
	if err := s.repo.UpdateJob(job); err != nil {
		return
	}
	s.auditJob(job, "toolchain_job_completed", sourceName, sourceFallback)
}

func commandTemplate(item *model.ToolchainDefinition, operation string) (string, error) {
	var command string
	switch operation {
	case "install":
		command = item.InstallTemplate
	case "upgrade":
		command = item.UpgradeTemplate
	case "uninstall":
		command = item.UninstallTemplate
	case "switch":
		command = item.SwitchTemplate
	default:
		return "", ErrInvalidOperation
	}
	if strings.TrimSpace(command) == "" {
		return "", ErrMissingTemplate
	}
	return command, nil
}

func needsSource(operation, template string) bool {
	return (operation == "install" || operation == "upgrade") && strings.Contains(template, "{{source_url}}")
}

func renderCommand(template string, toolchain model.ToolchainDefinition, version, sourceURL string) string {
	replacer := strings.NewReplacer(
		"{{toolchain}}", toolchain.Name,
		"{{executable}}", toolchain.Executable,
		"{{version}}", version,
		"{{source_url}}", sourceURL,
	)
	return replacer.Replace(template)
}

func executeCommand(ctx context.Context, command string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", command)
	}
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	return output.String(), err
}

func redact(value string) string {
	value = secretAssignment.ReplaceAllString(value, "$1[REDACTED]")
	value = secretSpacedFlag.ReplaceAllString(value, "$1[REDACTED]")
	value = secretBearer.ReplaceAllString(value, "$1[REDACTED]")
	return secretURLUserInfo.ReplaceAllString(value, "$1[REDACTED]@")
}

func sanitizeJob(job *model.ToolchainInstallJob) {
	job.CommandSnapshot = redact(job.CommandSnapshot)
	job.LogText = redact(job.LogText)
	job.ErrorMessage = redact(job.ErrorMessage)
	job.RequestedVersion = redact(job.RequestedVersion)
	if job.Source != nil {
		job.Source.BaseURL = "[REDACTED]"
	}
	job.ToolchainDefinition.DetectCommand = redact(job.ToolchainDefinition.DetectCommand)
	job.ToolchainDefinition.InstallTemplate = redact(job.ToolchainDefinition.InstallTemplate)
	job.ToolchainDefinition.UpgradeTemplate = redact(job.ToolchainDefinition.UpgradeTemplate)
	job.ToolchainDefinition.UninstallTemplate = redact(job.ToolchainDefinition.UninstallTemplate)
	job.ToolchainDefinition.VersionsCommand = redact(job.ToolchainDefinition.VersionsCommand)
	job.ToolchainDefinition.SwitchTemplate = redact(job.ToolchainDefinition.SwitchTemplate)
}

func redactSourceURL(value, sourceURL string) string {
	if sourceURL != "" {
		value = strings.ReplaceAll(value, sourceURL, "[SOURCE_URL_REDACTED]")
	}
	return redact(value)
}

func (s *ToolchainService) auditJob(
	job *model.ToolchainInstallJob,
	action, sourceName string,
	sourceFallback bool,
) {
	if s.audit == nil {
		return
	}
	details := redact(fmt.Sprintf(
		"operation=%s status=%s source=%s source_fallback=%t",
		job.Operation,
		job.Status,
		sourceName,
		sourceFallback,
	))
	_ = s.audit.Write(
		job.CreatedBy,
		"",
		action,
		"toolchain_install_job",
		strconv.FormatUint(uint64(job.ID), 10),
		details,
		"",
	)
}

func parseUint(value string) (uint, bool) {
	id, err := strconv.ParseUint(value, 10, 64)
	return uint(id), err == nil && id > 0
}

func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
