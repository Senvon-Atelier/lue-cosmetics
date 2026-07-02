package shipping

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"go.uber.org/zap"
)

// Handlers holds the HTTP handlers for the shipping domain.
type Handlers struct {
	svc *Service
	log *zap.Logger
}

// NewHandlers creates a new Handlers wired to the given Service.
func NewHandlers(svc *Service, log *zap.Logger) *Handlers {
	return &Handlers{svc: svc, log: log}
}

// Mount registers the shipping routes on the given router.
func (h *Handlers) Mount(r chi.Router) {
	r.Get("/shipping/quote", h.quote)
}

// quote godoc
//
//	@Summary	Get shipping quote
//	@Tags		shipping
//	@Produce	json
//	@Param		subtotal	query		int				true	"Cart subtotal in pesewas (>= 0)"
//	@Success	200			{object}	shipping.Quote
//	@Failure	400			{object}	httpx.ErrorEnvelope
//	@Router		/shipping/quote [get]
func (h *Handlers) quote(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query().Get("subtotal")
	if v == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "subtotal required",
			map[string]string{"subtotal": "missing"})
		return
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n < 0 {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid subtotal",
			map[string]string{"subtotal": "must be a non-negative integer (pesewas)"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, h.svc.Quote(n))
}
