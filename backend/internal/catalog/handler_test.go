package catalog_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
)

func newHandlers(t *testing.T) (*catalog.Handlers, func()) {
	t.Helper()
	repo, _, cleanup := newRepo(t)
	return catalog.NewHandlers(repo), cleanup
}

func TestGetCategoriesReturnsList(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	r := chi.NewRouter()
	h.Mount(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/categories", nil).WithContext(context.Background()))
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body) != 2 {
		t.Errorf("got %d categories", len(body))
	}
}

func TestGetProductsAppliesFiltersAndSort(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	r := chi.NewRouter()
	h.Mount(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/products?category=skincare&sort=price_asc&limit=5", nil))
	if rec.Code != 200 {
		t.Fatalf("code = %d, body = %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Items []struct{ Slug string } `json:"items"`
		Total int                     `json:"total"`
		Page  int                     `json:"page"`
		Limit int                     `json:"limit"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("total = %d", resp.Total)
	}
	if resp.Page != 1 || resp.Limit != 5 {
		t.Errorf("pagination = page %d, limit %d", resp.Page, resp.Limit)
	}
	if len(resp.Items) == 0 || resp.Items[0].Slug != "gentle-cleanser" {
		t.Errorf("cheapest skincare should be gentle-cleanser, got %+v", resp.Items)
	}
}

func TestGetProductsRejectsBadSort(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	r := chi.NewRouter()
	h.Mount(r)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/products?sort=DROP%20TABLE", nil))
	if rec.Code != 400 {
		t.Errorf("code = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"code":"validation_failed"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestGetProductBySlug404(t *testing.T) {
	h, cleanup := newHandlers(t)
	defer cleanup()
	r := chi.NewRouter()
	h.Mount(r)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/products/no-such-thing", nil))
	if rec.Code != 404 {
		t.Errorf("code = %d", rec.Code)
	}
}
