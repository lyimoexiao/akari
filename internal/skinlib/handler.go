package skinlib

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/pkg/pagination"
	"github.com/lyimoexiao/akari/pkg/response"
	"go.uber.org/zap"
)

type PermissionChecker interface {
	Require() gin.HandlerFunc
	HasPermission(ctx context.Context, userID uint, object, action string) (bool, string, error)
}

type Handler struct {
	svc         *Service
	permissions PermissionChecker
	logger      *zap.SugaredLogger
}

func NewHandler(svc *Service, permissions PermissionChecker, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, permissions: permissions, logger: logger}
}

// RegisterPublicRoutes registers public skinlib routes (browsable without authentication).
// Mounted directly on the /api/v1 group, outside the auth-protected group.
func (h *Handler) RegisterPublicRoutes(v1 *gin.RouterGroup) {
	public := v1.Group("/skinlib")
	{
		public.GET("", h.List)
		public.GET("/:tid", h.Get)
		public.GET("/:tid/download", h.Download)
	}
}

// RegisterRoutes registers authenticated skinlib routes (require JWT + RBAC permission).
// Mounted inside the auth-protected group.
func (h *Handler) RegisterRoutes(protected *gin.RouterGroup) {
	authed := protected.Group("/skinlib")
	authed.Use(h.permissions.Require())
	{
		authed.GET("/manage", h.ManageList)
		authed.POST("", h.Upload)
		authed.PUT("/:tid", h.Update)
		authed.DELETE("/:tid", h.Delete)
	}
}

// List handles GET /api/v1/skinlib
func (h *Handler) List(ctx *gin.Context) {
	var req ListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "无效参数: "+err.Error())
		return
	}

	listQuery := ListQuery{
		Pagination: pagination.Paging{Page: req.Page, PageSize: req.PageSize},
		Type:       req.Type,
		Search:     req.Search,
		PublicOnly: true,
		Order:      req.Order,
	}

	result, err := h.svc.List(ctx.Request.Context(), listQuery)
	if err != nil {
		h.logger.Errorw("skinlib list failed", "error", err)
		response.InternalError(ctx, "获取皮肤列表失败")
		return
	}
	response.Success(ctx, result)
}

// Get handles GET /api/v1/skinlib/:tid
func (h *Handler) Get(ctx *gin.Context) {
	tid, err := strconv.ParseUint(ctx.Param("tid"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的纹理 ID")
		return
	}

	record, err := h.svc.GetByID(ctx.Request.Context(), uint(tid))
	if err != nil {
		response.NotFound(ctx, "纹理不存在")
		return
	}

	response.Success(ctx, TextureDetail{
		TextureItem: *h.svc.toItem(record),
	})
}

// Download handles GET /api/v1/skinlib/:tid/download
func (h *Handler) Download(ctx *gin.Context) {
	tid, err := strconv.ParseUint(ctx.Param("tid"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的纹理 ID")
		return
	}

	record, err := h.svc.GetByID(ctx.Request.Context(), uint(tid))
	if err != nil {
		response.NotFound(ctx, "纹理不存在")
		return
	}

	// Check if the texture is public or the requester owns it
	userID := auth.GetUserID(ctx)
	if !record.Public && record.Uploader != userID {
		response.Forbidden(ctx, "无权下载此纹理")
		return
	}

	// Redirect to raw file serving endpoint
	ctx.Redirect(302, "/api/v1/raw/"+record.Hash)
}

// Upload handles POST /api/v1/skinlib (multipart: file + name + type)
func (h *Handler) Upload(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		response.BadRequest(ctx, "需要上传文件")
		return
	}
	defer file.Close()

	name := ctx.PostForm("name")
	if name == "" {
		name = header.Filename
	}

	textureType := ctx.PostForm("type")
	if textureType == "" {
		textureType = "steve"
	}

	result, err := h.svc.Upload(ctx.Request.Context(), userID, name, textureType, file)
	if err != nil {
		switch {
		case err == ErrInvalidType:
			response.BadRequest(ctx, err.Error())
		case err == ErrAlreadyUploaded:
			response.Conflict(ctx, err.Error())
		case err == ErrForbidden:
			response.Forbidden(ctx, err.Error())
		default:
			h.logger.Errorw("skinlib upload failed", "user_id", userID, "error", err)
			response.InternalError(ctx, "上传失败")
		}
		return
	}

	response.Created(ctx, result)
}

// Update handles PUT /api/v1/skinlib/:tid
func (h *Handler) Update(ctx *gin.Context) {
	tid, err := strconv.ParseUint(ctx.Param("tid"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的纹理 ID")
		return
	}

	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.BadRequest(ctx, "无效请求: "+err.Error())
		return
	}

	userID := auth.GetUserID(ctx)
	force := false

	// STAFF can update any texture
	canManage, _, _ := h.permissions.HasPermission(ctx.Request.Context(), userID, "/api/v1/skinlib/manage", "GET")
	if canManage {
		force = true
	}

	if err := h.svc.Update(ctx.Request.Context(), uint(tid), userID, req.Name, req.Public, force); err != nil {
		switch {
		case err == ErrForbidden:
			response.Forbidden(ctx, "无权修改此纹理")
		case err == ErrTextureNotFound:
			response.NotFound(ctx, "纹理不存在")
		default:
			h.logger.Errorw("skinlib update failed", "tid", tid, "error", err)
			response.InternalError(ctx, "更新失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "已更新", nil)
}

// Delete handles DELETE /api/v1/skinlib/:tid
func (h *Handler) Delete(ctx *gin.Context) {
	tid, err := strconv.ParseUint(ctx.Param("tid"), 10, 64)
	if err != nil {
		response.BadRequest(ctx, "无效的纹理 ID")
		return
	}

	userID := auth.GetUserID(ctx)
	force := false

	// Check if user has manage permission (staff) — can delete any texture
	canManage, _, _ := h.permissions.HasPermission(ctx.Request.Context(), userID, "/api/v1/skinlib/manage", "GET")
	if canManage {
		force = true
	}

	if err := h.svc.Delete(ctx.Request.Context(), uint(tid), userID, force); err != nil {
		switch {
		case err == ErrForbidden:
			response.Forbidden(ctx, "无权删除此纹理")
		case err == ErrTextureNotFound:
			response.NotFound(ctx, "纹理不存在")
		default:
			h.logger.Errorw("skinlib delete failed", "tid", tid, "error", err)
			response.InternalError(ctx, "删除失败")
		}
		return
	}

	response.SuccessWithMsg(ctx, "已删除", nil)
}

// ManageList handles GET /api/v1/skinlib/manage — lists all textures including private ones (staff only)
func (h *Handler) ManageList(ctx *gin.Context) {
	var req ListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		response.BadRequest(ctx, "无效参数: "+err.Error())
		return
	}

	// Show all textures including private (PublicOnly: false), optionally filter by uploader
	listQuery := ListQuery{
		Pagination: pagination.Paging{Page: req.Page, PageSize: req.PageSize},
		Type:       req.Type,
		Search:     req.Search,
		PublicOnly: false,
		Order:      req.Order,
	}

	if req.Uploader > 0 {
		listQuery.Uploader = req.Uploader
	}

	result, err := h.svc.List(ctx.Request.Context(), listQuery)
	if err != nil {
		h.logger.Errorw("skinlib manage list failed", "error", err)
		response.InternalError(ctx, "获取皮肤列表失败")
		return
	}
	response.Success(ctx, result)
}
