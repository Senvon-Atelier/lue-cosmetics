package auth_test

// Tests for: VerifyEmail, ResendVerification, RequestPasswordReset, ConfirmPasswordReset.

import (
	"context"
	"errors"
	"net"
	"testing"

	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// capturingSender records every Send call so tests can inspect captured tokens.
type capturingSender struct {
	calls []sendCall
}
type sendCall struct {
	To, Template string
	Data         map[string]any
}

func (s *capturingSender) Send(_ context.Context, to, template string, data map[string]any) error {
	s.calls = append(s.calls, sendCall{To: to, Template: template, Data: data})
	return nil
}

// newServiceWithCapture is like newService but replaces the email sender so
// tokens can be captured for verification/reset round-trip tests.
func newServiceWithCapture(t *testing.T) (*auth.Service, db.Pool, *capturingSender, func()) {
	t.Helper()
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	logger := zap.NewNop()
	repo := auth.NewRepository(pool)
	cap := &capturingSender{}
	svc := auth.NewService(repo, logger, cap, nil)
	svc.Params = auth.TestParams
	return svc, pool, cap, cleanup
}

// findCall returns the first captured send call matching the template, or nil.
func findCall(cap *capturingSender, template string) *sendCall {
	for i := range cap.calls {
		if cap.calls[i].Template == template {
			return &cap.calls[i]
		}
	}
	return nil
}

// ── VerifyEmail ───────────────────────────────────────────────────────────────

func TestVerifyEmail_HappyPath(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	svc.Allowlist = []string{"vip@y.test"}
	ctx := context.Background()

	// Signup triggers email_verify token via Email.Send.
	sr, err := svc.Signup(ctx, auth.SignupInput{
		Email: "vip@y.test", Password: "hunter22", Name: "VIP",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if sr.EmailVerified {
		t.Fatal("allowlisted user should not be auto-verified")
	}

	// Capture the token from the signup send call.
	call := findCall(cap, "verify_email")
	if call == nil {
		t.Fatal("expected verify_email send call from Signup")
	}
	rawToken, ok := call.Data["token"].(string)
	if !ok || rawToken == "" {
		t.Fatalf("token not found in send call data: %v", call.Data)
	}

	// Verify with the raw token.
	if err := svc.VerifyEmail(ctx, rawToken); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}

	// Session should now show email_verified = true.
	view, err := svc.GetSession(ctx, sr.SessionToken)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if !view.EmailVerified {
		t.Error("expected email_verified = true after VerifyEmail")
	}
}

func TestVerifyEmail_BadToken(t *testing.T) {
	svc, _, _, cleanup := newServiceWithCapture(t)
	defer cleanup()
	err := svc.VerifyEmail(context.Background(), "notarealtoken")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("want ErrInvalidToken, got %v", err)
	}
}

func TestVerifyEmail_EmptyToken(t *testing.T) {
	svc, _, _, cleanup := newServiceWithCapture(t)
	defer cleanup()
	err := svc.VerifyEmail(context.Background(), "")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("want ErrInvalidToken on empty token, got %v", err)
	}
}

func TestVerifyEmail_ReusedToken(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	svc.Allowlist = []string{"vip2@y.test"}
	ctx := context.Background()

	_, err := svc.Signup(ctx, auth.SignupInput{
		Email: "vip2@y.test", Password: "hunter22",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	call := findCall(cap, "verify_email")
	if call == nil {
		t.Fatal("no verify_email call")
	}
	raw := call.Data["token"].(string)

	// First use succeeds.
	if err := svc.VerifyEmail(ctx, raw); err != nil {
		t.Fatalf("first VerifyEmail: %v", err)
	}
	// Second use returns ErrInvalidToken (used_at IS NOT NULL now).
	if err := svc.VerifyEmail(ctx, raw); !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("want ErrInvalidToken on reuse, got %v", err)
	}
}

// ── ResendVerification ────────────────────────────────────────────────────────

func TestResendVerification_AllowlistedCreatesNewToken(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	svc.Allowlist = []string{"vip3@y.test"}
	ctx := context.Background()

	sr, err := svc.Signup(ctx, auth.SignupInput{
		Email: "vip3@y.test", Password: "hunter22",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}

	beforeCount := len(cap.calls)
	if err := svc.ResendVerification(ctx, sr.UserID, "vip3@y.test"); err != nil {
		t.Fatalf("ResendVerification: %v", err)
	}
	if len(cap.calls) <= beforeCount {
		t.Error("expected a new email send call from ResendVerification")
	}
	// The new token should also work.
	newCall := &cap.calls[len(cap.calls)-1]
	raw := newCall.Data["token"].(string)
	if err := svc.VerifyEmail(ctx, raw); err != nil {
		t.Fatalf("VerifyEmail with resent token: %v", err)
	}
}

func TestResendVerification_NonAllowlistedIsNoop(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	ctx := context.Background()

	sr, err := svc.Signup(ctx, auth.SignupInput{
		Email: "regular@y.test", Password: "hunter22",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	beforeCount := len(cap.calls)
	if err := svc.ResendVerification(ctx, sr.UserID, "regular@y.test"); err != nil {
		t.Fatalf("ResendVerification (non-allowlisted): %v", err)
	}
	if len(cap.calls) != beforeCount {
		t.Error("expected no new send call for non-allowlisted user")
	}
}

// ── RequestPasswordReset ──────────────────────────────────────────────────────

func TestRequestPasswordReset_UnknownEmailIsNoop(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	// Must not panic or return error — just a no-op.
	beforeCount := len(cap.calls)
	svc.RequestPasswordReset(context.Background(), "nobody@unknown.test")
	if len(cap.calls) != beforeCount {
		t.Error("expected no send call for unknown email")
	}
}

func TestRequestPasswordReset_KnownEmailSendsResetEmail(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	ctx := context.Background()

	_, err := svc.Signup(ctx, auth.SignupInput{
		Email: "reset@y.test", Password: "hunter22",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}

	beforeCount := len(cap.calls)
	svc.RequestPasswordReset(ctx, "reset@y.test")
	if len(cap.calls) <= beforeCount {
		t.Error("expected a password_reset send call")
	}
	last := cap.calls[len(cap.calls)-1]
	if last.Template != "password_reset" {
		t.Errorf("template = %q, want password_reset", last.Template)
	}
}

// ── ConfirmPasswordReset ──────────────────────────────────────────────────────

func TestConfirmPasswordReset_HappyPath_InvalidatesAllSessions(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	ctx := context.Background()

	// Sign up and create two sessions.
	sr, err := svc.Signup(ctx, auth.SignupInput{
		Email: "pr@y.test", Password: "oldpassword",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	lr, err := svc.Login(ctx, auth.LoginInput{Email: "pr@y.test", Password: "oldpassword"}, net.IPv4(127, 0, 0, 1), "go-test")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// Request and capture reset token.
	svc.RequestPasswordReset(ctx, "pr@y.test")
	call := findCall(cap, "password_reset")
	if call == nil {
		t.Fatal("no password_reset send call")
	}
	rawToken := call.Data["token"].(string)

	// Confirm.
	if err := svc.ConfirmPasswordReset(ctx, rawToken, "newpassword99"); err != nil {
		t.Fatalf("ConfirmPasswordReset: %v", err)
	}

	// All old sessions must be invalid.
	if _, err := svc.GetSession(ctx, sr.SessionToken); !errors.Is(err, auth.ErrNoSession) {
		t.Errorf("signup session still valid after reset, want ErrNoSession, got %v", err)
	}
	if _, err := svc.GetSession(ctx, lr.SessionToken); !errors.Is(err, auth.ErrNoSession) {
		t.Errorf("login session still valid after reset, want ErrNoSession, got %v", err)
	}

	// New password must work.
	lr2, err := svc.Login(ctx, auth.LoginInput{Email: "pr@y.test", Password: "newpassword99"}, nil, "")
	if err != nil {
		t.Fatalf("Login with new password: %v", err)
	}
	if lr2.SessionToken == "" {
		t.Error("expected a new session token")
	}
}

func TestConfirmPasswordReset_PasswordTooShort(t *testing.T) {
	svc, _, _, cleanup := newServiceWithCapture(t)
	defer cleanup()
	err := svc.ConfirmPasswordReset(context.Background(), "anytoken", "short")
	if !errors.Is(err, auth.ErrInvalidCreds) {
		t.Errorf("want ErrInvalidCreds, got %v", err)
	}
}

func TestConfirmPasswordReset_BadToken(t *testing.T) {
	svc, _, _, cleanup := newServiceWithCapture(t)
	defer cleanup()
	err := svc.ConfirmPasswordReset(context.Background(), "notarealtoken", "validpassword")
	if !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("want ErrInvalidToken, got %v", err)
	}
}

func TestConfirmPasswordReset_ReusedToken(t *testing.T) {
	svc, _, cap, cleanup := newServiceWithCapture(t)
	defer cleanup()
	ctx := context.Background()

	_, err := svc.Signup(ctx, auth.SignupInput{
		Email: "pr2@y.test", Password: "oldpassword",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}

	svc.RequestPasswordReset(ctx, "pr2@y.test")
	call := findCall(cap, "password_reset")
	if call == nil {
		t.Fatal("no password_reset call")
	}
	raw := call.Data["token"].(string)

	if err := svc.ConfirmPasswordReset(ctx, raw, "newpassword99"); err != nil {
		t.Fatalf("first Confirm: %v", err)
	}
	if err := svc.ConfirmPasswordReset(ctx, raw, "newpassword99"); !errors.Is(err, auth.ErrInvalidToken) {
		t.Errorf("want ErrInvalidToken on reuse, got %v", err)
	}
}
