// Package api assembles the full production HTTP router from an Application.
// It lives outside package app because internal/health imports app.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oti-adjei/ruecosmetics/internal/addresses"
	"github.com/oti-adjei/ruecosmetics/internal/admin"
	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/cart"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/health"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/me"
	"github.com/oti-adjei/ruecosmetics/internal/orders"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

// New builds the production router. main.go and integration tests share it.
func New(a *app.Application) http.Handler {
	cfg := a.Config
	r := chi.NewRouter()
	r.Use(httpx.Recovery(a.Logger))
	r.Use(httpx.RequestID)
	r.Use(httpx.RequestLogger(a.Logger))
	r.Use(httpx.CORS(cfg.CORSOrigins))

	// /healthz stays at the root for uptime monitoring.
	r.Get("/healthz", health.Handler(a))

	// All public + future protected APIs mount under /api/v1.
	catalogHandlers := catalog.NewHandlers(a.Catalog, a.Logger)
	shippingHandlers := shipping.NewHandlers(a.Shipping, a.Logger)
	r.Route("/api/v1", func(apiRouter chi.Router) {
		catalogHandlers.Mount(apiRouter)
		shippingHandlers.Mount(apiRouter)

		secure := cfg.Env != "development"
		authHandlers := auth.NewHandlers(a.Auth, cfg.SessionCookieName, cfg.SessionCookieDomain, secure)
		authHandlers.GoogleClientID = cfg.GoogleClientID
		authHandlers.GoogleClientSecret = cfg.GoogleClientSecret
		authHandlers.GoogleRedirectURL = cfg.GoogleRedirectURL
		authHandlers.FrontendBaseURL = cfg.FrontendBaseURL
		authHandlers.Mount(apiRouter) // public: /auth/signup, /auth/login, /auth/logout, /auth/session, /auth/google/start, /auth/google/callback

		cartHandlers := cart.NewHandlers(a.Cart, a.Auth, cfg.SessionCookieName, cfg.SessionCookieDomain, secure)
		cartHandlers.Mount(apiRouter) // public: GET /cart, POST/PATCH/DELETE /cart/items

		ordersHandlers := orders.NewHandlers(a.Orders, cfg.PaystackSecretKey, a.Logger)
		ordersHandlers.MountPublic(apiRouter) // public: POST /webhooks/paystack

		// Admin routes (require admin role)
		adminHandlers := admin.NewHandlers(a.Admin, authHandlers, a.Logger)
		adminHandlers.MountPublic(apiRouter) // renamed to Mount in Task 4

		// Auth-gated routes (one Group with RequireSession middleware)
		apiRouter.Group(func(r chi.Router) {
			r.Use(authHandlers.RequireSession)
			ordersRepo := orders.NewRepository(a.Pool)
			profileSvc := me.NewProfileService(auth.NewRepository(a.Pool), a.Logger)
			meHandlers := me.NewHandlers(ordersRepo, profileSvc, a.Logger)
			authHandlers.MountAuthGated(r)   // POST /auth/verify-email/resend
			cartHandlers.MountAuthGated(r)   // POST /cart/merge
			ordersHandlers.MountAuthGated(r) // POST /checkout/init, GET /checkout/verify/{reference}

			addressesHandlers := addresses.NewHandlers(a.Addresses, a.Logger)
			r.Route("/me", func(meRouter chi.Router) {
				meHandlers.MountRoutes(meRouter)  // GET /me, GET /me/orders, GET /me/orders/:id, PATCH /me
				addressesHandlers.Mount(meRouter) // POST/GET/PATCH/DELETE /me/addresses*, POST /me/addresses/{id}/default
			})
		})
	})

	return r
}
