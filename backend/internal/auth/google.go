package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

const oauthStateCookie = "rue_oauth_state"
const oauthStateMaxAge = 10 * 60

func (h *Handlers) oauthConfig() (*oauth2.Config, bool) {
	if h.GoogleClientID == "" || h.GoogleClientSecret == "" {
		return nil, false
	}
	return &oauth2.Config{
		ClientID:     h.GoogleClientID,
		ClientSecret: h.GoogleClientSecret,
		RedirectURL:  h.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}, true
}

// googleStart godoc
//
// @Summary  Begin Google OAuth login
// @Tags     auth
// @Success  302 "redirect to Google"
// @Failure  503 {object} httpx.ErrorEnvelope
// @Router   /auth/google/start [get]
func (h *Handlers) googleStart(w http.ResponseWriter, r *http.Request) {
	cfg, ok := h.oauthConfig()
	if !ok {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeNotConfigured, "google oauth not configured", nil)
		return
	}
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "state generation failed", nil)
		return
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    state,
		Path:     "/",
		Domain:   h.CookieDomain,
		MaxAge:   oauthStateMaxAge,
		HttpOnly: true,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// afterOAuthRedirect returns the frontend URL to redirect to after a successful OAuth callback.
func (h *Handlers) afterOAuthRedirect() string {
	if h.FrontendBaseURL == "" {
		return "/"
	}
	u, err := url.Parse(h.FrontendBaseURL)
	if err != nil {
		return "/"
	}
	u.Path = "/account"
	return u.String()
}

// IDTokenValidator abstracts Google's idtoken.Validate for testability.
type IDTokenValidator interface {
	Validate(ctx context.Context, token, audience string) (*Payload, error)
}

// Payload holds the extracted claims from a verified Google ID token.
type Payload struct {
	Subject string
	Email   string
	Name    string
}

type googleValidator struct{}

func (googleValidator) Validate(ctx context.Context, token, audience string) (*Payload, error) {
	p, err := idtoken.Validate(ctx, token, audience)
	if err != nil {
		return nil, err
	}
	return &Payload{
		Subject: p.Subject,
		Email:   asString(p.Claims["email"]),
		Name:    asString(p.Claims["name"]),
	}, nil
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

// googleCallback godoc
//
// @Summary  Handle Google OAuth callback
// @Tags     auth
// @Success  302
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  503 {object} httpx.ErrorEnvelope
// @Router   /auth/google/callback [get]
func (h *Handlers) googleCallback(w http.ResponseWriter, r *http.Request) {
	cfg, ok := h.oauthConfig()
	if !ok {
		httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeNotConfigured, "google oauth not configured", nil)
		return
	}
	q := r.URL.Query()
	if errParam := q.Get("error"); errParam != "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "oauth denied: "+errParam, nil)
		return
	}
	state := q.Get("state")
	code := q.Get("code")
	if state == "" || code == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "missing state or code", nil)
		return
	}
	c, err := r.Cookie(oauthStateCookie)
	// Clear state cookie immediately whether or not it validates — never allow replay.
	http.SetCookie(w, &http.Cookie{
		Name: oauthStateCookie, Value: "", Path: "/", Domain: h.CookieDomain,
		MaxAge: -1, HttpOnly: true, Secure: h.Secure, SameSite: http.SameSiteLaxMode,
	})
	if err != nil || c.Value == "" || c.Value != state {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid oauth state", nil)
		return
	}

	tok, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "oauth exchange", "err", err)
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "code exchange failed", nil)
		return
	}
	idToken, _ := tok.Extra("id_token").(string)
	if idToken == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "missing id_token", nil)
		return
	}
	validator := h.Validator
	if validator == nil {
		validator = googleValidator{}
	}
	payload, err := validator.Validate(r.Context(), idToken, h.GoogleClientID)
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "id token", "err", err)
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid id_token", nil)
		return
	}
	if payload.Subject == "" || payload.Email == "" {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "incomplete id_token claims", nil)
		return
	}
	res, err := h.Svc.LoginWithGoogle(r.Context(), payload.Subject, payload.Email, payload.Name, clientIP(r), r.UserAgent())
	if err != nil {
		h.Svc.Log.ErrorContext(r.Context(), "login with google", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "login failed", nil)
		return
	}
	h.setSessionCookie(w, res.SessionToken, res.SessionExpires)
	http.Redirect(w, r, h.afterOAuthRedirect(), http.StatusFound)
}
