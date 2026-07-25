package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/captcha"
	"github.com/lyimoexiao/akari/internal/response"
	"go.uber.org/zap"
)

// Handler provides HTTP handlers for authentication endpoints.
type Handler struct {
	svc        *Service
	mw         *Middleware
	cfg        *HandlerConfig
	captchaSvc *captcha.Service
	logger     *zap.SugaredLogger
}

// HandlerConfig exposes auth configuration for the handler.
type HandlerConfig struct {
	RegistrationEnabled      bool
	EmailVerificationEnabled bool
}

// NewHandler creates a new auth HTTP handler.
func NewHandler(svc *Service, mw *Middleware, cfg *HandlerConfig, captchaSvc *captcha.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		svc:        svc,
		mw:         mw,
		cfg:        cfg,
		captchaSvc: captchaSvc,
		logger:     logger,
	}
}

// verifyCaptcha checks captcha if the service is enabled.
func (h *Handler) verifyCaptcha(ctx *gin.Context, captchaID string, userAnswer map[string]any, captchaToken string) bool {
	if h.captchaSvc == nil || !h.captchaSvc.IsEnabled() {
		return true
	}
	if captchaToken != "" {
		pass, err := h.captchaSvc.VerifyToken(ctx.Request.Context(), captchaToken)
		if err != nil || !pass {
			response.BadRequest(ctx, "验证码验证失败")
			return false
		}
		return true
	}
	if captchaID == "" || userAnswer == nil {
		response.BadRequest(ctx, "需要验证码")
		return false
	}
	pass, err := h.captchaSvc.Verify(ctx.Request.Context(), captchaID, userAnswer)
	if err != nil || !pass {
		response.BadRequest(ctx, "验证码验证失败")
		return false
	}
	return true
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(ctx *gin.Context) {
	if !h.cfg.RegistrationEnabled {
		response.Forbidden(ctx, "注册已关闭")
		return
	}

	var req RegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

if !h.verifyCaptcha(ctx, req.CaptchaID, req.UserAnswer, req.CaptchaToken) {
			return
		}

		resp, err := h.svc.Register(ctx.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrRegistrationClosed):
			response.Forbidden(ctx, err.Error())
		case errors.Is(err, ErrEmailAlreadyExists):
			response.Conflict(ctx, err.Error())
		case errors.Is(err, ErrUsernameAlreadyUsed):
			response.Conflict(ctx, err.Error())
		default:
			h.logger.Errorw("register failed", "error", err)
			response.InternalError(ctx, "注册失败")
		}
		return
	}

	response.Created(ctx, resp)
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(ctx *gin.Context) {
	var req LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

if !h.verifyCaptcha(ctx, req.CaptchaID, req.UserAnswer, req.CaptchaToken) {
			return
		}

		resp, err := h.svc.Login(ctx.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(ctx, err.Error())
		} else {
			h.logger.Errorw("login failed", "error", err)
			response.InternalError(ctx, "登录失败")
		}
		return
	}

	response.Success(ctx, resp)
}

// Logout handles POST /api/v1/auth/logout (requires authentication)
func (h *Handler) Logout(ctx *gin.Context) {
	tokenString := extractBearerToken(ctx)
	if tokenString == "" {
		response.Unauthorized(ctx, "缺少认证令牌")
		return
	}

	if err := h.svc.Logout(ctx.Request.Context(), tokenString); err != nil {
		h.logger.Errorw("logout failed", "error", err)
		response.InternalError(ctx, "登出失败")
		return
	}

	response.SuccessWithMsg(ctx, "已成功登出", nil)
}

// Me handles GET /api/v1/auth/me (requires authentication)
func (h *Handler) Me(ctx *gin.Context) {
	userID := GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	user, err := h.svc.GetCurrentUser(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			response.NotFound(ctx, err.Error())
		} else {
			h.logger.Errorw("get current user failed", "error", err)
			response.InternalError(ctx, "获取用户失败")
		}
		return
	}

	response.Success(ctx, user)
}

// SendVerificationEmail handles POST /api/v1/auth/verify-email/send (requires authentication)
func (h *Handler) SendVerificationEmail(ctx *gin.Context) {
	if !h.cfg.EmailVerificationEnabled {
		response.Forbidden(ctx, "邮箱验证已禁用")
		return
	}

	userID := GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	if err := h.svc.SendVerificationEmail(ctx.Request.Context(), userID); err != nil {
		switch {
		case errors.Is(err, ErrEmailVerificationDisabled):
			response.Forbidden(ctx, err.Error())
		case errors.Is(err, ErrUserNotFound):
			response.NotFound(ctx, err.Error())
		case errors.Is(err, ErrEmailAlreadyVerified):
			response.Conflict(ctx, err.Error())
		default:
			h.logger.Errorw("send verification email failed", "error", err)
			response.InternalError(ctx, "发送验证邮件失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "验证邮件已发送", nil)
}

// VerifyEmail handles POST /api/v1/auth/verify-email (requires authentication)
func (h *Handler) VerifyEmail(ctx *gin.Context) {
	if !h.cfg.EmailVerificationEnabled {
		response.Forbidden(ctx, "邮箱验证已禁用")
		return
	}

	userID := GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	var req VerifyEmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "需要提供令牌")
		return
	}

	if err := h.svc.VerifyEmail(ctx.Request.Context(), userID, req.Token); err != nil {
		switch {
		case errors.Is(err, ErrEmailVerificationDisabled):
			response.Forbidden(ctx, err.Error())
		case errors.Is(err, ErrInvalidToken):
			response.BadRequest(ctx, err.Error())
		default:
			h.logger.Errorw("verify email failed", "error", err)
			response.InternalError(ctx, "验证失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "邮箱验证成功", nil)
}

// RegisterRoutes registers all auth routes under the given group.
func (h *Handler) RegisterRoutes(v1 *gin.RouterGroup) {
	auth := v1.Group("/auth")
	{
		// Public routes
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/password-reset/request", h.ForgotPassword)
		auth.POST("/password-reset", h.ResetPassword)

		// Authenticated routes
		authed := auth.Group("")
		authed.Use(h.mw.RequireAuth())
		{
			authed.GET("/me", h.Me)
			authed.POST("/logout", h.Logout)
			authed.POST("/verify-email/send", h.SendVerificationEmail)
			authed.POST("/verify-email", h.VerifyEmail)
		}
	}
}

// ForgotPassword handles POST /api/v1/auth/password-reset/request
func (h *Handler) ForgotPassword(ctx *gin.Context) {
	var req ForgotPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if !h.verifyCaptcha(ctx, req.CaptchaID, req.UserAnswer, req.CaptchaToken) {
		return
	}

	if err := h.svc.ForgotPassword(ctx.Request.Context(), &req); err != nil {
		if errors.Is(err, ErrPasswordResetDisabled) {
			response.Forbidden(ctx, err.Error())
		} else {
			h.logger.Errorw("forgot password failed", "error", err)
			response.InternalError(ctx, "发送重置邮件失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "如果该邮箱存在，重置邮件已发送", nil)
}

// ResetPassword handles POST /api/v1/auth/password-reset
func (h *Handler) ResetPassword(ctx *gin.Context) {
	var req ResetPasswordReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.ResetPassword(ctx.Request.Context(), &req); err != nil {
		switch {
		case errors.Is(err, ErrPasswordResetDisabled):
			response.Forbidden(ctx, err.Error())
		case errors.Is(err, ErrInvalidToken):
			response.BadRequest(ctx, err.Error())
		default:
			h.logger.Errorw("reset password failed", "error", err)
			response.InternalError(ctx, "密码重置失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "密码已重置", nil)
}
