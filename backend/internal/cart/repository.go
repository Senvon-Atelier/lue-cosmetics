package cart

import (
	"context"
	"errors"

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

// ErrNotFound is returned when a cart or item lookup fails.
var ErrNotFound = errors.New("cart: not found")

func (r *Repository) GetCartByUserID(ctx context.Context, userID uuid.UUID) (sqlcq.Cart, error) {
	c, err := r.q.GetCartByUserID(ctx, pgtype.UUID{Bytes: userID, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Cart{}, ErrNotFound
	}
	return c, err
}

func (r *Repository) GetCartByGuestToken(ctx context.Context, token string) (sqlcq.Cart, error) {
	c, err := r.q.GetCartByGuestToken(ctx, &token)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Cart{}, ErrNotFound
	}
	return c, err
}

func (r *Repository) CreateCartForUser(ctx context.Context, userID uuid.UUID) (sqlcq.Cart, error) {
	return r.q.CreateCartForUser(ctx, pgtype.UUID{Bytes: userID, Valid: true})
}

func (r *Repository) CreateCartForGuest(ctx context.Context, token string) (sqlcq.Cart, error) {
	return r.q.CreateCartForGuest(ctx, &token)
}

func (r *Repository) DeleteCart(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteCart(ctx, id)
}

func (r *Repository) TouchCart(ctx context.Context, id uuid.UUID) error {
	return r.q.TouchCart(ctx, id)
}

func (r *Repository) ListCartItems(ctx context.Context, cartID uuid.UUID) ([]sqlcq.CartItem, error) {
	return r.q.ListCartItems(ctx, cartID)
}

func (r *Repository) GetCartItemByID(ctx context.Context, itemID, cartID uuid.UUID) (sqlcq.CartItem, error) {
	i, err := r.q.GetCartItemByID(ctx, sqlcq.GetCartItemByIDParams{ID: itemID, CartID: cartID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.CartItem{}, ErrNotFound
	}
	return i, err
}

func (r *Repository) GetCartItemByProduct(ctx context.Context, cartID, productID uuid.UUID) (sqlcq.CartItem, error) {
	i, err := r.q.GetCartItemByProduct(ctx, sqlcq.GetCartItemByProductParams{CartID: cartID, ProductID: productID})
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.CartItem{}, ErrNotFound
	}
	return i, err
}

func (r *Repository) UpsertCartItemAddQty(ctx context.Context, cartID, productID uuid.UUID, qty int32, unitPrice int64) (sqlcq.CartItem, error) {
	return r.q.UpsertCartItemAddQty(ctx, sqlcq.UpsertCartItemAddQtyParams{
		CartID:          cartID,
		ProductID:       productID,
		Qty:             qty,
		UnitPriceGhsMinor: unitPrice,
	})
}

func (r *Repository) SetCartItemQty(ctx context.Context, itemID, cartID uuid.UUID, qty int32) error {
	return r.q.SetCartItemQty(ctx, sqlcq.SetCartItemQtyParams{ID: itemID, CartID: cartID, Qty: qty})
}

func (r *Repository) DeleteCartItem(ctx context.Context, itemID, cartID uuid.UUID) error {
	return r.q.DeleteCartItem(ctx, sqlcq.DeleteCartItemParams{ID: itemID, CartID: cartID})
}
