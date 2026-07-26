package middleware

import (
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

const (
	apiPathPrefix     = "/api/"
	requestIDHeader   = "X-Request-ID"
	maxRequestIDBytes = 128
)

func RequestID() gin.HandlerFunc {
	handler := requestid.New()
	return func(ctx *gin.Context) {
		if !strings.HasPrefix(ctx.Request.URL.Path, apiPathPrefix) {
			ctx.Next()
			return
		}
		if len(ctx.GetHeader(requestIDHeader)) > maxRequestIDBytes {
			ctx.Request.Header.Del(requestIDHeader)
		}
		handler(ctx)
	}
}

func GetRequestID(ctx *gin.Context) string {
	if !strings.HasPrefix(ctx.Request.URL.Path, apiPathPrefix) {
		return ""
	}
	return requestid.Get(ctx)
}
