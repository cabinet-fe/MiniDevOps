package router

import (
	"server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRoleRoutes 设置角色管理路由
func SetupRoleRoutes(router fiber.Router, db *gorm.DB) {
	roleService := service.NewRoleService(db)
	roleGroup := router.Group("/roles")

	roleGroup.Get("/", roleService.GetRoles)
	roleGroup.Post("/", roleService.CreateRole)
	roleGroup.Get("/:id", roleService.GetRole)
	roleGroup.Put("/:id", roleService.UpdateRole)
	roleGroup.Delete("/:id", roleService.DeleteRole)
}
