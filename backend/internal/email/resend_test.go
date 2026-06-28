package email

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestNewResendSender_EmptyAPIKey(t *testing.T) {
	r, err := NewResendSender("", "from@example.com", nil, zap.NewNop())
	if r != nil {
		t.Errorf("expected nil sender, got %+v", r)
	}
	if !errors.Is(err, ErrResendNotConfigured) {
		t.Errorf("err = %v, want ErrResendNotConfigured", err)
	}
}

func TestNewResendSender_EmptyFromEmail(t *testing.T) {
	r, err := NewResendSender("re_key", "", nil, zap.NewNop())
	if r != nil {
		t.Errorf("expected nil sender, got %+v", r)
	}
	if !errors.Is(err, ErrResendNotConfigured) {
		t.Errorf("err = %v, want ErrResendNotConfigured", err)
	}
}

func TestNewResendSender_Configured(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	s, err := NewResendSender("re_test_key", "noreply@rue.example.com", renderer, zap.NewNop())
	if err != nil {
		t.Fatalf("NewResendSender: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil sender")
	}
	if s.FromEmail != "noreply@rue.example.com" {
		t.Errorf("FromEmail = %q", s.FromEmail)
	}
	if s.Client == nil {
		t.Error("expected resend client to be constructed")
	}
}

func TestSubjectFor(t *testing.T) {
	cases := []struct {
		template string
		data     map[string]any
		want     string
	}{
		{"verify_email", nil, "Verify your email for Rue Cosmetics"},
		{"password_reset", nil, "Reset your password"},
		{"welcome", nil, "Welcome to Rue Cosmetics"},
		{"order_confirmation", map[string]any{"paystack_reference": "RUE-DEADBEEF"}, "Your Rue Cosmetics order RUE-DEADBEEF"},
		{"order_confirmation", nil, "Your Rue Cosmetics order"},
		{"unknown_template", nil, "Rue Cosmetics"},
	}
	for _, tc := range cases {
		if got := subjectFor(tc.template, tc.data); got != tc.want {
			t.Errorf("subjectFor(%q, %v) = %q, want %q", tc.template, tc.data, got, tc.want)
		}
	}
}
