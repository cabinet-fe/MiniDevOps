package router

import (
	"minidevops/server/internal/router/auth"
	"minidevops/server/internal/router/config"
	"minidevops/server/internal/router/permission"
	"minidevops/server/internal/router/remote"
	"minidevops/server/internal/router/repository"
	"minidevops/server/internal/router/role"
	"minidevops/server/internal/router/task"
	"minidevops/server/internal/router/user"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRoutes 设置所有路由
func SetupRoutes(app *fiber.App, db *gorm.DB) {
	// API前缀
	api := app.Group("/api/v1")

	// 认证相关路由（无需认证）
	auth.SetupAuthRoutes(api, db)

	// 需要认证的路由组
	protected := api.Group("")
	// TODO: 添加JWT中间件

	// 用户管理路由
	user.SetupUserRoutes(protected, db)

	// 角色管理路由
	role.SetupRoleRoutes(protected, db)

	// 权限管理路由
	permission.SetupPermissionRoutes(protected, db)

	// 仓库管理路由
	repository.SetupRepositoryRoutes(protected, db)

	// 任务管理路由
	task.SetupTaskRoutes(protected, db)

	// 远程服务器路由
	remote.SetupRemoteRoutes(protected, db)

	// 系统配置路由
	config.SetupConfigRoutes(protected, db)
}
