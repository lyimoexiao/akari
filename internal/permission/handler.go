package permission

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/response"
	"go.uber.org/zap"
)

type Handler struct {
	svc        *Service
	middleware *Middleware
	logger     *zap.SugaredLogger
}

func NewHandler(svc *Service, middleware *Middleware, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, middleware: middleware, logger: logger}
}

func (h *Handler) RegisterRoutes(routes *gin.RouterGroup) {
	routes.GET("/auth/permission", h.CurrentUser)
	routes.GET("/permissions", h.middleware.Require(), h.List)
}

func (h *Handler) Require() gin.HandlerFunc {
	return h.middleware.Require()
}

func (h *Handler) List(ctx *gin.Context) {
	snapshot, err := h.svc.Snapshot()
	if err != nil {
		h.logger.Errorw("permission list failed", "error", err)
		response.InternalError(ctx, "获取权限列表失败")
		return
	}
	response.Success(ctx, snapshot)
}

func (h *Handler) CurrentUser(ctx *gin.Context) {
	identifiers, err := h.svc.IdentifiersForUser(ctx.Request.Context(), auth.GetUserID(ctx))
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.Unauthorized(ctx, err.Error())
			return
		}
		h.logger.Errorw("current user permission list failed", "error", err)
		response.InternalError(ctx, "获取当前用户权限失败")
		return
	}
	response.Success(ctx, IdentifierList{Permissions: identifiers})
}
