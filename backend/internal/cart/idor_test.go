package cart

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// TestCartIDORMatrix exercises the cross-tenant access matrix at the full HTTP
// stack: cart_item_ids from one cart MUST NOT be addressable from another
// cart's identity. Mismatches surface as 404 (not 403) to avoid enumeration.
func TestCartIDORMatrix(t *testing.T) {
	ctx, pool, cleanup := testsupport.StartPool(t, "../../migrations")
	defer cleanup()
	logger := zap.NewNop()

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, logger, email.LogSender{Log: logger}, nil)
	authSvc.Params = auth.TestParams
	authHandlers := auth.NewHandlers(authSvc, "rue_session", "", false)
	cartSvc := newCartService(t, pool)
	cartHandlers := NewHandlers(cartSvc, authSvc, "rue_session", "", false)

	router := chi.NewRouter()
	authHandlers.Mount(router)
	cartHandlers.Mount(router)
	router.Group(func(r chi.Router) {
		r.Use(authHandlers.RequireSession)
		authHandlers.MountAuthGated(r)
		cartHandlers.MountAuthGated(r)
	})

	productID := seedTestProduct(t, ctx, pool)

	// ── set up two users with items ─────────────────────────────────────────
	sessionA := signupGetCookie(t, router, "idor-a@cart.test")
	sessionB := signupGetCookie(t, router, "idor-b@cart.test")

	addItem := func(t *testing.T, session *http.Cookie) string {
		t.Helper()
		rr := doJSON(t, router, http.MethodPost, "/cart/items", map[string]any{
			"product_id": productID.String(), "qty": 1,
		}, session)
		if rr.Code != http.StatusOK {
			t.Fatalf("user add = %d; %s", rr.Code, rr.Body.String())
		}
		var body cartResponse
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(body.Items) == 0 {
			t.Fatal("expected at least one item after add")
		}
		return body.Items[0].ID
	}
	userAItemID := addItem(t, sessionA)
	userBItemID := addItem(t, sessionB)

	// ── set up two guest carts with items ───────────────────────────────────
	addGuestItem := func(t *testing.T) (*http.Cookie, string) {
		t.Helper()
		rr := doJSON(t, router, http.MethodPost, "/cart/items", map[string]any{
			"product_id": productID.String(), "qty": 1,
		})
		if rr.Code != http.StatusOK {
			t.Fatalf("guest add = %d; %s", rr.Code, rr.Body.String())
		}
		cookie := testsupport.FindCookie(rr.Result(), GuestCookieName)
		if cookie == nil {
			t.Fatal("no guest cookie minted")
		}
		var body cartResponse
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		return cookie, body.Items[0].ID
	}
	guestACookie, guestAItemID := addGuestItem(t)
	guestBCookie, guestBItemID := addGuestItem(t)
	_ = guestAItemID
	_ = userBItemID

	// ── matrix ──────────────────────────────────────────────────────────────

	t.Run("UserA_PATCH_UserB_item_404", func(t *testing.T) {
		rr := doJSON(t, router, http.MethodPatch, "/cart/items/"+userBItemID, map[string]any{
			"qty": 9,
		}, sessionA)
		if rr.Code != http.StatusNotFound {
			t.Errorf("PATCH = %d, want 404; %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("UserA_DELETE_UserB_item_404", func(t *testing.T) {
		rr := doJSON(t, router, http.MethodDelete, "/cart/items/"+userBItemID, nil, sessionA)
		if rr.Code != http.StatusNotFound {
			t.Errorf("DELETE = %d, want 404; %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("GuestA_PATCH_GuestB_item_404", func(t *testing.T) {
		rr := doJSON(t, router, http.MethodPatch, "/cart/items/"+guestBItemID, map[string]any{
			"qty": 9,
		}, guestACookie)
		if rr.Code != http.StatusNotFound {
			t.Errorf("PATCH = %d, want 404; %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("GuestA_DELETE_GuestB_item_404", func(t *testing.T) {
		rr := doJSON(t, router, http.MethodDelete, "/cart/items/"+guestBItemID, nil, guestACookie)
		if rr.Code != http.StatusNotFound {
			t.Errorf("DELETE = %d, want 404; %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("Anonymous_PATCH_any_item_404", func(t *testing.T) {
		// No cookies — resolveIdentity returns empty, service finds no cart,
		// returns ErrItemNotFound → 404. Crucially, the request must not 200
		// or 5xx — no fishing.
		rr := doJSON(t, router, http.MethodPatch, "/cart/items/"+userAItemID, map[string]any{
			"qty": 9,
		})
		if rr.Code != http.StatusNotFound {
			t.Errorf("anon PATCH = %d, want 404; %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("UserA_session_plus_GuestB_cookie_GET_returns_UserA_cart", func(t *testing.T) {
		// resolveIdentity must prefer the session cookie. The guest cookie is
		// ignored unless /cart/merge is invoked. UserA's cart contains 1 item;
		// guestB's cart contains 1 item. With both cookies, the response must
		// reflect UserA's cart — not guestB's.
		rr := doJSON(t, router, http.MethodGet, "/cart", nil, sessionA, guestBCookie)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET = %d; %s", rr.Code, rr.Body.String())
		}
		var body cartResponse
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(body.Items) != 1 {
			t.Fatalf("UserA cart len = %d, want 1", len(body.Items))
		}
		if body.Items[0].ID != userAItemID {
			t.Errorf("returned item id = %s, want UserA's %s (session must win over guest cookie)",
				body.Items[0].ID, userAItemID)
		}
	})
}
