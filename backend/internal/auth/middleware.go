package auth

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

type ctxKey int

const (
	sessionKey ctxKey = iota + 1
	userIDKey
	roleKey
)

func (h *Handlers) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(h.CookieName)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
			return
		}
		view, err := h.Svc.GetSession(r.Context(), c.Value)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "authentication required", nil)
			return
		}
		ctx := context.WithValue(r.Context(), sessionKey, view)
		ctx = context.WithValue(ctx, userIDKey, view.UserID)
		ctx = context.WithValue(ctx, roleKey, view.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(userIDKey).(uuid.UUID)
	return v, ok
}

func GetRole(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(roleKey).(string)
	return v, ok
}

func GetSessionView(ctx context.Context) (SessionView, bool) {
	v, ok := ctx.Value(sessionKey).(SessionView)
	return v, ok
}

func (h *Handlers) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got, ok := GetRole(r.Context())
			if !ok || got != role {
				httpx.WriteError(w, http.StatusForbidden, httpx.CodeForbidden, "forbidden", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// MustBeAdmin is the belt-and-suspenders check called inside admin handlers.
// If the caller is somehow not admin, it writes 403 and returns false.
func MustBeAdmin(w http.ResponseWriter, r *http.Request) bool {
	if role, ok := GetRole(r.Context()); !ok || role != "admin" {
		httpx.WriteError(w, http.StatusForbidden, httpx.CodeForbidden, "admin required", nil)
		return false
	}
	return true
}
