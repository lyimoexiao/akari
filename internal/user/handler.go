package user

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/pkg/response"
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
	routes.GET("/users", h.permissions.Require(), h.ListUsers)
	routes.POST("/users/verify-email", h.permissions.Require(), h.VerifyEmail)
	routes.POST("/users/reset-password", h.permissions.Require(), h.ResetPassword)
	routes.DELETE("/users/:id", h.permissions.Require(), h.DeleteUser)
}

func (h *Handler) ListUsers(ctx *gin.Context) {
	var req ListUsersReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "无效参数: "+err.Error())
		return
	}

	result, err := h.svc.ListUsers(ctx.Request.Context(), req)
	if err != nil {
		h.logger.Errorw("user list failed", "error", err)
		response.InternalError(ctx, "获取用户列表失败")
		return
	}
	response.Success(ctx, result)
}

func (h *Handler) VerifyEmail(ctx *gin.Context) {
	var req VerifyEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.VerifyEmail(ctx.Request.Context(), auth.GetUserID(ctx), req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(ctx, err.Error())
		case errors.Is(err, ErrCannotModifyAdmin):
			response.Forbidden(ctx, err.Error())
		default:
			h.logger.Errorw("user verify email failed", "error", err)
			response.InternalError(ctx, "邮箱验证失败")
		}
		return
	}
	response.SuccessWithMsg(ctx, "邮箱已验证", nil)
}

func (h *Handler) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.ResetPassword(ctx.Request.Context(), auth.GetUserID(ctx), req); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(ctx, err.Error())
		case errors.Is(err, ErrCannotModifyAdmin):
			response.Forbidden(ctx, err.Error())
		default:
			h.logger.Errorw("user reset password failed", "error", err)
			response.InternalError(ctx, "密码重置失败")
		}
		return
	}
	response.SuccessWithMsg(ctx, "密码已重置", nil)
}

func (h *Handler) DeleteUser(ctx *gin.Context) {
	targetID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的用户 ID")
		return
	}

	command := DeleteUserCommand{
		CallerID: auth.GetUserID(ctx),
		UserID:   uint(targetID),
	}
	if err := h.svc.DeleteUser(ctx.Request.Context(), command); err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(ctx, err.Error())
		case errors.Is(err, ErrCannotDeleteSelf):
			response.BadRequest(ctx, err.Error())
		case errors.Is(err, ErrCannotModifyAdmin):
			response.Forbidden(ctx, err.Error())
		default:
			h.logger.Errorw("user delete failed", "error", err)
			response.InternalError(ctx, "删除用户失败")
		}
		return
	}
	response.SuccessWithMsg(ctx, "用户已删除", nil)
}

