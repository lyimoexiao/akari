//go:build wireinject

package main

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lyimoexiao/akari/internal/admin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/captcha"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/logger"
	"github.com/lyimoexiao/akari/internal/router"
	"github.com/lyimoexiao/akari/internal/smtp"
	"github.com/lyimoexiao/akari/internal/yggdrasil"
	"github.com/lyimoexiao/akari/pkg/jwt"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Engine     *gin.Engine
	Logger     *zap.SugaredLogger
	Config     *config.Config
	Mailer     *smtp.Mailer
	CaptchaSvc *captcha.Service
}

func Initialize(cfgPath string) (*App, func(), error) {
	wire.Build(
		ProvideConfig,
		ProvideLogger,
		ProvideDB,
		ProvideCache,
		ProvideMailer,
		ProvideCaptcha,
		ProvideJWTManager,
		ProvideAuthService,
		ProvideAuthHandler,
		ProvideAuthMiddleware,
		ProvideAdminService,
		ProvideAdminHandler,
		ProvideYggdrasilKeyManager,
		ProvideYggdrasilService,
		ProvideYggdrasilHandler,
		ProvideEngine,
		wire.Struct(new(App), "Config", "Engine", "Logger", "Mailer", "CaptchaSvc"),
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

func ProvideMailer(cfg *config.Config) (*smtp.Mailer, func(), error) {
	m := smtp.New(&cfg.SMTP)
	cleanup := func() {
		_ = m.Close()
	}
	return m, cleanup, nil
}

func ProvideEngine(
	cfg *config.Config,
	db *gorm.DB,
	c cache.Cache,
	captchaSvc *captcha.Service,
	authHandler *auth.Handler,
	adminHandler *admin.Handler,
	yggdrasilHandler *yggdrasil.Handler,
	authMw *auth.Middleware,
	logger *zap.SugaredLogger,
) *gin.Engine {
	setGinMode(cfg.Server.Mode)
	return router.Setup(db, c, captchaSvc, authHandler, adminHandler, yggdrasilHandler, authMw, logger)
}

func ProvideCaptcha(cfg *config.Config, c cache.Cache) *captcha.Service {
	return captcha.New(&cfg.Captcha, c)
}

func ProvideJWTManager(cfg *config.Config) *jwt.Manager {
	exp, _ := time.ParseDuration(cfg.JWT.Expiration)
	return jwt.New(&jwt.Config{
		Secret:     cfg.JWT.Secret,
		Issuer:     cfg.JWT.Issuer,
		Expiration: exp,
	})
}

func ProvideAuthService(
	db *gorm.DB,
	c cache.Cache,
	jwtManager *jwt.Manager,
	mailer *smtp.Mailer,
	cfg *config.Config,
) *auth.Service {
	return auth.NewService(db, c, jwtManager, mailer, &cfg.Auth, &cfg.JWT)
}

func ProvideAuthHandler(
	svc *auth.Service,
	mw *auth.Middleware,
	cfg *config.Config,
	captchaSvc *captcha.Service,
	logger *zap.SugaredLogger,
) *auth.Handler {
	return auth.NewHandler(svc, mw, &auth.HandlerConfig{
		RegistrationEnabled:      cfg.Auth.RegistrationEnabled,
		EmailVerificationEnabled: cfg.Auth.EmailVerificationEnabled,
	}, captchaSvc, logger)
}

func ProvideAuthMiddleware(svc *auth.Service) *auth.Middleware {
	return auth.NewMiddleware(svc)
}

func ProvideAdminService(db *gorm.DB) *admin.Service {
	return admin.NewService(db)
}

func ProvideAdminHandler(svc *admin.Service, mw *auth.Middleware, logger *zap.SugaredLogger) *admin.Handler {
	return admin.NewHandler(svc, mw, logger)
}

func ProvideYggdrasilKeyManager() (*yggdrasil.KeyManager, error) {
	// ponytail: hardcoded to data/; could be config-driven if needed
	return yggdrasil.NewKeyManager("data")
}

func ProvideYggdrasilService(db *gorm.DB, c cache.Cache, cfg *config.Config, logger *zap.SugaredLogger, km *yggdrasil.KeyManager) *yggdrasil.Service {
	return yggdrasil.NewService(db, c, &cfg.Auth, logger, km, &cfg.Yggdrasil)
}

func ProvideYggdrasilHandler(svc *yggdrasil.Service, logger *zap.SugaredLogger) *yggdrasil.Handler {
	return yggdrasil.NewHandler(svc, logger)
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
