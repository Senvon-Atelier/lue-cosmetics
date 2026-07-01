// Package admin provides the Repository layer for admin dashboard queries.
// It wraps sqlc-generated queries and provides domain-specific error handling.
package admin

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

var (
	// ErrNotFound is returned when a requested resource is not found.
	// This matches the pattern used in other repositories.
	ErrNotFound = errors.New("admin: resource not found")
)

// Repository wraps sqlc queries with domain-specific error handling.
type Repository struct {
	q    *sqlcq.Queries
	pool db.Pool
}

// NewRepository creates a new admin Repository.
func NewRepository(pool db.Pool) *Repository {
	return &Repository{
		q:    sqlcq.New(pool),
		pool: pool,
	}
}

// Pool exposes the underlying pool so the service can run transactions.
func (r *Repository) Pool() db.Pool { return r.pool }

// Dashboard

// GetDashboardStats returns summary statistics for the admin dashboard.
func (r *Repository) GetDashboardStats(ctx context.Context) (sqlcq.GetDashboardStatsRow, error) {
	return r.q.GetDashboardStats(ctx)
}

// GetRecentOrders returns the most recent orders with customer information.
func (r *Repository) GetRecentOrders(ctx context.Context, limit int32) ([]sqlcq.GetRecentOrdersRow, error) {
	return r.q.GetRecentOrders(ctx, limit)
}

// Orders

// ListAllOrdersParams contains parameters for listing all orders.
type ListAllOrdersParams struct {
	Status   *string
	DateFrom pgtype.Timestamptz
	DateTo   pgtype.Timestamptz
	Limit    *int32
	Offset   *int32
}

// ListAllOrders returns a paginated list of all orders with optional filters.
func (r *Repository) ListAllOrders(ctx context.Context, params ListAllOrdersParams) ([]sqlcq.ListAllOrdersRow, error) {
	// Convert the repository params to sqlc params
	sqlcParams := sqlcq.ListAllOrdersParams{
		Status:   params.Status,
		DateFrom: params.DateFrom,
		DateTo:   params.DateTo,
		Limit:   params.Limit,
		Offset:  params.Offset,
	}
	return r.q.ListAllOrders(ctx, sqlcParams)
}

// CountAllOrders counts orders with optional status and date filters.
func (r *Repository) CountAllOrders(ctx context.Context, status *string, dateFrom, dateTo pgtype.Timestamptz) (int64, error) {
	return r.q.CountAllOrders(ctx, sqlcq.CountAllOrdersParams{
		Status:   status,
		DateFrom: dateFrom,
		DateTo:   dateTo,
	})
}

// GetOrderAnalytics returns order counts and revenue grouped by status.
func (r *Repository) GetOrderAnalytics(ctx context.Context) (sqlcq.GetOrderAnalyticsRow, error) {
	return r.q.GetOrderAnalytics(ctx)
}

// GetOrderByID returns a single order by ID.
func (r *Repository) GetOrderByID(ctx context.Context, orderID uuid.UUID) (sqlcq.Order, error) {
	order, err := r.q.GetOrderByID(ctx, orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Order{}, ErrNotFound
	}
	return order, err
}

// Customers

// ListAllCustomersParams contains parameters for listing all customers.
type ListAllCustomersParams struct {
	Offset *int32
	Limit  *int32
}

// ListAllCustomers returns a paginated list of all customers.
func (r *Repository) ListAllCustomers(ctx context.Context, params ListAllCustomersParams) ([]sqlcq.ListAllCustomersRow, error) {
	return r.q.ListAllCustomers(ctx, sqlcq.ListAllCustomersParams{
		Offset: params.Offset,
		Limit:  params.Limit,
	})
}

// CountAllCustomers counts total customers.
func (r *Repository) CountAllCustomers(ctx context.Context) (int64, error) {
	return r.q.CountAllCustomers(ctx)
}

// GetCustomerStats returns customer statistics.
func (r *Repository) GetCustomerStats(ctx context.Context) (sqlcq.GetCustomerStatsRow, error) {
	return r.q.GetCustomerStats(ctx)
}

// GetCustomerByID returns a customer (user) by ID.
func (r *Repository) GetCustomerByID(ctx context.Context, userID uuid.UUID) (sqlcq.User, error) {
	user, err := r.q.GetUserByID(ctx, userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.User{}, ErrNotFound
	}
	return user, err
}

// Products

// GetProductStats returns product statistics.
func (r *Repository) GetProductStats(ctx context.Context) (sqlcq.GetProductStatsRow, error) {
	return r.q.GetProductStats(ctx)
}

// GetTopProducts returns top-selling products by revenue.
func (r *Repository) GetTopProducts(ctx context.Context, limit int32) ([]sqlcq.GetTopProductsRow, error) {
	return r.q.GetTopProducts(ctx, limit)
}

// GetProductByID returns a product by ID.
func (r *Repository) GetProductByID(ctx context.Context, productID uuid.UUID) (sqlcq.Product, error) {
	product, err := r.q.GetProductByID(ctx, productID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Product{}, ErrNotFound
	}
	return product, err
}

// Analytics

// GetRevenueByDateParams contains parameters for revenue by date query.
type GetRevenueByDateParams struct {
	Granularity *string // 'day', 'week', or 'month'
	DateFrom    pgtype.Timestamptz
	DateTo      pgtype.Timestamptz
}

// GetRevenueByDate returns revenue aggregated by date.
func (r *Repository) GetRevenueByDate(ctx context.Context, params GetRevenueByDateParams) ([]sqlcq.GetRevenueByDateRow, error) {
	return r.q.GetRevenueByDate(ctx, sqlcq.GetRevenueByDateParams{
		Granularity: params.Granularity,
		DateFrom:    params.DateFrom,
		DateTo:      params.DateTo,
	})
}

// GetRevenueByCategory returns revenue breakdown by product category.
func (r *Repository) GetRevenueByCategory(ctx context.Context) ([]sqlcq.GetRevenueByCategoryRow, error) {
	return r.q.GetRevenueByCategory(ctx)
}
