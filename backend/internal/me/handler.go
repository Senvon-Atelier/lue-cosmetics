package me

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/orders"
)

type Handlers struct {
	OrdersRepo *orders.Repository
}

func NewHandlers(ordersRepo *orders.Repository) *Handlers {
	return &Handlers{OrdersRepo: ordersRepo}
}

func (h *Handlers) Mount(r chi.Router) {
	r.Get("/me", h.get)
	r.Get("/me/orders", h.listOrders)
	r.Get("/me/orders/{id}", h.getOrder)
	r.Patch("/me", h.updateProfile)
}

func (h *Handlers) MountRoutes(r chi.Router) {
	r.Get("/", h.get)
}

type meResponse struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

// get godoc
//
// @Summary  Get the current user
// @Tags     me
// @Produce  json
// @Success  200 {object} meResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Router   /me [get]
func (h *Handlers) get(w http.ResponseWriter, r *http.Request) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, meResponse{
		UserID: view.UserID.String(), Email: view.Email, Name: view.Name,
		Role: view.Role, EmailVerified: view.EmailVerified,
	})
}

// Order response types
type orderResponse struct {
	ID                 string  `json:"id"`
	UserID             string  `json:"user_id"`
	Status             string  `json:"status"`
	Subtotal           float64 `json:"subtotal_ghs"`
	Shipping           float64 `json:"shipping_ghs"`
	Total              float64 `json:"total_ghs"`
	PaystackReference  string  `json:"paystack_reference"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type orderDetailResponse struct {
	orderResponse
	ShippingAddress string           `json:"shipping_address"`
	Items           []orderItemResponse `json:"items"`
}

type orderItemResponse struct {
	ID               string  `json:"id"`
	ProductID        string  `json:"product_id"`
	Qty              int32   `json:"qty"`
	UnitPrice        float64 `json:"unit_price_ghs"`
	ProductName      string  `json:"product_name_snapshot"`
	ProductBrand     string  `json:"product_brand_snapshot"`
	ProductImage     string  `json:"product_image_snapshot"`
}

type listOrdersResponse struct {
	Orders  []orderResponse `json:"orders"`
	Total   int64            `json:"total"`
	Limit   int32            `json:"limit"`
	Offset  int32            `json:"offset"`
}

// listOrders godoc
//
// @Summary  List user's orders
// @Tags     me
// @Produce  json
// @Param    status query string false "Filter by status"
// @Param    limit query int false "Pagination limit" default(20)
// @Param    offset query int false "Pagination offset" default(0)
// @Success  200 {object} listOrdersResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me/orders [get]
func (h *Handlers) listOrders(w http.ResponseWriter, r *http.Request) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	// Check if orders repository is available (for testing)
	if h.OrdersRepo == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeInternal, "orders service not available", nil)
		return
	}

	// Parse query params
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(20) // default
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	offset := int32(0) // default
	if offsetStr != "" {
		if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	// Get orders and total count
	ordersList, err := h.OrdersRepo.ListOrdersByUserID(r.Context(), view.UserID, status, limit, offset)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list orders", nil)
		return
	}

	total, err := h.OrdersRepo.CountOrdersByUserID(r.Context(), view.UserID, status)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to count orders", nil)
		return
	}

	// Convert to response format
	resp := listOrdersResponse{
		Orders: make([]orderResponse, len(ordersList)),
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	for i, order := range ordersList {
		resp.Orders[i] = orderResponse{
			ID:                order.ID.String(),
			UserID:            order.UserID.String(),
			Status:            order.Status,
			Subtotal:          float64(order.SubtotalGhsMinor) / 100,
			Shipping:          float64(order.ShippingGhsMinor) / 100,
			Total:             float64(order.TotalGhsMinor) / 100,
			PaystackReference: order.PaystackReference,
			CreatedAt:         order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:         order.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

// getOrder godoc
//
// @Summary  Get a specific order
// @Tags     me
// @Produce  json
// @Param    id path string true "Order ID"
// @Success  200 {object} orderDetailResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  404 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me/orders/{id} [get]
func (h *Handlers) getOrder(w http.ResponseWriter, r *http.Request) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	// Check if orders repository is available (for testing)
	if h.OrdersRepo == nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeInternal, "orders service not available", nil)
		return
	}

	// Parse order ID from URL
	orderIDStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid order id", nil)
		return
	}

	// Get order
	order, err := h.OrdersRepo.GetOrderByID(r.Context(), orderID)
	if err != nil {
		if err == orders.ErrNotFound {
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "order not found", nil)
		} else {
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get order", nil)
		}
		return
	}

	// Check ownership - users can only view their own orders
	if order.UserID != view.UserID {
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "order not found", nil)
		return
	}

	// Get order items
	items, err := h.OrdersRepo.ListOrderItems(r.Context(), orderID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get order items", nil)
		return
	}

	// Convert to response format
	resp := orderDetailResponse{
		orderResponse: orderResponse{
			ID:                order.ID.String(),
			UserID:            order.UserID.String(),
			Status:            order.Status,
			Subtotal:          float64(order.SubtotalGhsMinor) / 100,
			Shipping:          float64(order.ShippingGhsMinor) / 100,
			Total:             float64(order.TotalGhsMinor) / 100,
			PaystackReference: order.PaystackReference,
			CreatedAt:         order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:         order.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		},
		ShippingAddress: string(order.ShippingAddress),
		Items:           make([]orderItemResponse, len(items)),
	}

	for i, item := range items {
		resp.Items[i] = orderItemResponse{
			ID:           item.ID.String(),
			ProductID:    item.ProductID.String(),
			Qty:          item.Qty,
			UnitPrice:    float64(item.UnitPriceGhsMinor) / 100,
			ProductName:  item.ProductNameSnapshot,
			ProductBrand: item.ProductBrandSnapshot,
			ProductImage: item.ProductImageSnapshot,
		}
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

// updateProfile godoc
//
// @Summary  Update user profile
// @Tags     me
// @Accept   json
// @Produce  json
// @Param    request body updateProfileRequest true "Profile update data"
// @Success  200 {object} meResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  501 {object} httpx.ErrorEnvelope
// @Router   /me [patch]
func (h *Handlers) updateProfile(w http.ResponseWriter, r *http.Request) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	var req updateProfileRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid request body", nil)
		return
	}

	// For now, email updates are not supported (require verification flow)
	if req.Email != nil && *req.Email != view.Email {
		httpx.WriteError(w, http.StatusNotImplemented, httpx.CodeInternal, "email updates require verification flow (not yet implemented)", nil)
		return
	}

	// Name updates require auth repository which isn't injected here yet
	// TODO: Inject auth repository and implement name updates
	httpx.WriteError(w, http.StatusNotImplemented, httpx.CodeInternal, "profile updates not yet implemented (requires auth repository injection)", nil)
}

type updateProfileRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
}
