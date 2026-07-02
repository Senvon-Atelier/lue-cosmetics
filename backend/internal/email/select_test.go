package email

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("renderer: %v", err)
	}
	return r
}

func TestSelectWithoutKeyInDevelopmentUsesLogSender(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	s, err := Select("development", "", "noreply@rue.example.com", nil, newTestRenderer(t), zap.New(core))
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if s == nil {
		t.Fatal("nil sender")
	}
	if logs.Len() != 1 || !strings.Contains(logs.All()[0].Message, "log-only") {
		t.Fatalf("expected one log-only warning, got %v", logs.All())
	}
}

func TestSelectWithoutKeyInProductionFails(t *testing.T) {
	_, err := Select("production", "", "noreply@rue.example.com", nil, newTestRenderer(t), zap.NewNop())
	if err == nil {
		t.Fatal("expected error when production has no email provider")
	}
}

func TestSelectWithKeyUsesResend(t *testing.T) {
	// NewResendSender creates a resend.Client via resend.NewClient which does
	// no network I/O at construction time — it's safe to use a fake key here.
	s, err := Select("production", "re_test_key", "noreply@rue.example.com", nil, newTestRenderer(t), zap.NewNop())
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	// AllowlistSender is a value type (not pointer), so assert accordingly.
	allow, ok := s.(AllowlistSender)
	if !ok {
		t.Fatalf("outermost sender = %T, want AllowlistSender", s)
	}
	if _, ok := allow.Inner.(*ResendSender); !ok {
		t.Fatalf("inner sender = %T, want *ResendSender", allow.Inner)
	}
}
