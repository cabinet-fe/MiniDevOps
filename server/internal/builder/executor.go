package builder

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"minidevops/internal/model"
)

// Logger 日志接口
type Logger interface {
	Write(p []byte) (n int, err error)
	Close() error
}

// FileLogger 基于文件的日志实现
type FileLogger struct {
	file    *os.File
	writers []io.Writer
	mu      sync.Mutex
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(logPath string) (*FileLogger, error) {
	// 确保目录存在
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	file, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}

	return &FileLogger{
		file:    file,
		writers: []io.Writer{file},
	}, nil
}

// Write 实现io.Writer接口，写入日志
func (l *FileLogger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 广播到所有写入器
	for _, w := range l.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// Close 关闭日志文件
func (l *FileLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}

// AddWriter 添加额外的写入器（如WebSocket）
func (l *FileLogger) AddWriter(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writers = append(l.writers, w)
}

// Task 表示一个构建任务
type Task struct {
	Project   *model.Project
	BuildTask *model.BuildTask
	Logger    *FileLogger
	client    *model.Client
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewTask 创建新的构建任务
func NewTask(ctx context.Context, project *model.Project, buildTask *model.BuildTask, client *model.Client) (*Task, error) {
	// 设置超时控制
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)

	// 创建日志目录和文件
	logDir := filepath.Join("./data/logs", fmt.Sprintf("project_%d", project.ID))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		cancel()
		return nil, err
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("build_%d.log", buildTask.ID))
	logger, err := NewFileLogger(logPath)
	if err != nil {
		cancel()
		return nil, err
	}

	// 更新日志路径
	buildTask, err = buildTask.Update().SetLogPath(logPath).Save(ctx)
	if err != nil {
		cancel()
		logger.Close()
		return nil, err
	}

	return &Task{
		Project:   project,
		BuildTask: buildTask,
		Logger:    logger,
		client:    client,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Execute 执行构建任务
func (t *Task) Execute() error {
	defer t.cancel()
	defer t.Logger.Close()

	// 更新任务状态
	startTime := time.Now()
	t.BuildTask, _ = t.BuildTask.Update().
		SetStatus("running").
		SetStartedAt(startTime).
		Save(t.ctx)

	// 获取项目目录
	projectDir := filepath.Join("./data/projects", fmt.Sprintf("%d", t.Project.ID))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return t.finishWithError(err)
	}

	// 克隆仓库
	if err := t.cloneRepo(projectDir); err != nil {
		return t.finishWithError(err)
	}

	// 执行构建命令
	if err := t.runBuildCommand(projectDir); err != nil {
		return t.finishWithError(err)
	}

	// 构建成功
	endTime := time.Now()
	duration := int(endTime.Sub(startTime).Seconds())

	t.BuildTask, _ = t.BuildTask.Update().
		SetStatus("success").
		SetFinishedAt(endTime).
		SetDuration(duration).
		Save(t.ctx)

	// 更新项目的最后构建时间
	t.Project, _ = t.Project.Update().
		SetLastBuildAt(endTime).
		Save(t.ctx)

	fmt.Fprintf(t.Logger, "\n✅ 构建成功, 耗时: %d 秒\n", duration)
	return nil
}

// cloneRepo 克隆Git仓库
func (t *Task) cloneRepo(projectDir string) error {
	// 先检查目录是否已存在
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		// 如果目录存在，先执行git fetch --all
		fmt.Fprintf(t.Logger, "代码目录已存在，更新代码...\n")
		fetchCmd := exec.CommandContext(t.ctx, "git", "-C", projectDir, "fetch", "--all")
		fetchCmd.Stdout = t.Logger
		fetchCmd.Stderr = t.Logger

		if err := fetchCmd.Run(); err != nil {
			fmt.Fprintf(t.Logger, "❌ 更新代码失败: %v\n", err)
			return err
		}

		// 切换到指定分支并拉取最新代码
		checkoutCmd := exec.CommandContext(t.ctx, "git", "-C", projectDir, "checkout", t.Project.Branch)
		checkoutCmd.Stdout = t.Logger
		checkoutCmd.Stderr = t.Logger

		if err := checkoutCmd.Run(); err != nil {
			fmt.Fprintf(t.Logger, "❌ 切换分支失败: %v\n", err)
			return err
		}

		pullCmd := exec.CommandContext(t.ctx, "git", "-C", projectDir, "pull", "origin", t.Project.Branch)
		pullCmd.Stdout = t.Logger
		pullCmd.Stderr = t.Logger

		if err := pullCmd.Run(); err != nil {
			fmt.Fprintf(t.Logger, "❌ 拉取代码失败: %v\n", err)
			return err
		}

		fmt.Fprintf(t.Logger, "✅ 代码更新成功\n")
		return nil
	}

	// 目录不存在，克隆仓库
	fmt.Fprintf(t.Logger, "克隆仓库: %s 分支: %s\n", t.Project.RepoURL, t.Project.Branch)

	cloneCmd := exec.CommandContext(t.ctx, "git", "clone", "-b", t.Project.Branch, t.Project.RepoURL, projectDir)
	cloneCmd.Stdout = t.Logger
	cloneCmd.Stderr = t.Logger

	if err := cloneCmd.Run(); err != nil {
		fmt.Fprintf(t.Logger, "❌ 克隆仓库失败: %v\n", err)
		return err
	}

	fmt.Fprintf(t.Logger, "✅ 代码克隆成功\n")
	return nil
}

// runBuildCommand 执行构建命令
func (t *Task) runBuildCommand(projectDir string) error {
	fmt.Fprintf(t.Logger, "执行构建命令: %s\n", t.Project.BuildCmd)

	buildCmd := exec.CommandContext(t.ctx, "sh", "-c", t.Project.BuildCmd)
	buildCmd.Dir = projectDir
	buildCmd.Stdout = t.Logger
	buildCmd.Stderr = t.Logger

	if err := buildCmd.Run(); err != nil {
		fmt.Fprintf(t.Logger, "❌ 构建命令执行失败: %v\n", err)
		return err
	}

	fmt.Fprintf(t.Logger, "✅ 构建命令执行成功\n")
	return nil
}

// finishWithError 处理构建失败
func (t *Task) finishWithError(err error) error {
	endTime := time.Now()
	duration := int(endTime.Sub(t.BuildTask.StartedAt).Seconds())

	t.BuildTask, _ = t.BuildTask.Update().
		SetStatus("failed").
		SetFinishedAt(endTime).
		SetDuration(duration).
		Save(t.ctx)

	fmt.Fprintf(t.Logger, "\n❌ 构建失败: %v, 耗时: %d 秒\n", err, duration)
	return err
}