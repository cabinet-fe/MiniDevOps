package task

import (
	"context"
	"errors"
	"sync"
	"time"

	"minidevops/internal/builder"
	"minidevops/internal/model"
)

// Manager 任务管理器
type Manager struct {
	client     *model.Client
	runningMap map[int]context.CancelFunc
	mu         sync.Mutex
}

// NewManager 创建任务管理器
func NewManager(client *model.Client) *Manager {
	return &Manager{
		client:     client,
		runningMap: make(map[int]context.CancelFunc),
	}
}

// StartBuild 启动构建任务
func (m *Manager) StartBuild(ctx context.Context, projectID int) (*model.BuildTask, error) {
	// 查找项目
	project, err := m.client.Project.Get(ctx, projectID)
	if err != nil {
		return nil, errors.New("项目不存在")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已有正在运行的任务
	runningTasks, err := m.client.BuildTask.
		Query().
		Where(
			model.BuildTaskStatusEQ("running"),
			model.HasProjectWith(model.ProjectIDEQ(projectID)),
		).
		All(ctx)

	if err != nil {
		return nil, err
	}

	if len(runningTasks) > 0 {
		return nil, errors.New("该项目已有正在运行的构建任务")
	}

	// 创建新的构建任务
	buildTask, err := m.client.BuildTask.
		Create().
		SetStatus("pending").
		SetCreatedAt(time.Now()).
		SetProject(project).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	// 创建任务上下文
	taskCtx, cancel := context.WithCancel(ctx)
	m.runningMap[buildTask.ID] = cancel

	// 异步执行构建
	go func() {
		defer delete(m.runningMap, buildTask.ID)

		task, err := builder.NewTask(taskCtx, project, buildTask, m.client)
		if err != nil {
			return
		}

		_ = task.Execute()
	}()

	return buildTask, nil
}

// StopBuild 停止构建任务
func (m *Manager) StopBuild(ctx context.Context, buildID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查任务是否存在且正在运行
	cancel, exists := m.runningMap[buildID]
	if !exists {
		return errors.New("任务不存在或已结束")
	}

	// 取消任务
	cancel()
	delete(m.runningMap, buildID)

	// 更新任务状态
	_, err := m.client.BuildTask.
		UpdateOneID(buildID).
		SetStatus("failed").
		SetFinishedAt(time.Now()).
		Save(ctx)

	return err
}

// GetTaskStatus 获取任务状态
func (m *Manager) GetTaskStatus(ctx context.Context, buildID int) (string, error) {
	task, err := m.client.BuildTask.Get(ctx, buildID)
	if err != nil {
		return "", errors.New("任务不存在")
	}
	return task.Status, nil
}