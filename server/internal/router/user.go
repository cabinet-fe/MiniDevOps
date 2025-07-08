package router

import (
	"minidevops/server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupUserRoutes 设置用户管理路由
func SetupUserRoutes(router fiber.Router, db *gorm.DB) {
	userService := service.NewUserService(db)
	userGroup := router.Group("/users")

	userGroup.Get("/", userService.GetUsers)
	userGroup.Post("/", userService.CreateUser)
	userGroup.Get("/:id", userService.GetUser)
	userGroup.Put("/:id", userService.UpdateUser)
	userGroup.Delete("/:id", userService.DeleteUser)
}
