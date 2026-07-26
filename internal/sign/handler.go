package sign

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/response"
	"go.uber.org/zap"
)

type Handler struct {
	svc    *Service
	logger *zap.SugaredLogger
}

func NewHandler(svc *Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.POST("/sign", h.Sign)
	protected.GET("/sign/status", h.Status)
}

// Sign handles POST /api/v1/sign
func (h *Handler) Sign(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)

	acquired, err := h.svc.Sign(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrAlreadySigned) {
			response.Error(ctx, http.StatusConflict, "今天已签到")
			return
		}
		if errors.Is(err, ErrSignTooEarly) {
			response.Error(ctx, http.StatusConflict, "签到间隔不足")
			return
		}
		h.logger.Errorw("sign failed", "user_id", userID, "error", err)
		response.Error(ctx, http.StatusInternalServerError, "签到失败")
		return
	}

	response.Success(ctx, gin.H{
		"acquired": acquired,
	})
}

// Status handles GET /api/v1/sign/status
func (h *Handler) Status(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)

	signedToday, score, err := h.svc.Status(ctx.Request.Context(), userID)
	if err != nil {
		h.logger.Errorw("sign status failed", "user_id", userID, "error", err)
		response.Error(ctx, http.StatusInternalServerError, "获取签到状态失败")
		return
	}

	response.Success(ctx, gin.H{
		"signed_today": signedToday,
		"score":        score,
	})
}