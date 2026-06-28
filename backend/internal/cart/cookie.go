package cart

import "net/http"

// GuestCookieName is the cookie used to identify a guest cart owner.
const GuestCookieName = "rue_guest_cart"

// GuestCookieMaxAge is the lifetime of the guest cookie in seconds (30 days).
const GuestCookieMaxAge = 30 * 24 * 3600

// SetGuestCookie installs the rue_guest_cart cookie. It is intentionally NOT
// HttpOnly so the frontend can read it and mirror the token to localStorage
// for the post-login merge call.
func SetGuestCookie(w http.ResponseWriter, token, domain string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     GuestCookieName,
		Value:    token,
		Path:     "/",
		Domain:   domain,
		MaxAge:   GuestCookieMaxAge,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearGuestCookie emits a zero-Max-Age cookie that instructs the browser to
// drop rue_guest_cart.
func ClearGuestCookie(w http.ResponseWriter, domain string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     GuestCookieName,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
