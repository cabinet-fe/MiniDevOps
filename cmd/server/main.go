package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"buildflow/internal/config"
	"buildflow/internal/engine"
	"buildflow/internal/handler"
	"buildflow/internal/middleware"
	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
	"buildflow/internal/service"
	"buildflow/internal/ws"
)

//go:embed all:dist
var webFS embed.FS

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Load config
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Init logger
	var logger *zap.Logger
	if gin.Mode() == gin.ReleaseMode {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// Init encryption
	if err := pkg.InitEncryption(cfg.Encryption.Key); err != nil {
		logger.Fatal("Failed to init encryption", zap.Error(err))
	}

	// Init database
	db, err := model.InitDB()
	if err != nil {
		logger.Fatal("Failed to init database", zap.Error(err))
	}

	// Init repositories
	userRepo := repository.NewUserRepository(db)
	serverRepo := repository.NewServerRepository(db)
	projectRepo := repository.NewProjectRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	buildRepo := repository.NewBuildRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Init services
	authService, err := service.NewAuthService(cfg)
	if err != nil {
		logger.Fatal("Failed to init auth service", zap.Error(err))
	}
	userService := service.NewUserService(userRepo)
	serverService := service.NewServerService(serverRepo, envRepo)
	projectService := service.NewProjectService(projectRepo, envRepo, buildRepo)
	buildService := service.NewBuildService(buildRepo, projectRepo, envRepo, userRepo)
	notifService := service.NewNotificationService(notifRepo)
	auditService := service.NewAuditService(auditRepo)

	// Init WebSocket hub
	hub := ws.NewHub()

	// Init build pipeline and scheduler
	pipeline := engine.NewPipeline(
		buildRepo, projectRepo, envRepo, serverRepo, notifRepo,
		hub, logger,
		cfg.Build.WorkspaceDir, cfg.Build.ArtifactDir, cfg.Build.LogDir,
	)
	scheduler := engine.NewScheduler(cfg.Build.MaxConcurrent, pipeline, logger)
	scheduler.Start()

	// Init cron scheduler for timed builds
	cronScheduler := engine.NewCronScheduler(envRepo, buildRepo, scheduler, logger)
	if err := cronScheduler.Start(); err != nil {
		logger.Error("Failed to start cron scheduler", zap.Error(err))
	}

	// Init handlers
	authHandler := handler.NewAuthHandler(userService, authService)
	userHandler := handler.NewUserHandler(userService)
	serverHandler := handler.NewServerHandler(serverService)
	projectHandler := handler.NewProjectHandler(projectService, cronScheduler)
	buildHandler := handler.NewBuildHandler(buildService, scheduler)
	webhookHandler := handler.NewWebhookHandler(projectService, buildService, envRepo, scheduler)
	notifHandler := handler.NewNotificationHandler(notifService)
	systemHandler := handler.NewSystemHandler(auditService)
	wsHandler := handler.NewWSHandler(authService, buildRepo, projectRepo, hub)

	// Setup Gin
	r := gin.Default()
	r.Use(middleware.CORSGin(middleware.CORSConfig{}))

	// API routes
	api := r.Group("/api/v1")
	{
		// Auth (public)
		api.POST("/auth/login", authHandler.Login)

		// Auth (authenticated)
		auth := api.Group("", middleware.Auth(authService))
		{
			auth.POST("/auth/logout", authHandler.Logout)
			auth.POST("/auth/refresh", authHandler.Refresh)
			auth.GET("/auth/me", authHandler.Me)
			auth.PUT("/auth/profile", authHandler.UpdateProfile)

			// Users (admin only)
			users := auth.Group("/users", middleware.RequireRole("admin"))
			{
				users.GET("", userHandler.List)
				users.POST("", userHandler.Create)
				users.GET("/:id", userHandler.GetByID)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			// Servers
			auth.GET("/servers", serverHandler.List)
			servers := auth.Group("/servers", middleware.RequireRole("ops", "admin"))
			{
				servers.POST("", serverHandler.Create)
				servers.GET("/:id", serverHandler.GetByID)
				servers.PUT("/:id", serverHandler.Update)
				servers.DELETE("/:id", serverHandler.Delete)
				servers.POST("/:id/test", serverHandler.TestConnection)
			}

			// Projects
			auth.GET("/projects", projectHandler.List)
			auth.POST("/projects", projectHandler.Create)
			auth.GET("/projects/:id", projectHandler.GetByID)
			auth.PUT("/projects/:id", projectHandler.Update)
			auth.DELETE("/projects/:id", projectHandler.Delete)
			auth.GET("/projects/:id/export", middleware.RequireRole("admin"), projectHandler.Export)
			auth.POST("/projects/import", middleware.RequireRole("admin"), projectHandler.Import)

			// Environments
			auth.GET("/projects/:id/envs", projectHandler.ListEnvironments)
			auth.POST("/projects/:id/envs", projectHandler.CreateEnvironment)
			auth.PUT("/projects/:id/envs/:envId", projectHandler.UpdateEnvironment)
			auth.DELETE("/projects/:id/envs/:envId", projectHandler.DeleteEnvironment)

			// Builds
			auth.GET("/builds", buildHandler.ListAll)
			auth.GET("/projects/:id/builds", buildHandler.ListByProject)
			auth.POST("/projects/:id/builds", buildHandler.TriggerBuild)
			auth.GET("/builds/:id", buildHandler.GetByID)
			auth.GET("/builds/:id/log", buildHandler.GetLog)
			auth.POST("/builds/:id/cancel", buildHandler.Cancel)
			auth.POST("/builds/:id/deploy", buildHandler.Deploy)
			auth.GET("/builds/:id/artifact", buildHandler.DownloadArtifact)
			auth.POST("/builds/:id/rollback", middleware.RequireRole("ops", "admin"), buildHandler.Rollback)
			auth.POST("/builds/:id/retry", buildHandler.Retry)

			// Dashboard
			auth.GET("/dashboard/stats", buildHandler.DashboardStats)
			auth.GET("/dashboard/active-builds", buildHandler.DashboardActiveBuilds)
			auth.GET("/dashboard/recent-builds", buildHandler.DashboardRecentBuilds)
			auth.GET("/dashboard/trend", buildHandler.DashboardTrend)

			// Notifications
			auth.GET("/notifications", notifHandler.List)
			auth.PUT("/notifications/:id/read", notifHandler.MarkRead)
			auth.PUT("/notifications/read-all", notifHandler.MarkAllRead)

			// System (admin/ops)
			auth.GET("/system/audit-logs", middleware.RequireRole("admin", "ops"), systemHandler.AuditLogs)
			auth.POST("/system/backup", middleware.RequireRole("admin"), systemHandler.Backup)
			auth.POST("/system/restore", middleware.RequireRole("admin"), systemHandler.Restore)
		}

		// Webhook (public, secret-verified)
		api.POST("/webhook/:projectId/:secret", webhookHandler.Handle)
	}

	// Audit middleware on state-changing routes
	r.Use(middleware.Audit(db))

	// WebSocket routes
	r.GET("/ws/builds/:id/logs", wsHandler.HandleBuildLogs)
	r.GET("/ws/notifications", wsHandler.HandleNotifications)

	// Serve frontend SPA
	serveSPA(r)

	// Create data directories
	os.MkdirAll(cfg.Build.WorkspaceDir, 0755)
	os.MkdirAll(cfg.Build.ArtifactDir, 0755)
	os.MkdirAll(cfg.Build.LogDir, 0755)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Info("BuildFlow starting", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")
	cronScheduler.Stop()
	scheduler.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced shutdown", zap.Error(err))
	}
	logger.Info("BuildFlow stopped")
}

func serveSPA(r *gin.Engine) {
	distFS, err := fs.Sub(webFS, "dist")
	if err != nil {
		return
	}
	staticServer := http.FileServer(http.FS(distFS))

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws/") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
			return
		}

		trimmedPath := strings.TrimPrefix(path, "/")
		if trimmedPath == "" {
			c.Request.URL.Path = "/"
			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Serve real static assets directly; fallback all other routes to SPA entry.
		if fileInfo, err := fs.Stat(distFS, trimmedPath); err == nil && !fileInfo.IsDir() {
			c.Request.URL.Path = path
			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Request.URL.Path = "/"
		staticServer.ServeHTTP(c.Writer, c.Request)
	})
}
