package email_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/email"
)

func TestLogSenderReturnsNil(t *testing.T) {
	s := email.LogSender{Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if err := s.Send(context.Background(), "x@y.test", "welcome", nil); err != nil {
		t.Fatal(err)
	}
}
