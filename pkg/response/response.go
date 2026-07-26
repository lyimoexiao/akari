package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/pkg/middleware"
)

// Response is the unified API response envelope.
type Response struct {
	Code      int    `json:"code"`
	Success   bool   `json:"success"`
	Msg       string `json:"msg"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// New creates a Response with the given fields.
func New(code int, success bool, msg string, data any) Response {
	return Response{
		Code:    code,
		Success: success,
		Msg:     msg,
		Data:    data,
	}
}

func writeResponse(ctx *gin.Context, httpStatus int, resp Response) {
	resp.RequestID = middleware.GetRequestID(ctx)
	ctx.JSON(httpStatus, resp)
}

// Success writes a success response with data.
func Success(ctx *gin.Context, data any) {
	writeResponse(ctx, http.StatusOK, New(0, true, "ok", data))
}

// SuccessWithMsg writes a success response with custom message and data.
func SuccessWithMsg(ctx *gin.Context, msg string, data any) {
	writeResponse(ctx, http.StatusOK, New(0, true, msg, data))
}

// Created writes a 201 created response.
func Created(ctx *gin.Context, data any) {
	writeResponse(ctx, http.StatusCreated, New(0, true, "created", data))
}

// Error writes an error response with the given HTTP status and message.
func Error(ctx *gin.Context, httpStatus int, msg string) {
	writeResponse(ctx, httpStatus, New(-1, false, msg, nil))
}

// ErrorWithCode writes an error response with a custom code and message.
func ErrorWithCode(ctx *gin.Context, httpStatus int, code int, msg string) {
	writeResponse(ctx, httpStatus, Response{
		Code:    code,
		Success: false,
		Msg:     msg,
	})
}

// BadRequest writes a 400 error.
func BadRequest(ctx *gin.Context, msg string) {
	Error(ctx, http.StatusBadRequest, msg)
}

// Unauthorized writes a 401 error.
func Unauthorized(ctx *gin.Context, msg string) {
	Error(ctx, http.StatusUnauthorized, msg)
}

// Forbidden writes a 403 error.
func Forbidden(ctx *gin.Context, msg string) {
	Error(ctx, http.StatusForbidden, msg)
}

// NotFound writes a 404 error.
func NotFound(ctx *gin.Context, msg string) {
	Error(ctx, http.StatusNotFound, msg)
}

// Conflict writes a 409 error.
func Conflict(ctx *gin.Context, msg string) {
	Error(ctx, http.StatusConflict, msg)
}

// InternalError writes a 500 error.
func InternalError(ctx *gin.Context, msg string) {
	Error(ctx, http.StatusInternalServerError, msg)
}
