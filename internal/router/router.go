package router

import (
	"io/fs"
	"log"
	"mime"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/admin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/captcha"
	"github.com/lyimoexiao/akari/internal/logger"
	"github.com/lyimoexiao/akari/internal/middleware"
	"github.com/lyimoexiao/akari/internal/response"
	"github.com/lyimoexiao/akari/internal/yggdrasil"
	"github.com/lyimoexiao/akari/web"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, c cache.Cache, captchaSvc *captcha.Service, authHandler *auth.Handler, adminHandler *admin.Handler, yggdrasilHandler *yggdrasil.Handler, authMw *auth.Middleware, l *zap.SugaredLogger) *gin.Engine {
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.Use(middleware.TraceID())
	r.Use(logger.GinLogger(), logger.GinRecovery())

	r.Use(func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Set("cache", c)
		ctx.Next()
	})

	r.GET("/health", func(ctx *gin.Context) {
		response.Success(ctx, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	// Captcha routes (only added if enabled)
	if captchaSvc != nil && captchaSvc.IsEnabled() {
		h := captcha.NewHandler(captchaSvc)
		v1.GET("/captcha", h.Generate)
		v1.POST("/captcha/verify", h.Verify)
	}

	// Auth routes
	if authHandler != nil {
		authHandler.RegisterRoutes(v1)
	} else {
		l.Warn("auth handler is nil, auth routes not registered")
	}

	// Admin routes
	if adminHandler != nil {
		adminHandler.RegisterRoutes(v1)
	} else {
		l.Warn("admin handler is nil, admin routes not registered")
	}

	// Yggdrasil API
	if yggdrasilHandler != nil {
		yg := v1.Group("/yggdrasil")
		yggdrasilHandler.RegisterRoutes(yg)
		// User status endpoint uses app JWT auth, not Yggdrasil token auth
		if authMw != nil {
			yg.GET("/user/status", authMw.RequireAuth(), yggdrasilHandler.UserStatus)
		}
	} else {
		l.Warn("yggdrasil handler is nil, yggdrasil routes not registered")
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
