package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/mattn/go-sqlite3"

	"minidevops/internal/config"
	"minidevops/internal/router"
	"minidevops/internal/db"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	client, err := db.NewClient(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("创建数据库连接失败: %v", err)
	}
	defer client.Close()

	// 确保数据库目录存在
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}

	// 初始化项目和输出目录
	if err := os.MkdirAll("./data/projects", 0755); err != nil {
		log.Fatalf("创建项目目录失败: %v", err)
	}
	if err := os.MkdirAll("./data/output", 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	// 中间件
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// 初始化路由
	router.Setup(app, client)

	// 优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("正在关闭服务...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Fatalf("服务关闭出错: %v", err)
		}
	}()

	// 启动服务
	log.Printf("服务启动在 %s", cfg.ServerAddr)
	if err := app.Listen(cfg.ServerAddr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}