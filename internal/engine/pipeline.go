package engine

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
	"buildflow/internal/ws"
)

type Pipeline struct {
	buildRepo      *repository.BuildRepository
	buildDistRepo  *repository.BuildDistributionRepository
	projectRepo    *repository.ProjectRepository
	credentialRepo *repository.CredentialRepository
	envRepo        *repository.EnvironmentRepository
	distRepo       *repository.DistributionRepository
	envVarRepo     *repository.EnvVarRepository
	varGroupRepo   *repository.VarGroupRepository
	serverRepo     *repository.ServerRepository
	notifRepo      *repository.NotificationRepository
	hub            *ws.Hub
	logger         *zap.Logger
	workspaceDir   string
	artifactDir    string
	logDir         string
	cacheDir       string
}

type notificationMessage struct {
	ID            uint      `json:"id"`
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	Message       string    `json:"message"`
	BuildID       *uint     `json:"build_id"`
	ProjectID     uint      `json:"project_id"`
	EnvironmentID uint      `json:"environment_id"`
	BuildStatus   string    `json:"build_status"`
	IsRead        bool      `json:"is_read"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewPipeline(
	buildRepo *repository.BuildRepository,
	buildDistRepo *repository.BuildDistributionRepository,
	projectRepo *repository.ProjectRepository,
	credentialRepo *repository.CredentialRepository,
	envRepo *repository.EnvironmentRepository,
	distRepo *repository.DistributionRepository,
	envVarRepo *repository.EnvVarRepository,
	varGroupRepo *repository.VarGroupRepository,
	serverRepo *repository.ServerRepository,
	notifRepo *repository.NotificationRepository,
	hub *ws.Hub,
	logger *zap.Logger,
	workspaceDir, artifactDir, logDir, cacheDir string,
) *Pipeline {
	return &Pipeline{
		buildRepo:      buildRepo,
		buildDistRepo:  buildDistRepo,
		projectRepo:    projectRepo,
		credentialRepo: credentialRepo,
		envRepo:        envRepo,
		distRepo:       distRepo,
		envVarRepo:     envVarRepo,
		varGroupRepo:   varGroupRepo,
		serverRepo:     serverRepo,
		notifRepo:      notifRepo,
		hub:            hub,
		logger:         logger,
		workspaceDir:   workspaceDir,
		artifactDir:    artifactDir,
		logDir:         logDir,
		cacheDir:       cacheDir,
	}
}

func (p *Pipeline) Execute(ctx context.Context, buildID uint) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error("pipeline panic recovered",
				zap.Uint("build_id", buildID),
				zap.Any("panic", r),
			)
			b, err := p.buildRepo.FindByID(buildID)
			if err == nil && b.Status == "success" {
				return
			}
			p.failBuild(&model.Build{ID: buildID}, fmt.Sprintf("internal panic: %v", r))
		}
	}()

	build, err := p.buildRepo.FindByID(buildID)
	if err != nil {
		p.logger.Error("build not found", zap.Uint("id", buildID))
		return
	}
	project, _ := p.projectRepo.FindByID(build.ProjectID)
	if project == nil {
		p.failBuild(build, "project not found")
		return
	}
	env, _ := p.envRepo.FindByID(build.EnvironmentID)
	if env == nil {
		p.failBuild(build, "environment not found")
		return
	}

	now := time.Now()
	onlyDeploy := build.TriggerType == "deploy" || build.TriggerType == "redistribute"
	if !onlyDeploy || build.Status != "success" {
		build.StartedAt = &now
	}
	if build.TriggerType == "redistribute" {
		_ = p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{"current_stage": "distributing"})
		build.Status = "success"
		build.CurrentStage = "distributing"
	} else if build.TriggerType == "deploy" {
		p.updateStage(build, "deploying")
	} else {
		p.updateStage(build, "cloning")
	}

	// Setup log file
	logDir := filepath.Join(p.logDir, fmt.Sprintf("project-%d", project.ID))
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("build-%03d.log", build.BuildNumber))
	var logFile *os.File
	if build.TriggerType == "redistribute" && build.LogPath != "" {
		logPath = build.LogPath
		logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		logFile, err = os.Create(logPath)
	}
	if err != nil {
		p.failBuild(build, "无法创建日志文件: "+err.Error())
		return
	}
	defer logFile.Close()
	build.LogPath = logPath
	p.buildRepo.UpdateStatus(build.ID, build.Status, map[string]interface{}{
		"log_path":      logPath,
		"started_at":    build.StartedAt,
		"current_stage": build.CurrentStage,
	})

	channel := fmt.Sprintf("build:%d", build.ID)
	var logMu sync.Mutex
	writeLine := func(line string) {
		logMu.Lock()
		defer logMu.Unlock()
		logFile.WriteString(line + "\n")
		p.hub.BroadcastToChannel(channel, []byte(line))
	}

	if build.TriggerType == "deploy" || build.TriggerType == "redistribute" {
		p.executeRedistributeOnly(ctx, build, project, env, writeLine)
		return
	}

	// Stage 1: Git clone/pull
	writeLine("=== Stage: Cloning ===")
	workDir := filepath.Join(p.workspaceDir, fmt.Sprintf("project-%d", project.ID), fmt.Sprintf("env-%d", env.ID))

	authType, repoUsername, repoPassword, err := p.resolveProjectGitAuth(project)
	if err != nil {
		p.failBuild(build, "仓库凭证错误: "+err.Error())
		writeLine("ERROR: " + err.Error())
		return
	}

	// Use build-level branch override if specified, otherwise use env default
	branch := env.Branch
	if build.Branch != "" {
		branch = build.Branch
	}

	err = GitCloneOrPull(ctx, workDir, project.RepoURL, authType, repoUsername, repoPassword, branch, writeLine)
	if err != nil {
		if ctx.Err() != nil {
			p.cancelBuild(build)
			return
		}
		p.failBuild(build, "Git操作失败: "+err.Error())
		writeLine("ERROR: " + err.Error())
		return
	}

	// Checkout specific commit if specified
	if build.CommitHash != "" {
		writeLine("Checking out commit: " + build.CommitHash)
		if err := runGit(ctx, workDir, writeLine, "checkout", build.CommitHash); err != nil {
			if ctx.Err() != nil {
				p.cancelBuild(build)
				return
			}
			p.failBuild(build, "Checkout commit 失败: "+err.Error())
			writeLine("ERROR: " + err.Error())
			return
		}
	}

	// Stage 1.5: Restore cache
	cachePaths := parseCachePaths(env.CachePaths)
	if len(cachePaths) > 0 && p.cacheDir != "" {
		writeLine("=== Stage: Restoring Cache ===")
		envCacheDir := filepath.Join(p.cacheDir, fmt.Sprintf("project-%d", project.ID), fmt.Sprintf("env-%d", env.ID))
		restoredCount := 0
		for _, cp := range cachePaths {
			src := filepath.Join(envCacheDir, cp)
			dst := filepath.Join(workDir, cp)
			if _, err := os.Stat(src); err == nil {
				os.MkdirAll(filepath.Dir(dst), 0755)
				if err := copyDir(src, dst); err != nil {
					writeLine(fmt.Sprintf("WARNING: 恢复缓存 %s 失败: %s", cp, err.Error()))
				} else {
					restoredCount++
					writeLine(fmt.Sprintf("Restored cache: %s", cp))
				}
			}
		}
		if restoredCount == 0 {
			writeLine("No cache found (first build or cache cleared)")
		} else {
			writeLine(fmt.Sprintf("Restored %d cache entries", restoredCount))
		}
	}

	// Stage 2: Build
	if ctx.Err() != nil {
		p.cancelBuild(build)
		return
	}
	p.updateStage(build, "building")
	writeLine("=== Stage: Building ===")

	// Inject env vars
	envVars := os.Environ()
	resolvedEnvVars, err := p.resolveEnvironmentVars(env.ID)
	if err != nil {
		p.failBuild(build, "解析环境变量失败: "+err.Error())
		writeLine("ERROR: " + err.Error())
		return
	}
	if len(resolvedEnvVars) > 0 {
		envVars = append(envVars, resolvedEnvVars...)
	}

	cmd, cleanupScript, err := newBuildScriptCommand(ctx, workDir, env.BuildScriptType, env.BuildScript)
	if err != nil {
		p.failBuild(build, "构建脚本配置无效: "+err.Error())
		writeLine("ERROR: " + err.Error())
		return
	}
	defer cleanupScript()
	cmd.Dir = workDir
	cmd.Env = envVars
	configureBuildCmdProc(cmd)
	cmd.Cancel = func() error {
		return killBuildCmdProcess(cmd)
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		p.failBuild(build, "启动构建脚本失败: "+err.Error())
		writeLine("ERROR: " + err.Error())
		return
	}

	// Ensure any background processes spawned by the script are killed
	defer func() {
		_ = killBuildCmdProcess(cmd)
	}()

	// Stream output line by line
	var scanWg sync.WaitGroup
	scanWg.Add(1)
	go func() {
		defer scanWg.Done()
		scanLines(stdout, writeLine)
	}()
	scanLines(stderr, writeLine)
	scanWg.Wait()

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			p.cancelBuild(build)
			return
		}
		p.failBuild(build, "构建失败: "+err.Error())
		writeLine("ERROR: Build failed with " + err.Error())
		return
	}

	writeLine("=== Build completed successfully ===")

	// Stage 2.5: Save cache
	if len(cachePaths) > 0 && p.cacheDir != "" {
		writeLine("=== Stage: Saving Cache ===")
		envCacheDir := filepath.Join(p.cacheDir, fmt.Sprintf("project-%d", project.ID), fmt.Sprintf("env-%d", env.ID))
		savedCount := 0
		for _, cp := range cachePaths {
			src := filepath.Join(workDir, cp)
			dst := filepath.Join(envCacheDir, cp)
			if info, err := os.Stat(src); err == nil && info.IsDir() {
				os.MkdirAll(filepath.Dir(dst), 0755)
				// Remove old cache entry first to avoid stale files
				os.RemoveAll(dst)
				if err := copyDir(src, dst); err != nil {
					writeLine(fmt.Sprintf("WARNING: 保存缓存 %s 失败: %s", cp, err.Error()))
				} else {
					savedCount++
					writeLine(fmt.Sprintf("Saved cache: %s", cp))
				}
			}
		}
		if savedCount > 0 {
			writeLine(fmt.Sprintf("Saved %d cache entries", savedCount))
		}
	}

	// Stage 3: Collect artifact
	if env.BuildOutputDir != "" {
		writeLine("=== Stage: Collecting Artifact ===")
		outputPath := filepath.Join(workDir, env.BuildOutputDir)
		artifactDir := filepath.Join(p.artifactDir, fmt.Sprintf("project-%d", project.ID))
		os.MkdirAll(artifactDir, 0755)
		artifactFormat := normalizeArtifactFormat(project.ArtifactFormat)
		artifactPath := filepath.Join(artifactDir, artifactArchiveName(build.BuildNumber, artifactFormat))

		if err := createArtifactArchive(artifactPath, outputPath, artifactFormat); err != nil {
			writeLine("WARNING: 打包构建产物失败: " + err.Error())
		} else {
			build.ArtifactPath = artifactPath
			p.buildRepo.UpdateStatus(build.ID, build.Status, map[string]interface{}{"artifact_path": artifactPath})
			writeLine("Artifact saved: " + artifactPath)
		}

		// Cleanup old artifacts
		p.cleanupArtifacts(project)
	}

	sourceDir := filepath.Join(workDir, env.BuildOutputDir)
	if env.BuildOutputDir == "" {
		sourceDir = workDir
	}
	dists, _ := p.distRepo.ListByEnvironmentID(env.ID)
	hasDist := len(dists) > 0
	p.markBuildArtifactSuccess(build, writeLine, hasDist)
	if ctx.Err() != nil {
		p.cancelBuild(build)
		return
	}
	if hasDist {
		p.runDistributions(ctx, build, project, env, sourceDir, writeLine, nil)
	}
}

func decryptServerSecrets(server *model.Server) (password, privateKey, agentToken string, err error) {
	if server.Password != "" {
		password, err = pkg.Decrypt(server.Password)
		if err != nil {
			return "", "", "", fmt.Errorf("decrypt server password: %w", err)
		}
	}
	if server.PrivateKey != "" {
		privateKey, err = pkg.Decrypt(server.PrivateKey)
		if err != nil {
			return "", "", "", fmt.Errorf("decrypt server private key: %w", err)
		}
	}
	if server.AgentToken != "" {
		agentToken, err = pkg.Decrypt(server.AgentToken)
		if err != nil {
			return "", "", "", fmt.Errorf("decrypt agent token: %w", err)
		}
	}
	return password, privateKey, agentToken, nil
}

func (p *Pipeline) updateStage(build *model.Build, stage string) {
	build.Status = stage
	build.CurrentStage = stage
	p.buildRepo.UpdateStatus(build.ID, stage, map[string]interface{}{"current_stage": stage})
}

func (p *Pipeline) failBuild(build *model.Build, errMsg string) {
	latest, err := p.buildRepo.FindByID(build.ID)
	if err == nil && latest.Status == "success" {
		return
	}
	finished := time.Now()
	build.Status = "failed"
	build.ErrorMessage = errMsg
	build.FinishedAt = &finished
	if build.StartedAt != nil {
		build.DurationMs = finished.Sub(*build.StartedAt).Milliseconds()
	}
	p.buildRepo.UpdateStatus(build.ID, "failed", map[string]interface{}{
		"error_message": errMsg,
		"finished_at":   build.FinishedAt,
		"duration_ms":   build.DurationMs,
		"current_stage": build.CurrentStage,
	})
	p.notify(build, "failed")
}

func (p *Pipeline) cancelBuild(build *model.Build) {
	latest, err := p.buildRepo.FindByID(build.ID)
	if err == nil && latest.Status == "success" {
		_ = p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
			"current_stage":            "success",
			"distribution_summary":     "cancelled",
			"redistribute_filter_json": "",
		})
		return
	}
	finished := time.Now()
	build.Status = "cancelled"
	build.FinishedAt = &finished
	if build.StartedAt != nil {
		build.DurationMs = finished.Sub(*build.StartedAt).Milliseconds()
	}
	p.buildRepo.UpdateStatus(build.ID, "cancelled", map[string]interface{}{
		"finished_at":   build.FinishedAt,
		"duration_ms":   build.DurationMs,
		"current_stage": build.CurrentStage,
	})
}

func (p *Pipeline) notify(build *model.Build, status string) {
	notifType := "build_" + status
	statusLabel := map[string]string{
		"success":   "成功",
		"failed":    "失败",
		"cancelled": "已取消",
	}[status]
	if statusLabel == "" {
		statusLabel = status
	}
	title := fmt.Sprintf("构建 #%d 已%s", build.BuildNumber, statusLabel)
	message := strings.TrimSpace(build.ErrorMessage)
	if message == "" {
		message = fmt.Sprintf("项目 #%d / 环境 #%d", build.ProjectID, build.EnvironmentID)
	}
	notif := &model.Notification{
		UserID:  build.TriggeredBy,
		Type:    notifType,
		Title:   title,
		Message: message,
		BuildID: &build.ID,
	}
	if err := p.notifRepo.Create(notif); err != nil {
		if p.logger != nil {
			p.logger.Warn("create notification failed", zap.Uint("build_id", build.ID), zap.Error(err))
		}
		return
	}

	msg, err := json.Marshal(notificationMessage{
		ID:            notif.ID,
		Type:          notif.Type,
		Title:         notif.Title,
		Message:       notif.Message,
		BuildID:       notif.BuildID,
		ProjectID:     build.ProjectID,
		EnvironmentID: build.EnvironmentID,
		BuildStatus:   status,
		IsRead:        notif.IsRead,
		CreatedAt:     notif.CreatedAt,
	})
	if err != nil {
		if p.logger != nil {
			p.logger.Warn("marshal notification failed", zap.Uint("build_id", build.ID), zap.Error(err))
		}
		return
	}
	p.hub.BroadcastToUser(build.TriggeredBy, msg)
}

func (p *Pipeline) cleanupArtifacts(project *model.Project) {
	builds, _ := p.buildRepo.FindArtifactsByProject(project.ID)
	maxArtifacts := project.MaxArtifacts
	if maxArtifacts <= 0 {
		maxArtifacts = 5
	}
	if len(builds) <= maxArtifacts {
		return
	}
	// Remove oldest
	toRemove := builds[maxArtifacts:]
	for _, b := range toRemove {
		if b.ArtifactPath != "" {
			os.Remove(b.ArtifactPath)
			p.buildRepo.UpdateStatus(b.ID, b.Status, map[string]interface{}{"artifact_path": ""})
		}
	}
}

// scanLines reads lines from reader and calls fn for each
func scanLines(r io.Reader, fn func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fn(scanner.Text())
	}
}

func (p *Pipeline) resolveEnvironmentVars(environmentID uint) ([]string, error) {
	groupItems, err := p.varGroupRepo.ListItemsByEnvironmentID(environmentID)
	if err != nil {
		return nil, err
	}
	envVars, err := p.envVarRepo.ListByEnvironmentID(environmentID)
	if err != nil {
		return nil, err
	}
	merged := make(map[string]string)
	order := make([]string, 0, len(groupItems)+len(envVars))
	for _, item := range groupItems {
		value, err := decryptPipelineValue(item.Value, item.IsSecret)
		if err != nil {
			return nil, err
		}
		if _, exists := merged[item.Key]; !exists {
			order = append(order, item.Key)
		}
		merged[item.Key] = value
	}
	for _, item := range envVars {
		value, err := decryptPipelineValue(item.Value, item.IsSecret)
		if err != nil {
			return nil, err
		}
		if _, exists := merged[item.Key]; !exists {
			order = append(order, item.Key)
		}
		merged[item.Key] = value
	}
	result := make([]string, 0, len(order))
	for _, key := range order {
		result = append(result, key+"="+merged[key])
	}
	return result, nil
}

func decryptPipelineValue(value string, isSecret bool) (string, error) {
	if !isSecret {
		return value, nil
	}
	return pkg.Decrypt(value)
}

func (p *Pipeline) resolveProjectGitAuth(project *model.Project) (string, string, string, error) {
	switch strings.ToLower(strings.TrimSpace(project.RepoAuthType)) {
	case "", "none":
		return "none", "", "", nil
	case "credential":
		if project.CredentialID == nil || *project.CredentialID == 0 {
			return "", "", "", fmt.Errorf("project credential is empty")
		}
		credential, err := p.credentialRepo.FindByID(*project.CredentialID)
		if err != nil {
			return "", "", "", err
		}
		secret := ""
		if credential.Password != "" {
			secret, err = pkg.Decrypt(credential.Password)
			if err != nil {
				return "", "", "", err
			}
		}
		authType := "password"
		if strings.ToLower(strings.TrimSpace(credential.Type)) == "token" {
			authType = "token"
		}
		return authType, credential.Username, secret, nil
	default:
		return "none", "", "", nil
	}
}

func artifactArchiveName(buildNumber int, format string) string {
	if normalizeArtifactFormat(format) == "zip" {
		return fmt.Sprintf("build-%03d.zip", buildNumber)
	}
	return fmt.Sprintf("build-%03d.tar.gz", buildNumber)
}

func normalizeArtifactFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "zip":
		return "zip"
	default:
		return "gzip"
	}
}

func createArtifactArchive(targetPath, sourceDir, format string) error {
	if normalizeArtifactFormat(format) == "zip" {
		return createZip(targetPath, sourceDir)
	}
	return createTarGz(targetPath, sourceDir)
}

// createTarGz creates a tar.gz archive from a directory
func createTarGz(targetPath, sourceDir string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

func createZip(targetPath, sourceDir string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		dst, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		src, err := os.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(dst, src)
		return err
	})
}

func extractArtifactArchive(archivePath, destDir, format string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}
	if normalizeArtifactFormat(format) == "zip" {
		return extractZipArchiveFile(archivePath, destDir)
	}
	return extractTarGzArchiveFile(archivePath, destDir)
}

func extractTarGzArchiveFile(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destDir, filepath.Clean(header.Name))
		relPath, err := filepath.Rel(destDir, targetPath)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("illegal archive path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tarReader); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		}
	}
}

func extractZipArchiveFile(archivePath, destDir string) error {
	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		targetPath := filepath.Join(destDir, filepath.Clean(file.Name))
		relPath, err := filepath.Rel(destDir, targetPath)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("illegal archive path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, file.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		reader, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
		if err != nil {
			reader.Close()
			return err
		}
		if _, err := io.Copy(dst, reader); err != nil {
			dst.Close()
			reader.Close()
			return err
		}
		if err := dst.Close(); err != nil {
			reader.Close()
			return err
		}
		if err := reader.Close(); err != nil {
			return err
		}
	}

	return nil
}

// parseCachePaths parses a JSON array string into a list of paths.
// Supports both JSON array format (["node_modules", ".npm"]) and
// newline-separated format.
func parseCachePaths(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	// Try JSON array first
	var paths []string
	if err := json.Unmarshal([]byte(raw), &paths); err == nil {
		var result []string
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}
	// Fallback: newline-separated
	var result []string
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

// copyDir recursively copies a directory from src to dst.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		return copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
