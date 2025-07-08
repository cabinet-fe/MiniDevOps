package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupConfigRoutes 设置系统配置管理路由
func SetupConfigRoutes(router fiber.Router, db *gorm.DB) {
	configGroup := router.Group("/configs")

	configGroup.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "系统配置列表"})
	})

	configGroup.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建系统配置"})
	})

	configGroup.Get("/:key", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "获取配置详情"})
	})

	configGroup.Put("/:key", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新配置"})
	})

	configGroup.Delete("/:key", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除配置"})
	})
}
