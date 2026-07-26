package requestlog

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/pkg/response"
	"go.uber.org/zap"
)

type Handler struct {
	reader Reader
	logger *zap.SugaredLogger
}

func NewHandler(reader Reader, logger *zap.SugaredLogger) *Handler {
	return &Handler{reader: reader, logger: logger}
}

func (h *Handler) RegisterRoutes(routes *gin.RouterGroup) {
	routes.GET("/request-logs/:request_id", h.Get)
}

func (h *Handler) Get(ctx *gin.Context) {
	record, err := h.reader.GetByRequestID(ctx.Request.Context(), ctx.Param("request_id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			response.NotFound(ctx, err.Error())
			return
		}
		h.logger.Errorw("request log query failed", "error", err, "request_id", ctx.Param("request_id"))
		response.InternalError(ctx, "查询请求日志失败")
		return
	}
	response.Success(ctx, record)
}
