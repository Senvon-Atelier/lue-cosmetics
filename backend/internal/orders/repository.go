package orders

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

// ErrNotFound is returned when an order lookup misses.
var ErrNotFound = errors.New("orders: not found")

func (r *Repository) CreateOrder(ctx context.Context, params sqlcq.CreateOrderParams) (sqlcq.Order, error) {
	return r.q.CreateOrder(ctx, params)
}

func (r *Repository) CreateOrderItem(ctx context.Context, params sqlcq.CreateOrderItemParams) (sqlcq.OrderItem, error) {
	return r.q.CreateOrderItem(ctx, params)
}

func (r *Repository) GetOrderByReference(ctx context.Context, ref string) (sqlcq.Order, error) {
	o, err := r.q.GetOrderByReference(ctx, ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Order{}, ErrNotFound
	}
	return o, err
}

func (r *Repository) GetOrderByID(ctx context.Context, id uuid.UUID) (sqlcq.Order, error) {
	o, err := r.q.GetOrderByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Order{}, ErrNotFound
	}
	return o, err
}

func (r *Repository) ListOrderItems(ctx context.Context, orderID uuid.UUID) ([]sqlcq.OrderItem, error) {
	return r.q.ListOrderItems(ctx, orderID)
}

func (r *Repository) CountOrdersByStatus(ctx context.Context, status string) (int64, error) {
	return r.q.CountOrdersByStatus(ctx, status)
}

// ListOrdersByUserID returns a paginated list of orders for a user, optionally filtered by status.
func (r *Repository) ListOrdersByUserID(ctx context.Context, userID uuid.UUID, status string, limit, offset int32) ([]sqlcq.Order, error) {
	// Convert status to pointer if provided
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	return r.q.ListOrdersByUserID(ctx, sqlcq.ListOrdersByUserIDParams{
		UserID:  userID,
		Status:  statusPtr,
		Limit:   &limit,
		Offset:  &offset,
	})
}

// CountOrdersByUserID returns the total count of orders for a user, optionally filtered by status.
func (r *Repository) CountOrdersByUserID(ctx context.Context, userID uuid.UUID, status string) (int64, error) {
	// Convert status to pointer if provided
	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	return r.q.CountOrdersByUserID(ctx, sqlcq.CountOrdersByUserIDParams{
		UserID: userID,
		Status: statusPtr,
	})
}
