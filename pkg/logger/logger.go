package logger

import (
	"io"
	"os"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/pkg/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L = zap.NewNop().Sugar()

func Init(cfg *Config) {
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
	base := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.RedirectStdLog(base)
	gin.DefaultWriter = io.Discard
	L = base.Sugar()
}

func newEncoder(format string) zapcore.Encoder {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.RFC3339TimeEncoder
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(config)
}

func GinRecovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				L.Errorw(
					"panic recovered",
					"error", recovered,
					"request", ctx.Request.URL.String(),
					"request_id", middleware.GetRequestID(ctx),
					"stack", string(debug.Stack()),
				)
				ctx.AbortWithStatus(500)
			}
		}()
		ctx.Next()
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
