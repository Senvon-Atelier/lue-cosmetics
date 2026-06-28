package logging_test

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/logging"
)

func TestFromReturnsContextLoggerWhenPresent(t *testing.T) {
	base := zap.NewNop()
	scoped := base.With(zap.String("test", "yes"))
	ctx := logging.WithLogger(context.Background(), scoped)
	got := logging.From(ctx, base)
	if got != scoped {
		t.Errorf("From returned %p, want %p", got, scoped)
	}
}

func TestFromReturnsFallbackWhenAbsent(t *testing.T) {
	base := zap.NewNop()
	got := logging.From(context.Background(), base)
	if got != base {
		t.Errorf("From returned %p, want fallback %p", got, base)
	}
}
