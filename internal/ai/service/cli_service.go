package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"bedrock/internal/ai/model"
	"bedrock/internal/ai/repository"
)

type AuditWriter interface {
	Write(userID uint, username, action, resourceType, resourceID, details, ip string) error
}

type CLIService struct {
	repo  *repository.AIRepository
	audit AuditWriter
}

func NewCLIService(repo *repository.AIRepository, audit ...AuditWriter) *CLIService {
	svc := &CLIService{repo: repo}
	if len(audit) > 0 {
		svc.audit = audit[0]
	}
	return svc
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
	Detected   bool   `json:"detected"`
	Output     string `json:"output"`
	Path       string `json:"path"`
	Version    string `json:"version"`
	Healthy    bool   `json:"healthy"`
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
		RiskNotice: model.RiskNoticeSameUID,
		Output:     strings.TrimSpace(output),
	}
	if runErr != nil {
		result.Detected = false
		result.Healthy = false
		cli.InstallStatus = "missing"
		cli.InstalledPath = ""
		cli.InstalledVersion = ""
		cli.Healthy = false
		if result.Output == "" {
			result.Output = runErr.Error()
		}
		_ = s.repo.UpdateCLI(cli)
		return result, nil
	}
	version := extractCLIVersion(output, cli.BinaryName)
	if version == "" {
		version = probeCLIVersion(cli.BinaryName)
	}
	if isPathLine(version, cli.BinaryName) {
		version = ""
	}
	path := ""
	if p, lookErr := exec.LookPath(cli.BinaryName); lookErr == nil {
		path = p
	} else if p := extractCLIPath(output, cli.BinaryName); p != "" {
		path = p
	}
	result.Detected = true
	result.Healthy = true
	result.Path = path
	result.Version = version
	cli.InstallStatus = "installed"
	cli.InstalledPath = path
	cli.InstalledVersion = version
	cli.Healthy = true
	_ = s.repo.UpdateCLI(cli)
	return result, nil
}

type ExecuteInput struct {
	Version string `json:"version"`
}

type ExecuteResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

func (s *CLIService) Execute(ctx context.Context, key, operation string, input ExecuteInput, createdBy uint) (*ExecuteResult, error) {
	cli, err := s.repo.FindCLIByKey(key)
	if err != nil {
		return nil, err
	}
	template := templateFor(cli, operation)
	if template == "" {
		return nil, errors.New("该 CLI 未配置此操作命令")
	}
	version := strings.TrimSpace(input.Version)

	if needsSource(operation, template) {
		return s.executeWithSources(ctx, key, operation, version, template, createdBy)
	}
	command := renderCLICommand(template, version, "")
	output, runErr := executeShell(ctx, command)
	if runErr != nil {
		s.auditExecute(key, operation, createdBy, false)
		return &ExecuteResult{Success: false, Output: output, Error: runErr.Error()}, nil
	}
	s.auditExecute(key, operation, createdBy, true)
	return &ExecuteResult{Success: true, Output: output}, nil
}

func (s *CLIService) executeWithSources(ctx context.Context, key, operation, version, template string, createdBy uint) (*ExecuteResult, error) {
	sources, err := s.repo.ListEnabledSources(key)
	if err != nil {
		return nil, err
	}
	if len(sources) == 0 {
		return nil, errors.New("没有可用安装源")
	}
	var log strings.Builder
	for i, source := range sources {
		command := renderCLICommand(template, version, source.BaseURL)
		log.WriteString(fmt.Sprintf("trying source %q (priority %d)\n", source.Name, source.Priority))
		output, runErr := executeShell(ctx, command)
		log.WriteString(output)
		if runErr == nil {
			if i > 0 {
				log.WriteString("multi-source fallback succeeded after earlier failures\n")
			}
			s.auditExecute(key, operation, createdBy, true)
			return &ExecuteResult{Success: true, Output: log.String()}, nil
		}
		log.WriteString(fmt.Sprintf("source %q failed: %v\n", source.Name, runErr))
	}
	s.auditExecute(key, operation, createdBy, false)
	return &ExecuteResult{Success: false, Output: log.String(), Error: "所有安装源均失败"}, nil
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

func (s *CLIService) auditExecute(cliKey, operation string, createdBy uint, success bool) {
	if s.audit == nil {
		return
	}
	status := "failed"
	if success {
		status = "success"
	}
	_ = s.audit.Write(createdBy, "", "cli_execute", "cli_runtime", cliKey,
		fmt.Sprintf("cli=%s op=%s status=%s", cliKey, operation, status), "")
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

var versionPattern = regexp.MustCompile(`(?i)(?:version\s+)?v?(\d+\.\d+(?:\.\d+)?(?:[-+][\w.]+)?)`)

func probeCLIVersion(binaryName string) string {
	for _, probe := range []string{
		binaryName + " --version",
		binaryName + " -v",
		binaryName + " version",
	} {
		out, err := executeShell(context.Background(), probe)
		if err != nil {
			continue
		}
		if v := extractCLIVersion(out, binaryName); v != "" && !isPathLine(v, binaryName) {
			return v
		}
	}
	return ""
}

func extractCLIVersion(output, binaryName string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || isPathLine(line, binaryName) {
			continue
		}
		if m := versionPattern.FindStringSubmatch(line); len(m) > 1 {
			return m[1]
		}
		return line
	}
	return ""
}

func extractCLIPath(output, binaryName string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if isPathLine(line, binaryName) {
			return line
		}
	}
	return ""
}

func isPathLine(line, binaryName string) bool {
	if strings.HasPrefix(line, "/") {
		return true
	}
	if len(line) >= 2 && line[1] == ':' {
		return true
	}
	if strings.Contains(line, "/") && !strings.Contains(line, " ") {
		return true
	}
	if filepath.Base(line) == binaryName {
		return true
	}
	return false
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
	if cli.InstalledPath != "" {
		dir := filepath.Dir(cli.InstalledPath)
		env = append(env, "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
	return env
}
