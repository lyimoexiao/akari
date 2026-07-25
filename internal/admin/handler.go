package admin

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/response"
	"go.uber.org/zap"
)

// Handler provides HTTP handlers for admin user management.
type Handler struct {
	svc    *Service
	mw     *auth.Middleware
	logger *zap.SugaredLogger
}

// NewHandler creates an admin HTTP handler.
func NewHandler(svc *Service, mw *auth.Middleware, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, mw: mw, logger: logger}
}

// RegisterRoutes registers admin routes under the given group.
func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	admin := v1.Group("/admin")
	admin.Use(h.mw.RequireAuth())
	admin.Use(h.mw.RequireRole("super_admin", "staff"))
	{
		admin.GET("/users", h.ListUsers)
		admin.POST("/users/verify-email", h.VerifyEmail)
		admin.POST("/users/set-role", h.SetRole)
		admin.POST("/users/reset-password", h.ResetPassword)
		admin.DELETE("/users/:id", h.DeleteUser)
	}
}

// ListUsers handles GET /api/v1/admin/users
func (h *Handler) ListUsers(ctx *gin.Context) {
	var req ListUsersReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "无效参数: "+err.Error())
		return
	}

	resp, err := h.svc.ListUsers(ctx.Request.Context(), &req)
	if err != nil {
		h.logger.Errorw("admin list users failed", "error", err)
		response.InternalError(ctx, "获取用户列表失败")
		return
	}

	response.Success(ctx, resp)
}

// VerifyEmail handles POST /api/v1/admin/users/verify-email
func (h *Handler) VerifyEmail(ctx *gin.Context) {
	var req AdminVerifyEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.VerifyEmail(ctx.Request.Context(), &req); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(ctx, err.Error())
		} else {
			h.logger.Errorw("admin verify email failed", "error", err)
			response.InternalError(ctx, "邮箱验证失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "邮箱已验证", nil)
}

// SetRole handles POST /api/v1/admin/users/set-role
func (h *Handler) SetRole(ctx *gin.Context) {
	callerRole := auth.GetRole(ctx)
	callerIsSuperAdmin := callerRole == "super_admin"

	var req SetRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.SetRole(ctx.Request.Context(), callerIsSuperAdmin, &req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(ctx, err.Error())
		case errors.Is(err, ErrRoleNotAllowed):
			response.Forbidden(ctx, err.Error())
		default:
			h.logger.Errorw("admin set role failed", "error", err)
			response.InternalError(ctx, "设置角色失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "角色已更新", nil)
}

// ResetPassword handles POST /api/v1/admin/users/reset-password
func (h *Handler) ResetPassword(ctx *gin.Context) {
	callerRole := auth.GetRole(ctx)
	callerIsSuperAdmin := callerRole == "super_admin"

	var req AdminResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.ResetPassword(ctx.Request.Context(), callerIsSuperAdmin, &req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(ctx, err.Error())
		case errors.Is(err, ErrCannotModifyAdmin):
			response.Forbidden(ctx, err.Error())
		default:
			h.logger.Errorw("admin reset password failed", "error", err)
			response.InternalError(ctx, "密码重置失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "密码已重置", nil)
}

// DeleteUser handles DELETE /api/v1/admin/users/:id
func (h *Handler) DeleteUser(ctx *gin.Context) {
	callerID := auth.GetUserID(ctx)
	callerRole := auth.GetRole(ctx)
	callerIsSuperAdmin := callerRole == "super_admin"

	idParam := ctx.Param("id")
	targetID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的用户 ID")
		return
	}

	req := &DeleteUserReq{UserID: uint(targetID)}
	if err := h.svc.DeleteUser(ctx.Request.Context(), callerID, callerIsSuperAdmin, req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(ctx, err.Error())
		case errors.Is(err, ErrCannotDeleteSelf):
			response.BadRequest(ctx, err.Error())
		case errors.Is(err, ErrCannotModifyAdmin):
			response.Forbidden(ctx, err.Error())
		default:
			h.logger.Errorw("admin delete user failed", "error", err)
			response.InternalError(ctx, "删除用户失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "用户已删除", nil)
}
