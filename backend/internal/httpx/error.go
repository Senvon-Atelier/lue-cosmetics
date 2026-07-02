package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/oti-adjei/ruecosmetics/internal/logging"
	"go.uber.org/zap"
)

type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

const (
	CodeBadRequest    = "bad_request"
	CodeUnauthorized  = "unauthorized"
	CodeForbidden     = "forbidden"
	CodeNotFound      = "not_found"
	CodeConflict      = "conflict"
	CodeInternal      = "internal_error"
	CodeValidation    = "validation_failed"
	CodeNotConfigured = "not_configured"
	CodeUpstream      = "upstream_error"
)

func WriteError(w http.ResponseWriter, status int, code, message string, fields map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorEnvelope{
		Error: ErrorBody{Code: code, Message: message, Fields: fields},
	})
}

// WriteInternal logs err (with the request-scoped logger, so request_id is
// attached) and writes the standard 500 envelope with a generic public
// message. Use this for every unexpected-error branch instead of a bare
// WriteError, so no 500 is ever silent.
func WriteInternal(w http.ResponseWriter, r *http.Request, fallback *zap.Logger, publicMsg string, err error) {
	logging.From(r.Context(), fallback).Error(publicMsg,
		zap.Error(err),
		zap.String("path", r.URL.Path),
	)
	WriteError(w, http.StatusInternalServerError, CodeInternal, publicMsg, nil)
}
