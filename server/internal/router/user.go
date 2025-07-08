package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupUserRoutes 设置用户管理路由
func SetupUserRoutes(router fiber.Router, db *gorm.DB) {
	userGroup := router.Group("/users")

	userGroup.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "用户列表"})
	})

	userGroup.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建用户"})
	})

	userGroup.Get("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "用户详情"})
	})

	userGroup.Put("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新用户"})
	})

	userGroup.Delete("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除用户"})
	})
}
