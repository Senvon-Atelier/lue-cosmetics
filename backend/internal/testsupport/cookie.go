package testsupport

import "net/http"

// FindCookie returns the cookie with the given name from a response, or nil
// if absent. Use this in tests instead of inline range-over-Cookies() loops.
func FindCookie(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
