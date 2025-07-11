package router

import (
	auth "server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupLoginRoutes 设置登录路由（无需认证）
func SetupLoginRoutes(router fiber.Router, db *gorm.DB) {
	authService := auth.NewAuthService(db)
	router.Post("/login", authService.Login)
}

// SetupAuthRoutes 设置认证相关路由（需要认证）
func SetupAuthRoutes(router fiber.Router, db *gorm.DB) {
	authService := auth.NewAuthService(db)

	router.Post("/logout", authService.Logout)
	router.Get("/profile", authService.GetProfile)
}
