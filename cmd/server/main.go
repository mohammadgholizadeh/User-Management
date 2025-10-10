package main

import (
	"context"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/user-management/configs"
	db "github.com/mohammadgholizadeh/user-management/pkg/database"
	"github.com/mohammadgholizadeh/user-management/pkg/jwt"
	logger "github.com/mohammadgholizadeh/user-management/pkg/logger"

	"github.com/mohammadgholizadeh/user-management/internal/controller"
	"github.com/mohammadgholizadeh/user-management/internal/middleware"
	"github.com/mohammadgholizadeh/user-management/internal/service"
	"github.com/mohammadgholizadeh/user-management/internal/storage"
)

func main() {
	// Logger
	logger := logger.NewLogger()
	defer func() { _ = logger.Sync() }()
	zap.ReplaceGlobals(logger)

	// Config
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}
	cfg := configs.LoadConfig(cfgPath)

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
	bm := middleware.NewRequestBodyMiddleware(1048576)
	tm := middleware.NewTokenMiddleware(jwtMgr, logger)

	v1 := e.Group("/api/v1", lm.Log)

	v1.GET("/swagger.json", func(c echo.Context) error {
		return c.File("docs/swagger.json")
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
	logger.Info("starting http server", zap.String("addr", addr))
	if err := e.Start(addr); err != nil {
		logger.Fatal("server stopped", zap.Error(err))
	}
}
