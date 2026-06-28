package httpx

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/oti-adjei/ruecosmetics/internal/logging"
)

type ctxKey int

const RequestIDKey ctxKey = 1

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", rid)
		ctx := context.WithValue(r.Context(), RequestIDKey, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDKey).(string); ok {
		return v
	}
	return ""
}

// RequestLogger creates a per-request *zap.Logger scoped with the request_id
// from context (set by RequestID middleware) and stashes it on r.Context()
// via logging.WithLogger. Downstream code should call logging.From(ctx, base)
// to retrieve it.
//
// RequestLogger MUST be installed AFTER RequestID so the request_id is in
// context.
func RequestLogger(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			fields := []zap.Field{}
			if rid := GetRequestID(ctx); rid != "" {
				fields = append(fields, zap.String("request_id", rid))
			}
			reqLog := base.With(fields...)
			ctx = logging.WithLogger(ctx, reqLog)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Recovery(base *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log := logging.From(r.Context(), base)
					log.Error("panic recovered",
						zap.Any("panic", rec),
						zap.String("path", r.URL.Path),
					)
					WriteError(w, http.StatusInternalServerError, CodeInternal, "internal server error", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
