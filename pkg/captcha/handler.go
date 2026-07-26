package captcha

import (
	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/pkg/response"
)

// Handler provides HTTP handlers for captcha endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new captcha HTTP handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Generate handles GET /api/v1/captcha and returns a new captcha.
func (h *Handler) Generate(ctx *gin.Context) {
	if !h.svc.IsEnabled() {
		response.Success(ctx, gin.H{
			"enabled": false,
			"message": "验证码已禁用",
		})
		return
	}

	result, err := h.svc.Generate(ctx.Request.Context())
	if err != nil {
		response.InternalError(ctx, "生成验证码失败")
		return
	}

	response.Success(ctx, gin.H{
		"enabled": true,
		"data":    result,
	})
}

// VerifyReq is the expected JSON body for captcha verification.
type VerifyReq struct {
	CaptchaID  string         `json:"captcha_id" binding:"required"`
	UserAnswer map[string]any `json:"user_answer" binding:"required"`
}

// VerifyTurnstileReq is the expected JSON body for Turnstile verification.
type VerifyTurnstileReq struct {
	Token string `json:"token" binding:"required"`
}

// TurnstileVerify handles POST /api/v1/captcha/turnstile-verify for Turnstile tokens.
func (h *Handler) TurnstileVerify(ctx *gin.Context) {
	var req VerifyTurnstileReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	pass, err := h.svc.VerifyToken(ctx.Request.Context(), req.Token)
	if err != nil {
		response.Success(ctx, gin.H{
			"enabled": true,
			"pass":    false,
			"message": err.Error(),
		})
		return
	}

	response.Success(ctx, gin.H{
		"enabled": true,
		"pass":    pass,
	})
}

// Verify handles POST /api/v1/captcha/verify and checks the user's answer.
func (h *Handler) Verify(ctx *gin.Context) {
	if !h.svc.IsEnabled() {
		response.Success(ctx, gin.H{
			"enabled": true,
			"pass":    true,
			"message": "验证码已禁用，直接通过",
		})
		return
	}

	var req VerifyReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	pass, err := h.svc.Verify(ctx.Request.Context(), req.CaptchaID, req.UserAnswer)
	if err != nil {
		// Captcha expired or not found - treat as failure
		response.Success(ctx, gin.H{
			"enabled": true,
			"pass":    false,
			"message": err.Error(),
		})
		return
	}

	response.Success(ctx, gin.H{
		"enabled": true,
		"pass":    pass,
	})
}
