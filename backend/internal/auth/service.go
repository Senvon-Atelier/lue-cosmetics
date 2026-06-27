package auth

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/email"
)

var (
	ErrEmailInUse   = errors.New("auth: email already in use")
	ErrInvalidCreds = errors.New("auth: invalid credentials")
	ErrNoSession    = errors.New("auth: no session")
	ErrInvalidToken = errors.New("auth: invalid token")
)

const DefaultSessionLifetime = 30 * 24 * time.Hour
const sessionRollThreshold = 24 * time.Hour

type Service struct {
	Repo            *Repository
	Email           email.Sender
	Log             *slog.Logger
	Params          Params
	Allowlist       []string
	SessionLifetime time.Duration
	Now             func() time.Time
}

func NewService(repo *Repository, log *slog.Logger, sender email.Sender, allowlist []string) *Service {
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

func (s *Service) isAllowlisted(emailAddr string) bool {
	addr := strings.ToLower(strings.TrimSpace(emailAddr))
	for _, a := range s.Allowlist {
		if a == "*" || a == addr {
			return true
		}
	}
	return false
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
	allow := s.isAllowlisted(in.Email)
	emailVerified := !allow // non-allowlisted → auto-verified at signup

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

	// Email side effect AFTER tx commits.
	if allow {
		verifyRaw, _ := NewToken()
		verifyHash := HashToken(verifyRaw)
		_, _ = s.Repo.CreateVerificationToken(ctx, sqlcq.CreateVerificationTokenParams{
			UserID:    result.UserID,
			Kind:      "email_verify",
			TokenHash: verifyHash[:],
			ExpiresAt: pgtype.Timestamptz{Time: s.Now().Add(24 * time.Hour), Valid: true},
		})
		_ = s.Email.Send(ctx, in.Email, "verify_email", map[string]any{
			"token": verifyRaw, "name": in.Name,
		})
	} else {
		_ = s.Email.Send(ctx, in.Email, "welcome", map[string]any{"name": in.Name})
	}
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
	if sess.ExpiresAt.Time.Sub(now) < s.SessionLifetime-sessionRollThreshold {
		_ = s.Repo.RollSessionExpiry(ctx, sess.ID, now.Add(s.SessionLifetime))
	}
	return SessionView{
		SessionID:     sess.ID,
		UserID:        user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Role:          primaryRole(roles),
		EmailVerified: user.EmailVerified,
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
