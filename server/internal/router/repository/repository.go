package repository

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRepositoryRoutes 设置仓库管理路由
func SetupRepositoryRoutes(router fiber.Router, db *gorm.DB) {
	repoGroup := router.Group("/repositories")

	repoGroup.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "仓库列表"})
	})

	repoGroup.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建仓库"})
	})

	repoGroup.Get("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "仓库详情"})
	})

	repoGroup.Put("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新仓库"})
	})

	repoGroup.Delete("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除仓库"})
	})
}
