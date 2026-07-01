package me

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/orders"
	"go.uber.org/zap"
)

type Handlers struct {
	OrdersRepo     *orders.Repository
	ProfileService *ProfileService
	Log            *zap.Logger
}

func NewHandlers(ordersRepo *orders.Repository, profileService *ProfileService, log *zap.Logger) *Handlers {
	return &Handlers{
		OrdersRepo:     ordersRepo,
		ProfileService: profileService,
		Log:            log,
	}
}

func (h *Handlers) Mount(r chi.Router) {
	r.Get("/me", h.get)
	r.Get("/me/orders", h.listOrders)
	r.Get("/me/orders/{id}", h.getOrder)
	r.Patch("/me", h.updateProfile)
}

func (h *Handlers) MountRoutes(r chi.Router) {
	r.Get("/", h.get)
	r.Get("/orders", h.listOrders)
	r.Get("/orders/{id}", h.getOrder)
	r.Patch("/", h.updateProfile)
}

type meResponse struct {
	UserID        string  `json:"user_id"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	Image         *string `json:"image,omitempty"`
	Role          string  `json:"role"`
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
	ID                string  `json:"id"`
	UserID            string  `json:"user_id"`
	Status            string  `json:"status"`
	Subtotal          float64 `json:"subtotal_ghs"`
	Shipping          float64 `json:"shipping_ghs"`
	Total             float64 `json:"total_ghs"`
	PaystackReference string  `json:"paystack_reference"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type orderDetailResponse struct {
	orderResponse
	ShippingAddress string              `json:"shipping_address"`
	Items           []orderItemResponse `json:"items"`
}

type orderItemResponse struct {
	ID           string  `json:"id"`
	ProductID    string  `json:"product_id"`
	Qty          int32   `json:"qty"`
	UnitPrice    float64 `json:"unit_price_ghs"`
	ProductName  string  `json:"product_name_snapshot"`
	ProductBrand string  `json:"product_brand_snapshot"`
	ProductImage string  `json:"product_image_snapshot"`
}

type listOrdersResponse struct {
	Orders []orderResponse `json:"orders"`
	Total  int64           `json:"total"`
	Limit  int32           `json:"limit"`
	Offset int32           `json:"offset"`
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
// @Router   /api/v1/me/orders [get]
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
// @Router   /api/v1/me/orders/{id} [get]
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
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me [patch]
func (h *Handlers) updateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	view, ok := auth.GetSessionView(ctx)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	var req updateProfileRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid request body", nil)
		return
	}

	// Email updates still require verification flow (future work)
	if req.Email != nil && *req.Email != view.Email {
		httpx.WriteError(w, http.StatusNotImplemented, httpx.CodeInternal, "email updates require verification flow (not yet implemented)", nil)
		return
	}

	// Image updates not supported in v1
	if req.Image != nil {
		httpx.WriteError(w, http.StatusNotImplemented, httpx.CodeInternal, "image updates not yet implemented", nil)
		return
	}

	// Delegate to ProfileService
	result, err := h.ProfileService.UpdateProfile(ctx, UpdateProfileParams{
		UserID: view.UserID,
		Name:   req.Name,
		Image:  req.Image,
	})
	if err != nil {
		// Map service errors to HTTP responses
		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid name") || strings.Contains(errMsg, "invalid image") {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, errMsg, nil)
		} else {
			h.Log.Error("failed to update user profile", zap.Error(err))
			httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to update profile", nil)
		}
		return
	}

	// Return updated profile
	httpx.WriteJSON(w, http.StatusOK, meResponse{
		UserID:        result.UserID.String(),
		Email:         result.Email,
		Name:          result.Name,
		Image:         result.Image,
		Role:          result.Role,
		EmailVerified: result.EmailVerified,
	})
}

type updateProfileRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Image *string `json:"image,omitempty"`
}
