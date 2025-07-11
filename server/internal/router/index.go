package router

import (
	"server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRoutes 设置所有路由
func SetupRoutes(app *fiber.App, db *gorm.DB) {
	// API前缀
	api := app.Group("/api/v1")

	// 登录路由（无需认证）
	SetupLoginRoutes(api, db)

	// 需要认证的路由组
	protected := api.Group("", utils.JWTMiddleware())

	// 认证相关路由
	SetupAuthRoutes(protected, db)

	// 用户管理路由
	SetupUserRoutes(protected, db)

	// 角色管理路由
	SetupRoleRoutes(protected, db)

	// 权限管理路由
	SetupPermissionRoutes(protected, db)

	// 仓库管理路由
	SetupRepositoryRoutes(protected, db)

	// 任务管理路由
	SetupTaskRoutes(protected, db)

	// 远程服务器路由
	SetupRemoteRoutes(protected, db)

	// 系统配置路由
	SetupConfigRoutes(protected, db)
}
