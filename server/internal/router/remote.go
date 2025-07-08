package router

import (
	"minidevops/server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRemoteRoutes 设置远程服务器管理路由
func SetupRemoteRoutes(router fiber.Router, db *gorm.DB) {
	remoteService := service.NewRemoteService(db)
	remoteGroup := router.Group("/remotes")

	remoteGroup.Get("/", remoteService.GetRemotes)
	remoteGroup.Post("/", remoteService.CreateRemote)
	remoteGroup.Get("/:id", remoteService.GetRemote)
	remoteGroup.Put("/:id", remoteService.UpdateRemote)
	remoteGroup.Delete("/:id", remoteService.DeleteRemote)
	remoteGroup.Post("/:id/test", remoteService.TestConnection)
}
