package role

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/response"
	"go.uber.org/zap"
)

type RouteGuard interface {
	Require() gin.HandlerFunc
}

type Handler struct {
	svc         *Service
	permissions RouteGuard
	logger      *zap.SugaredLogger
}

func NewHandler(svc *Service, permissions RouteGuard, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, permissions: permissions, logger: logger}
}

func (h *Handler) RegisterRoutes(routes *gin.RouterGroup) {
	routes.GET("/roles", h.permissions.Require(), h.List)
	routes.POST("/roles", h.permissions.Require(), h.Create)
	routes.PUT("/roles/:id", h.permissions.Require(), h.Update)
	routes.DELETE("/roles/:id", h.permissions.Require(), h.Delete)
	routes.PUT("/roles/:id/default", h.permissions.Require(), h.SetDefault)
	routes.POST("/roles/assign", h.permissions.Require(), h.SetRole)
}

func (h *Handler) List(ctx *gin.Context) {
	roles, err := h.svc.List(ctx.Request.Context())
	if err != nil {
		h.logger.Errorw("role list failed", "error", err)
		response.InternalError(ctx, "获取角色列表失败")
		return
	}
	response.Success(ctx, gin.H{"items": roles})
}

func (h *Handler) Create(ctx *gin.Context) {
	var req CreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}
	created, err := h.svc.Create(ctx.Request.Context(), auth.GetUserID(ctx), req)
	if err != nil {
		h.writeError(ctx, "创建角色失败", err)
		return
	}
	response.Created(ctx, created)
}

func (h *Handler) Update(ctx *gin.Context) {
	roleID, ok := parseRoleID(ctx)
	if !ok {
		return
	}
	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}
	updated, err := h.svc.Update(ctx.Request.Context(), auth.GetUserID(ctx), roleID, req)
	if err != nil {
		h.writeError(ctx, "更新角色失败", err)
		return
	}
	response.Success(ctx, updated)
}

func (h *Handler) Delete(ctx *gin.Context) {
	roleID, ok := parseRoleID(ctx)
	if !ok {
		return
	}
	if err := h.svc.Delete(ctx.Request.Context(), auth.GetUserID(ctx), roleID); err != nil {
		h.writeError(ctx, "删除角色失败", err)
		return
	}
	response.SuccessWithMsg(ctx, "角色已删除", nil)
}

func (h *Handler) SetDefault(ctx *gin.Context) {
	roleID, ok := parseRoleID(ctx)
	if !ok {
		return
	}
	if err := h.svc.SetDefault(ctx.Request.Context(), auth.GetUserID(ctx), roleID); err != nil {
		h.writeError(ctx, "设置默认注册角色失败", err)
		return
	}
	response.SuccessWithMsg(ctx, "默认注册角色已更新", nil)
}

func (h *Handler) SetRole(ctx *gin.Context) {
	var req SetRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}
	if err := h.svc.SetRole(ctx.Request.Context(), auth.GetUserID(ctx), req); err != nil {
		h.writeError(ctx, "设置角色失败", err)
		return
	}
	response.SuccessWithMsg(ctx, "角色已更新", nil)
}

func (h *Handler) writeError(ctx *gin.Context, operation string, err error) {
	switch {
	case errors.Is(err, ErrCannotModifySuperAdmin):
		response.Forbidden(ctx, err.Error())
	case errors.Is(err, ErrUserNotFound), errors.Is(err, ErrRoleNotFound):
		response.NotFound(ctx, err.Error())
	case errors.Is(err, ErrInvalidRole), errors.Is(err, ErrInvalidRoleName), errors.Is(err, ErrInvalidPermission):
		response.BadRequest(ctx, err.Error())
	case errors.Is(err, ErrRoleExists), errors.Is(err, ErrRoleInUse), errors.Is(err, ErrProtectedRole), errors.Is(err, ErrDefaultRole):
		response.Conflict(ctx, err.Error())
	default:
		h.logger.Errorw("role operation failed", "operation", operation, "error", err)
		response.InternalError(ctx, operation)
	}
}

func parseRoleID(ctx *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil || value == 0 {
		response.BadRequest(ctx, "无效的角色 ID")
		return 0, false
	}
	return uint(value), true
}
