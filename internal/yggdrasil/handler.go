package yggdrasil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/response"
	"go.uber.org/zap"
)

// ───────────────────────────── Handler ─────────────────────────────

type Handler struct {
	svc    *Service
	logger *zap.SugaredLogger
}

func NewHandler(svc *Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// ── Error helpers ──

// writeYggdrasilError writes an error in Yggdrasil spec format.
func writeYggdrasilError(ctx *gin.Context, status int, err, msg string) {
	ctx.AbortWithStatusJSON(status, ErrorResp{
		Error:        err,
		ErrorMessage: msg,
	})
}

func writeForbiddenOperation(ctx *gin.Context, msg string) {
	writeYggdrasilError(ctx, http.StatusForbidden, "ForbiddenOperationException", msg)
}

func writeIllegalArgument(ctx *gin.Context, msg string) {
	writeYggdrasilError(ctx, http.StatusBadRequest, "IllegalArgumentException", msg)
}

func writeNoContent(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusNoContent)
}

// ───────────────────────────── Authserver ─────────────────────────────

// POST /authserver/authenticate
func (h *Handler) Authenticate(ctx *gin.Context) {
	var req AuthenticateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeIllegalArgument(ctx, "无效请求内容")
		return
	}

	resp, err := h.svc.Authenticate(ctx.Request.Context(), &req, ctx.ClientIP())
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			writeForbiddenOperation(ctx, "Invalid credentials. Invalid username or password.")
		case errors.Is(err, ErrEmailNotVerified):
			writeForbiddenOperation(ctx, "Email not verified.")
		default:
			h.logger.Errorw("yggdrasil authenticate failed", "error", err)
			writeForbiddenOperation(ctx, "Invalid credentials. Invalid username or password.")
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// POST /authserver/refresh
func (h *Handler) Refresh(ctx *gin.Context) {
	var req RefreshReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeIllegalArgument(ctx, "无效请求内容")
		return
	}

	resp, err := h.svc.Refresh(ctx.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidToken), errors.Is(err, ErrTokenExpired):
			writeForbiddenOperation(ctx, "Invalid token.")
		default:
			h.logger.Errorw("yggdrasil refresh failed", "error", err)
			writeForbiddenOperation(ctx, "Invalid token.")
		}
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// POST /authserver/validate
func (h *Handler) Validate(ctx *gin.Context) {
	var req ValidateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeIllegalArgument(ctx, "无效请求内容")
		return
	}

	if err := h.svc.Validate(ctx.Request.Context(), &req); err != nil {
		writeForbiddenOperation(ctx, "Invalid token.")
		return
	}

	writeNoContent(ctx)
}

// POST /authserver/invalidate
func (h *Handler) Invalidate(ctx *gin.Context) {
	var req InvalidateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeNoContent(ctx)
		return
	}

	_ = h.svc.Invalidate(ctx.Request.Context(), &req)
	writeNoContent(ctx)
}

// POST /authserver/signout
func (h *Handler) Signout(ctx *gin.Context) {
	var req SignoutReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeNoContent(ctx)
		return
	}

	_ = h.svc.Signout(ctx.Request.Context(), &req)
	writeNoContent(ctx)
}

// ───────────────────────────── Sessionserver ─────────────────────────────

// POST /sessionserver/session/minecraft/join
func (h *Handler) JoinServer(ctx *gin.Context) {
	var req JoinReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeIllegalArgument(ctx, "无效请求内容")
		return
	}

	clientIP := ctx.ClientIP()
	if err := h.svc.JoinServer(ctx.Request.Context(), &req, clientIP); err != nil {
		writeForbiddenOperation(ctx, "Invalid token.")
		return
	}

	writeNoContent(ctx)
}

// GET /sessionserver/session/minecraft/hasJoined
func (h *Handler) HasJoined(ctx *gin.Context) {
	username := ctx.Query("username")
	serverID := ctx.Query("serverId")
	ip := ctx.Query("ip")

	resp, err := h.svc.HasJoined(ctx.Request.Context(), username, serverID, ip)
	if err != nil {
		writeNoContent(ctx)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GET /sessionserver/session/minecraft/profile/:uuid
func (h *Handler) GetProfile(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	unsigned := ctx.DefaultQuery("unsigned", "true") == "true"

	resp, err := h.svc.GetProfile(ctx.Request.Context(), uuid, unsigned)
	if err != nil {
		writeNoContent(ctx)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ───────────────────────────── API Profiles ─────────────────────────────

// POST /api/profiles/minecraft
func (h *Handler) GetProfilesByName(ctx *gin.Context) {
	var names []string
	if err := ctx.ShouldBindJSON(&names); err != nil {
		writeIllegalArgument(ctx, "无效请求内容")
		return
	}

	// Limit batch size to prevent abuse (spec says at least 2, we use 100)
	const maxBatch = 100
	if len(names) > maxBatch {
		names = names[:maxBatch]
	}
	if len(names) == 0 {
		ctx.JSON(http.StatusOK, []ProfileResp{})
		return
	}

	resp, err := h.svc.GetProfilesByName(ctx.Request.Context(), names)
	if err != nil {
		ctx.JSON(http.StatusOK, []ProfileResp{})
		return
	}
	if resp == nil {
		resp = []ProfileResp{}
	}

	ctx.JSON(http.StatusOK, resp)
}

// ───────────────────────────── User Status (app JWT auth) ─────────────────────────────

// GET /user/status — requires app JWT auth (middleware applied in router)
func (h *Handler) UserStatus(ctx *gin.Context) {
	userID := auth.GetUserID(ctx)
	if userID == 0 {
		response.Unauthorized(ctx, "需要认证")
		return
	}

	resp, err := h.svc.UserStatus(ctx.Request.Context(), userID)
	if err != nil {
		h.logger.Errorw("yggdrasil user status failed", "error", err)
		response.InternalError(ctx, "获取 Yggdrasil 状态失败")
		return
	}

	response.Success(ctx, resp)
}

// ───────────────────────────── API Metadata ─────────────────────────────

// GET /
func (h *Handler) Metadata(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, h.svc.Metadata())
}

// ───────────────────────────── Route Registration ─────────────────────────────

// RegisterRoutes registers all Yggdrasil API routes under the given group.
// Expected to be mounted at /api/v1/yggdrasil.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Authserver
	r.POST("/authserver/authenticate", h.Authenticate)
	r.POST("/authserver/refresh", h.Refresh)
	r.POST("/authserver/validate", h.Validate)
	r.POST("/authserver/invalidate", h.Invalidate)
	r.POST("/authserver/signout", h.Signout)

	// Sessionserver
	r.POST("/sessionserver/session/minecraft/join", h.JoinServer)
	r.GET("/sessionserver/session/minecraft/hasJoined", h.HasJoined)
	r.GET("/sessionserver/session/minecraft/profile/:uuid", h.GetProfile)

	// API profiles
	r.POST("/api/profiles/minecraft", h.GetProfilesByName)

	// Metadata at the root of the Yggdrasil API (with and without trailing slash)
	r.GET("", h.Metadata)
	r.GET("/", h.Metadata)
}
