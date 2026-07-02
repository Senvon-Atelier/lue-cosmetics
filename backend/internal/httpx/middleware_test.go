package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

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
	logger := zap.NewNop()
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

func TestAccessLogEmitsOneEntryPerRequest(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	base := zap.New(core)

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("short and stout"))
	})
	h := httpx.RequestID(httpx.AccessLog(base)(inner))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(entries))
	}
	got := entries[0].ContextMap()
	if got["method"] != "GET" || got["path"] != "/api/v1/products" {
		t.Fatalf("context = %+v", got)
	}
	if got["status"] != int64(http.StatusTeapot) {
		t.Fatalf("status = %v, want 418", got["status"])
	}
	if _, ok := got["request_id"]; !ok {
		t.Fatal("missing request_id")
	}
}
