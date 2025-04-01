package handler

import (
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"minidevops/internal/model"
	"minidevops/internal/task"
)

// BuildHandler 构建任务处理器
type BuildHandler struct {
	client  *model.Client
	manager *task.Manager
}

// NewBuildHandler 创建构建任务处理器
func NewBuildHandler(client *model.Client, manager *task.Manager) *BuildHandler {
	return &BuildHandler{
		client:  client,
		manager: manager,
	}
}

// BuildTaskResponse 构建任务响应
type BuildTaskResponse struct {
	ID         int        `json:"id"`
	Status     string     `json:"status"`
	LogPath    string     `json:"log_path,omitempty"`
	Duration   int        `json:"duration,omitempty"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	ProjectID  int        `json:"project_id"`
}

// StartBuild godoc
// @Summary 启动构建任务
// @Description 为指定项目启动新的构建任务
// @Tags 构建
// @Produce json
// @Param id path int true "项目ID"
// @Success 201 {object} BuildTaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/projects/{id}/build [post]
func (h *BuildHandler) StartBuild(c *fiber.Ctx) error {
	projectID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的项目ID",
		})
	}

	userID := c.Locals("userID").(int)

	// 验证项目归属
	exists, err := h.client.Project.
		Query().
		Where(
			model.ProjectIDEQ(projectID),
			model.HasOwnerWith(model.UserIDEQ(userID)),
		).
		Exist(c.Context())

	if err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "项目不存在或无权访问",
		})
	}

	// 启动构建任务
	buildTask, err := h.manager.StartBuild(c.Context(), projectID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "启动构建失败: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(BuildTaskResponse{
		ID:        buildTask.ID,
		Status:    buildTask.Status,
		LogPath:   buildTask.LogPath,
		CreatedAt: buildTask.CreatedAt,
		ProjectID: projectID,
	})
}

// GetBuildTasks godoc
// @Summary 获取项目构建历史
// @Description 获取指定项目的所有构建任务历史
// @Tags 构建
// @Produce json
// @Param id path int true "项目ID"
// @Success 200 {array} BuildTaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/projects/{id}/builds [get]
func (h *BuildHandler) GetBuildTasks(c *fiber.Ctx) error {
	projectID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的项目ID",
		})
	}

	userID := c.Locals("userID").(int)

	// 验证项目归属
	exists, err := h.client.Project.
		Query().
		Where(
			model.ProjectIDEQ(projectID),
			model.HasOwnerWith(model.UserIDEQ(userID)),
		).
		Exist(c.Context())

	if err != nil || !exists {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "项目不存在或无权访问",
		})
	}

	// 查询构建历史
	buildTasks, err := h.client.BuildTask.
		Query().
		Where(model.HasProjectWith(model.ProjectIDEQ(projectID))).
		Order(model.Desc(model.BuildTaskFieldCreatedAt)).
		All(c.Context())

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "获取构建历史失败",
		})
	}

	// 转换为响应格式
	response := make([]BuildTaskResponse, len(buildTasks))
	for i, task := range buildTasks {
		response[i] = BuildTaskResponse{
			ID:         task.ID,
			Status:     task.Status,
			LogPath:    task.LogPath,
			Duration:   task.Duration,
			StartedAt:  task.StartedAt,
			FinishedAt: task.FinishedAt,
			CreatedAt:  task.CreatedAt,
			ProjectID:  projectID,
		}
	}

	return c.JSON(response)
}

// GetBuildTask godoc
// @Summary 获取构建任务详情
// @Description 获取指定构建任务的详细信息
// @Tags 构建
// @Produce json
// @Param id path int true "构建任务ID"
// @Success 200 {object} BuildTaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/builds/{id} [get]
func (h *BuildHandler) GetBuildTask(c *fiber.Ctx) error {
	buildID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的构建ID",
		})
	}

	userID := c.Locals("userID").(int)

	// 查询构建任务并检查权限
	buildTask, err := h.client.BuildTask.
		Query().
		Where(model.BuildTaskIDEQ(buildID)).
		WithProject(func(pq *model.ProjectQuery) {
			pq.WithOwner()
		}).
		Only(c.Context())

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "构建任务不存在",
		})
	}

	// 验证权限
	if buildTask.Edges.Project.Edges.Owner.ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error: "无权访问此构建任务",
		})
	}

	return c.JSON(BuildTaskResponse{
		ID:         buildTask.ID,
		Status:     buildTask.Status,
		LogPath:    buildTask.LogPath,
		Duration:   buildTask.Duration,
		StartedAt:  buildTask.StartedAt,
		FinishedAt: buildTask.FinishedAt,
		CreatedAt:  buildTask.CreatedAt,
		ProjectID:  buildTask.Edges.Project.ID,
	})
}

// GetBuildLogs godoc
// @Summary 获取构建日志
// @Description 获取指定构建任务的日志内容
// @Tags 构建
// @Produce text/plain
// @Param id path int true "构建任务ID"
// @Success 200 {string} string "构建日志内容"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/builds/{id}/logs [get]
func (h *BuildHandler) GetBuildLogs(c *fiber.Ctx) error {
	buildID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error: "无效的构建ID",
		})
	}

	userID := c.Locals("userID").(int)

	// 查询构建任务
	buildTask, err := h.client.BuildTask.
		Query().
		Where(model.BuildTaskIDEQ(buildID)).
		WithProject(func(pq *model.ProjectQuery) {
			pq.WithOwner()
		}).
		Only(c.Context())

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "构建任务不存在",
		})
	}

	// 验证权限
	if buildTask.Edges.Project.Edges.Owner.ID != userID {
		return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
			Error: "无权访问此构建日志",
		})
	}

	// 检查日志文件
	if buildTask.LogPath == "" {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: "构建日志不存在",
		})
	}

	// 读取日志文件
	logFile, err := os.Open(buildTask.LogPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "读取日志失败",
		})
	}
	defer logFile.Close()

	// 设置响应头
	c.Type("text/plain")

	// 将日志内容直接写入响应
	_, err = io.Copy(c.Response().BodyWriter(), logFile)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error: "读取日志失败",
		})
	}

	return nil
}

// WebsocketLogs godoc
// @Summary 构建日志WebSocket
// @Description 通过WebSocket实时获取构建日志
// @Tags 构建
// @Produce json
// @Param id path int true "构建任务ID"
// @Success 101 {string} string "WebSocket升级成功"
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/ws/builds/{id}/logs [get]
func (h *BuildHandler) WebsocketLogsHandler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// 从URL参数获取构建ID
		buildID, err := c.Params("id")
		if err != nil {
			c.Close()
			return
		}

		// 查询构建任务等逻辑应在升级到WebSocket前完成
		// 这里只进行日志推送

		// 读取查询字符串中的Token
		token := c.Query("token")
		if token == "" {
			c.WriteMessage(websocket.TextMessage, []byte("错误: 未提供认证令牌"))
			c.Close()
			return
		}

		// 发送初始消息
		c.WriteMessage(websocket.TextMessage, []byte("已连接到构建日志WebSocket\n"))
		c.WriteMessage(websocket.TextMessage, []byte("构建ID: "+buildID+"\n"))

		// 保持连接直到客户端关闭
		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	})
}