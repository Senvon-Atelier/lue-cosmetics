package httpx_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
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

func TestWriteInternalLogsAndWritesEnvelope(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	base := zap.New(core)

	req := httptest.NewRequest(http.MethodGet, "/things", nil)
	rec := httptest.NewRecorder()

	httpx.WriteInternal(rec, req, base, "failed to list things", errors.New("pg: connection refused"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	var env httpx.ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error.Code != httpx.CodeInternal || env.Error.Message != "failed to list things" {
		t.Fatalf("envelope = %+v", env.Error)
	}
	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	if entries[0].Message != "failed to list things" {
		t.Fatalf("log message = %q", entries[0].Message)
	}
	found := false
	for _, f := range entries[0].Context {
		if f.Key == "error" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected zap error field on log entry")
	}
}
