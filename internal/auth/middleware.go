package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/response"
)

const (
	// Context keys
	CtxKeyUserID   = "auth_user_id"
	CtxKeyUsername = "auth_username"
	CtxKeyRole     = "auth_role"
)

// Middleware provides auth-related Gin middleware.
type Middleware struct {
	svc *Service
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware(svc *Service) *Middleware {
	return &Middleware{svc: svc}
}

// RequireAuth requires a valid JWT token to access the route.
// It extracts the token from the Authorization header (Bearer scheme),
// validates it, and sets user info in the Gin context.
func (m *Middleware) RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := extractBearerToken(ctx)
		if tokenString == "" {
			response.Unauthorized(ctx, "缺少认证令牌")
			ctx.Abort()
			return
		}

		claims, err := m.svc.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(ctx, "令牌无效或已过期")
			ctx.Abort()
			return
		}

		// Check token blacklist (logout)
		if m.svc.IsTokenBlacklisted(ctx.Request.Context(), tokenString) {
			response.Unauthorized(ctx, "令牌已被吊销")
			ctx.Abort()
			return
		}

		// Set user info in context
		ctx.Set(CtxKeyUserID, claims.UserID)
		ctx.Set(CtxKeyUsername, claims.Username)
		ctx.Set(CtxKeyRole, claims.Role)

		ctx.Next()
	}
}

// RequireEmailVerified requires the authenticated user to have verified
// their email address. If email verification is disabled, this check is skipped.
func (m *Middleware) RequireEmailVerified() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, exists := ctx.Get(CtxKeyUserID)
		if !exists {
			response.Unauthorized(ctx, "需要认证")
			ctx.Abort()
			return
		}

		uid, ok := userID.(uint)
		if !ok {
			response.Forbidden(ctx, "用户数据无效")
			ctx.Abort()
			return
		}

		verified, err := m.svc.HasVerifiedEmail(ctx.Request.Context(), uid)
		if err != nil {
			response.InternalError(ctx, "检查邮箱验证状态失败")
			ctx.Abort()
			return
		}

		if !verified {
			response.Forbidden(ctx, "需要邮箱验证")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// GetUserID extracts the authenticated user ID from the Gin context.
func GetUserID(ctx *gin.Context) uint {
	id, exists := ctx.Get(CtxKeyUserID)
	if !exists {
		return 0
	}
	uid, ok := id.(uint)
	if !ok {
		return 0
	}
	return uid
}

// GetRole extracts the authenticated user role from the Gin context.
func GetRole(ctx *gin.Context) string {
	role, exists := ctx.Get(CtxKeyRole)
	if !exists {
		return ""
	}
	r, ok := role.(string)
	if !ok {
		return ""
	}
	return r
}

// extractBearerToken extracts a Bearer token from the Authorization header.
func extractBearerToken(ctx *gin.Context) string {
	auth := ctx.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
