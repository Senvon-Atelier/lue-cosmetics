package auth

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/logging"
)

var (
	ErrEmailInUse   = errors.New("auth: email already in use")
	ErrInvalidCreds = errors.New("auth: invalid credentials")
	ErrNoSession    = errors.New("auth: no session")
	ErrInvalidToken = errors.New("auth: invalid token")
)

const DefaultSessionLifetime = 30 * 24 * time.Hour
const sessionRollThreshold = 24 * time.Hour

const (
	emailVerifyTTL   = 24 * time.Hour
	passwordResetTTL = 1 * time.Hour
)

type Service struct {
	Repo            *Repository
	Email           email.Sender
	Log             *zap.Logger
	Params          Params
	Allowlist       []string
	SessionLifetime time.Duration
	Now             func() time.Time
}

func NewService(repo *Repository, log *zap.Logger, sender email.Sender, allowlist []string) *Service {
	return &Service{
		Repo:            repo,
		Log:             log,
		Email:           sender,
		Params:          DefaultParams,
		Allowlist:       normalizeAllowlist(allowlist),
		SessionLifetime: DefaultSessionLifetime,
		Now:             time.Now,
	}
}

func normalizeAllowlist(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(strings.ToLower(s))
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// autoVerifyDecision reports whether a signup for emailAddr should mark
// email_verified=true immediately and skip issuing a verify token. The
// historical "allowlist" semantics: addresses NOT on the allowlist are
// dev/local addresses we cannot actually deliver mail to, so we auto-verify
// them. Addresses ON the allowlist will receive a real verify email via the
// wrapper, so they must verify the link the normal way.
func (s *Service) autoVerifyDecision(emailAddr string) bool {
	addr := strings.ToLower(strings.TrimSpace(emailAddr))
	for _, a := range s.Allowlist {
		if a == "*" || a == addr {
			return false
		}
	}
	return true
}

type SignupInput struct {
	Email    string
	Password string
	Name     string
}

type SignupResult struct {
	UserID         uuid.UUID
	SessionToken   string
	SessionExpires time.Time
	EmailVerified  bool
}

func (s *Service) Signup(ctx context.Context, in SignupInput, ip net.IP, ua string) (SignupResult, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	if !validEmail(in.Email) || len(in.Password) < 8 {
		return SignupResult{}, ErrInvalidCreds
	}
	if _, err := s.Repo.GetUserByEmail(ctx, in.Email); err == nil {
		return SignupResult{}, ErrEmailInUse
	} else if !errors.Is(err, ErrNotFound) {
		return SignupResult{}, err
	}
	hash, err := Hash(in.Password, s.Params)
	if err != nil {
		return SignupResult{}, err
	}
	emailVerified := s.autoVerifyDecision(in.Email)

	rawToken, err := NewToken()
	if err != nil {
		return SignupResult{}, err
	}
	tokenHash := HashToken(rawToken)
	expires := s.Now().Add(s.SessionLifetime)

	var result SignupResult
	err = db.WithTx(ctx, s.Repo.Pool(), func(tx pgx.Tx) error {
		q := sqlcq.New(tx)
		user, err := q.CreateUser(ctx, sqlcq.CreateUserParams{
			Email: in.Email, Name: in.Name, EmailVerified: emailVerified,
		})
		if err != nil {
			return err
		}
		if err := q.UpsertPasswordCredential(ctx, sqlcq.UpsertPasswordCredentialParams{
			UserID: user.ID, PasswordHash: hash,
		}); err != nil {
			return err
		}
		if err := q.AddUserRole(ctx, sqlcq.AddUserRoleParams{UserID: user.ID, Role: "customer"}); err != nil {
			return err
		}
		ipa := netToAddr(ip)
		_, err = q.CreateSession(ctx, sqlcq.CreateSessionParams{
			UserID:    user.ID,
			TokenHash: tokenHash[:],
			ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
			Ip:        ipa,
			UserAgent: ua,
		})
		if err != nil {
			return err
		}
		result = SignupResult{
			UserID:         user.ID,
			SessionToken:   rawToken,
			SessionExpires: expires,
			EmailVerified:  emailVerified,
		}
		return nil
	})
	if err != nil {
		return SignupResult{}, err
	}

	// Email side effect AFTER tx commits. Token row is created only when the
	// user is NOT auto-verified (i.e., they need to click a verify link). The
	// Send call is unconditional — the AllowlistSender wrapper installed in
	// app.New decides whether the message is actually delivered or suppressed.
	verifyData := map[string]any{"name": in.Name}
	if !emailVerified {
		verifyRaw, _ := NewToken()
		verifyHash := HashToken(verifyRaw)
		_, _ = s.Repo.CreateVerificationToken(ctx, sqlcq.CreateVerificationTokenParams{
			UserID:    result.UserID,
			Kind:      "email_verify",
			TokenHash: verifyHash[:],
			ExpiresAt: pgtype.Timestamptz{Time: s.Now().Add(emailVerifyTTL), Valid: true},
		})
		verifyData["token"] = verifyRaw
	}
	_ = s.Email.Send(ctx, in.Email, "verify_email", verifyData)
	return result, nil
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginResult struct {
	UserID         uuid.UUID
	Role           string
	SessionToken   string
	SessionExpires time.Time
}

func (s *Service) Login(ctx context.Context, in LoginInput, ip net.IP, ua string) (LoginResult, error) {
	in.Email = strings.TrimSpace(strings.ToLower(in.Email))
	user, err := s.Repo.GetUserByEmail(ctx, in.Email)
	if errors.Is(err, ErrNotFound) {
		// Burn argon2 anyway to keep timing constant.
		_, _ = Hash("decoy-password-for-constant-time", s.Params)
		return LoginResult{}, ErrInvalidCreds
	}
	if err != nil {
		return LoginResult{}, err
	}
	cred, err := s.Repo.GetPasswordCredential(ctx, user.ID)
	if errors.Is(err, ErrNotFound) {
		_, _ = Hash("decoy-password-for-constant-time", s.Params)
		return LoginResult{}, ErrInvalidCreds
	}
	if err != nil {
		return LoginResult{}, err
	}
	ok, err := Verify(in.Password, cred.PasswordHash)
	if err != nil || !ok {
		return LoginResult{}, ErrInvalidCreds
	}
	roles, err := s.Repo.ListRolesForUser(ctx, user.ID)
	if err != nil {
		return LoginResult{}, err
	}
	role := primaryRole(roles)

	rawToken, err := NewToken()
	if err != nil {
		return LoginResult{}, err
	}
	tokenHash := HashToken(rawToken)
	expires := s.Now().Add(s.SessionLifetime)
	_, err = s.Repo.CreateSession(ctx, sqlcq.CreateSessionParams{
		UserID:    user.ID,
		TokenHash: tokenHash[:],
		ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
		Ip:        netToAddr(ip),
		UserAgent: ua,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		UserID:         user.ID,
		Role:           role,
		SessionToken:   rawToken,
		SessionExpires: expires,
	}, nil
}

func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	h := HashToken(rawToken)
	return s.Repo.DeleteSession(ctx, h[:])
}

type SessionView struct {
	SessionID     uuid.UUID
	UserID        uuid.UUID
	Email, Name   string
	Role          string
	EmailVerified bool
	Rolled        bool      // true when GetSession rolled the DB expiry forward
	NewExpires    time.Time // the new expiry when Rolled=true
}

func (s *Service) GetSession(ctx context.Context, rawToken string) (SessionView, error) {
	if rawToken == "" {
		return SessionView{}, ErrNoSession
	}
	h := HashToken(rawToken)
	sess, err := s.Repo.GetSessionByTokenHash(ctx, h[:])
	if errors.Is(err, ErrNotFound) {
		return SessionView{}, ErrNoSession
	}
	if err != nil {
		return SessionView{}, err
	}
	user, err := s.Repo.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return SessionView{}, err
	}
	roles, err := s.Repo.ListRolesForUser(ctx, user.ID)
	if err != nil {
		return SessionView{}, err
	}
	// Best-effort touch; ignore errors (cookie still valid even if write fails).
	now := s.Now()
	_ = s.Repo.RefreshSessionLastUsed(ctx, sess.ID)
	rolled := false
	var newExpires time.Time
	if sess.ExpiresAt.Time.Sub(now) < s.SessionLifetime-sessionRollThreshold {
		newExpires = now.Add(s.SessionLifetime)
		if err := s.Repo.RollSessionExpiry(ctx, sess.ID, newExpires); err == nil {
			rolled = true
		}
	}
	return SessionView{
		SessionID:     sess.ID,
		UserID:        user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Role:          primaryRole(roles),
		EmailVerified: user.EmailVerified,
		Rolled:        rolled,
		NewExpires:    newExpires,
	}, nil
}

func primaryRole(roles []string) string {
	for _, r := range roles {
		if r == "admin" {
			return "admin"
		}
	}
	return "customer"
}

func validEmail(s string) bool {
	if len(s) < 3 || len(s) > 254 {
		return false
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	if strings.IndexByte(s[at+1:], '.') < 0 {
		return false
	}
	return true
}

// LoginWithGoogle logs in or creates a user from a verified Google ID token.
// Fast path: if an oauth_account row exists for (google, providerSub), mint a session.
// Slow path: link by email or create a new user, then upsert the oauth_account row.
func (s *Service) LoginWithGoogle(ctx context.Context, providerSub, emailAddr, name string, ip net.IP, ua string) (LoginResult, error) {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	// Fast path: existing oauth_account.
	if acc, err := s.Repo.GetOAuthAccount(ctx, "google", providerSub); err == nil {
		return s.mintSessionForUser(ctx, acc.UserID, ip, ua)
	} else if !errors.Is(err, ErrNotFound) {
		return LoginResult{}, err
	}
	// Slow path: link or create.
	var userID uuid.UUID
	err := db.WithTx(ctx, s.Repo.Pool(), func(tx pgx.Tx) error {
		q := sqlcq.New(tx)
		user, err := q.GetUserByEmail(ctx, emailAddr)
		if errors.Is(err, pgx.ErrNoRows) {
			user, err = q.CreateUser(ctx, sqlcq.CreateUserParams{
				Email: emailAddr, Name: name, EmailVerified: true,
			})
			if err != nil {
				return err
			}
			if err := q.AddUserRole(ctx, sqlcq.AddUserRoleParams{UserID: user.ID, Role: "customer"}); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := q.UpsertOAuthAccount(ctx, sqlcq.UpsertOAuthAccountParams{
			UserID: user.ID, Provider: "google", ProviderAccountID: providerSub,
		}); err != nil {
			return err
		}
		userID = user.ID
		return nil
	})
	if err != nil {
		return LoginResult{}, err
	}
	return s.mintSessionForUser(ctx, userID, ip, ua)
}

func (s *Service) mintSessionForUser(ctx context.Context, userID uuid.UUID, ip net.IP, ua string) (LoginResult, error) {
	roles, err := s.Repo.ListRolesForUser(ctx, userID)
	if err != nil {
		return LoginResult{}, err
	}
	rawToken, err := NewToken()
	if err != nil {
		return LoginResult{}, err
	}
	tokenHash := HashToken(rawToken)
	expires := s.Now().Add(s.SessionLifetime)
	_, err = s.Repo.CreateSession(ctx, sqlcq.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash[:],
		ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
		Ip:        netToAddr(ip),
		UserAgent: ua,
	})
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{
		UserID:         userID,
		Role:           primaryRole(roles),
		SessionToken:   rawToken,
		SessionExpires: expires,
	}, nil
}

// VerifyEmail marks a user's email as verified using the raw token from the
// verification email. Returns ErrInvalidToken when the token is missing,
// not found, already used, or expired.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return ErrInvalidToken
	}
	h := HashToken(rawToken)
	tok, err := s.Repo.GetUnusedVerificationToken(ctx, h[:], "email_verify")
	if errors.Is(err, ErrNotFound) {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	return db.WithTx(ctx, s.Repo.Pool(), func(tx pgx.Tx) error {
		q := sqlcq.New(tx)
		if err := q.UpdateUserEmailVerified(ctx, sqlcq.UpdateUserEmailVerifiedParams{
			ID: tok.UserID, EmailVerified: true,
		}); err != nil {
			return err
		}
		return q.MarkVerificationTokenUsed(ctx, tok.ID)
	})
}

// ResendVerification issues a new email_verify token and sends the verification
// email. If the user is already email_verified, this is a no-op (no token
// created, no Send call). The AllowlistSender wrapper installed in app.New
// decides whether the Send actually delivers.
func (s *Service) ResendVerification(ctx context.Context, userID uuid.UUID, emailAddr string) error {
	user, err := s.Repo.GetUserByID(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if user.EmailVerified {
		return nil
	}
	raw, err := NewToken()
	if err != nil {
		return err
	}
	h := HashToken(raw)
	_, err = s.Repo.CreateVerificationToken(ctx, sqlcq.CreateVerificationTokenParams{
		UserID:    userID,
		Kind:      "email_verify",
		TokenHash: h[:],
		ExpiresAt: pgtype.Timestamptz{Time: s.Now().Add(emailVerifyTTL), Valid: true},
	})
	if err != nil {
		return err
	}
	return s.Email.Send(ctx, emailAddr, "verify_email", map[string]any{"token": raw})
}

// RequestPasswordReset looks up the user by email and, if found, creates a
// password_reset token and sends the reset email. Errors are silently logged
// and never surfaced to the caller — preventing email enumeration.
func (s *Service) RequestPasswordReset(ctx context.Context, emailAddr string) {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	user, err := s.Repo.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		logging.From(ctx, s.Log).Info("password-reset request for unknown email", zap.String("email", emailAddr))
		return
	}
	raw, err := NewToken()
	if err != nil {
		logging.From(ctx, s.Log).Error("password-reset token gen", zap.Error(err))
		return
	}
	h := HashToken(raw)
	if _, err := s.Repo.CreateVerificationToken(ctx, sqlcq.CreateVerificationTokenParams{
		UserID:    user.ID,
		Kind:      "password_reset",
		TokenHash: h[:],
		ExpiresAt: pgtype.Timestamptz{Time: s.Now().Add(passwordResetTTL), Valid: true},
	}); err != nil {
		logging.From(ctx, s.Log).Error("password-reset token insert", zap.Error(err))
		return
	}
	if err := s.Email.Send(ctx, emailAddr, "password_reset", map[string]any{"token": raw}); err != nil {
		logging.From(ctx, s.Log).Error("password-reset email send", zap.Error(err))
	}
}

// ConfirmPasswordReset validates the reset token, rotates the password, marks
// the token used, and deletes ALL sessions for the user — all inside a single
// transaction. Returns ErrInvalidCreds when newPassword is shorter than 8 chars
// and ErrInvalidToken when the token is absent, expired, or already used.
func (s *Service) ConfirmPasswordReset(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrInvalidCreds
	}
	if rawToken == "" {
		return ErrInvalidToken
	}
	h := HashToken(rawToken)
	tok, err := s.Repo.GetUnusedVerificationToken(ctx, h[:], "password_reset")
	if errors.Is(err, ErrNotFound) {
		return ErrInvalidToken
	}
	if err != nil {
		return err
	}
	hash, err := Hash(newPassword, s.Params)
	if err != nil {
		return err
	}
	return db.WithTx(ctx, s.Repo.Pool(), func(tx pgx.Tx) error {
		q := sqlcq.New(tx)
		if err := q.UpsertPasswordCredential(ctx, sqlcq.UpsertPasswordCredentialParams{
			UserID: tok.UserID, PasswordHash: hash,
		}); err != nil {
			return err
		}
		if err := q.MarkVerificationTokenUsed(ctx, tok.ID); err != nil {
			return err
		}
		return q.DeleteSessionsForUser(ctx, tok.UserID)
	})
}

// netToAddr converts a net.IP to *netip.Addr for use in CreateSessionParams.
// Returns nil if ip is nil or cannot be parsed.
func netToAddr(ip net.IP) *netip.Addr {
	if ip == nil {
		return nil
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return nil
	}
	addr = addr.Unmap()
	return &addr
}
