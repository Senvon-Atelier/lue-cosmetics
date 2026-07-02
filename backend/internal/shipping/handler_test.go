package shipping_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
	"go.uber.org/zap"
)

func newHandlers(t *testing.T) *shipping.Handlers {
	t.Helper()
	p := writeConfig(t, 2500, 50000)
	svc, err := shipping.NewService(p)
	if err != nil {
		t.Fatalf("svc: %v", err)
	}
	return shipping.NewHandlers(svc, zap.NewNop())
}

func TestQuoteHandlerReturnsAppliedShipping(t *testing.T) {
	h := newHandlers(t)
	r := chi.NewRouter()
	h.Mount(r)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/shipping/quote?subtotal=10000", nil))
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var q shipping.Quote
	if err := json.Unmarshal(rec.Body.Bytes(), &q); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if q.AppliedCostGhsMinor != 2500 || q.FreeShippingRemainderGhsMinor != 40000 {
		t.Errorf("quote = %+v", q)
	}
}

func TestQuoteHandlerValidatesSubtotal(t *testing.T) {
	h := newHandlers(t)
	r := chi.NewRouter()
	h.Mount(r)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/shipping/quote", nil))
	if rec.Code != 400 {
		t.Errorf("missing subtotal: code = %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/shipping/quote?subtotal=-1", nil))
	if rec.Code != 400 {
		t.Errorf("negative subtotal: code = %d", rec.Code)
	}
}
