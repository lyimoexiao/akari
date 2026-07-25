package logger

import (
	"bytes"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L *zap.SugaredLogger

var isDebug bool

func Init(cfg *config.LoggerConfig) {
	isDebug = cfg.Level == "debug"

	atom := zap.NewAtomicLevelAt(parseLevel(cfg.Level))

	enc := newEncoder(cfg.Format)

	writers := []zapcore.WriteSyncer{zapcore.AddSync(os.Stdout)}
	if cfg.OutputPath != "" {
		writers = append(writers, zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    cfg.MaxSize,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			LocalTime:  cfg.LocalTime,
			Compress:   cfg.Compress,
		}))
	}

	core := zapcore.NewCore(enc, zapcore.NewMultiWriteSyncer(writers...), atom)
	l := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	zap.RedirectStdLog(l)
	gin.DefaultWriter = io.Discard // silence Gin default logger

	L = l.Sugar()
}

func newEncoder(format string) zapcore.Encoder {
	ec := zap.NewProductionEncoderConfig()
	ec.EncodeTime = zapcore.RFC3339TimeEncoder
	ec.EncodeLevel = zapcore.CapitalLevelEncoder

	if format == "json" {
		return zapcore.NewJSONEncoder(ec)
	}

	ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(ec)
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		traceID := middleware.GetTraceID(c)

		// Read and restore request body for debug logging
		var bodyBytes []byte
		if isDebug && (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Wrap response writer to capture response body
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []any{
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency", latency,
			"trace_id", traceID,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"errors", c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}

		if isDebug && len(bodyBytes) > 0 && len(bodyBytes) < 4096 {
			fields = append(fields, "req_body", string(bodyBytes))
		}

		if isDebug && blw.body.Len() > 0 && blw.body.Len() < 4096 {
			fields = append(fields, "resp_body", blw.body.String())
		}

		if status >= 500 {
			L.Errorw("request", fields...)
		} else if status >= 400 {
			L.Warnw("request", fields...)
		} else {
			L.Infow("request", fields...)
		}
	}
}

// bodyLogWriter captures the response body for logging.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				L.Errorw("panic recovered",
					"error", err,
					"request", c.Request.URL.String(),
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
