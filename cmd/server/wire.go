//go:build wireinject

package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/authadapter"
	"github.com/lyimoexiao/akari/internal/closet"
	"github.com/lyimoexiao/akari/internal/closetadapter"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/permission"
	"github.com/lyimoexiao/akari/internal/rbacadapter"
	"github.com/lyimoexiao/akari/internal/requestlog"
	"github.com/lyimoexiao/akari/internal/requestlogadapter"
	"github.com/lyimoexiao/akari/internal/role"
	"github.com/lyimoexiao/akari/internal/router"
	"github.com/lyimoexiao/akari/internal/routeradapter"
	"github.com/lyimoexiao/akari/internal/score"
	"github.com/lyimoexiao/akari/internal/scoreadapter"
	"github.com/lyimoexiao/akari/internal/sign"
	"github.com/lyimoexiao/akari/internal/signadapter"
	"github.com/lyimoexiao/akari/internal/skinlib"
	"github.com/lyimoexiao/akari/internal/textureadapter"
	"github.com/lyimoexiao/akari/internal/user"
	"github.com/lyimoexiao/akari/internal/useradapter"
	"github.com/lyimoexiao/akari/internal/yggdrasil"
	"github.com/lyimoexiao/akari/internal/yggdrasiladapter"
	"github.com/lyimoexiao/akari/pkg/cache"
	"github.com/lyimoexiao/akari/pkg/captcha"
	"github.com/lyimoexiao/akari/pkg/jwt"
	"github.com/lyimoexiao/akari/pkg/logger"
	"github.com/lyimoexiao/akari/pkg/response"
	"github.com/lyimoexiao/akari/pkg/smtp"
	"github.com/lyimoexiao/akari/pkg/util"
	"github.com/lyimoexiao/akari/pkg/version"
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
		ProvideRBACManager,
		ProvidePermissionService,
		ProvidePermissionMiddleware,
		ProvidePermissionHandler,
		ProvideRequestLogRepository,
		ProvideRequestLogService,
		ProvideRequestLogMiddleware,
		ProvideRequestLogHandler,
		ProvideUserService,
		ProvideUserHandler,
		ProvideRoleService,
		ProvideRoleHandler,
		ProvideHealthChecker,
		ProvideYggdrasilKeyManager,
		ProvideYggdrasilService,
		ProvideYggdrasilHandler,
		wire.Struct(new(router.Handlers), "*"),
		ProvideScoreOperator,
		ProvideSignRepository,
		ProvideSignService,
		ProvideSignHandler,
		ProvideScoreHandler,
		ProvideSkinlibService,
		ProvideSkinlibHandler,
		ProvideTextureFileStorage,
		ProvideTextureChecker,
		ProvideClosetService,
		ProvideClosetHandler,
		ProvideTextureServer,
		wire.Struct(new(router.Dependencies), "*"),
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
	if err := yggdrasiladapter.Migrate(db); err != nil {
		_ = database.Close(db)
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

func ProvideMailer(cfg *config.Config, logger *zap.SugaredLogger) (*smtp.Mailer, func(), error) {
	m := smtp.New(&cfg.SMTP, logger)
	cleanup := func() {
		_ = m.Close()
	}
	return m, cleanup, nil
}

func ProvideEngine(
	cfg *config.Config,
	deps *router.Dependencies,
) *gin.Engine {
	setGinMode(cfg.Server.Mode)
	return router.Setup(deps)
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
	roles *role.Service,
	logger *zap.SugaredLogger,
) *auth.Service {
	return auth.NewService(auth.Dependencies{
		Users:  authadapter.NewUserRepository(db),
		Roles:  roles,
		Tokens: authadapter.NewJWTManager(jwtManager),
		Store:  authadapter.NewTokenStore(c, logger),
		Mailer: mailer,
		Settings: auth.Settings{
			RegistrationEnabled:      cfg.Auth.RegistrationEnabled,
			EmailVerificationEnabled: cfg.Auth.EmailVerificationEnabled,
			PasswordResetEnabled:     cfg.Auth.PasswordResetEnabled,
			VerifyEmailTokenTTL:      util.ParseDuration(cfg.Auth.VerifyEmailTokenTTL, 2*time.Hour),
			PasswordResetTokenTTL:    util.ParseDuration(cfg.Auth.PasswordResetTokenTTL, 30*time.Minute),
		},
	})
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

func ProvideRBACManager(db *gorm.DB) (*rbacadapter.Manager, error) {
	return rbacadapter.NewManager(db)
}

func ProvidePermissionService(manager *rbacadapter.Manager) *permission.Service {
	return permission.NewService(manager)
}

func ProvidePermissionMiddleware(svc *permission.Service) *permission.Middleware {
	return permission.NewMiddleware(svc)
}

func ProvidePermissionHandler(svc *permission.Service, middleware *permission.Middleware, logger *zap.SugaredLogger) *permission.Handler {
	return permission.NewHandler(svc, middleware, logger)
}

func ProvideRequestLogRepository(db *gorm.DB) *requestlogadapter.Repository {
	return requestlogadapter.NewRepository(db)
}

func ProvideRequestLogService(repository *requestlogadapter.Repository) *requestlog.Service {
	return requestlog.NewService(repository)
}

func ProvideRequestLogMiddleware(repository *requestlogadapter.Repository, logger *zap.SugaredLogger) gin.HandlerFunc {
	return requestlog.Middleware(repository, logger)
}

func ProvideRequestLogHandler(svc *requestlog.Service, logger *zap.SugaredLogger) *requestlog.Handler {
	return requestlog.NewHandler(svc, logger)
}

func ProvideUserService(db *gorm.DB) *user.Service {
	return user.NewService(user.Dependencies{
		Repository: useradapter.NewRepository(db),
		Clock:      useradapter.Clock{},
		Hasher:     useradapter.PasswordHasher{},
	})
}

func ProvideUserHandler(svc *user.Service, permissions *permission.Middleware, logger *zap.SugaredLogger) *user.Handler {
	return user.NewHandler(svc, permissions, logger)
}

func ProvideRoleService(manager *rbacadapter.Manager) *role.Service {
	return role.NewService(role.Dependencies{Repository: manager, Policies: manager})
}

func ProvideRoleHandler(svc *role.Service, permissions *permission.Middleware, logger *zap.SugaredLogger) *role.Handler {
	return role.NewHandler(svc, permissions, logger)
}

func ProvideScoreOperator(db *gorm.DB) *scoreadapter.Repository {
	return scoreadapter.NewRepository(db)
}

func ProvideSignRepository(db *gorm.DB) *signadapter.Repository {
	return signadapter.NewRepository(db)
}

func ProvideSignService(cfg *config.Config, repo *signadapter.Repository, ops *scoreadapter.Repository) *sign.Service {
	return sign.NewService(sign.Dependencies{
		Repository: repo,
		ScoreOps:   ops,
		Config: sign.SignConfig{
			GapHours:  cfg.Score.SignGapHours,
			ScoreMin:  cfg.Score.SignScoreMin,
			ScoreMax:  cfg.Score.SignScoreMax,
			AfterZero: cfg.Score.SignAfterZero,
		},
		Clock: signadapter.Clock{},
	})
}

func ProvideSignHandler(svc *sign.Service, logger *zap.SugaredLogger) *sign.Handler {
	return sign.NewHandler(svc, logger)
}

func ProvideScoreHandler(ops *scoreadapter.Repository, permissions *permission.Middleware, logger *zap.SugaredLogger) *score.Handler {
	return score.NewHandler(ops, permissions, logger)
}

func ProvideHealthChecker(db *gorm.DB, cacheBackend cache.Cache) router.HealthChecker {
	return routeradapter.NewHealthChecker(db, cacheBackend)
}

func ProvideYggdrasilKeyManager() (*yggdrasiladapter.KeyManager, error) {
	// ponytail: hardcoded to data/; could be config-driven if needed
	return yggdrasiladapter.NewKeyManager("data")
}

func ProvideYggdrasilService(db *gorm.DB, c cache.Cache, cfg *config.Config, logger *zap.SugaredLogger, km *yggdrasiladapter.KeyManager) *yggdrasil.Service {
	return yggdrasil.NewService(yggdrasil.Dependencies{
		Repository: yggdrasiladapter.NewRepository(db),
		Sessions:   yggdrasiladapter.NewSessionStore(c),
		Signer:     km,
		Reporter:   yggdrasiladapter.NewSigningFailureReporter(logger),
		Settings: yggdrasil.Settings{
			EmailVerificationEnabled: cfg.Auth.EmailVerificationEnabled,
			ServerName:               cfg.Yggdrasil.ServerName,
			ImplementationName:       cfg.Yggdrasil.ImplementationName,
			ImplementationVersion:    version.Version,
			TextureBaseURL:           cfg.Server.BaseURL,
		},
	})
}

func ProvideYggdrasilHandler(svc *yggdrasil.Service, logger *zap.SugaredLogger) *yggdrasil.Handler {
	return yggdrasil.NewHandler(svc, logger)
}

func ProvideTextureFileStorage(cfg *config.Config) *textureadapter.FileStorage {
	return textureadapter.NewFileStorage(cfg.Storage.Dir)
}

func ProvideTextureChecker(db *gorm.DB) *textureadapter.TextureChecker {
	return textureadapter.NewTextureChecker(db)
}

func ProvideSkinlibService(
	db *gorm.DB,
	storage *textureadapter.FileStorage,
	scoreOps *scoreadapter.Repository,
	cfg *config.Config,
) *skinlib.Service {
	return skinlib.NewService(skinlib.Dependencies{
		Repository:    textureadapter.NewRepository(db),
		Storage:       storage,
		ScoreOps:      scoreOps,
		ClosetAdder:   closetadapter.NewRepository(db),
		ClosetCleaner: closetadapter.NewRepository(db),
		BaseURL:       cfg.Server.BaseURL,
		AwardUpload:   cfg.Score.AwardPerUpload,
	})
}

func ProvideSkinlibHandler(svc *skinlib.Service, permissions *permission.Middleware, logger *zap.SugaredLogger) *skinlib.Handler {
	return skinlib.NewHandler(svc, permissions, logger)
}

func ProvideClosetService(
	db *gorm.DB,
	textureChecker *textureadapter.TextureChecker,
	scoreOps *scoreadapter.Repository,
	cfg *config.Config,
) *closet.Service {
	return closet.NewService(closet.Dependencies{
		Repository:          closetadapter.NewRepository(db),
		TextureRepo:         textureChecker,
		TextureDeleter:      textureChecker,
		ProfileCleaner:      yggdrasiladapter.NewRepository(db),
		ScoreOps:            scoreOps,
		CostPerItem:         cfg.Score.CostPerClosetItem,
		ReturnScoreOnRemove: cfg.Score.ReturnScoreOnRemove,
		AwardPerLike:        cfg.Score.AwardPerLike,
	})
}

func ProvideClosetHandler(svc *closet.Service, permissions *permission.Middleware, logger *zap.SugaredLogger) *closet.Handler {
	return closet.NewHandler(svc, permissions, logger)
}

func ProvideTextureServer(cfg *config.Config) router.TextureFileHandler {
	dir := cfg.Storage.Dir
	return func(ctx *gin.Context) {
		hash := ctx.Param("hash")
		if hash == "" {
			response.BadRequest(ctx, "缺少 hash 参数")
			return
		}
		if len(hash) < 2 {
			response.BadRequest(ctx, "无效的 hash")
			return
		}

		// Build file path: dir/hash[:2]/hash
		filePath := filepath.Join(dir, hash[:2], hash)
		// Set correct content type for Minecraft client
		ctx.Header("Content-Type", "image/png")
		// Add CORS headers for cross-origin texture loading
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.File(filePath)
	}
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
