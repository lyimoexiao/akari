package score

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/pkg/pagination"
	"github.com/lyimoexiao/akari/pkg/response"
	"go.uber.org/zap"
)

type Handler struct {
	ops         Operator
	permissions RouteGuard
	logger      *zap.SugaredLogger
}

func NewHandler(ops Operator, permissions RouteGuard, logger *zap.SugaredLogger) *Handler {
	return &Handler{ops: ops, permissions: permissions, logger: logger}
}

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	protected.GET("/score", h.ScoreInfo)
	protected.GET("/score/history", h.ScoreHistory)
	protected.GET("/score/history/:id", h.permissions.Require(), h.ScoreHistoryByUser)
}

func (h *Handler) ScoreInfo(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)

	balance, lastSignAt, err := h.ops.ScoreInfo(ctx.Request.Context(), userID)
	if err != nil {
		h.logger.Errorw("score info failed", "user_id", userID, "error", err)
		response.Error(ctx, 500, "获取积分信息失败")
		return
	}

	resp := gin.H{"score": balance}
	if lastSignAt != nil {
		resp["last_sign_at"] = lastSignAt.Format("2006-01-02T15:04:05+08:00")
	}
	response.Success(ctx, resp)
}

func (h *Handler) ScoreHistory(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)

	var req ScoreHistoryReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "无效参数: "+err.Error())
		return
	}

	list, err := h.ops.ListLogs(ctx.Request.Context(), userID, LogQuery{
		Paging: pagination.Paging{Page: req.Page, PageSize: req.PageSize},
	})
	if err != nil {
		h.logger.Errorw("score history failed", "user_id", userID, "error", err)
		response.Error(ctx, 500, "获取积分历史失败")
		return
	}

	response.Success(ctx, list)
}

func (h *Handler) ScoreHistoryByUser(ctx *gin.Context) {
	targetID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的用户 ID")
		return
	}

	var req ScoreHistoryReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "无效参数: "+err.Error())
		return
	}

	list, err := h.ops.ListLogs(ctx.Request.Context(), uint(targetID), LogQuery{
		Paging: pagination.Paging{Page: req.Page, PageSize: req.PageSize},
	})
	if err != nil {
		h.logger.Errorw("score history by user failed", "target_id", targetID, "error", err)
		response.Error(ctx, 500, "获取积分历史失败")
		return
	}

	response.Success(ctx, list)
}
