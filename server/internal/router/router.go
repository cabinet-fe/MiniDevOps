package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"minidevops/internal/handler"
	"minidevops/internal/middleware"
	"minidevops/internal/model"
	"minidevops/internal/task"
)

// Setup 设置API路由
func Setup(app *fiber.App, client *model.Client) {
	// 初始化任务管理器
	taskManager := task.NewManager(client)

	// 创建各处理器
	authHandler := handler.NewAuthHandler(client)
	projectHandler := handler.NewProjectHandler(client)
	buildHandler := handler.NewBuildHandler(client, taskManager)

	// API根路由
	api := app.Group("/api")

	// 认证相关路由（无需认证）
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)

	// 需要认证的API
	protected := api.Group("", middleware.JWTAuth())

	// 项目管理路由
	protected.Get("/projects", projectHandler.GetProjects)
	protected.Post("/projects", projectHandler.CreateProject)
	protected.Get("/projects/:id", projectHandler.GetProject)
	protected.Put("/projects/:id", projectHandler.UpdateProject)
	protected.Delete("/projects/:id", projectHandler.DeleteProject)

	// 构建任务路由
	protected.Post("/projects/:id/build", buildHandler.StartBuild)
	protected.Get("/projects/:id/builds", buildHandler.GetBuildTasks)
	protected.Get("/builds/:id", buildHandler.GetBuildTask)
	protected.Get("/builds/:id/logs", buildHandler.GetBuildLogs)

	// WebSocket路由（认证在WebSocket处理器中进行）
	api.Get("/ws/builds/:id/logs", buildHandler.WebsocketLogsHandler())

	// Swagger文档
	app.Get("/swagger/*", swagger.HandlerDefault)
}