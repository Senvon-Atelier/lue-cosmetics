package addresses

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// Pool exposes the underlying pool so the service can run transactions.
func (r *Repository) Pool() db.Pool { return r.pool }

// ErrNotFound is returned when an address lookup misses.
var ErrNotFound = errors.New("addresses: not found")

func (r *Repository) CreateAddress(ctx context.Context, params sqlcq.CreateAddressParams) (sqlcq.Address, error) {
	return r.q.CreateAddress(ctx, params)
}

func (r *Repository) GetAddressByID(ctx context.Context, id uuid.UUID) (sqlcq.Address, error) {
	addr, err := r.q.GetAddressByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Address{}, ErrNotFound
	}
	return addr, err
}

func (r *Repository) ListAddressesByUserID(ctx context.Context, userID uuid.UUID) ([]sqlcq.Address, error) {
	return r.q.ListAddressesByUserID(ctx, userID)
}

func (r *Repository) UpdateAddress(ctx context.Context, params sqlcq.UpdateAddressParams) (sqlcq.Address, error) {
	addr, err := r.q.UpdateAddress(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Address{}, ErrNotFound
	}
	return addr, err
}

func (r *Repository) DeleteAddress(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteAddress(ctx, id)
}

func (r *Repository) CountAddressesByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.q.CountAddressesByUserID(ctx, userID)
}
