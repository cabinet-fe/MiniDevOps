package engine

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
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
	serverRepo   *repository.ServerRepository
	notifRepo    *repository.NotificationRepository
	hub          *ws.Hub
	logger       *zap.Logger
	workspaceDir string
	artifactDir  string
	logDir       string
}

func NewPipeline(
	buildRepo *repository.BuildRepository,
	projectRepo *repository.ProjectRepository,
	envRepo *repository.EnvironmentRepository,
	serverRepo *repository.ServerRepository,
	notifRepo *repository.NotificationRepository,
	hub *ws.Hub,
	logger *zap.Logger,
	workspaceDir, artifactDir, logDir string,
) *Pipeline {
	return &Pipeline{
		buildRepo:    buildRepo, projectRepo: projectRepo, envRepo: envRepo,
		serverRepo: serverRepo, notifRepo: notifRepo,
		hub: hub, logger: logger,
		workspaceDir: workspaceDir, artifactDir: artifactDir, logDir: logDir,
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
	p.updateStatus(build, "cloning")

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
	p.buildRepo.UpdateStatus(build.ID, build.Status, map[string]interface{}{"log_path": logPath, "started_at": build.StartedAt})

	channel := fmt.Sprintf("build:%d", build.ID)
	writeLine := func(line string) {
		logFile.WriteString(line + "\n")
		p.hub.BroadcastToChannel(channel, []byte(line))
	}

	// Stage 1: Git clone/pull
	writeLine("=== Stage: Cloning ===")
	workDir := filepath.Join(p.workspaceDir, fmt.Sprintf("project-%d", project.ID))

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

	// Stage 2: Build
	if ctx.Err() != nil {
		p.cancelBuild(build)
		return
	}
	p.updateStatus(build, "building")
	writeLine("=== Stage: Building ===")

	// Inject env vars
	envVars := os.Environ()
	if env.EnvVars != "" {
		envVars = append(envVars, parseEnvVars(env.EnvVars)...)
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", env.BuildScript)
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

	// Stage 3: Collect artifact
	if env.BuildOutputDir != "" {
		writeLine("=== Stage: Collecting Artifact ===")
		outputPath := filepath.Join(workDir, env.BuildOutputDir)
		artifactDir := filepath.Join(p.artifactDir, fmt.Sprintf("project-%d", project.ID))
		os.MkdirAll(artifactDir, 0755)
		artifactPath := filepath.Join(artifactDir, fmt.Sprintf("build-%03d.tar.gz", build.BuildNumber))

		if err := createTarGz(artifactPath, outputPath); err != nil {
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
		p.updateStatus(build, "deploying")
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
			SourceDir: sourceDir,
			Server: deployer.ServerInfo{
				Host:       server.Host,
				Port:       server.Port,
				Username:   server.Username,
				AuthType:   server.AuthType,
				Password:   password,
				PrivateKey: privateKey,
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
			if err := deployer.ExecuteRemoteScript(ctx, deployOpts.Server, env.PostDeployScript, writeLine); err != nil {
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
	p.updateStatus(build, "success")
	p.buildRepo.UpdateStatus(build.ID, "success", map[string]interface{}{
		"finished_at": build.FinishedAt,
		"duration_ms": build.DurationMs,
	})
	writeLine(fmt.Sprintf("=== Build #%d finished in %dms ===", build.BuildNumber, build.DurationMs))

	// Notify
	p.notify(build, "success")
}

func (p *Pipeline) updateStatus(build *model.Build, status string) {
	build.Status = status
	p.buildRepo.UpdateStatus(build.ID, status, nil)
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
		"finished_at": build.FinishedAt,
		"duration_ms": build.DurationMs,
	})
}

func (p *Pipeline) notify(build *model.Build, status string) {
	notifType := "build_" + status
	title := fmt.Sprintf("构建 #%d %s", build.BuildNumber, status)
	p.notifRepo.Create(&model.Notification{
		UserID:  build.TriggeredBy,
		Type:    notifType,
		Title:   title,
		Message: build.ErrorMessage,
		BuildID: &build.ID,
	})
	// Broadcast via WebSocket
	msg := fmt.Sprintf(`{"type":"%s","build_id":%d,"title":"%s"}`, notifType, build.ID, title)
	p.hub.BroadcastToUser(build.TriggeredBy, []byte(msg))
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

// parseEnvVars parses JSON env vars to KEY=VALUE format
func parseEnvVars(jsonStr string) []string {
	var result []string
	jsonStr = strings.TrimSpace(jsonStr)
	if jsonStr == "" || jsonStr == "{}" {
		return result
	}
	jsonStr = strings.Trim(jsonStr, "{}")
	parts := strings.Split(jsonStr, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			key := strings.Trim(strings.TrimSpace(kv[0]), `"`)
			val := strings.Trim(strings.TrimSpace(kv[1]), `"`)
			result = append(result, key+"="+val)
		}
	}
	return result
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
