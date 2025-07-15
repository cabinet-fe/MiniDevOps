package router

import (
	"server/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SetupRepositoryRoutes 设置仓库管理路由
func SetupRepositoryRoutes(router fiber.Router, db *gorm.DB) {
	repositoryService := service.NewRepositoryService(db)
	repoGroup := router.Group("/repositories")

	repoGroup.Get("/", repositoryService.GetRepositories)
	repoGroup.Post("/", repositoryService.CreateRepository)
	repoGroup.Get("/page", repositoryService.GetRepositoryPage)
	repoGroup.Get("/:id", repositoryService.GetRepository)
	repoGroup.Put("/:id", repositoryService.UpdateRepository)
	repoGroup.Delete("/:id", repositoryService.DeleteRepository)
}
