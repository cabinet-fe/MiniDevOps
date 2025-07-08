package router

import (
	"minidevops/server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupTaskRoutes 设置任务管理路由
func SetupTaskRoutes(router fiber.Router, db *gorm.DB) {
	configService := service.NewConfigService(db)
	taskService := service.NewTaskService(db, configService)
	taskGroup := router.Group("/tasks")

	taskGroup.Get("/", taskService.GetTasks)
	taskGroup.Post("/", taskService.CreateTask)
	taskGroup.Get("/:id", taskService.GetTask)
	taskGroup.Put("/:id", taskService.UpdateTask)
	taskGroup.Delete("/:id", taskService.DeleteTask)

	// 任务操作路由
	taskGroup.Post("/:id/build", taskService.BuildTask)
	taskGroup.Post("/:id/push", taskService.PushTask)
	taskGroup.Get("/:id/download", taskService.DownloadBuildArtifacts)
}
