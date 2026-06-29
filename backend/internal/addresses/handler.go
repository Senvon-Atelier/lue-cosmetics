package addresses

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

type Handlers struct {
	Svc *Service
	Log *zap.Logger
}

func NewHandlers(svc *Service, log *zap.Logger) *Handlers {
	return &Handlers{
		Svc: svc,
		Log: log,
	}
}

func (h *Handlers) Mount(r chi.Router) {
	r.Route("/me/addresses", func(r chi.Router) {
		r.Post("/", h.create)
		r.Get("/", h.list)
		r.Patch("/{id}", h.update)
		r.Delete("/{id}", h.delete)
		r.Post("/{id}/default", h.setDefault)
	})
}

// createAddressRequest represents the JSON request body for creating an address.
type createAddressRequest struct {
	Label  string `json:"label"`
	Line1  string `json:"line1"`
	Line2  string `json:"line2"`
	City   string `json:"city"`
	Region string `json:"region"`
	Phone  string `json:"phone"`
}

// addressResponse represents a full address row returned to the client.
type addressResponse struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Line1     string `json:"line1"`
	Line2     string `json:"line2"`
	City      string `json:"city"`
	Region    string `json:"region"`
	Phone     string `json:"phone"`
	IsDefault bool   `json:"is_default"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// listAddressesResponse represents the response for listing addresses.
type listAddressesResponse struct {
	Addresses []addressResponse `json:"addresses"`
}

// create godoc
//
// @Summary  Create an address
// @Tags     addresses
// @Accept   json
// @Produce  json
// @Param    request body createAddressRequest true "Address data"
// @Success  201 {object} addressResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me/addresses [post]
func (h *Handlers) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	var req createAddressRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid request body", nil)
		return
	}

	addr, err := h.Svc.Create(r.Context(), userID, AddressInput{
		Label:  req.Label,
		Line1:  req.Line1,
		Line2:  req.Line2,
		City:   req.City,
		Region: req.Region,
		Phone:  req.Phone,
	})
	if err != nil {
		h.mapCreateError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toAddressResponse(addr))
}

// list godoc
//
// @Summary  List addresses
// @Tags     addresses
// @Produce  json
// @Success  200 {object} listAddressesResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me/addresses [get]
func (h *Handlers) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	addrs, err := h.Svc.List(r.Context(), userID)
	if err != nil {
		h.Log.Error("list addresses", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list addresses", nil)
		return
	}

	resp := listAddressesResponse{
		Addresses: make([]addressResponse, len(addrs)),
	}
	for i, addr := range addrs {
		resp.Addresses[i] = toAddressResponse(addr)
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}

// patchAddressRequest represents the JSON request body for patching an address.
// All fields are optional; nil means leave unchanged.
type patchAddressRequest struct {
	Label  *string `json:"label,omitempty"`
	Line1  *string `json:"line1,omitempty"`
	Line2  *string `json:"line2,omitempty"`
	City   *string `json:"city,omitempty"`
	Region *string `json:"region,omitempty"`
	Phone  *string `json:"phone,omitempty"`
}

// update godoc
//
// @Summary  Update an address
// @Tags     addresses
// @Accept   json
// @Produce  json
// @Param    id path string true "Address ID"
// @Param    request body patchAddressRequest true "Patch data"
// @Success  200 {object} addressResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  404 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me/addresses/{id} [patch]
func (h *Handlers) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	addrID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid address id", nil)
		return
	}

	var req patchAddressRequest
	if err := httpx.ReadJSON(r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid request body", nil)
		return
	}

	addr, err := h.Svc.Update(r.Context(), userID, addrID, AddressPatch{
		Label:  req.Label,
		Line1:  req.Line1,
		Line2:  req.Line2,
		City:   req.City,
		Region: req.Region,
		Phone:  req.Phone,
	})
	if err != nil {
		h.mapUpdateError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAddressResponse(addr))
}

// delete godoc
//
// @Summary  Delete an address
// @Tags     addresses
// @Produce  json
// @Param    id path string true "Address ID"
// @Success  204
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  404 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me/addresses/{id} [delete]
func (h *Handlers) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	addrID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid address id", nil)
		return
	}

	if err := h.Svc.Delete(r.Context(), userID, addrID); err != nil {
		h.mapDeleteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// setDefault godoc
//
// @Summary  Set an address as default
// @Tags     addresses
// @Produce  json
// @Param    id path string true "Address ID"
// @Success  200 {object} addressResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  404 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /me/addresses/{id}/default [post]
func (h *Handlers) setDefault(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
		return
	}

	addrID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid address id", nil)
		return
	}

	addr, err := h.Svc.SetDefault(r.Context(), userID, addrID)
	if err != nil {
		h.mapSetDefaultError(w, err)
		return
	}

	httpx.WriteJSON(w, http.StatusOK, toAddressResponse(addr))
}

// Error mapping helpers

func (h *Handlers) mapCreateError(w http.ResponseWriter, err error) {
	switch {
	case err == ErrInvalidAddress:
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid address data", nil)
	case err == ErrAddressLimitReached:
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "address limit reached", map[string]string{"code": "address_limit_reached"})
	default:
		h.Log.Error("create address", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to create address", nil)
	}
}

func (h *Handlers) mapUpdateError(w http.ResponseWriter, err error) {
	switch {
	case err == ErrInvalidAddress:
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid address data", nil)
	case err == ErrNotFound || err == ErrNotOwned:
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "address not found", nil)
	default:
		h.Log.Error("update address", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to update address", nil)
	}
}

func (h *Handlers) mapDeleteError(w http.ResponseWriter, err error) {
	switch {
	case err == ErrNotFound || err == ErrNotOwned:
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "address not found", nil)
	default:
		h.Log.Error("delete address", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to delete address", nil)
	}
}

func (h *Handlers) mapSetDefaultError(w http.ResponseWriter, err error) {
	switch {
	case err == ErrNotFound || err == ErrNotOwned:
		httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "address not found", nil)
	default:
		h.Log.Error("set default address", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to set default address", nil)
	}
}

// Response builder

func toAddressResponse(addr sqlcq.Address) addressResponse {
	return addressResponse{
		ID:        addr.ID.String(),
		Label:     addr.Label,
		Line1:     addr.Line1,
		Line2:     addr.Line2,
		City:      addr.City,
		Region:    addr.Region,
		Phone:     addr.Phone,
		IsDefault: addr.IsDefault,
		CreatedAt: addr.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: addr.UpdatedAt.Time.Format(time.RFC3339),
	}
}

// getUserID extracts the authenticated user's ID from the request context.
// Returns false if the user is not authenticated.
func getUserID(r *http.Request) (uuid.UUID, bool) {
	view, ok := auth.GetSessionView(r.Context())
	if !ok {
		return uuid.Nil, false
	}
	return view.UserID, true
}

// parseUUID parses a UUID string and returns an error if invalid.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
