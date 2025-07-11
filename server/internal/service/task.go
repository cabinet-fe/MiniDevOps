package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"server/internal/models"
	"server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// TaskService 任务服务
type TaskService struct {
	*CrudService[models.Task]
	configService *ConfigService
}

// NewTaskService 创建任务服务实例
func NewTaskService(db *gorm.DB, configService *ConfigService) *TaskService {
	return &TaskService{
		CrudService:   NewCrudService[models.Task](db),
		configService: configService,
	}
}

// CreateTaskRequest 创建任务请求结构
type CreateTaskRequest struct {
	Name            string `json:"name" validate:"required"`          // 任务名称（必填）
	RepositoryID    uint   `json:"repository_id" validate:"required"` // 所属仓库（必填）
	Code            string `json:"code" validate:"required"`          // 任务标识（必填，唯一）
	Branch          string `json:"branch" validate:"required"`        // 分支（必填）
	BuildScript     string `json:"build_script"`                      // 构建脚本
	BuildPath       string `json:"build_path"`                        // 构建物路径
	AutoPush        bool   `json:"auto_push"`                         // 构建后是否自动推送
	RemoteServerIDs []uint `json:"remote_server_ids"`                 // 远程服务器ID列表
}

// UpdateTaskRequest 更新任务请求结构
type UpdateTaskRequest struct {
	Name            *string `json:"name"`              // 任务名称
	RepositoryID    *uint   `json:"repository_id"`     // 所属仓库
	Code            *string `json:"code"`              // 任务标识
	Branch          *string `json:"branch"`            // 分支
	BuildScript     *string `json:"build_script"`      // 构建脚本
	BuildPath       *string `json:"build_path"`        // 构建物路径
	AutoPush        *bool   `json:"auto_push"`         // 构建后是否自动推送
	RemoteServerIDs []uint  `json:"remote_server_ids"` // 远程服务器ID列表
}

// TaskListResponse 任务列表响应结构
type TaskListResponse struct {
	Total int64         `json:"total"`
	Items []models.Task `json:"items"`
}

// BuildRequest 构建请求结构
type BuildRequest struct {
	AutoPush bool `json:"auto_push"` // 是否自动推送
}

// GetTasks 获取任务列表
func (s *TaskService) GetTasks(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))
	keyword := c.Query("keyword", "")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	query := s.DB.Model(&models.Task{}).Preload("Repository").Preload("RemoteServers")

	// 关键词搜索
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR branch LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务总数失败", err)
	}

	// 获取任务列表
	var tasks []models.Task
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务列表失败", err)
	}

	response := TaskListResponse{
		Total: total,
		Items: tasks,
	}

	return utils.SuccessWithData(c, "获取任务列表成功", response)
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(c *fiber.Ctx) error {
	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查任务标识是否已存在
	var existTask models.Task
	if err := s.DB.Where("code = ?", req.Code).First(&existTask).Error; err == nil {
		return utils.Error(c, fiber.StatusBadRequest, "任务标识已存在", nil)
	}

	// 检查仓库是否存在
	var repository models.Repository
	if err := s.DB.First(&repository, req.RepositoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusBadRequest, "所属仓库不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询仓库失败", err)
	}

	// 创建任务
	task := models.Task{
		Name:         req.Name,
		RepositoryID: req.RepositoryID,
		Code:         req.Code,
		Branch:       req.Branch,
		BuildScript:  req.BuildScript,
		BuildPath:    req.BuildPath,
		AutoPush:     req.AutoPush,
	}

	if err := s.Create(c, &task); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "创建任务失败", err)
	}

	// 分配远程服务器
	if len(req.RemoteServerIDs) > 0 {
		if err := s.assignRemoteServers(task.ID, req.RemoteServerIDs); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "分配远程服务器失败", err)
		}
	}

	// 重新加载任务信息
	if err := s.DB.Preload("Repository").Preload("RemoteServers").First(&task, task.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载任务信息失败", err)
	}

	return utils.SuccessWithData(c, "创建任务成功", task)
}

// GetTask 获取任务详情
func (s *TaskService) GetTask(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "任务ID格式错误", err)
	}

	var task models.Task
	if err := s.DB.Preload("Repository").Preload("RemoteServers").First(&task, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "任务不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务失败", err)
	}

	return utils.SuccessWithData(c, "获取任务成功", task)
}

// UpdateTask 更新任务
func (s *TaskService) UpdateTask(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "任务ID格式错误", err)
	}

	var req UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "请求参数解析失败", err)
	}

	// 检查任务是否存在
	task, err := s.GetByID(c, uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "任务不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务失败", err)
	}

	// 检查任务标识是否已存在（排除自己）
	if req.Code != nil {
		var existTask models.Task
		if err := s.DB.Where("code = ? AND id != ?", *req.Code, id).First(&existTask).Error; err == nil {
			return utils.Error(c, fiber.StatusBadRequest, "任务标识已存在", nil)
		}
	}

	// 检查仓库是否存在
	if req.RepositoryID != nil {
		var repository models.Repository
		if err := s.DB.First(&repository, *req.RepositoryID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return utils.Error(c, fiber.StatusBadRequest, "所属仓库不存在", nil)
			}
			return utils.Error(c, fiber.StatusInternalServerError, "查询仓库失败", err)
		}
	}

	// 更新任务信息
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.RepositoryID != nil {
		updates["repository_id"] = *req.RepositoryID
	}
	if req.Code != nil {
		updates["code"] = *req.Code
	}
	if req.Branch != nil {
		updates["branch"] = *req.Branch
	}
	if req.BuildScript != nil {
		updates["build_script"] = *req.BuildScript
	}
	if req.BuildPath != nil {
		updates["build_path"] = *req.BuildPath
	}
	if req.AutoPush != nil {
		updates["auto_push"] = *req.AutoPush
	}

	if len(updates) > 0 {
		if err := s.DB.Model(&task).Updates(updates).Error; err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新任务失败", err)
		}
	}

	// 更新远程服务器关联
	if req.RemoteServerIDs != nil {
		if err := s.assignRemoteServers(task.ID, req.RemoteServerIDs); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "更新远程服务器失败", err)
		}
	}

	// 重新加载任务信息
	if err := s.DB.Preload("Repository").Preload("RemoteServers").First(&task, task.ID).Error; err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "加载任务信息失败", err)
	}

	return utils.SuccessWithData(c, "更新任务成功", task)
}

// DeleteTask 删除任务
func (s *TaskService) DeleteTask(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "任务ID格式错误", err)
	}

	// 检查任务是否存在
	task, err := s.GetByID(c, uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "任务不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务失败", err)
	}

	// 删除任务相关文件
	mountPath, err := s.getMountPath()
	if err == nil {
		taskPath := filepath.Join(mountPath, "repos", task.Code)
		os.RemoveAll(taskPath)
	}

	// 删除任务
	if err := s.Delete(c, uint(id)); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "删除任务失败", err)
	}

	return utils.Success(c, "删除任务成功")
}

// BuildTask 构建任务
func (s *TaskService) BuildTask(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "任务ID格式错误", err)
	}

	var req BuildRequest
	c.BodyParser(&req) // 忽略错误，使用默认值

	// 获取任务信息
	var task models.Task
	if err := s.DB.Preload("Repository").Preload("RemoteServers").First(&task, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "任务不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务失败", err)
	}

	// 获取挂载路径
	mountPath, err := s.getMountPath()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "获取挂载路径失败", err)
	}

	// 任务代码目录
	taskPath := filepath.Join(mountPath, "repos", task.Code)

	// 克隆或更新代码
	if err := s.cloneOrUpdateCode(task, taskPath); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "代码拉取失败", err)
	}

	// 执行构建脚本
	if task.BuildScript != "" {
		if err := s.executeBuildScript(task, taskPath); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "构建失败", err)
		}
	}

	// 如果设置了自动推送或请求中指定了推送
	autoPush := task.AutoPush || req.AutoPush
	if autoPush && len(task.RemoteServers) > 0 {
		if err := s.pushToRemoteServers(task, taskPath); err != nil {
			return utils.Error(c, fiber.StatusInternalServerError, "推送失败", err)
		}
	}

	message := "构建完成"
	if autoPush {
		message = "构建并推送完成"
	}

	return utils.Success(c, message)
}

// PushTask 推送任务
func (s *TaskService) PushTask(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "任务ID格式错误", err)
	}

	// 获取任务信息
	var task models.Task
	if err := s.DB.Preload("Repository").Preload("RemoteServers").First(&task, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "任务不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务失败", err)
	}

	if len(task.RemoteServers) == 0 {
		return utils.Error(c, fiber.StatusBadRequest, "任务未配置远程服务器", nil)
	}

	// 获取挂载路径
	mountPath, err := s.getMountPath()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "获取挂载路径失败", err)
	}

	taskPath := filepath.Join(mountPath, "repos", task.Code)

	// 检查任务目录是否存在
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		return utils.Error(c, fiber.StatusBadRequest, "任务尚未构建，请先执行构建", nil)
	}

	// 推送到远程服务器
	if err := s.pushToRemoteServers(task, taskPath); err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "推送失败", err)
	}

	return utils.Success(c, "推送完成")
}

// DownloadBuildArtifacts 下载构建物
func (s *TaskService) DownloadBuildArtifacts(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return utils.Error(c, fiber.StatusBadRequest, "任务ID格式错误", err)
	}

	// 获取任务信息
	var task models.Task
	if err := s.DB.Preload("Repository").First(&task, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.Error(c, fiber.StatusNotFound, "任务不存在", nil)
		}
		return utils.Error(c, fiber.StatusInternalServerError, "查询任务失败", err)
	}

	if task.BuildPath == "" {
		return utils.Error(c, fiber.StatusBadRequest, "任务未配置构建物路径", nil)
	}

	// 获取挂载路径
	mountPath, err := s.getMountPath()
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "获取挂载路径失败", err)
	}

	taskPath := filepath.Join(mountPath, "repos", task.Code)
	buildPath := filepath.Join(taskPath, task.BuildPath)

	// 检查构建物是否存在
	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		return utils.Error(c, fiber.StatusNotFound, "构建物不存在，请先执行构建", nil)
	}

	// 检查是文件还是目录
	fileInfo, err := os.Stat(buildPath)
	if err != nil {
		return utils.Error(c, fiber.StatusInternalServerError, "获取构建物信息失败", err)
	}

	if fileInfo.IsDir() {
		// 如果是目录，打包成压缩文件
		return s.downloadDirectory(c, buildPath, task.Code)
	} else {
		// 如果是文件，直接下载
		return c.SendFile(buildPath)
	}
}

// assignRemoteServers 分配远程服务器
func (s *TaskService) assignRemoteServers(taskID uint, remoteServerIDs []uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		// 删除现有的远程服务器关联
		if err := tx.Where("task_id = ?", taskID).Delete(&models.TaskRemote{}).Error; err != nil {
			return err
		}

		// 创建新的远程服务器关联
		for _, remoteServerID := range remoteServerIDs {
			taskRemote := models.TaskRemote{
				TaskID:         taskID,
				RemoteServerID: remoteServerID,
			}
			if err := tx.Create(&taskRemote).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// getMountPath 获取挂载路径
func (s *TaskService) getMountPath() (string, error) {
	var config models.SystemConfig
	err := s.DB.Where("key = ?", models.ConfigKeyMountPath).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 返回默认路径
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, "dev-ops"), nil
	} else if err != nil {
		return "", err
	}

	return config.Value, nil
}

// cloneOrUpdateCode 克隆或更新代码
func (s *TaskService) cloneOrUpdateCode(task models.Task, taskPath string) error {
	if _, err := os.Stat(taskPath); os.IsNotExist(err) {
		// 目录不存在，克隆代码
		parentDir := filepath.Dir(taskPath)
		os.MkdirAll(parentDir, 0755)

		cmd := exec.Command("git", "clone", "-b", task.Branch, task.Repository.URL, taskPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("克隆代码失败: %v", err)
		}
	} else {
		// 目录存在，更新代码
		cmd := exec.Command("git", "pull", "origin", task.Branch)
		cmd.Dir = taskPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("更新代码失败: %v", err)
		}
	}

	return nil
}

// executeBuildScript 执行构建脚本
func (s *TaskService) executeBuildScript(task models.Task, taskPath string) error {
	// 将构建脚本分行执行
	lines := strings.Split(task.BuildScript, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 创建上下文，设置超时
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		var cmd *exec.Cmd
		if strings.Contains(line, "&&") || strings.Contains(line, "||") || strings.Contains(line, "|") {
			// 复杂命令使用shell执行
			cmd = exec.CommandContext(ctx, "/bin/sh", "-c", line)
		} else {
			// 简单命令直接执行
			parts := strings.Fields(line)
			if len(parts) > 0 {
				cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
			}
		}

		if cmd != nil {
			cmd.Dir = taskPath
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("执行构建脚本失败: %s, 错误: %v", line, err)
			}
		}
	}

	return nil
}

// pushToRemoteServers 推送到远程服务器
func (s *TaskService) pushToRemoteServers(task models.Task, taskPath string) error {
	// TODO: 实现实际的推送逻辑
	// 这里需要使用SSH客户端库来连接远程服务器并推送文件
	// 可以使用golang.org/x/crypto/ssh包

	for _, remote := range task.RemoteServers {
		// 模拟推送过程
		fmt.Printf("推送到远程服务器: %s@%s:%d%s\n", remote.Username, remote.Host, remote.Port, remote.Path)
	}

	return nil
}

// downloadDirectory 下载目录（打包成压缩文件）
func (s *TaskService) downloadDirectory(c *fiber.Ctx, dirPath, taskCode string) error {
	// TODO: 实现目录打包下载功能
	// 可以使用archive/tar或archive/zip包来打包目录

	return utils.Error(c, fiber.StatusNotImplemented, "目录下载功能暂未实现", nil)
}
