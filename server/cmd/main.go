package main

import (
	"log"
	"server/internal/db"
	"server/internal/router"
	"server/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 初始化数据库
	database, err := db.InitDB()
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	// 初始化默认配置
	configService := service.NewConfigService(database)
	if err := configService.InitDefaultConfigs(); err != nil {
		log.Fatal("默认配置初始化失败:", err)
	}

	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	// 中间件
	app.Use(logger.New())

	// 设置路由
	router.SetupRoutes(app, database)

	// 启动服务器
	log.Fatal(app.Listen(":8080"))
}
