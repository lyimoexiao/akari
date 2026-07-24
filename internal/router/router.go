package router

import (
	"io/fs"
	"log"
	"mime"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/logger"
	"github.com/lyimoexiao/akari/web"
	"gorm.io/gorm"
)

func Setup(db *gorm.DB, c cache.Cache) *gin.Engine {
	r := gin.New()
	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false
	r.Use(logger.GinLogger(), logger.GinRecovery())

	r.Use(func(ctx *gin.Context) {
		ctx.Set("db", db)
		ctx.Set("cache", c)
		ctx.Next()
	})

	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	_ = v1

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
			ctx.JSON(404, gin.H{"error": "not found"})
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
		ctx.String(404, "not found")
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