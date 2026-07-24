package logger

import (
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L *zap.SugaredLogger

func Init(cfg *config.LoggerConfig) {
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

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		L.Infow("request",
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency", latency,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
			"errors", c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
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
