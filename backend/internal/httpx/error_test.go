package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

func TestWriteErrorShape(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(rec, http.StatusBadRequest, httpx.CodeBadRequest, "boom", map[string]string{"x": "missing"})
	if rec.Code != 400 {
		t.Errorf("status = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}
	var env httpx.ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error.Code != "bad_request" || env.Error.Message != "boom" {
		t.Errorf("envelope = %+v", env)
	}
	if env.Error.Fields["x"] != "missing" {
		t.Errorf("fields = %+v", env.Error.Fields)
	}
}

func TestWriteErrorOmitsFieldsWhenNil(t *testing.T) {
	rec := httptest.NewRecorder()
	httpx.WriteError(rec, http.StatusInternalServerError, httpx.CodeInternal, "oops", nil)
	if got := rec.Body.String(); got == "" {
		t.Fatal("empty body")
	}
	// fields should not appear in JSON when nil
	if got := rec.Body.String(); strings.Contains(got, "fields") {
		t.Errorf("expected no fields key, got %s", got)
	}
}
