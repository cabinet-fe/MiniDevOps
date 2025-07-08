package role

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRoleRoutes 设置角色管理路由
func SetupRoleRoutes(router fiber.Router, db *gorm.DB) {
	roleGroup := router.Group("/roles")

	roleGroup.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "角色列表"})
	})

	roleGroup.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建角色"})
	})

	roleGroup.Get("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "角色详情"})
	})

	roleGroup.Put("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新角色"})
	})

	roleGroup.Delete("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除角色"})
	})
}
