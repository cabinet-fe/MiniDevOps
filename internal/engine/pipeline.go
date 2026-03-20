package engine

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"

	"buildflow/internal/deployer"
	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
	"buildflow/internal/ws"
)

type Pipeline struct {
	buildRepo    *repository.BuildRepository
	projectRepo  *repository.ProjectRepository
	envRepo      *repository.EnvironmentRepository
	envVarRepo   *repository.EnvVarRepository
	varGroupRepo *repository.VarGroupRepository
	serverRepo   *repository.ServerRepository
	notifRepo    *repository.NotificationRepository
	hub          *ws.Hub
	logger       *zap.Logger
	workspaceDir string
	artifactDir  string
	logDir       string
	cacheDir     string
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
	projectRepo *repository.ProjectRepository,
	envRepo *repository.EnvironmentRepository,
	envVarRepo *repository.EnvVarRepository,
	varGroupRepo *repository.VarGroupRepository,
	serverRepo *repository.ServerRepository,
	notifRepo *repository.NotificationRepository,
	hub *ws.Hub,
	logger *zap.Logger,
	workspaceDir, artifactDir, logDir, cacheDir string,
) *Pipeline {
	return &Pipeline{
		buildRepo: buildRepo, projectRepo: projectRepo, envRepo: envRepo,
		envVarRepo: envVarRepo, varGroupRepo: varGroupRepo,
		serverRepo: serverRepo, notifRepo: notifRepo,
		hub: hub, logger: logger,
		workspaceDir: workspaceDir, artifactDir: artifactDir, logDir: logDir,
		cacheDir: cacheDir,
	}
}

func (p *Pipeline) Execute(ctx context.Context, buildID uint) {
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
	build.StartedAt = &now
	p.updateStage(build, "cloning")

	// Setup log file
	logDir := filepath.Join(p.logDir, fmt.Sprintf("project-%d", project.ID))
	os.MkdirAll(logDir, 0755)
	logPath := filepath.Join(logDir, fmt.Sprintf("build-%03d.log", build.BuildNumber))
	logFile, err := os.Create(logPath)
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
	writeLine := func(line string) {
		logFile.WriteString(line + "\n")
		p.hub.BroadcastToChannel(channel, []byte(line))
	}

	// Stage 1: Git clone/pull
	writeLine("=== Stage: Cloning ===")
	workDir := filepath.Join(p.workspaceDir, fmt.Sprintf("project-%d", project.ID), fmt.Sprintf("env-%d", env.ID))

	repoPassword := ""
	if project.RepoPassword != "" {
		repoPassword, _ = pkg.Decrypt(project.RepoPassword)
	}

	// Use build-level branch override if specified, otherwise use env default
	branch := env.Branch
	if build.Branch != "" {
		branch = build.Branch
	}

	err = GitCloneOrPull(ctx, workDir, project.RepoURL, project.RepoAuthType, project.RepoUsername, repoPassword, branch, writeLine)
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

	// Select interpreter based on build script type
	interpreter := "sh"
	interpreterArgs := []string{"-c", env.BuildScript}
	switch env.BuildScriptType {
	case "node":
		interpreter = "node"
		interpreterArgs = []string{"-e", env.BuildScript}
	case "python":
		interpreter = "python3"
		interpreterArgs = []string{"-c", env.BuildScript}
	default: // "bash" or empty
		interpreter = "sh"
		interpreterArgs = []string{"-c", env.BuildScript}
	}
	cmd := exec.CommandContext(ctx, interpreter, interpreterArgs...)
	cmd.Dir = workDir
	cmd.Env = envVars

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		p.failBuild(build, "启动构建脚本失败: "+err.Error())
		writeLine("ERROR: " + err.Error())
		return
	}

	// Stream output line by line
	go scanLines(stdout, writeLine)
	scanLines(stderr, writeLine)

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

	// Stage 4: Deploy
	if env.DeployServerID != nil && env.DeployPath != "" {
		if ctx.Err() != nil {
			p.cancelBuild(build)
			return
		}
		p.updateStage(build, "deploying")
		writeLine("=== Stage: Deploying ===")

		server, err := p.serverRepo.FindByID(*env.DeployServerID)
		if err != nil {
			p.failBuild(build, "服务器不存在")
			writeLine("ERROR: Server not found")
			return
		}

		password := ""
		if server.Password != "" {
			password, _ = pkg.Decrypt(server.Password)
		}
		privateKey := ""
		if server.PrivateKey != "" {
			privateKey, _ = pkg.Decrypt(server.PrivateKey)
		}

		sourceDir := filepath.Join(workDir, env.BuildOutputDir)
		if env.BuildOutputDir == "" {
			sourceDir = workDir
		}

		d := deployer.NewDeployer(env.DeployMethod)
		deployOpts := deployer.DeployOptions{
			SourceDir:     sourceDir,
			ArchiveFormat: normalizeArtifactFormat(project.ArtifactFormat),
			Server: deployer.ServerInfo{
				Host:       server.Host,
				Port:       server.Port,
				OSType:     server.OSType,
				Username:   server.Username,
				AuthType:   server.AuthType,
				Password:   password,
				PrivateKey: privateKey,
				AgentURL:   server.AgentURL,
				AgentToken: server.AgentToken,
			},
			RemotePath: env.DeployPath,
			Logger:     writeLine,
		}

		if err := d.Deploy(ctx, deployOpts); err != nil {
			if ctx.Err() != nil {
				p.cancelBuild(build)
				return
			}
			p.failBuild(build, "部署失败: "+err.Error())
			writeLine("ERROR: Deploy failed: " + err.Error())
			return
		}
		writeLine("Deploy completed successfully")

		// Post-deploy script
		if env.PostDeployScript != "" {
			writeLine("=== Executing post-deploy script ===")
			if err := deployer.ExecuteRemoteScriptInDir(ctx, deployOpts.Server, deployOpts.RemotePath, env.PostDeployScript, writeLine); err != nil {
				if ctx.Err() != nil {
					p.cancelBuild(build)
					return
				}
				p.failBuild(build, "部署后脚本失败: "+err.Error())
				writeLine("ERROR: Post-deploy script failed: " + err.Error())
				return
			}
			writeLine("Post-deploy script completed")
		}
	}

	// Success
	finished := time.Now()
	build.FinishedAt = &finished
	build.DurationMs = finished.Sub(*build.StartedAt).Milliseconds()
	build.Status = "success"
	build.CurrentStage = "success"
	p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
		"finished_at":   build.FinishedAt,
		"duration_ms":   build.DurationMs,
		"current_stage": build.CurrentStage,
	})
	writeLine(fmt.Sprintf("=== Build #%d finished in %dms ===", build.BuildNumber, build.DurationMs))

	// Notify
	p.notify(build, "success")
}

func (p *Pipeline) updateStage(build *model.Build, stage string) {
	build.Status = stage
	build.CurrentStage = stage
	p.buildRepo.UpdateStatus(build.ID, stage, map[string]interface{}{"current_stage": stage})
}

func (p *Pipeline) failBuild(build *model.Build, errMsg string) {
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
