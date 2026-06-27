package auth

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

type Handlers struct {
	Svc          *Service
	CookieName   string
	CookieDomain string
	Secure       bool
}

func NewHandlers(svc *Service, cookieName, cookieDomain string, secure bool) *Handlers {
	if cookieName == "" {
		cookieName = "rue_session"
	}
	return &Handlers{Svc: svc, CookieName: cookieName, CookieDomain: cookieDomain, Secure: secure}
}

func (h *Handlers) Mount(r chi.Router) {
	r.Post("/auth/signup", h.signup)
	r.Post("/auth/login", h.login)
	r.Post("/auth/logout", h.logout)
	r.Get("/auth/session", h.session)
}

type signupBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// signup godoc
//
// @Summary  Sign up with email and password
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body signupBody true "Signup payload"
// @Success  201 {object} sessionResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  409 {object} httpx.ErrorEnvelope
// @Router   /auth/signup [post]
func (h *Handlers) signup(w http.ResponseWriter, r *http.Request) {
	var body signupBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil)
		return
	}
	res, err := h.Svc.Signup(r.Context(), SignupInput(body), clientIP(r), r.UserAgent())
	switch {
	case errors.Is(err, ErrEmailInUse):
		httpx.WriteError(w, http.StatusConflict, httpx.CodeConflict, "email already in use", nil)
		return
	case errors.Is(err, ErrInvalidCreds):
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid email or password (min 8 chars)", nil)
		return
	case err != nil:
		h.Svc.Log.ErrorContext(r.Context(), "signup", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "signup failed", nil)
		return
	}
	h.setSessionCookie(w, res.SessionToken, res.SessionExpires)
	httpx.WriteJSON(w, http.StatusCreated, sessionResponse{
		UserID:        res.UserID.String(),
		Email:         strings.ToLower(body.Email),
		Role:          "customer",
		EmailVerified: res.EmailVerified,
	})
}

type loginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// login godoc
//
// @Summary  Log in with email and password
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body loginBody true "Login payload"
// @Success  200 {object} sessionResponse
// @Failure  401 {object} httpx.ErrorEnvelope
// @Router   /auth/login [post]
func (h *Handlers) login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := httpx.ReadJSON(r, &body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeBadRequest, "invalid body", nil)
		return
	}
	res, err := h.Svc.Login(r.Context(), LoginInput(body), clientIP(r), r.UserAgent())
	if errors.Is(err, ErrInvalidCreds) {
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "invalid email or password", nil)
		return
	}
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "login", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "login failed", nil)
		return
	}
	h.setSessionCookie(w, res.SessionToken, res.SessionExpires)
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{
		UserID: res.UserID.String(), Email: strings.ToLower(body.Email),
		Role: res.Role,
	})
}

// logout godoc
//
// @Summary  Log out (clear session)
// @Tags     auth
// @Produce  json
// @Success  204
// @Router   /auth/logout [post]
func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(h.CookieName); err == nil {
		_ = h.Svc.Logout(r.Context(), c.Value)
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

type sessionResponse struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	Name          string `json:"name,omitempty"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

// session godoc
//
// @Summary  Get current session
// @Tags     auth
// @Produce  json
// @Success  200 {object} sessionResponse
// @Success  204 "no active session"
// @Router   /auth/session [get]
func (h *Handlers) session(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(h.CookieName)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	view, err := h.Svc.GetSession(r.Context(), c.Value)
	if errors.Is(err, ErrNoSession) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "session check failed", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, sessionResponse{
		UserID: view.UserID.String(), Email: view.Email, Name: view.Name,
		Role: view.Role, EmailVerified: view.EmailVerified,
	})
}

func (h *Handlers) setSessionCookie(w http.ResponseWriter, token string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   h.CookieDomain,
		Expires:  expires,
		MaxAge:   int(time.Until(expires).Seconds()),
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handlers) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clientIP(r *http.Request) net.IP {
	// X-Forwarded-For first hop, fallback to RemoteAddr.
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		if i := strings.IndexByte(x, ','); i > 0 {
			x = x[:i]
		}
		return net.ParseIP(strings.TrimSpace(x))
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil
	}
	return net.ParseIP(host)
}
