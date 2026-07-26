package router

import (
	"context"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/captcha"
	"github.com/lyimoexiao/akari/internal/logger"
	"github.com/lyimoexiao/akari/internal/middleware"
	"github.com/lyimoexiao/akari/internal/permission"
	"github.com/lyimoexiao/akari/internal/requestlog"
	"github.com/lyimoexiao/akari/internal/response"
	"github.com/lyimoexiao/akari/internal/role"
	"github.com/lyimoexiao/akari/internal/user"
	"github.com/lyimoexiao/akari/internal/yggdrasil"
	"github.com/lyimoexiao/akari/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Auth       *auth.Handler
	User       *user.Handler
	Role       *role.Handler
	Permission *permission.Handler
	RequestLog *requestlog.Handler
	Yggdrasil  *yggdrasil.Handler
}

type Dependencies struct {
	Health         HealthChecker
	RequestLogging gin.HandlerFunc
	Captcha        *captcha.Service
	Handlers       *Handlers
	Auth           *auth.Middleware
	Logger         *zap.SugaredLogger
}

func Setup(deps *Dependencies) *gin.Engine {
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.Use(middleware.RequestID())
	if deps.RequestLogging != nil {
		r.Use(deps.RequestLogging)
	}
	r.Use(logger.GinRecovery())

	r.GET("/health", func(ctx *gin.Context) {
		probeContext, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		defer cancel()

		if deps.Health == nil || deps.Health.Check(probeContext) != nil {
			response.Error(ctx, http.StatusServiceUnavailable, "service unavailable")
			return
		}
		response.Success(ctx, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// Captcha routes (only added if enabled)
	if deps.Captcha != nil && deps.Captcha.IsEnabled() {
		h := captcha.NewHandler(deps.Captcha)
		v1.GET("/captcha", h.Generate)
		v1.POST("/captcha/verify", h.Verify)
		v1.POST("/captcha/turnstile-verify", h.TurnstileVerify)
	}

	// Auth routes
	if deps.Handlers.Auth != nil {
		deps.Handlers.Auth.RegisterRoutes(v1)
	} else {
		deps.Logger.Warn("auth handler is nil, auth routes not registered")
	}

	if deps.Auth != nil {
		protected := v1.Group("")
		protected.Use(deps.Auth.RequireAuth())
		if deps.Handlers.User != nil {
			deps.Handlers.User.RegisterRoutes(protected)
		}
		if deps.Handlers.Role != nil {
			deps.Handlers.Role.RegisterRoutes(protected)
		}
		if deps.Handlers.Permission != nil {
			deps.Handlers.Permission.RegisterRoutes(protected)
		}
		if deps.Handlers.RequestLog != nil && deps.Handlers.Permission != nil {
			requestLogs := protected.Group("")
			requestLogs.Use(deps.Handlers.Permission.Require())
			deps.Handlers.RequestLog.RegisterRoutes(requestLogs)
		}
	} else {
		deps.Logger.Warn("auth middleware is nil, protected routes not registered")
	}

	// Yggdrasil API
	if deps.Handlers.Yggdrasil != nil {
		yg := v1.Group("/yggdrasil")
		deps.Handlers.Yggdrasil.RegisterRoutes(yg)
		// User status endpoint uses app JWT auth, not Yggdrasil token auth
		if deps.Auth != nil && deps.Handlers.Permission != nil {
			yg.GET("/user/status", deps.Auth.RequireAuth(), deps.Handlers.Permission.Require(), deps.Handlers.Yggdrasil.UserStatus)
		}
	} else {
		deps.Logger.Warn("yggdrasil handler is nil, yggdrasil routes not registered")
	}

	// Yggdrasil ALI (API Location Indication): serve header on all frontend paths
	// so authlib-injector clients can auto-discover the API root.
	aliPath := "/api/v1/yggdrasil/"
	aliHeader := "X-Authlib-Injector-API-Location"
	r.Use(func(ctx *gin.Context) {
		// Only on non-API paths (SPA / static files)
		if !strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.Header(aliHeader, aliPath)
		}
		ctx.Next()
	})

	serveSPA(r)

	return r
}

func serveSPA(r *gin.Engine) {
	sub, err := fs.Sub(web.FS, "dist")
	if err != nil {
		log.Printf("serveSPA: embedded FS not available: %v", err)
		return
	}

	// Middleware: intercept root "/" EARLY, before Gin's router
	// performs its radix-tree redirect logic. Gin's root node
	// (which exists due to /health etc.) can return 301 for "/"
	// even with RedirectTrailingSlash=false.
	r.Use(func(ctx *gin.Context) {
		if ctx.Request.URL.Path != "/" {
			ctx.Next()
			return
		}
		serveEmbeddedFile(ctx, sub, "index.html")
		ctx.Abort()
	})

	// NoRoute handles all other unmatched paths (static files + SPA fallback).
	r.NoRoute(func(ctx *gin.Context) {
		filePath := strings.TrimPrefix(ctx.Request.URL.Path, "/")

		// Do not hijack API routes
		if strings.HasPrefix(filePath, "api/") {
			response.NotFound(ctx, "路由未找到")
			return
		}

		// Try to serve the exact file from embedded FS
		if filePath != "" {
			if data, readErr := fs.ReadFile(sub, filePath); readErr == nil {
				writeFileResponse(ctx, filePath, data)
				return
			}
		}

		// Fallback to index.html for SPA client-side routing
		serveEmbeddedFile(ctx, sub, "index.html")
	})
}

// serveEmbeddedFile reads a file from the embedded FS and writes it
// directly to the response. Avoids http.FS / FileFromFS to eliminate
// any risk of Gin/Golang redirect behavior.
func serveEmbeddedFile(ctx *gin.Context, sub fs.FS, name string) {
	data, err := fs.ReadFile(sub, name)
	if err != nil {
		ctx.String(404, "未找到")
		return
	}
	writeFileResponse(ctx, name, data)
}

func writeFileResponse(ctx *gin.Context, name string, data []byte) {
	ctype := mime.TypeByExtension(filepath.Ext(name))
	if ctype == "" {
		ctype = "application/octet-stream"
	}
	ctx.Data(200, ctype, data)
}
