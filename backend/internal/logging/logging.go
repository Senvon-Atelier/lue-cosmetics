// Package logging plumbs a per-request zap.Logger through context so handlers
// and services can log with request_id (and later user_id/role) attached
// automatically — without having to remember the attributes at every call site.
package logging

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey int

const loggerKey ctxKey = 1

// WithLogger returns a context that carries the given logger.
func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// From returns the logger stored on ctx, or the fallback if none is set.
// Services should pass their own long-lived logger as the fallback so
// background-task logging still works.
func From(ctx context.Context, fallback *zap.Logger) *zap.Logger {
	if l, ok := ctx.Value(loggerKey).(*zap.Logger); ok && l != nil {
		return l
	}
	return fallback
}
