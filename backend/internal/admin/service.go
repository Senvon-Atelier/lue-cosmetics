// Package admin provides the Service layer for admin dashboard business logic.
package admin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

// Service provides admin dashboard business logic.
type Service struct {
	Repo        *Repository
	Log         *zap.Logger
	stateMachine *OrderStatusTransition
}

// NewService creates a new admin Service.
func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{
		Repo:        repo,
		Log:         log,
		stateMachine: NewOrderStatusTransition(),
	}
}

// Dashboard

// GetDashboard returns summary statistics for the admin dashboard.
func (s *Service) GetDashboard(ctx context.Context) (Dashboard, error) {
	stats, err := s.Repo.GetDashboardStats(ctx)
	if err != nil {
		s.Log.Error("failed to get dashboard stats", zap.Error(err))
		return Dashboard{}, fmt.Errorf("get dashboard stats: %w", err)
	}

	recentOrders, err := s.Repo.GetRecentOrders(ctx, 10) // Last 10 orders
	if err != nil {
		s.Log.Error("failed to get recent orders", zap.Error(err))
		return Dashboard{}, fmt.Errorf("get recent orders: %w", err)
	}

	return Dashboard{
		Stats:         stats,
		RecentOrders:  recentOrders,
	}, nil
}

// Dashboard contains the complete dashboard data.
type Dashboard struct {
	Stats        sqlcq.GetDashboardStatsRow
	RecentOrders []sqlcq.GetRecentOrdersRow
}

// Orders

// ListOrdersParams contains parameters for listing orders.
type ListOrdersParams struct {
	Status   *string
	DateFrom pgtype.Timestamptz
	DateTo   pgtype.Timestamptz
	Page     int
	PageSize int
}

// PaginatedOrders contains paginated order results.
type PaginatedOrders struct {
	Orders    []sqlcq.ListAllOrdersRow
	Total     int64
	Page      int
	PageSize  int
	TotalPages int
}

// ListOrders returns a paginated list of all orders with optional filters.
func (s *Service) ListOrders(ctx context.Context, params ListOrdersParams) (PaginatedOrders, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	// Calculate offset
	offset := int32((params.Page - 1) * params.PageSize)
	limit := int32(params.PageSize)

	// Get total count
	total, err := s.Repo.CountAllOrders(ctx, params.Status, params.DateFrom, params.DateTo)
	if err != nil {
		s.Log.Error("failed to count orders", zap.Error(err))
		return PaginatedOrders{}, fmt.Errorf("count orders: %w", err)
	}

	// Get orders
	orders, err := s.Repo.ListAllOrders(ctx, ListAllOrdersParams{
		Status:   params.Status,
		DateFrom: params.DateFrom,
		DateTo:   params.DateTo,
		Limit:    &limit,
		Offset:   &offset,
	})
	if err != nil {
		s.Log.Error("failed to list orders", zap.Error(err))
		return PaginatedOrders{}, fmt.Errorf("list orders: %w", err)
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return PaginatedOrders{
		Orders:     orders,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetOrderDetail returns detailed information about a single order.
func (s *Service) GetOrderDetail(ctx context.Context, orderID uuid.UUID) (OrderDetail, error) {
	order, err := s.Repo.GetOrderByID(ctx, orderID)
	if err != nil {
		s.Log.Error("failed to get order", zap.Error(err), zap.String("order_id", orderID.String()))
		return OrderDetail{}, fmt.Errorf("get order: %w", err)
	}

	// Get order items using the orders repository (reuse existing query)
	q := sqlcq.New(s.Repo.Pool())
	items, err := q.ListOrderItems(ctx, orderID)
	if err != nil {
		s.Log.Error("failed to get order items", zap.Error(err), zap.String("order_id", orderID.String()))
		return OrderDetail{}, fmt.Errorf("get order items: %w", err)
	}

	return OrderDetail{
		Order: order,
		Items: items,
	}, nil
}

// OrderDetail contains detailed order information including items.
type OrderDetail struct {
	Order sqlcq.Order
	Items []sqlcq.OrderItem
}

// UpdateOrderStatusParams contains parameters for updating an order status.
type UpdateOrderStatusParams struct {
	OrderID uuid.UUID
	Status  string
}

// OrderStatusTransition validates and manages order status transitions
type OrderStatusTransition struct {
	transitions map[OrderStatus][]OrderStatus
}

// OrderStatus represents the status of an order
type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusFulfilled OrderStatus = "fulfilled"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
	StatusRefunded  OrderStatus = "refunded"
)

// NewOrderStatusTransition creates a new state machine with default transitions
func NewOrderStatusTransition() *OrderStatusTransition {
	return &OrderStatusTransition{
		transitions: defaultTransitions,
	}
}

// CanTransition checks if a status transition is valid
func (sm *OrderStatusTransition) CanTransition(from, to OrderStatus) bool {
	// Same-status transitions are allowed (idempotency)
	if from == to {
		return true
	}

	allowed, ok := sm.transitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}
	return false
}

// defaultTransitions defines the valid order status workflow
var defaultTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusPaid, StatusCancelled},
	StatusPaid:      {StatusFulfilled, StatusCancelled, StatusRefunded},
	StatusFulfilled: {StatusShipped, StatusCancelled},
	StatusShipped:   {StatusDelivered, StatusCancelled},
	StatusDelivered: {StatusRefunded},
	StatusCancelled: {}, // terminal state
	StatusRefunded:  {}, // terminal state
}

// UpdateOrderStatus updates the status of an order.
// This is a simplified version for v1 - in future we'd use proper state machine and transactions.
func (s *Service) UpdateOrderStatus(ctx context.Context, params UpdateOrderStatusParams) error {
	// Validate status transition
	order, err := s.Repo.GetOrderByID(ctx, params.OrderID)
	if err != nil {
		s.Log.Error("failed to get order for status update", zap.Error(err), zap.String("order_id", params.OrderID.String()))
		return fmt.Errorf("get order: %w", err)
	}

	// Validate transition using state machine
	if !s.stateMachine.CanTransition(OrderStatus(order.Status), OrderStatus(params.Status)) {
		s.Log.Warn("invalid status transition attempted",
			zap.String("order_id", params.OrderID.String()),
			zap.String("old_status", order.Status),
			zap.String("new_status", params.Status))
		return fmt.Errorf("invalid status transition from %s to %s", order.Status, params.Status)
	}

	// For v1, we're using a simple UPDATE via the repository
	// In a full implementation, this would use transactions and proper state management
	s.Log.Info("updating order status",
		zap.String("order_id", params.OrderID.String()),
		zap.String("old_status", order.Status),
		zap.String("new_status", params.Status))

	// TODO: Implement the actual UPDATE query
	// For now, this is a placeholder that validates the transition
	return fmt.Errorf("status update not yet implemented - requires migration")
}

// Customers

// ListCustomersParams contains parameters for listing customers.
type ListCustomersParams struct {
	Page     int
	PageSize int
}

// PaginatedCustomers contains paginated customer results.
type PaginatedCustomers struct {
	Customers   []sqlcq.ListAllCustomersRow
	Total       int64
	Page        int
	PageSize    int
	TotalPages  int
}

// ListCustomers returns a paginated list of all customers.
func (s *Service) ListCustomers(ctx context.Context, params ListCustomersParams) (PaginatedCustomers, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	// Calculate offset
	offset := int32((params.Page - 1) * params.PageSize)
	limit := int32(params.PageSize)

	// Get total count
	total, err := s.Repo.CountAllCustomers(ctx)
	if err != nil {
		s.Log.Error("failed to count customers", zap.Error(err))
		return PaginatedCustomers{}, fmt.Errorf("count customers: %w", err)
	}

	// Get customers
	customers, err := s.Repo.ListAllCustomers(ctx, ListAllCustomersParams{
		Offset: &offset,
		Limit:  &limit,
	})
	if err != nil {
		s.Log.Error("failed to list customers", zap.Error(err))
		return PaginatedCustomers{}, fmt.Errorf("list customers: %w", err)
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return PaginatedCustomers{
		Customers:   customers,
		Total:       total,
		Page:        params.Page,
		PageSize:    params.PageSize,
		TotalPages:  totalPages,
	}, nil
}

// GetCustomerDetail returns detailed information about a single customer.
func (s *Service) GetCustomerDetail(ctx context.Context, userID uuid.UUID) (CustomerDetail, error) {
	customer, err := s.Repo.GetCustomerByID(ctx, userID)
	if err != nil {
		s.Log.Error("failed to get customer", zap.Error(err), zap.String("user_id", userID.String()))
		return CustomerDetail{}, fmt.Errorf("get customer: %w", err)
	}

	// Get customer's orders (reuse existing orders query)
	q := sqlcq.New(s.Repo.Pool())
	limit := int32(50) // Show last 50 orders
	orders, err := q.ListOrdersByUserID(ctx, sqlcq.ListOrdersByUserIDParams{
		UserID: userID,
		Limit:  &limit,
		Offset: nil,
		Status: nil,
	})
	if err != nil {
		s.Log.Error("failed to get customer orders", zap.Error(err), zap.String("user_id", userID.String()))
		return CustomerDetail{}, fmt.Errorf("get customer orders: %w", err)
	}

	return CustomerDetail{
		Customer: customer,
		Orders:   orders,
	}, nil
}

// CustomerDetail contains detailed customer information including order history.
type CustomerDetail struct {
	Customer sqlcq.User
	Orders   []sqlcq.Order
}

// Products (read-only in v1)

// ListProductsParams contains parameters for listing products.
type ListProductsParams struct {
	Page     int
	PageSize int
}

// PaginatedProducts contains paginated product results.
type PaginatedProducts struct {
	Products    []sqlcq.Product
	Total       int64
	Page        int
	PageSize    int
	TotalPages  int
}

// ListProducts returns a paginated list of all products.
// Note: This reuses the existing catalog ListProductsByNewest query.
func (s *Service) ListProducts(ctx context.Context, params ListProductsParams) (PaginatedProducts, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	// Calculate offset
	offset := int32((params.Page - 1) * params.PageSize)
	limit := int32(params.PageSize)

	// Get total count (reuse existing CountProducts)
	q := sqlcq.New(s.Repo.Pool())
	total, err := q.CountProducts(ctx, sqlcq.CountProductsParams{
		CategorySlug: nil,
		BrandSlug:    nil,
		Tag:          nil,
		Q:            nil,
	})
	if err != nil {
		s.Log.Error("failed to count products", zap.Error(err))
		return PaginatedProducts{}, fmt.Errorf("count products: %w", err)
	}

	// Get products (reuse existing ListProductsByNewest)
	products, err := q.ListProductsByNewest(ctx, sqlcq.ListProductsByNewestParams{
		Limit:        limit,
		Offset:       offset,
		CategorySlug: nil,
		BrandSlug:    nil,
		Tag:          nil,
		Q:            nil,
	})
	if err != nil {
		s.Log.Error("failed to list products", zap.Error(err))
		return PaginatedProducts{}, fmt.Errorf("list products: %w", err)
	}

	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return PaginatedProducts{
		Products:   products,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetProductDetail returns detailed information about a single product.
func (s *Service) GetProductDetail(ctx context.Context, productID uuid.UUID) (sqlcq.Product, error) {
	return s.Repo.GetProductByID(ctx, productID)
}

// Analytics

// GetRevenueAnalyticsParams contains parameters for revenue analytics.
type GetRevenueAnalyticsParams struct {
	Granularity string // 'day', 'week', or 'month'
	DateFrom    pgtype.Timestamptz
	DateTo      pgtype.Timestamptz
}

// RevenueAnalytics contains revenue analytics data.
type RevenueAnalytics struct {
	ByDate      []sqlcq.GetRevenueByDateRow
	ByCategory  []sqlcq.GetRevenueByCategoryRow
	OrderStats  sqlcq.GetOrderAnalyticsRow
}

// GetRevenueAnalytics returns revenue analytics data.
func (s *Service) GetRevenueAnalytics(ctx context.Context, params GetRevenueAnalyticsParams) (RevenueAnalytics, error) {
	// Validate granularity
	validGranularities := map[string]bool{"day": true, "week": true, "month": true}
	if !validGranularities[params.Granularity] {
		return RevenueAnalytics{}, fmt.Errorf("invalid granularity: %s", params.Granularity)
	}

	// Get revenue by date
	granularityPtr := &params.Granularity
	revenueByDate, err := s.Repo.GetRevenueByDate(ctx, GetRevenueByDateParams{
		Granularity: granularityPtr,
		DateFrom:    params.DateFrom,
		DateTo:      params.DateTo,
	})
	if err != nil {
		s.Log.Error("failed to get revenue by date", zap.Error(err))
		return RevenueAnalytics{}, fmt.Errorf("get revenue by date: %w", err)
	}

	// Get revenue by category
	revenueByCategory, err := s.Repo.GetRevenueByCategory(ctx)
	if err != nil {
		s.Log.Error("failed to get revenue by category", zap.Error(err))
		return RevenueAnalytics{}, fmt.Errorf("get revenue by category: %w", err)
	}

	// Get order analytics
	orderStats, err := s.Repo.GetOrderAnalytics(ctx)
	if err != nil {
		s.Log.Error("failed to get order analytics", zap.Error(err))
		return RevenueAnalytics{}, fmt.Errorf("get order analytics: %w", err)
	}

	return RevenueAnalytics{
		ByDate:     revenueByDate,
		ByCategory: revenueByCategory,
		OrderStats: orderStats,
	}, nil
}

// GetProductStats returns product statistics.
func (s *Service) GetProductStats(ctx context.Context) (sqlcq.GetProductStatsRow, error) {
	return s.Repo.GetProductStats(ctx)
}

// GetCustomerStats returns customer statistics.
func (s *Service) GetCustomerStats(ctx context.Context) (sqlcq.GetCustomerStatsRow, error) {
	return s.Repo.GetCustomerStats(ctx)
}

// GetTopProducts returns top-selling products by revenue.
func (s *Service) GetTopProducts(ctx context.Context, limit int32) ([]sqlcq.GetTopProductsRow, error) {
	return s.Repo.GetTopProducts(ctx, limit)
}
