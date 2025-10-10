package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/configs"
	"github.com/mohammadgholizadeh/user-management/internal/controller"
	"github.com/mohammadgholizadeh/user-management/internal/middleware"
	"github.com/mohammadgholizadeh/user-management/internal/service"
	"github.com/mohammadgholizadeh/user-management/internal/storage"
	db "github.com/mohammadgholizadeh/user-management/pkg/database"
	"github.com/mohammadgholizadeh/user-management/pkg/jwt"
	logger "github.com/mohammadgholizadeh/user-management/pkg/logger"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Logger
		logger := logger.NewLogger()
		defer func() { _ = logger.Sync() }()
		zap.ReplaceGlobals(logger)

		// Config
		cfg := configs.LoadConfig(CfgPath())

		// Database
		pool, err := db.NewPostgresPool(cfg.DB, logger)
		if err != nil {
			logger.Fatal("postgres connection failed", zap.Error(err))
		}
		defer pool.Close()

		// Redis
		rc := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
		if err := rc.Ping(context.Background()).Err(); err != nil {
			logger.Fatal("redis connection failed", zap.Error(err))
		}

		// JWT
		jwtMgr := jwt.NewManager(cfg.JWT.SecretKey)
		exp := time.Duration(cfg.JWT.ExpirationHours) * time.Hour

		// Storage
		userStore := storage.NewUserStore(pool, logger)
		adminStore := storage.NewAdminStorage(pool, logger)

		// Services
		userSvc := service.NewUserService(userStore, logger)
		adminSvc := service.NewAdminService(adminStore)
		authSvc := service.NewAuthenticationService(userSvc, logger, rc, exp, jwtMgr)

		e := echo.New()

		// Middlewares
		lm := middleware.NewLoggerMiddleware(logger)
		bm := middleware.NewRequestBodyMiddleware(BodyLimit())
		tm := middleware.NewTokenMiddleware(jwtMgr, logger)

		v1 := e.Group("/api/v1", lm.Log)

		swaggerPath := SwaggerPath()
		v1.GET("/swagger.json", func(c echo.Context) error {
			return c.File(filepath.Clean(swaggerPath))
		})

		// Public routes
		pub := v1.Group("", bm.LimitBodySize)
		controller.AuthenticationPublicRoutes(pub, authSvc, logger)

		// Protected routes
		prot := v1.Group("", tm.ParseAndVerify, bm.LimitBodySize)
		controller.AuthenticationProtectedRoutes(prot, authSvc, logger)
		controller.UserRoutes(prot, userSvc, logger)
		controller.AdminRoutes(prot, adminSvc, logger)

		addr := ":" + cfg.Server.Port
		if PortOverride() != "" {
			addr = ":" + PortOverride()
		}
		logger.Info("starting http server", zap.String("addr", addr))
		if err := e.Start(addr); err != nil {
			logger.Fatal("server stopped", zap.Error(err))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
