//go:build wireinject

package main

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/logger"
	"github.com/lyimoexiao/akari/internal/router"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Engine *gin.Engine
	Logger *zap.SugaredLogger
	Config *config.Config
}

func Initialize(cfgPath string) (*App, func(), error) {
	wire.Build(
		ProvideConfig,
		ProvideLogger,
		ProvideDB,
		ProvideCache,
		ProvideEngine,
		wire.Struct(new(App), "Config", "Engine", "Logger"),
	)
	return nil, nil, nil
}

func ProvideConfig(cfgPath string) (*config.Config, error) {
	return config.Load(cfgPath)
}

func ProvideLogger(cfg *config.Config) (*zap.SugaredLogger, func(), error) {
	logger.Init(&cfg.Logger)
	cleanup := func() {
		_ = logger.L.Sync()
	}
	return logger.L, cleanup, nil
}

func ProvideDB(cfg *config.Config) (*gorm.DB, func(), error) {
	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = database.Close(db)
	}
	return db, cleanup, nil
}

func ProvideCache(cfg *config.Config) (cache.Cache, func(), error) {
	c := cache.New(&cfg.Cache)
	cleanup := func() {
		_ = c.Close()
	}
	return c, cleanup, nil
}

func ProvideEngine(cfg *config.Config, db *gorm.DB, c cache.Cache) *gin.Engine {
	setGinMode(cfg.Server.Mode)
	return router.Setup(db, c)
}

func setGinMode(mode string) {
	switch strings.ToLower(mode) {
	case "release", "prod", "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}
}
