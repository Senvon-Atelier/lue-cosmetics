package app_test

import (
	"log/slog"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/app"
)

func TestNewLoggerLevels(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"garbage", slog.LevelInfo},
	}
	for _, c := range cases {
		l := app.NewLogger(c.in, "development")
		if !l.Enabled(nil, c.want) {
			t.Errorf("level %s: not enabled at want %v", c.in, c.want)
		}
	}
}

func TestNewLoggerNotNil(t *testing.T) {
	if app.NewLogger("info", "production") == nil {
		t.Fatal("nil logger")
	}
}
