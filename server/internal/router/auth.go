package router

import (
	auth "server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupAuthRoutes 设置认证相关路由
func SetupAuthRoutes(router fiber.Router, db *gorm.DB) {
	authService := auth.NewAuthService(db)

	router.Post("/login", authService.Login)
	router.Post("/logout", authService.Logout)
	router.Get("/profile", authService.GetProfile)
}
