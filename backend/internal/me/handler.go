package me

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

type Handlers struct{}

func NewHandlers() *Handlers { return &Handlers{} }

func (h *Handlers) Mount(r chi.Router) {
	r.Get("/me", h.get)
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
