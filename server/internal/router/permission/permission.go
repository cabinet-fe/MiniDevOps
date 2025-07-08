package permission

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupPermissionRoutes 设置权限管理路由
func SetupPermissionRoutes(router fiber.Router, db *gorm.DB) {
	permissionGroup := router.Group("/permissions")

	permissionGroup.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "权限列表"})
	})

	permissionGroup.Post("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "创建权限"})
	})

	permissionGroup.Get("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "权限详情"})
	})

	permissionGroup.Put("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "更新权限"})
	})

	permissionGroup.Delete("/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "删除权限"})
	})
}
