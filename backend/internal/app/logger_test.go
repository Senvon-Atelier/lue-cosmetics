package app_test

import (
	"testing"

	"go.uber.org/zap/zapcore"

	"github.com/oti-adjei/ruecosmetics/internal/app"
)

func TestNewLoggerHonorsLevel(t *testing.T) {
	cases := []struct {
		in     string
		enable zapcore.Level
		deny   zapcore.Level
	}{
		{"debug", zapcore.DebugLevel, -1},
		{"info", zapcore.InfoLevel, zapcore.DebugLevel},
		{"warn", zapcore.WarnLevel, zapcore.InfoLevel},
		{"error", zapcore.ErrorLevel, zapcore.WarnLevel},
		{"garbage", zapcore.InfoLevel, zapcore.DebugLevel}, // unknown → info default
	}
	for _, c := range cases {
		l := app.NewLogger(c.in, "development")
		if !l.Core().Enabled(c.enable) {
			t.Errorf("level=%s: expected %v to be enabled", c.in, c.enable)
		}
		if c.deny >= 0 && l.Core().Enabled(c.deny) {
			t.Errorf("level=%s: did not expect %v to be enabled", c.in, c.deny)
		}
	}
}

func TestNewLoggerNotNil(t *testing.T) {
	if app.NewLogger("info", "production") == nil {
		t.Fatal("nil logger")
	}
}
