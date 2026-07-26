package closet

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/pkg/pagination"
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

func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	closet := protected.Group("/closet")
	closet.Use(h.permissions.Require())
	{
		closet.GET("", h.List)
		closet.GET("/all-ids", h.AllIDs)
		closet.POST("", h.Add)
		closet.PUT("/:tid", h.Rename)
		closet.DELETE("/:tid", h.Remove)
	}
}

// List handles GET /api/v1/closet
func (h *Handler) List(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	var req ListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "无效参数: "+err.Error())
		return
	}

	listQuery := ListQuery{
		Pagination: pagination.Paging{Page: req.Page, PageSize: req.PageSize},
		Type:       req.Type,
		Search:     req.Search,
	}

	result, err := h.svc.List(ctx.Request.Context(), userID, listQuery)
	if err != nil {
		h.logger.Errorw("closet list failed", "user_id", userID, "error", err)
		response.InternalError(ctx, "获取衣橱列表失败")
		return
	}
	response.Success(ctx, result)
}

// AllIDs handles GET /api/v1/closet/all-ids
func (h *Handler) AllIDs(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	ids, err := h.svc.AllTextureIDs(ctx.Request.Context(), userID)
	if err != nil {
		h.logger.Errorw("closet all-ids failed", "user_id", userID, "error", err)
		response.InternalError(ctx, "获取衣橱 ID 列表失败")
		return
	}
	response.Success(ctx, gin.H{"ids": ids})
}

// Add handles POST /api/v1/closet
func (h *Handler) Add(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	var req AddReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.Add(ctx.Request.Context(), userID, req.TID, req.Name); err != nil {
		switch {
		case err == ErrClosetItemNotFound:
			response.NotFound(ctx, "纹理不存在")
		case err == ErrAlreadyInCloset:
			response.Conflict(ctx, "已在衣橱中")
		case err == ErrNotEnoughScore:
			response.Error(ctx, 402, "积分不足")
		default:
			h.logger.Errorw("closet add failed", "user_id", userID, "tid", req.TID, "error", err)
			response.InternalError(ctx, "添加失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "已添加到衣橱", nil)
}

// Rename handles PUT /api/v1/closet/:tid
func (h *Handler) Rename(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	tid, err := strconv.ParseUint(ctx.Param("tid"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的纹理 ID")
		return
	}

	var req RenameReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	if err := h.svc.Rename(ctx.Request.Context(), userID, uint(tid), req.Name); err != nil {
		switch {
		case err == ErrClosetItemNotFound:
			response.NotFound(ctx, "衣橱物品不存在")
		case err == ErrNotTextureUploader:
			response.Forbidden(ctx, "只有材质上传者才能重命名")
		default:
			h.logger.Errorw("closet rename failed", "user_id", userID, "tid", tid, "error", err)
			response.InternalError(ctx, "重命名失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "已重命名", nil)
}

// Remove handles DELETE /api/v1/closet/:tid
func (h *Handler) Remove(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	tid, err := strconv.ParseUint(ctx.Param("tid"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的纹理 ID")
		return
	}

	if err := h.svc.Remove(ctx.Request.Context(), userID, uint(tid)); err != nil {
		if err == ErrClosetItemNotFound {
			response.NotFound(ctx, "衣橱物品不存在")
		} else {
			h.logger.Errorw("closet remove failed", "user_id", userID, "tid", tid, "error", err)
			response.InternalError(ctx, "移除失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "已从衣橱移除", nil)
}
