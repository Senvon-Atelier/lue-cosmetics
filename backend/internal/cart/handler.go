package cart

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/logging"
)

// Handlers binds the cart service to HTTP. AuthSvc is used to resolve session
// cookies into a CartIdentity; CookieName is the session cookie's name (e.g.
// "rue_session") so the cart handler can probe it directly without depending
// on auth middleware (these endpoints are public).
type Handlers struct {
	Svc          *Service
	AuthSvc      *auth.Service
	CookieName   string
	CookieDomain string
	Secure       bool
}

// NewHandlers constructs a Handlers. sessionCookieName defaults to
// "rue_session" if empty.
func NewHandlers(svc *Service, authSvc *auth.Service, sessionCookieName, cookieDomain string, secure bool) *Handlers {
	if sessionCookieName == "" {
		sessionCookieName = "rue_session"
	}
	return &Handlers{
		Svc:          svc,
		AuthSvc:      authSvc,
		CookieName:   sessionCookieName,
		CookieDomain: cookieDomain,
		Secure:       secure,
	}
}

// Mount registers the public cart endpoints. /cart/merge is intentionally NOT
// here — it lives behind RequireSession via MountAuthGated (Bundle 3).
func (h *Handlers) Mount(r chi.Router) {
	r.Get("/cart", h.getCart)
	r.Post("/cart/items", h.postItem)
	r.Patch("/cart/items/{id}", h.patchItem)
	r.Delete("/cart/items/{id}", h.deleteItem)
}

// MountAuthGated is a placeholder for Bundle 3 (POST /cart/merge). Left empty
// here so cmd/api can call it unconditionally once the wiring lands.
func (h *Handlers) MountAuthGated(r chi.Router) {
	// Bundle 3 will register POST /cart/merge here.
}

// resolveIdentity picks (in order): a valid session cookie → guest cookie →
// empty. Session always wins; an authenticated user with a stale guest cookie
// keeps using their user cart (merge is opt-in via POST /cart/merge).
func (h *Handlers) resolveIdentity(r *http.Request) CartIdentity {
	if h.AuthSvc != nil {
		if c, err := r.Cookie(h.CookieName); err == nil {
			if view, err := h.AuthSvc.GetSession(r.Context(), c.Value); err == nil {
				return CartIdentity{UserID: view.UserID}
			}
		}
	}
	if c, err := r.Cookie(GuestCookieName); err == nil && c.Value != "" {
		return CartIdentity{GuestToken: c.Value}
	}
	return CartIdentity{}
}

// ── JSON shapes ─────────────────────────────────────────────────────────────

type cartItemResponse struct {
	ID                string `json:"id"`
	ProductID         string `json:"product_id"`
	ProductSlug       string `json:"product_slug"`
	ProductName       string `json:"product_name"`
	ProductImagePath  string `json:"product_image_path"`
	Qty               int32  `json:"qty"`
	UnitPriceGhsMinor int64  `json:"unit_price_ghs_minor"`
	LineTotalGhsMinor int64  `json:"line_total_ghs_minor"`
}

type cartResponse struct {
	CartID                        string             `json:"cart_id"`
	GuestToken                    string             `json:"guest_token,omitempty"`
	Items                         []cartItemResponse `json:"items"`
	SubtotalGhsMinor              int64              `json:"subtotal_ghs_minor"`
	ShippingCostGhsMinor          int64              `json:"shipping_cost_ghs_minor"`
	FreeShippingRemainderGhsMinor int64              `json:"free_shipping_remainder_ghs_minor"`
	TotalGhsMinor                 int64              `json:"total_ghs_minor"`
}

func viewToResponse(v View) cartResponse {
	items := make([]cartItemResponse, 0, len(v.Items))
	for _, it := range v.Items {
		items = append(items, cartItemResponse{
			ID:                it.ID.String(),
			ProductID:         it.ProductID.String(),
			ProductSlug:       it.ProductSlug,
			ProductName:       it.ProductName,
			ProductImagePath:  it.ProductImagePath,
			Qty:               it.Qty,
			UnitPriceGhsMinor: it.UnitPriceGhsMinor,
			LineTotalGhsMinor: it.LineTotalGhsMinor,
		})
	}
	return cartResponse{
		CartID:                        v.CartID.String(),
		GuestToken:                    v.GuestToken,
		Items:                         items,
		SubtotalGhsMinor:              v.SubtotalGhsMinor,
		ShippingCostGhsMinor:          v.ShippingCostGhsMinor,
		FreeShippingRemainderGhsMinor: v.FreeShippingRemainderGhsMinor,
		TotalGhsMinor:                 v.TotalGhsMinor,
	}
}

// ── endpoints ───────────────────────────────────────────────────────────────

// getCart godoc
//
// @Summary  Get the caller's cart (mints a guest cart on first call)
// @Tags     cart
// @Produce  json
// @Success  200 {object} cartResponse
// @Router   /cart [get]
func (h *Handlers) getCart(w http.ResponseWriter, r *http.Request) {
	id := h.resolveIdentity(r)
	view, minted, err := h.Svc.GetOrCreate(r.Context(), id)
	if err != nil {
		logging.From(r.Context(), h.Svc.Log).Error("cart: get", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "cart unavailable", nil)
		return
	}
	if minted != "" {
		SetGuestCookie(w, minted, h.CookieDomain, h.Secure)
	}
	httpx.WriteJSON(w, http.StatusOK, viewToResponse(view))
}

type postItemBody struct {
	ProductID string `json:"product_id"`
	Qty       int32  `json:"qty"`
}

// postItem godoc
//
// @Summary  Add an item to the cart (upserts: qty is summed on conflict)
// @Tags     cart
// @Accept   json
// @Produce  json
// @Param    body body postItemBody true "Item payload"
// @Success  200 {object} cartResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Router   /cart/items [post]
func (h *Handlers) postItem(w http.ResponseWriter, r *http.Request) {
	var body postItemBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil)
		return
	}
	if body.Qty < 1 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "qty must be >= 1", map[string]string{"qty": "must be >= 1"})
		return
	}
	productID, err := uuid.Parse(body.ProductID)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid product_id", map[string]string{"product_id": "must be a uuid"})
		return
	}
	id := h.resolveIdentity(r)
	// Mint the cart up front so we can set the cookie regardless of any
	// AddItem outcome that would otherwise leave the guest without a token.
	_, minted, err := h.Svc.GetOrCreate(r.Context(), id)
	if err != nil {
		logging.From(r.Context(), h.Svc.Log).Error("cart: get-or-create", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "cart unavailable", nil)
		return
	}
	if minted != "" {
		id = CartIdentity{GuestToken: minted}
		SetGuestCookie(w, minted, h.CookieDomain, h.Secure)
	}
	view, err := h.Svc.AddItem(r.Context(), id, productID, body.Qty)
	switch {
	case errors.Is(err, ErrUnknownProduct):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "unknown product", map[string]string{"product_id": "not found"})
		return
	case errors.Is(err, ErrInvalidQty):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "qty must be >= 1", map[string]string{"qty": "must be >= 1"})
		return
	case err != nil:
		logging.From(r.Context(), h.Svc.Log).Error("cart: add item", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "add failed", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, viewToResponse(view))
}

type patchItemBody struct {
	Qty int32 `json:"qty"`
}

// patchItem godoc
//
// @Summary  Update an item's qty
// @Tags     cart
// @Accept   json
// @Produce  json
// @Param    id   path string        true "Cart item id"
// @Param    body body patchItemBody true "Qty payload"
// @Success  200 {object} cartResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  404 {object} httpx.ErrorEnvelope
// @Router   /cart/items/{id} [patch]
func (h *Handlers) patchItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "item not found", nil)
		return
	}
	var body patchItemBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil)
		return
	}
	if body.Qty < 1 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "qty must be >= 1", map[string]string{"qty": "must be >= 1"})
		return
	}
	id := h.resolveIdentity(r)
	view, err := h.Svc.UpdateQty(r.Context(), id, itemID, body.Qty)
	switch {
	case errors.Is(err, ErrItemNotFound):
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "item not found", nil)
		return
	case errors.Is(err, ErrInvalidQty):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "qty must be >= 1", map[string]string{"qty": "must be >= 1"})
		return
	case err != nil:
		logging.From(r.Context(), h.Svc.Log).Error("cart: update qty", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "update failed", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, viewToResponse(view))
}

// deleteItem godoc
//
// @Summary  Remove an item from the cart
// @Tags     cart
// @Produce  json
// @Param    id path string true "Cart item id"
// @Success  204
// @Failure  404 {object} httpx.ErrorEnvelope
// @Router   /cart/items/{id} [delete]
func (h *Handlers) deleteItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "item not found", nil)
		return
	}
	id := h.resolveIdentity(r)
	if _, err := h.Svc.RemoveItem(r.Context(), id, itemID); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "item not found", nil)
			return
		}
		logging.From(r.Context(), h.Svc.Log).Error("cart: remove item", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "remove failed", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
