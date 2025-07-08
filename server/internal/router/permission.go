package router

import (
	"minidevops/server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupPermissionRoutes 设置权限管理路由
func SetupPermissionRoutes(router fiber.Router, db *gorm.DB) {
	permissionService := service.NewPermissionService(db)
	permissionGroup := router.Group("/permissions")

	permissionGroup.Get("/", permissionService.GetPermissions)
	permissionGroup.Get("/tree", permissionService.GetPermissionTree)
	permissionGroup.Post("/", permissionService.CreatePermission)
	permissionGroup.Get("/:id", permissionService.GetPermission)
	permissionGroup.Put("/:id", permissionService.UpdatePermission)
	permissionGroup.Delete("/:id", permissionService.DeletePermission)
}
