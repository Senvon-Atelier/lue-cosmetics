package httpx_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func TestRequestIDPropagatesIncoming(t *testing.T) {
	var seen string
	h := httpx.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = httpx.GetRequestID(r.Context())
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-Id", "rid-abc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if seen != "rid-abc" {
		t.Errorf("seen = %q", seen)
	}
	if rec.Header().Get("X-Request-Id") != "rid-abc" {
		t.Errorf("response header = %q", rec.Header().Get("X-Request-Id"))
	}
}

func TestRequestIDGeneratesWhenMissing(t *testing.T) {
	var seen string
	h := httpx.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = httpx.GetRequestID(r.Context())
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if seen == "" {
		t.Fatal("expected generated request id")
	}
	if rec.Header().Get("X-Request-Id") != seen {
		t.Errorf("response header mismatch: %q vs %q", rec.Header().Get("X-Request-Id"), seen)
	}
}

func TestRecoveryReturnsEnvelope(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := httpx.Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != 500 {
		t.Errorf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"internal_error"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}
