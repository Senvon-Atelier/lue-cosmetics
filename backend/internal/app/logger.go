package app

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger returns a *zap.Logger. In development it uses zap's console
// encoder (human-readable, colored levels). In production it uses zap's
// JSON encoder (structured for log aggregation).
//
// The returned logger is the BASE logger. Per-request scoping (request_id,
// user_id, role) happens via the RequestLogger middleware in package httpx
// — which calls `logger.With(...)` and stashes the scoped logger on
// r.Context() through internal/logging.WithLogger.
func NewLogger(level, env string) *zap.Logger {
	lvl := zapcore.InfoLevel
	switch strings.ToLower(level) {
	case "debug":
		lvl = zapcore.DebugLevel
	case "warn":
		lvl = zapcore.WarnLevel
	case "error":
		lvl = zapcore.ErrorLevel
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.MessageKey = "msg"
	encCfg.LevelKey = "level"
	encCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	var enc zapcore.Encoder
	if env == "development" {
		encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enc = zapcore.NewConsoleEncoder(encCfg)
	} else {
		enc = zapcore.NewJSONEncoder(encCfg)
	}

	core := zapcore.NewCore(enc, zapcore.AddSync(os.Stdout), lvl)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}
