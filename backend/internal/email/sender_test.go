package email_test

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/email"
)

func TestLogSenderReturnsNil(t *testing.T) {
	s := email.LogSender{Log: zap.NewNop()}
	if err := s.Send(context.Background(), "x@y.test", "welcome", nil); err != nil {
		t.Fatal(err)
	}
}
