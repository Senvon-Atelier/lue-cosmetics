package httpx_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func TestReadJSONOK(t *testing.T) {
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	var v struct{ Name string }
	if err := httpx.ReadJSON(req, &v); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}
	if v.Name != "x" {
		t.Errorf("Name = %q", v.Name)
	}
}

func TestReadJSONRejectsUnknownFields(t *testing.T) {
	req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"name":"x","extra":1}`))
	var v struct{ Name string }
	if err := httpx.ReadJSON(req, &v); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestWriteJSONSetsHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteJSON(rec, http.StatusOK, map[string]string{"ok": "yes"})
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("ct = %q", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), `"ok":"yes"`) {
		t.Errorf("body = %s", rec.Body.String())
	}
}

func TestReadJSONRejectsOversizedBody(t *testing.T) {
	// Body slightly larger than the 1 MiB cap enforced by ReadJSON.
	oversized := `{"name":"` + strings.Repeat("x", 1<<20) + `"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	var v struct{ Name string }
	if err := httpx.ReadJSON(req, &v); err == nil {
		t.Fatal("expected error for oversized body")
	}
}
