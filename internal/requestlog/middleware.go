package requestlog

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/pkg/middleware"
	"github.com/lyimoexiao/akari/internal/model"
	"go.uber.org/zap"
)

const maxCapturedBodyBytes = 64 * 1024

type replayBody struct {
	io.Reader
	closer io.Closer
}

func (body *replayBody) Close() error {
	return body.closer.Close()
}

type responseCaptureWriter struct {
	gin.ResponseWriter
	body      bytes.Buffer
	truncated bool
}

func (writer *responseCaptureWriter) Write(data []byte) (int, error) {
	writer.capture(data)
	return writer.ResponseWriter.Write(data)
}

func (writer *responseCaptureWriter) WriteString(data string) (int, error) {
	writer.capture([]byte(data))
	return writer.ResponseWriter.WriteString(data)
}

func (writer *responseCaptureWriter) capture(data []byte) {
	remaining := maxCapturedBodyBytes - writer.body.Len()
	if remaining <= 0 {
		writer.truncated = true
		return
	}
	if len(data) > remaining {
		writer.body.Write(data[:remaining])
		writer.truncated = true
		return
	}
	writer.body.Write(data)
}

func Middleware(writer Writer, appLogger *zap.SugaredLogger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := middleware.GetRequestID(ctx)
		if requestID == "" {
			ctx.Next()
			return
		}

		requestBody, requestBodyTruncated := captureRequestBody(ctx)
		responseWriter := &responseCaptureWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = responseWriter
		startedAt := time.Now()
		ctx.Next()

		userID := auth.GetUserID(ctx)
		var operationUserID *uint
		if userID != 0 {
			operationUserID = &userID
		}
		record := model.RequestLog{
			RequestID:       requestID,
			UserID:          operationUserID,
			Module:          requestModule(ctx.Request.URL.Path),
			Method:          ctx.Request.Method,
			Path:            ctx.Request.URL.Path,
			Status:          ctx.Writer.Status(),
			LatencyMS:       time.Since(startedAt).Milliseconds(),
			IP:              ctx.ClientIP(),
			UserAgent:       ctx.Request.UserAgent(),
			RequestHeaders:  sanitizeHeaders(ctx.Request.Header),
			RequestBody:     sanitizeBody(ctx.GetHeader("Content-Type"), requestBody, requestBodyTruncated),
			ResponseHeaders: sanitizeHeaders(ctx.Writer.Header()),
			ResponseBody:    sanitizeBody(ctx.Writer.Header().Get("Content-Type"), responseWriter.body.Bytes(), responseWriter.truncated),
		}
		if err := writer.Save(ctx.Request.Context(), record); err != nil {
			appLogger.Errorw("request log persistence failed", "error", err, "request_id", requestID)
		}

		fields := []any{
			"method", record.Method,
			"path", record.Path,
			"module", record.Module,
			"status", record.Status,
			"latency_ms", record.LatencyMS,
			"request_id", record.RequestID,
			"ip", record.IP,
			"user_agent", record.UserAgent,
		}
		if record.UserID != nil {
			fields = append(fields, "user_id", *record.UserID)
		}
		if record.Status >= http.StatusInternalServerError {
			appLogger.Errorw("request", fields...)
		} else if record.Status >= http.StatusBadRequest {
			appLogger.Warnw("request", fields...)
		} else {
			appLogger.Infow("request", fields...)
		}
	}
}

func captureRequestBody(ctx *gin.Context) ([]byte, bool) {
	if ctx.Request.Body == nil {
		return nil, false
	}
	prefix, err := io.ReadAll(io.LimitReader(ctx.Request.Body, maxCapturedBodyBytes+1))
	if err != nil {
		ctx.Error(err)
	}
	ctx.Request.Body = &replayBody{
		Reader: io.MultiReader(bytes.NewReader(prefix), ctx.Request.Body),
		closer: ctx.Request.Body,
	}
	if len(prefix) > maxCapturedBodyBytes {
		return prefix[:maxCapturedBodyBytes], true
	}
	return prefix, false
}

func requestModule(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" && strings.HasPrefix(parts[1], "v") {
		return parts[2]
	}
	if len(parts) >= 2 && parts[0] == "api" {
		return parts[1]
	}
	return "api"
}
