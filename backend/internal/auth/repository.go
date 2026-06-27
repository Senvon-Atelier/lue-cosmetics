package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

type Repository struct {
	q    *sqlcq.Queries
	pool db.Pool
}

func NewRepository(pool db.Pool) *Repository {
	return &Repository{q: sqlcq.New(pool), pool: pool}
}

// Pool exposes the *pgxpool.Pool so the service can run transactions.
func (r *Repository) Pool() db.Pool { return r.pool }

// withQueries returns a Queries bound to a transaction.
func (r *Repository) withQueries(tx pgx.Tx) *sqlcq.Queries { return sqlcq.New(tx) }

// ErrNotFound is returned by Get*-style methods when no row matches.
var ErrNotFound = errors.New("not found")

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (sqlcq.User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.User{}, ErrNotFound
	}
	return u, err
}

func (r *Repository) GetUserByID(ctx context.Context, id uuid.UUID) (sqlcq.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.User{}, ErrNotFound
	}
	return u, err
}

func (r *Repository) GetPasswordCredential(ctx context.Context, userID uuid.UUID) (sqlcq.PasswordCredential, error) {
	c, err := r.q.GetPasswordCredentialByUserID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.PasswordCredential{}, ErrNotFound
	}
	return c, err
}

func (r *Repository) GetSessionByTokenHash(ctx context.Context, hash []byte) (sqlcq.Session, error) {
	s, err := r.q.GetSessionByTokenHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Session{}, ErrNotFound
	}
	return s, err
}

func (r *Repository) ListRolesForUser(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return r.q.ListRolesForUser(ctx, userID)
}

func (r *Repository) DeleteSession(ctx context.Context, hash []byte) error {
	return r.q.DeleteSession(ctx, hash)
}

func (r *Repository) RefreshSessionLastUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.RefreshSessionLastUsed(ctx, id)
}

func (r *Repository) RollSessionExpiry(ctx context.Context, id uuid.UUID, expires time.Time) error {
	return r.q.RollSessionExpiry(ctx, sqlcq.RollSessionExpiryParams{
		ID:        id,
		ExpiresAt: pgtype.Timestamptz{Time: expires, Valid: true},
	})
}

func (r *Repository) GetUnusedVerificationToken(ctx context.Context, hash []byte, kind string) (sqlcq.VerificationToken, error) {
	v, err := r.q.GetUnusedVerificationToken(ctx, sqlcq.GetUnusedVerificationTokenParams{TokenHash: hash, Kind: kind})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.VerificationToken{}, ErrNotFound
	}
	return v, err
}

func (r *Repository) MarkVerificationTokenUsed(ctx context.Context, id uuid.UUID) error {
	return r.q.MarkVerificationTokenUsed(ctx, id)
}

func (r *Repository) UpsertPasswordCredential(ctx context.Context, userID uuid.UUID, hash string) error {
	return r.q.UpsertPasswordCredential(ctx, sqlcq.UpsertPasswordCredentialParams{UserID: userID, PasswordHash: hash})
}

func (r *Repository) DeleteOtherSessionsForUser(ctx context.Context, userID, keep uuid.UUID) error {
	return r.q.DeleteOtherSessionsForUser(ctx, sqlcq.DeleteOtherSessionsForUserParams{UserID: userID, ID: keep})
}

func (r *Repository) UpsertOAuthAccount(ctx context.Context, userID uuid.UUID, provider, providerAccountID string) error {
	return r.q.UpsertOAuthAccount(ctx, sqlcq.UpsertOAuthAccountParams{
		UserID: userID, Provider: provider, ProviderAccountID: providerAccountID,
	})
}

func (r *Repository) GetOAuthAccount(ctx context.Context, provider, providerAccountID string) (sqlcq.OauthAccount, error) {
	a, err := r.q.GetOAuthAccount(ctx, sqlcq.GetOAuthAccountParams{Provider: provider, ProviderAccountID: providerAccountID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.OauthAccount{}, ErrNotFound
	}
	return a, err
}

func (r *Repository) CreateSession(ctx context.Context, params sqlcq.CreateSessionParams) (sqlcq.Session, error) {
	return r.q.CreateSession(ctx, params)
}

func (r *Repository) CreateVerificationToken(ctx context.Context, params sqlcq.CreateVerificationTokenParams) (sqlcq.VerificationToken, error) {
	return r.q.CreateVerificationToken(ctx, params)
}
