package permission

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/pkg/response"
)

type Middleware struct {
	authorizer interface {
		EnforceUser(context.Context, Check) (bool, string, error)
	}
}

func NewMiddleware(authorizer interface {
	EnforceUser(context.Context, Check) (bool, string, error)
}) *Middleware {
	return &Middleware{authorizer: authorizer}
}

func (middleware *Middleware) Require() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := auth.GetUserID(ctx)
		if userID == 0 {
			response.Unauthorized(ctx, "需要认证")
			ctx.Abort()
			return
		}

		allowed, roleName, err := middleware.authorizer.EnforceUser(ctx.Request.Context(), Check{
			UserID: userID,
			Object: ctx.Request.URL.Path,
			Action: ctx.Request.Method,
		})
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				response.Unauthorized(ctx, err.Error())
			} else {
				response.InternalError(ctx, "权限检查失败")
			}
			ctx.Abort()
			return
		}

		ctx.Set(auth.CtxKeyRole, roleName)
		if !allowed {
			response.Forbidden(ctx, "权限不足")
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// HasPermission checks whether the current user has a specific permission identifier.
// Unlike Require(), this does not abort the request — it's meant for use inside handlers
// where behavior varies by permission level (e.g., user vs staff).
func (middleware *Middleware) HasPermission(ctx context.Context, userID uint, object, action string) (bool, string, error) {
	return middleware.authorizer.EnforceUser(ctx, Check{
		UserID: userID,
		Object: object,
		Action: action,
	})
}
