package router

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRemoteRoutes 设置远程服务器管理路由
func SetupRemoteRoutes(router fiber.Router, db *gorm.DB) {
	remoteGroup := router.Group("/remotes")

	remoteGroup.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "远程服务器列表"})
	})

	remoteGroup.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建远程服务器"})
	})

	remoteGroup.Get("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "远程服务器详情"})
	})

	remoteGroup.Put("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新远程服务器"})
	})

	remoteGroup.Delete("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除远程服务器"})
	})
}
