package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupTaskRoutes 设置任务管理路由
func SetupTaskRoutes(router fiber.Router, db *gorm.DB) {
	taskGroup := router.Group("/tasks")

	taskGroup.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "任务列表"})
	})

	taskGroup.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建任务"})
	})

	taskGroup.Get("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "任务详情"})
	})

	taskGroup.Put("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新任务"})
	})

	taskGroup.Delete("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除任务"})
	})

	taskGroup.Post("/:id/build", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "构建任务"})
	})

	taskGroup.Post("/:id/push", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "推送任务"})
	})

	taskGroup.Get("/:id/download", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "下载构建物"})
	})
}
