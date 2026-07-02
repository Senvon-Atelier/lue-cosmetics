// Package admin provides HTTP handlers for the admin dashboard.
package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

// Handlers provides HTTP handlers for admin operations.
type Handlers struct {
	Svc          *Service
	AuthHandlers *auth.Handlers
}

// NewHandlers creates a new admin Handlers instance.
func NewHandlers(svc *Service, authHandlers *auth.Handlers) *Handlers {
	return &Handlers{
		Svc:          svc,
		AuthHandlers: authHandlers,
	}
}

// MountPublic mounts admin routes under /api/v1/admin.
// All routes require authentication + admin role.
func (h *Handlers) MountPublic(r chi.Router) {
	// All admin routes require session + admin role
	r.Group(func(r chi.Router) {
		r.Use(h.AuthHandlers.RequireSession)
		r.Use(h.AuthHandlers.RequireRole("admin"))

		// Dashboard
		r.Get("/dashboard", h.getDashboard)

		// Orders
		r.Get("/orders", h.listOrders)
		r.Get("/orders/{id}", h.getOrderDetail)
		r.Patch("/orders/{id}/status", h.updateOrderStatus)

		// Customers
		r.Get("/customers", h.listCustomers)
		r.Get("/customers/{id}", h.getCustomerDetail)

		// Products (read-only in v1)
		r.Get("/products", h.listProducts)
		r.Get("/products/{id}", h.getProductDetail)

		// Analytics
		r.Get("/analytics/revenue", h.getRevenueAnalytics)
		r.Get("/analytics/stats", h.getStats)

		// Placeholder endpoints for v1
		r.Get("/marketing", h.getMarketingPlaceholder)
		r.Get("/content", h.getContentPlaceholder)
		r.Get("/settings", h.getSettingsPlaceholder)
	})
}

// Dashboard

// getDashboard returns dashboard statistics and recent orders.
// @Summary Get admin dashboard
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} DashboardResponse
// @Router /admin/dashboard [get]
func (h *Handlers) getDashboard(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	dashboard, err := h.Svc.GetDashboard(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get dashboard", nil)
		return
	}

	// Convert dashboard stats
	stats := DashboardStats{
		TotalRevenueGhsMinor: dashboard.Stats.TotalRevenueGhsMinor,
		TotalOrders:          dashboard.Stats.TotalOrders,
		PendingOrders:        dashboard.Stats.PendingOrders,
		PaidOrders:           dashboard.Stats.PaidOrders,
		ShippedOrders:        dashboard.Stats.ShippedOrders,
		DeliveredOrders:      dashboard.Stats.DeliveredOrders,
		TotalCustomers:       dashboard.Stats.TotalCustomers,
		TotalProducts:        dashboard.Stats.TotalProducts,
	}

	// Convert recent orders
	recentOrders := make([]RecentOrder, len(dashboard.RecentOrders))
	for i, order := range dashboard.RecentOrders {
		createdAt := ""
		if order.CreatedAt.Valid {
			createdAt = order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		recentOrders[i] = RecentOrder{
			ID:            order.ID.String(),
			UserID:        order.UserID.String(),
			Status:        order.Status,
			TotalGhsMinor: order.TotalGhsMinor,
			PaystackRef:   order.PaystackReference,
			CreatedAt:     createdAt,
			CustomerEmail: order.CustomerEmail,
			CustomerName:  order.CustomerName,
		}
	}

	httpx.WriteJSON(w, http.StatusOK, DashboardResponse{
		Stats:        stats,
		RecentOrders: recentOrders,
	})
}

// DashboardResponse contains dashboard statistics and recent orders.
type DashboardResponse struct {
	Stats        DashboardStats `json:"stats"`
	RecentOrders []RecentOrder  `json:"recent_orders"`
}

// DashboardStats contains dashboard statistics.
type DashboardStats struct {
	TotalRevenueGhsMinor int64 `json:"total_revenue_ghs_minor"`
	TotalOrders          int32 `json:"total_orders"`
	PendingOrders        int32 `json:"pending_orders"`
	PaidOrders           int32 `json:"paid_orders"`
	ShippedOrders        int32 `json:"shipped_orders"`
	DeliveredOrders      int32 `json:"delivered_orders"`
	TotalCustomers       int32 `json:"total_customers"`
	TotalProducts        int32 `json:"total_products"`
}

// RecentOrder represents a recent order with customer information.
type RecentOrder struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	Status        string `json:"status"`
	TotalGhsMinor int64  `json:"total_ghs_minor"`
	PaystackRef   string `json:"paystack_reference"`
	CreatedAt     string `json:"created_at"`
	CustomerEmail string `json:"customer_email"`
	CustomerName  string `json:"customer_name"`
}

// Orders

// listOrders returns a paginated list of all orders with optional filters.
// @Summary List all orders
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} OrdersResponse
// @Param page      query int    false "Page number (1-based)"
// @Param page_size query int    false "Items per page"
// @Param status    query string false "Filter by order status"
// @Param date_from query string false "RFC3339 lower bound"
// @Param date_to   query string false "RFC3339 upper bound"
// @Router /admin/orders [get]
func (h *Handlers) listOrders(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	// Parse query parameters
	params := ListOrdersParams{
		Page:     httpx.QueryInt(r, "page", 1),
		PageSize: httpx.QueryInt(r, "page_size", 20),
	}

	// Parse optional status filter
	if status := r.URL.Query().Get("status"); status != "" {
		params.Status = &status
	}

	// Parse optional date filters
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if t, err := httpx.ParseTime(dateFrom); err == nil {
			params.DateFrom = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if t, err := httpx.ParseTime(dateTo); err == nil {
			params.DateTo = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	orders, err := h.Svc.ListOrders(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list orders", nil)
		return
	}

	response := OrdersResponse{
		Orders:     make([]OrderSummary, len(orders.Orders)),
		Total:      orders.Total,
		Page:       orders.Page,
		PageSize:   orders.PageSize,
		TotalPages: orders.TotalPages,
	}

	for i, order := range orders.Orders {
		createdAt := ""
		if order.CreatedAt.Valid {
			createdAt = order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		updatedAt := ""
		if order.UpdatedAt.Valid {
			updatedAt = order.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		// Handle CustomerPhone which might be nil
		customerPhone := ""
		if order.CustomerPhone != nil {
			if phoneStr, ok := order.CustomerPhone.(string); ok {
				customerPhone = phoneStr
			}
		}

		response.Orders[i] = OrderSummary{
			ID:               order.ID.String(),
			UserID:           order.UserID.String(),
			Status:           order.Status,
			SubtotalGhsMinor: order.SubtotalGhsMinor,
			ShippingGhsMinor: order.ShippingGhsMinor,
			TotalGhsMinor:    order.TotalGhsMinor,
			PaystackRef:      order.PaystackReference,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
			CustomerEmail:    order.CustomerEmail,
			CustomerName:     order.CustomerName,
			CustomerPhone:    customerPhone,
		}
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// OrdersResponse contains paginated order results.
type OrdersResponse struct {
	Orders     []OrderSummary `json:"orders"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// OrderSummary represents an order in a list.
type OrderSummary struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Status           string `json:"status"`
	SubtotalGhsMinor int64  `json:"subtotal_ghs_minor"`
	ShippingGhsMinor int64  `json:"shipping_ghs_minor"`
	TotalGhsMinor    int64  `json:"total_ghs_minor"`
	PaystackRef      string `json:"paystack_reference"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	CustomerEmail    string `json:"customer_email"`
	CustomerName     string `json:"customer_name"`
	CustomerPhone    string `json:"customer_phone,omitempty"`
}

// getOrderDetail returns detailed information about a single order.
// @Summary Get order details
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} OrderDetailResponse
// @Failure 400 {object} httpx.ErrorEnvelope
// @Failure 401 {object} httpx.ErrorEnvelope
// @Failure 404 {object} httpx.ErrorEnvelope
// @Router /admin/orders/{id} [get]
func (h *Handlers) getOrderDetail(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid order ID", nil)
		return
	}

	detail, err := h.Svc.GetOrderDetail(r.Context(), orderID)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "order not found", nil)
		} else {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get order detail", nil)
		}
		return
	}

	response := OrderDetailResponse{
		Order: OrderInfo{
			ID:               detail.Order.ID.String(),
			UserID:           detail.Order.UserID.String(),
			Status:           detail.Order.Status,
			SubtotalGhsMinor: detail.Order.SubtotalGhsMinor,
			ShippingGhsMinor: detail.Order.ShippingGhsMinor,
			TotalGhsMinor:    detail.Order.TotalGhsMinor,
			PaystackRef:      detail.Order.PaystackReference,
		},
		Items: make([]OrderItemInfo, len(detail.Items)),
	}

	// Handle timestamps
	if detail.Order.CreatedAt.Valid {
		response.Order.CreatedAt = detail.Order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}
	if detail.Order.UpdatedAt.Valid {
		response.Order.UpdatedAt = detail.Order.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
	}

	for i, item := range detail.Items {
		response.Items[i] = OrderItemInfo{
			ID:                   item.ID.String(),
			ProductID:            item.ProductID.String(),
			Qty:                  item.Qty,
			UnitPriceGhsMinor:    item.UnitPriceGhsMinor,
			ProductNameSnapshot:  item.ProductNameSnapshot,
			ProductBrandSnapshot: item.ProductBrandSnapshot,
			ProductImageSnapshot: item.ProductImageSnapshot,
		}
	}

	// Parse shipping address from JSONB (raw []byte)
	if len(detail.Order.ShippingAddress) > 0 {
		var shippingAddr ShippingAddress
		if err := json.Unmarshal(detail.Order.ShippingAddress, &shippingAddr); err == nil {
			response.Order.ShippingAddress = &shippingAddr
		}
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// OrderDetailResponse contains detailed order information.
type OrderDetailResponse struct {
	Order OrderInfo       `json:"order"`
	Items []OrderItemInfo `json:"items"`
}

// OrderInfo represents detailed order information.
type OrderInfo struct {
	ID               string           `json:"id"`
	UserID           string           `json:"user_id"`
	Status           string           `json:"status"`
	SubtotalGhsMinor int64            `json:"subtotal_ghs_minor"`
	ShippingGhsMinor int64            `json:"shipping_ghs_minor"`
	TotalGhsMinor    int64            `json:"total_ghs_minor"`
	PaystackRef      string           `json:"paystack_reference"`
	ShippingAddress  *ShippingAddress `json:"shipping_address,omitempty"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
}

// ShippingAddress represents a shipping address.
type ShippingAddress struct {
	Line1  string `json:"line1"`
	Line2  string `json:"line2,omitempty"`
	City   string `json:"city"`
	Region string `json:"region"`
	Phone  string `json:"phone"`
	Label  string `json:"label,omitempty"`
}

// OrderItemInfo represents an order item.
type OrderItemInfo struct {
	ID                   string `json:"id"`
	ProductID            string `json:"product_id"`
	Qty                  int32  `json:"qty"`
	UnitPriceGhsMinor    int64  `json:"unit_price_ghs_minor"`
	ProductNameSnapshot  string `json:"product_name_snapshot"`
	ProductBrandSnapshot string `json:"product_brand_snapshot"`
	ProductImageSnapshot string `json:"product_image_snapshot"`
}

// updateOrderStatus updates the status of an order.
// @Summary Update order status
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param body body UpdateOrderStatusRequest true "New status"
// @Success 204
// @Failure 400 {object} httpx.ErrorEnvelope
// @Failure 401 {object} httpx.ErrorEnvelope
// @Failure 404 {object} httpx.ErrorEnvelope
// @Router /admin/orders/{id}/status [patch]
func (h *Handlers) updateOrderStatus(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid order ID", nil)
		return
	}

	var req UpdateOrderStatusRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid request body", nil)
		return
	}

	if req.Status == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "status is required", map[string]string{
			"status": "cannot be empty",
		})
		return
	}

	err = h.Svc.UpdateOrderStatus(r.Context(), UpdateOrderStatusParams{
		OrderID: orderID,
		Status:  req.Status,
	})
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "order not found", nil)
		} else {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, err.Error(), nil)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateOrderStatusRequest contains the new order status.
type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

// Customers

// listCustomers returns a paginated list of all customers.
// @Summary List all customers
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} CustomersResponse
// @Param page      query int false "Page number (1-based)"
// @Param page_size query int false "Items per page"
// @Router /admin/customers [get]
func (h *Handlers) listCustomers(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	params := ListCustomersParams{
		Page:     httpx.QueryInt(r, "page", 1),
		PageSize: httpx.QueryInt(r, "page_size", 20),
	}

	customers, err := h.Svc.ListCustomers(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list customers", nil)
		return
	}

	response := CustomersResponse{
		Customers:  make([]CustomerSummary, len(customers.Customers)),
		Total:      customers.Total,
		Page:       customers.Page,
		PageSize:   customers.PageSize,
		TotalPages: customers.TotalPages,
	}

	for i, customer := range customers.Customers {
		lastOrderAt := ""
		if customer.LastOrderAt != nil {
			if t, ok := customer.LastOrderAt.(time.Time); ok {
				lastOrderAt = t.Format("2006-01-02T15:04:05Z07:00")
			}
		}

		response.Customers[i] = CustomerSummary{
			ID:                    customer.ID.String(),
			Email:                 customer.Email,
			Name:                  customer.Name,
			EmailVerified:         customer.EmailVerified,
			CreatedAt:             customer.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			OrderCount:            customer.OrderCount,
			LifetimeValueGhsMinor: customer.LifetimeValueGhsMinor,
			LastOrderAt:           lastOrderAt,
		}
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// CustomersResponse contains paginated customer results.
type CustomersResponse struct {
	Customers  []CustomerSummary `json:"customers"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// CustomerSummary represents a customer in a list.
type CustomerSummary struct {
	ID                    string `json:"id"`
	Email                 string `json:"email"`
	Name                  string `json:"name,omitempty"`
	EmailVerified         bool   `json:"email_verified"`
	CreatedAt             string `json:"created_at"`
	OrderCount            int32  `json:"order_count"`
	LifetimeValueGhsMinor int64  `json:"lifetime_value_ghs_minor"`
	LastOrderAt           string `json:"last_order_at,omitempty"`
}

// getCustomerDetail returns detailed information about a single customer.
// @Summary Get customer details
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} CustomerDetailResponse
// @Failure 400 {object} httpx.ErrorEnvelope
// @Failure 401 {object} httpx.ErrorEnvelope
// @Failure 404 {object} httpx.ErrorEnvelope
// @Router /admin/customers/{id} [get]
func (h *Handlers) getCustomerDetail(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid customer ID", nil)
		return
	}

	detail, err := h.Svc.GetCustomerDetail(r.Context(), userID)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "customer not found", nil)
		} else {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get customer detail", nil)
		}
		return
	}

	response := CustomerDetailResponse{
		Customer: CustomerInfo{
			ID:            detail.Customer.ID.String(),
			Email:         detail.Customer.Email,
			Name:          detail.Customer.Name,
			EmailVerified: detail.Customer.EmailVerified,
			CreatedAt:     detail.Customer.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     detail.Customer.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		},
		Orders: make([]OrderSummary, len(detail.Orders)),
	}

	// Handle image field
	if detail.Customer.Image != nil {
		response.Customer.Image = *detail.Customer.Image
	}

	for i, order := range detail.Orders {
		createdAt := ""
		if order.CreatedAt.Valid {
			createdAt = order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		updatedAt := ""
		if order.UpdatedAt.Valid {
			updatedAt = order.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		response.Orders[i] = OrderSummary{
			ID:               order.ID.String(),
			UserID:           order.UserID.String(),
			Status:           order.Status,
			SubtotalGhsMinor: order.SubtotalGhsMinor,
			ShippingGhsMinor: order.ShippingGhsMinor,
			TotalGhsMinor:    order.TotalGhsMinor,
			PaystackRef:      order.PaystackReference,
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		}
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// CustomerDetailResponse contains detailed customer information.
type CustomerDetailResponse struct {
	Customer CustomerInfo   `json:"customer"`
	Orders   []OrderSummary `json:"orders"`
}

// CustomerInfo represents detailed customer information.
type CustomerInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name,omitempty"`
	EmailVerified bool   `json:"email_verified"`
	Image         string `json:"image,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Products (read-only)

// listProducts returns a paginated list of all products.
// @Summary List all products
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ProductsResponse
// @Param page      query int false "Page number (1-based)"
// @Param page_size query int false "Items per page"
// @Router /admin/products [get]
func (h *Handlers) listProducts(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	params := ListProductsParams{
		Page:     httpx.QueryInt(r, "page", 1),
		PageSize: httpx.QueryInt(r, "page_size", 20),
	}

	products, err := h.Svc.ListProducts(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list products", nil)
		return
	}

	response := ProductsResponse{
		Products:   make([]ProductSummary, len(products.Products)),
		Total:      products.Total,
		Page:       products.Page,
		PageSize:   products.PageSize,
		TotalPages: products.TotalPages,
	}

	for i, product := range products.Products {
		wasPrice := int64(0)
		if product.WasPriceGhsMinor != nil {
			wasPrice = *product.WasPriceGhsMinor
		}

		rating := 0.0
		if product.Rating.Valid {
			var f float64
			if err := product.Rating.Scan(&f); err == nil {
				rating = f
			}
		}

		response.Products[i] = ProductSummary{
			ID:               product.ID.String(),
			Slug:             product.Slug,
			Name:             product.Name,
			BrandID:          product.BrandID.String(),
			CategoryID:       product.CategoryID.String(),
			PriceGhsMinor:    product.PriceGhsMinor,
			WasPriceGhsMinor: wasPrice,
			Tone:             product.Tone,
			Size:             product.Size,
			Rating:           rating,
			ReviewCount:      product.ReviewCount,
			Tags:             product.Tags,
			ImagePath:        product.ImagePath,
			CreatedAt:        product.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// ProductsResponse contains paginated product results.
type ProductsResponse struct {
	Products   []ProductSummary `json:"products"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// ProductSummary represents a product in a list.
type ProductSummary struct {
	ID               string   `json:"id"`
	Slug             string   `json:"slug"`
	Name             string   `json:"name"`
	BrandID          string   `json:"brand_id"`
	CategoryID       string   `json:"category_id"`
	PriceGhsMinor    int64    `json:"price_ghs_minor"`
	WasPriceGhsMinor int64    `json:"was_price_ghs_minor,omitempty"`
	Tone             string   `json:"tone"`
	Size             string   `json:"size"`
	Rating           float64  `json:"rating,omitempty"`
	ReviewCount      int32    `json:"review_count"`
	Tags             []string `json:"tags"`
	ImagePath        string   `json:"image_path"`
	CreatedAt        string   `json:"created_at"`
}

// getProductDetail returns detailed information about a single product.
// @Summary Get product details
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} ProductSummary
// @Failure 400 {object} httpx.ErrorEnvelope
// @Failure 401 {object} httpx.ErrorEnvelope
// @Failure 404 {object} httpx.ErrorEnvelope
// @Router /admin/products/{id} [get]
func (h *Handlers) getProductDetail(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid product ID", nil)
		return
	}

	product, err := h.Svc.GetProductDetail(r.Context(), productID)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "product not found", nil)
		} else {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get product detail", nil)
		}
		return
	}

	wasPrice := int64(0)
	if product.WasPriceGhsMinor != nil {
		wasPrice = *product.WasPriceGhsMinor
	}

	rating := 0.0
	if product.Rating.Valid {
		var f float64
		if err := product.Rating.Scan(&f); err == nil {
			rating = f
		}
	}

	response := ProductSummary{
		ID:               product.ID.String(),
		Slug:             product.Slug,
		Name:             product.Name,
		BrandID:          product.BrandID.String(),
		CategoryID:       product.CategoryID.String(),
		PriceGhsMinor:    product.PriceGhsMinor,
		WasPriceGhsMinor: wasPrice,
		Tone:             product.Tone,
		Size:             product.Size,
		Rating:           rating,
		ReviewCount:      product.ReviewCount,
		Tags:             product.Tags,
		ImagePath:        product.ImagePath,
		CreatedAt:        product.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// Analytics

// getRevenueAnalytics returns revenue analytics data.
// @Summary Get revenue analytics
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} RevenueAnalyticsResponse
// @Param granularity query string false "day|week|month"
// @Param date_from   query string false "RFC3339 lower bound"
// @Param date_to     query string false "RFC3339 upper bound"
// @Router /admin/analytics/revenue [get]
func (h *Handlers) getRevenueAnalytics(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	// Parse query parameters
	granularity := r.URL.Query().Get("granularity")
	if granularity == "" {
		granularity = "month"
	}

	dateFromStr := r.URL.Query().Get("date_from")
	if dateFromStr == "" {
		dateFromStr = "2024-01-01T00:00:00Z"
	}
	dateToStr := r.URL.Query().Get("date_to")
	if dateToStr == "" {
		dateToStr = "2024-12-31T23:59:59Z"
	}

	dateFrom, err := httpx.ParseTime(dateFromStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid date_from format", nil)
		return
	}
	dateTo, err := httpx.ParseTime(dateToStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid date_to format", nil)
		return
	}

	params := GetRevenueAnalyticsParams{
		Granularity: granularity,
		DateFrom:    pgtype.Timestamptz{Time: dateFrom, Valid: true},
		DateTo:      pgtype.Timestamptz{Time: dateTo, Valid: true},
	}

	analytics, err := h.Svc.GetRevenueAnalytics(r.Context(), params)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get revenue analytics", nil)
		return
	}

	response := RevenueAnalyticsResponse{
		ByDate:     make([]RevenueByDate, len(analytics.ByDate)),
		ByCategory: make([]RevenueByCategory, len(analytics.ByCategory)),
		OrderStats: OrderStats{
			PendingCount:                  analytics.OrderStats.PendingCount,
			PaidCount:                     analytics.OrderStats.PaidCount,
			FulfilledCount:                analytics.OrderStats.FulfilledCount,
			ShippedCount:                  analytics.OrderStats.ShippedCount,
			DeliveredCount:                analytics.OrderStats.DeliveredCount,
			CancelledCount:                analytics.OrderStats.CancelledCount,
			RefundedCount:                 analytics.OrderStats.RefundedCount,
			PaidRevenueGhsMinor:           analytics.OrderStats.PaidRevenueGhsMinor,
			DeliveredRevenueGhsMinor:      analytics.OrderStats.DeliveredRevenueGhsMinor,
			TotalCompletedRevenueGhsMinor: analytics.OrderStats.TotalCompletedRevenueGhsMinor,
		},
	}

	for i, item := range analytics.ByDate {
		response.ByDate[i] = RevenueByDate{
			Date:            item.RevenueDate.Time.Format("2006-01-02"),
			OrderCount:      item.OrderCount,
			RevenueGhsMinor: item.RevenueGhsMinor,
		}
	}

	for i, item := range analytics.ByCategory {
		response.ByCategory[i] = RevenueByCategory{
			CategoryID:      item.ID.String(),
			CategorySlug:    item.Slug,
			CategoryName:    item.CategoryName,
			OrderCount:      item.OrderCount,
			RevenueGhsMinor: item.RevenueGhsMinor,
		}
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// RevenueAnalyticsResponse contains revenue analytics data.
type RevenueAnalyticsResponse struct {
	ByDate     []RevenueByDate     `json:"by_date"`
	ByCategory []RevenueByCategory `json:"by_category"`
	OrderStats OrderStats          `json:"order_stats"`
}

// RevenueByDate represents revenue for a specific date.
type RevenueByDate struct {
	Date            string `json:"date"`
	OrderCount      int32  `json:"order_count"`
	RevenueGhsMinor int64  `json:"revenue_ghs_minor"`
}

// RevenueByCategory represents revenue for a specific category.
type RevenueByCategory struct {
	CategoryID      string `json:"category_id"`
	CategorySlug    string `json:"category_slug"`
	CategoryName    string `json:"category_name"`
	OrderCount      int32  `json:"order_count"`
	RevenueGhsMinor int64  `json:"revenue_ghs_minor"`
}

// OrderStats contains order statistics.
type OrderStats struct {
	PendingCount                  int32 `json:"pending_count"`
	PaidCount                     int32 `json:"paid_count"`
	FulfilledCount                int32 `json:"fulfilled_count"`
	ShippedCount                  int32 `json:"shipped_count"`
	DeliveredCount                int32 `json:"delivered_count"`
	CancelledCount                int32 `json:"cancelled_count"`
	RefundedCount                 int32 `json:"refunded_count"`
	PaidRevenueGhsMinor           int64 `json:"paid_revenue_ghs_minor"`
	DeliveredRevenueGhsMinor      int64 `json:"delivered_revenue_ghs_minor"`
	TotalCompletedRevenueGhsMinor int64 `json:"total_completed_revenue_ghs_minor"`
}

// getStats returns general statistics for the admin dashboard.
// @Summary Get admin statistics
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} StatsResponse
// @Router /admin/analytics/stats [get]
func (h *Handlers) getStats(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}

	productStats, err := h.Svc.GetProductStats(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get product stats", nil)
		return
	}

	customerStats, err := h.Svc.GetCustomerStats(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get customer stats", nil)
		return
	}

	topProducts, err := h.Svc.GetTopProducts(r.Context(), 10)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get top products", nil)
		return
	}

	response := StatsResponse{
		ProductStats: ProductStats{
			TotalProducts: productStats.TotalProducts,
			OutOfStock:    productStats.OutOfStockCount,
			LowStock:      productStats.LowStockCount,
			LavenderCount: productStats.LavenderCount,
			RoseCount:     productStats.RoseCount,
			CreamCount:    productStats.CreamCount,
			InkCount:      productStats.InkCount,
		},
		CustomerStats: CustomerStats{
			TotalCustomers:      customerStats.TotalCustomers,
			ActiveCustomers30d:  customerStats.ActiveCustomers30d,
			CustomersWithOrders: customerStats.CustomersWithOrders,
		},
		TopProducts: make([]TopProduct, len(topProducts)),
	}

	for i, product := range topProducts {
		brandName := ""
		if product.BrandName != nil {
			brandName = *product.BrandName
		}
		categoryLabel := ""
		if product.CategoryLabel != nil {
			categoryLabel = *product.CategoryLabel
		}

		response.TopProducts[i] = TopProduct{
			ID:              product.ID.String(),
			Slug:            product.Slug,
			Name:            product.Name,
			BrandID:         product.BrandID.String(),
			CategoryID:      product.CategoryID.String(),
			PriceGhsMinor:   product.PriceGhsMinor,
			Tone:            product.Tone,
			ImagePath:       product.ImagePath,
			BrandName:       brandName,
			CategoryLabel:   categoryLabel,
			TotalSold:       product.TotalSold,
			RevenueGhsMinor: product.TotalRevenueGhsMinor,
		}
	}

	httpx.WriteJSON(w, http.StatusOK, response)
}

// StatsResponse contains various statistics for the admin dashboard.
type StatsResponse struct {
	ProductStats  ProductStats  `json:"product_stats"`
	CustomerStats CustomerStats `json:"customer_stats"`
	TopProducts   []TopProduct  `json:"top_products"`
}

// ProductStats contains product statistics.
type ProductStats struct {
	TotalProducts int32 `json:"total_products"`
	OutOfStock    int32 `json:"out_of_stock"`
	LowStock      int32 `json:"low_stock"`
	LavenderCount int32 `json:"lavender_count"`
	RoseCount     int32 `json:"rose_count"`
	CreamCount    int32 `json:"cream_count"`
	InkCount      int32 `json:"ink_count"`
}

// CustomerStats contains customer statistics.
type CustomerStats struct {
	TotalCustomers      int32 `json:"total_customers"`
	ActiveCustomers30d  int32 `json:"active_customers_30d"`
	CustomersWithOrders int32 `json:"customers_with_orders"`
}

// TopProduct represents a top-selling product.
type TopProduct struct {
	ID              string `json:"id"`
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	BrandID         string `json:"brand_id"`
	CategoryID      string `json:"category_id"`
	PriceGhsMinor   int64  `json:"price_ghs_minor"`
	Tone            string `json:"tone"`
	ImagePath       string `json:"image_path"`
	BrandName       string `json:"brand_name"`
	CategoryLabel   string `json:"category_label"`
	TotalSold       int32  `json:"total_sold"`
	RevenueGhsMinor int64  `json:"revenue_ghs_minor"`
}

// Placeholder endpoints for v1

func (h *Handlers) getMarketingPlaceholder(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Marketing features coming soon",
	})
}

func (h *Handlers) getContentPlaceholder(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Content management features coming soon",
	})
}

func (h *Handlers) getSettingsPlaceholder(w http.ResponseWriter, r *http.Request) {
	if !auth.MustBeAdmin(w, r) {
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Settings features coming soon",
	})
}
