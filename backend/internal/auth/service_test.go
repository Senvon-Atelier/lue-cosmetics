package auth_test

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func newService(t *testing.T) (*auth.Service, db.Pool, func()) {
	t.Helper()
	_, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	logger := zap.NewNop()
	repo := auth.NewRepository(pool)
	svc := auth.NewService(repo, logger, email.LogSender{Log: logger}, nil)
	svc.Params = auth.TestParams // fast hashes in tests
	return svc, pool, cleanup
}

func TestSignupCreatesUserSessionAndAutoVerifies(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	res, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "user@demo.test", Password: "hunter22", Name: "Ada",
	}, net.IPv4(127, 0, 0, 1), "go-test")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if !res.EmailVerified {
		t.Errorf("non-allowlisted address should be auto-verified")
	}
	if res.SessionToken == "" || res.SessionExpires.IsZero() {
		t.Errorf("session not minted: %+v", res)
	}
}

func TestSignupRejectsDuplicateEmail(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	_, err := svc.Signup(ctx, auth.SignupInput{Email: "dup@demo.test", Password: "12345678"}, nil, "")
	if err != nil {
		t.Fatalf("first signup: %v", err)
	}
	_, err = svc.Signup(ctx, auth.SignupInput{Email: "DUP@demo.test", Password: "12345678"}, nil, "")
	if !errors.Is(err, auth.ErrEmailInUse) {
		t.Errorf("dup signup err = %v, want ErrEmailInUse", err)
	}
}

func TestSignupRejectsShortPassword(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	_, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "x@y.test", Password: "short",
	}, nil, "")
	if !errors.Is(err, auth.ErrInvalidCreds) {
		t.Errorf("got %v, want ErrInvalidCreds", err)
	}
}

func TestLoginHappyPath(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	_, err := svc.Signup(ctx, auth.SignupInput{Email: "ok@y.test", Password: "hunter22"}, nil, "")
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	res, err := svc.Login(ctx, auth.LoginInput{Email: "ok@y.test", Password: "hunter22"}, nil, "")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.Role != "customer" {
		t.Errorf("role = %s", res.Role)
	}
}

func TestLoginWrongPasswordSameErrAsMissingUser(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	_, _ = svc.Signup(ctx, auth.SignupInput{Email: "u@y.test", Password: "hunter22"}, nil, "")
	_, e1 := svc.Login(ctx, auth.LoginInput{Email: "u@y.test", Password: "WRONG"}, nil, "")
	_, e2 := svc.Login(ctx, auth.LoginInput{Email: "nobody@y.test", Password: "anything"}, nil, "")
	if !errors.Is(e1, auth.ErrInvalidCreds) || !errors.Is(e2, auth.ErrInvalidCreds) {
		t.Errorf("both errors must be ErrInvalidCreds: e1=%v e2=%v", e1, e2)
	}
}

func TestGetSessionRoundTrip(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	sr, _ := svc.Signup(ctx, auth.SignupInput{Email: "s@y.test", Password: "hunter22", Name: "Ann"}, nil, "")
	view, err := svc.GetSession(ctx, sr.SessionToken)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if view.UserID != sr.UserID {
		t.Errorf("user mismatch")
	}
	if !strings.EqualFold(view.Email, "s@y.test") {
		t.Errorf("email = %s", view.Email)
	}
	if view.Role != "customer" {
		t.Errorf("role = %s", view.Role)
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()
	sr, _ := svc.Signup(ctx, auth.SignupInput{Email: "lo@y.test", Password: "hunter22"}, nil, "")
	if err := svc.Logout(ctx, sr.SessionToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	_, err := svc.GetSession(ctx, sr.SessionToken)
	if !errors.Is(err, auth.ErrNoSession) {
		t.Errorf("after logout: want ErrNoSession, got %v", err)
	}
}

func TestSignupAllowlistedSendsVerifyToken(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	svc.Allowlist = []string{"vip@y.test"}
	res, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "vip@y.test", Password: "hunter22",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if res.EmailVerified {
		t.Errorf("allowlisted address should NOT be auto-verified — verify token issued")
	}
}

// ── LoginWithGoogle ──────────────────────────────────────────────────────────

func TestLoginWithGoogle_NewUser(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()

	res, err := svc.LoginWithGoogle(ctx, "sub-001", "google-new@g.test", "New User", nil, "")
	if err != nil {
		t.Fatalf("LoginWithGoogle: %v", err)
	}
	if res.SessionToken == "" {
		t.Error("expected session token")
	}
	if res.Role != "customer" {
		t.Errorf("role = %s, want customer", res.Role)
	}

	// Verify session is retrievable.
	view, err := svc.GetSession(ctx, res.SessionToken)
	if err != nil {
		t.Fatalf("GetSession: %v", err)
	}
	if !strings.EqualFold(view.Email, "google-new@g.test") {
		t.Errorf("email = %s", view.Email)
	}
}

func TestLoginWithGoogle_ExistingEmailLinked(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()

	// Sign up with password first.
	sr, err := svc.Signup(ctx, auth.SignupInput{
		Email: "link@g.test", Password: "hunter22", Name: "Link User",
	}, nil, "")
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}

	// Now sign in via Google with the same email — should link and return same user.
	res, err := svc.LoginWithGoogle(ctx, "sub-link-001", "link@g.test", "Link User", nil, "")
	if err != nil {
		t.Fatalf("LoginWithGoogle (link): %v", err)
	}
	if res.UserID != sr.UserID {
		t.Errorf("userID mismatch: got %s, want %s", res.UserID, sr.UserID)
	}
	if res.SessionToken == "" {
		t.Error("expected session token")
	}
}

func TestLoginWithGoogle_SameSubReturnsSameUser(t *testing.T) {
	svc, _, cleanup := newService(t)
	defer cleanup()
	ctx := context.Background()

	r1, err := svc.LoginWithGoogle(ctx, "sub-same-001", "same@g.test", "Same User", nil, "")
	if err != nil {
		t.Fatalf("first LoginWithGoogle: %v", err)
	}
	r2, err := svc.LoginWithGoogle(ctx, "sub-same-001", "same@g.test", "Same User", nil, "")
	if err != nil {
		t.Fatalf("second LoginWithGoogle: %v", err)
	}
	if r1.UserID != r2.UserID {
		t.Errorf("expected same user on repeated sub: got %s vs %s", r1.UserID, r2.UserID)
	}
	// Both sessions should be valid.
	if _, err := svc.GetSession(ctx, r1.SessionToken); err != nil {
		t.Errorf("first session invalid: %v", err)
	}
	if _, err := svc.GetSession(ctx, r2.SessionToken); err != nil {
		t.Errorf("second session invalid: %v", err)
	}
}
