package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	authhandler "bedrock/internal/auth/handler"
	authmiddleware "bedrock/internal/auth/middleware"
	authrepo "bedrock/internal/auth/repository"
	authservice "bedrock/internal/auth/service"
	cicdhandler "bedrock/internal/cicd/handler"
	cicdrepo "bedrock/internal/cicd/repository"
	cicdservice "bedrock/internal/cicd/service"
	"bedrock/internal/engine"
	"bedrock/internal/middleware"
	"bedrock/internal/pkg"
	"bedrock/internal/platform/config"
	"bedrock/internal/platform/db"
	"bedrock/internal/platform/migration"
	_ "bedrock/internal/platform/migration/migrations"
	"bedrock/internal/platform/seed"
	rbachandler "bedrock/internal/rbac/handler"
	rbacrepo "bedrock/internal/rbac/repository"
	rbacservice "bedrock/internal/rbac/service"
	systemhandler "bedrock/internal/system/handler"
	systemmw "bedrock/internal/system/middleware"
	systemrepo "bedrock/internal/system/repository"
	systemservice "bedrock/internal/system/service"
	"bedrock/internal/ws"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	var logger *zap.Logger
	if gin.Mode() == gin.ReleaseMode {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()
	logger.Info("Bedrock server",
		zap.String("version", version),
		zap.String("db_driver", cfg.Database.Driver),
	)
	logger.Info("database driver change does not migrate data; 2.0 supports fresh install only")
	logger.Info("build scripts execute as the same OS user as Bedrock (no sandbox isolation)")

	if err := pkg.InitEncryption(cfg.Encryption.Key); err != nil {
		logger.Fatal("Failed to init encryption", zap.Error(err))
	}

	gdb, err := db.Open(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to open database", zap.Error(err))
	}

	if err := migration.Up(context.Background(), gdb, migration.Driver(cfg.Database.Driver)); err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
	}
	if err := seed.EnsureSuperAdmin(gdb, cfg.Admin); err != nil {
		logger.Fatal("Failed to seed super-admin", zap.Error(err))
	}
	if err := seed.EnsureRBACResources(gdb); err != nil {
		logger.Fatal("Failed to seed RBAC resources", zap.Error(err))
	}

	userRepo := authrepo.NewUserRepository(gdb)
	roleRepo := rbacrepo.NewRoleRepository(gdb)
	resourceRepo := rbacrepo.NewResourceRepository(gdb)
	dictRepo := systemrepo.NewDictionaryRepository(gdb)
	logRepo := systemrepo.NewOperationLogRepository(gdb)

	permSvc := rbacservice.NewPermissionService(roleRepo, resourceRepo)
	roleSvc := rbacservice.NewRoleService(roleRepo)
	resourceSvc := rbacservice.NewResourceService(resourceRepo)
	userSvc := systemservice.NewUserService(userRepo, roleRepo)
	dictSvc := systemservice.NewDictionaryService(dictRepo)
	auditSvc := systemservice.NewAuditService(logRepo)

	authSvc, err := authservice.NewAuthService(cfg, userRepo, permSvc)
	if err != nil {
		logger.Fatal("Failed to init auth service", zap.Error(err))
	}

	authHandler := authhandler.NewAuthHandler(authSvc)
	userHandler := systemhandler.NewUserHandler(userSvc, permSvc)
	roleHandler := rbachandler.NewRoleHandler(roleSvc, permSvc)
	resourceHandler := rbachandler.NewResourceHandler(resourceSvc, permSvc)
	dictHandler := systemhandler.NewDictionaryHandler(dictSvc, permSvc)
	logHandler := systemhandler.NewOperationLogHandler(auditSvc, permSvc)

	credRepo := cicdrepo.NewCredentialRepository(gdb)
	repoRepo := cicdrepo.NewRepositoryRepository(gdb)
	serverRepo := cicdrepo.NewServerRepository(gdb)
	jobRepo := cicdrepo.NewBuildJobRepository(gdb)
	runRepo := cicdrepo.NewBuildRunRepository(gdb)
	deliveryRepo := cicdrepo.NewWebhookDeliveryRepository(gdb)

	credSvc := cicdservice.NewCredentialService(credRepo)
	repoSvc := cicdservice.NewRepositoryService(repoRepo, credSvc)
	serverSvc := cicdservice.NewServerService(serverRepo, credSvc)
	jobSvc := cicdservice.NewBuildJobService(jobRepo, repoRepo)
	runSvc := cicdservice.NewBuildRunService(runRepo, jobRepo)
	webhookSvc := cicdservice.NewWebhookService(repoRepo, jobRepo, deliveryRepo, runSvc)

	hub := ws.NewHub()
	pipeline := engine.NewPipeline(
		runRepo, jobRepo, repoRepo, serverRepo,
		cicdservice.NewCredentialSecretResolver(credSvc),
		hub, logger,
		cfg.Build.WorkspaceDir, cfg.Build.ArtifactDir, cfg.Build.LogDir, cfg.Build.CacheDir,
	)
	sched := engine.NewScheduler(cfg.Build.MaxConcurrent, pipeline, runRepo, logger)
	runSvc.SetScheduler(sched)
	cronSched := engine.NewCronScheduler(jobRepo, runRepo, runSvc, sched, logger)
	jobSvc.SetCron(cronSched)

	credHandler := cicdhandler.NewCredentialHandler(credSvc, permSvc)
	repoHandler := cicdhandler.NewRepositoryHandler(repoSvc, permSvc)
	serverHandler := cicdhandler.NewServerHandler(serverSvc, permSvc)
	jobHandler := cicdhandler.NewBuildJobHandler(jobSvc, runSvc, permSvc)
	runHandler := cicdhandler.NewBuildRunHandler(runSvc, permSvc)
	webhookHandler := cicdhandler.NewWebhookHandler(webhookSvc)

	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	corsCfg := middleware.DefaultCORSConfig()
	r.Use(middleware.CORSGin(corsCfg))

	api := r.Group("/api/v1")
	api.Use(systemmw.AuditWrite(auditSvc))
	authMW := authmiddleware.Auth(authSvc)
	authHandler.RegisterRoutes(api, authMW)
	userHandler.RegisterRoutes(api, authMW)
	roleHandler.RegisterRoutes(api, authMW)
	resourceHandler.RegisterRoutes(api, authMW)
	dictHandler.RegisterRoutes(api, authMW)
	logHandler.RegisterRoutes(api, authMW)
	credHandler.RegisterRoutes(api, authMW)
	repoHandler.RegisterRoutes(api, authMW)
	serverHandler.RegisterRoutes(api, authMW)
	jobHandler.RegisterRoutes(api, authMW)
	runHandler.RegisterRoutes(api, authMW)
	webhookHandler.RegisterRoutes(api)

	api.GET("/health", func(c *gin.Context) {
		pkg.Success(c, gin.H{
			"status":  "ok",
			"version": version,
			"driver":  cfg.Database.Driver,
		})
	})

	wsHandler := cicdhandler.NewWSHandler(authSvc, permSvc, runSvc, hub, corsCfg)
	wsHandler.RegisterRoutes(r)

	serveSPA(r, cfg.Encryption.Key)

	for _, dir := range []string{cfg.Build.WorkspaceDir, cfg.Build.ArtifactDir, cfg.Build.LogDir, cfg.Build.CacheDir} {
		if dir != "" {
			_ = os.MkdirAll(dir, 0o755)
		}
	}

	sched.Start()
	if err := sched.RecoverOnStartup(); err != nil {
		logger.Error("scheduler recovery failed", zap.Error(err))
	}
	if err := cronSched.Start(); err != nil {
		logger.Error("cron start failed", zap.Error(err))
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		logger.Info("listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down...")

	cronSched.Stop()
	sched.Shutdown()
	hub.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("HTTP server forced shutdown", zap.Error(err))
	}
	if sqlDB, err := gdb.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
