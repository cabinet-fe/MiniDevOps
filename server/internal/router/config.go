package router

import (
	"minidevops/server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupConfigRoutes 设置系统配置管理路由
func SetupConfigRoutes(router fiber.Router, db *gorm.DB) {
	configService := service.NewConfigService(db)
	configGroup := router.Group("/configs")

	configGroup.Get("/", configService.GetConfigs)
	configGroup.Get("/:key", configService.GetConfig)
	configGroup.Put("/:key", configService.UpdateConfig)
	configGroup.Delete("/:key", configService.DeleteConfig)

	// 挂载路径相关路由
	configGroup.Get("/mount-path/current", configService.GetMountPath)
	configGroup.Put("/mount-path/update", configService.UpdateMountPath)
}
