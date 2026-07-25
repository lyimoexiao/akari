package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const CtxKeyTraceID = "trace_id"

// TraceID injects a unique trace ID into each request context.
func TraceID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID := generateTraceID()
		ctx.Set(CtxKeyTraceID, traceID)
		ctx.Header("X-Trace-ID", traceID)
		ctx.Next()
	}
}

// GetTraceID retrieves the trace ID from the context.
func GetTraceID(ctx *gin.Context) string {
	id, _ := ctx.Get(CtxKeyTraceID)
	if id == nil {
		return ""
	}
	return id.(string)
}

func generateTraceID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
