package email

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

type capturingSender struct {
	calls []capturedCall
}

type capturedCall struct {
	To       string
	Template string
	Data     map[string]any
}

func (c *capturingSender) Send(ctx context.Context, to, template string, data map[string]any) error {
	c.calls = append(c.calls, capturedCall{To: to, Template: template, Data: data})
	return nil
}

func TestAllowlistSender_AllowlistedDelegates(t *testing.T) {
	inner := &capturingSender{}
	s := NewAllowlistSender(inner, []string{"vip@y.test"}, zap.NewNop())

	if err := s.Send(context.Background(), "vip@y.test", "verify_email", map[string]any{"token": "abc"}); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(inner.calls) != 1 {
		t.Fatalf("expected 1 inner call, got %d", len(inner.calls))
	}
	if inner.calls[0].To != "vip@y.test" || inner.calls[0].Template != "verify_email" {
		t.Errorf("inner call args wrong: %+v", inner.calls[0])
	}

	if err := s.Send(context.Background(), "rando@y.test", "verify_email", nil); err != nil {
		t.Fatalf("Send rando: %v", err)
	}
	if len(inner.calls) != 1 {
		t.Errorf("expected non-allowlisted Send to be suppressed; inner calls = %d", len(inner.calls))
	}
}

func TestAllowlistSender_WildcardAllowsAll(t *testing.T) {
	inner := &capturingSender{}
	s := NewAllowlistSender(inner, []string{"*"}, zap.NewNop())

	addrs := []string{"a@y.test", "b@y.test", "c@y.test"}
	for _, a := range addrs {
		if err := s.Send(context.Background(), a, "welcome", nil); err != nil {
			t.Fatalf("Send %s: %v", a, err)
		}
	}
	if len(inner.calls) != len(addrs) {
		t.Errorf("expected %d delegated calls, got %d", len(addrs), len(inner.calls))
	}
}

func TestAllowlistSender_EmptyAllowlistSuppressesAll(t *testing.T) {
	inner := &capturingSender{}
	s := NewAllowlistSender(inner, nil, zap.NewNop())

	if err := s.Send(context.Background(), "anyone@y.test", "verify_email", nil); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(inner.calls) != 0 {
		t.Errorf("expected empty allowlist to suppress all; got %d calls", len(inner.calls))
	}
}

func TestAllowlistSender_CaseInsensitive(t *testing.T) {
	inner := &capturingSender{}
	s := NewAllowlistSender(inner, []string{"VIP@Y.TEST"}, zap.NewNop())

	if err := s.Send(context.Background(), "vip@y.test", "verify_email", nil); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(inner.calls) != 1 {
		t.Errorf("expected case-insensitive match; got %d calls", len(inner.calls))
	}
}
