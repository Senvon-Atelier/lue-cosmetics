package orders

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/logging"
	"github.com/oti-adjei/ruecosmetics/internal/payments/paystack"
)

type Handlers struct {
	Svc            *Service
	PaystackSecret string
	Log            *zap.Logger
}

func NewHandlers(svc *Service, paystackSecret string, log *zap.Logger) *Handlers {
	return &Handlers{Svc: svc, PaystackSecret: paystackSecret, Log: log}
}

// MountAuthGated registers the session-protected checkout endpoints. The
// caller is expected to wrap the chi.Router with auth.RequireSession.
func (h *Handlers) MountAuthGated(r chi.Router) {
	r.Post("/checkout/init", h.initCheckout)
	r.Get("/checkout/verify/{reference}", h.verifyCheckout)
}

// MountPublic registers the public Paystack webhook. Must NOT live inside the
// RequireSession group — Paystack does not present a session cookie.
func (h *Handlers) MountPublic(r chi.Router) {
	r.Post("/webhooks/paystack", h.paystackWebhook)
}

// ── JSON shapes ─────────────────────────────────────────────────────────────

type initCheckoutBody struct {
	ShippingAddress ShippingAddress `json:"shipping_address"`
	ShippingMethod  string          `json:"shipping_method"`
}

type initCheckoutResponse struct {
	OrderID          string `json:"order_id"`
	Reference        string `json:"reference"`
	AuthorizationURL string `json:"authorization_url"`
	TotalGhsMinor    int64  `json:"total_ghs_minor"`
}

type verifyCheckoutResponse struct {
	Status string `json:"status"`
}

// ── handlers ────────────────────────────────────────────────────────────────

// initCheckout godoc
//
// @Summary  Initialize a Paystack-hosted checkout for the caller's cart
// @Tags     checkout
// @Accept   json
// @Produce  json
// @Param    body body initCheckoutBody true "Shipping address + method"
// @Success  200 {object} initCheckoutResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  503 {object} httpx.ErrorEnvelope
// @Router   /checkout/init [post]
func (h *Handlers) initCheckout(w http.ResponseWriter, r *http.Request) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}
	var body initCheckoutBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil)
		return
	}
	out, err := h.Svc.InitCheckout(r.Context(), InitCheckoutInput{
		UserID:          view.UserID,
		UserEmail:       view.Email,
		UserName:        view.Name,
		ShippingAddress: body.ShippingAddress,
		ShippingMethod:  body.ShippingMethod,
	})
	switch {
	case errors.Is(err, ErrInvalidAddress):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid shipping address",
			map[string]string{"shipping_address": "line1, city, region, phone are required"})
		return
	case errors.Is(err, ErrEmptyCart):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "cart is empty",
			map[string]string{"cart": "no items"})
		return
	case errors.Is(err, ErrPaystackNotReady):
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeUpstream, "paystack not configured", nil)
		return
	case errors.Is(err, paystack.ErrPaystackUpstream):
		logging.From(r.Context(), h.Log).Error("checkout: paystack init upstream error", zap.Error(err))
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeUpstream, "payment provider unavailable", nil)
		return
	case err != nil:
		logging.From(r.Context(), h.Log).Error("checkout: init", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "checkout init failed", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, initCheckoutResponse{
		OrderID:          out.OrderID.String(),
		Reference:        out.Reference,
		AuthorizationURL: out.AuthorizationURL,
		TotalGhsMinor:    out.TotalGhsMinor,
	})
}

// verifyCheckout godoc
//
// @Summary  Verify a checkout (poll path) — calls Paystack verify and converges with the webhook
// @Tags     checkout
// @Produce  json
// @Param    reference path string true "Paystack reference, e.g. RUE-XXXXXXXX"
// @Success  200 {object} verifyCheckoutResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  404 {object} httpx.ErrorEnvelope
// @Failure  503 {object} httpx.ErrorEnvelope
// @Router   /checkout/verify/{reference} [get]
func (h *Handlers) verifyCheckout(w http.ResponseWriter, r *http.Request) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}
	reference := chi.URLParam(r, "reference")
	// Owner check BEFORE asking Paystack — IDOR guard. 404 (not 403) to avoid
	// leaking existence to non-owners.
	order, err := h.Svc.Repo.GetOrderByReference(r.Context(), reference)
	if errors.Is(err, ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "order not found", nil)
		return
	}
	if err != nil {
		logging.From(r.Context(), h.Log).Error("checkout: get order", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "verify failed", nil)
		return
	}
	if order.UserID != view.UserID {
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "order not found", nil)
		return
	}

	status, err := h.Svc.VerifyCheckout(r.Context(), reference)
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "order not found", nil)
		return
	case errors.Is(err, ErrPaystackNotReady):
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeUpstream, "paystack not configured", nil)
		return
	case errors.Is(err, paystack.ErrPaystackUpstream):
		logging.From(r.Context(), h.Log).Error("checkout: verify upstream", zap.Error(err))
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeUpstream, "payment provider unavailable", nil)
		return
	case err != nil:
		logging.From(r.Context(), h.Log).Error("checkout: verify", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "verify failed", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, verifyCheckoutResponse{Status: status})
}
