// Package email exposes the Sender interface that auth + orders flow through.
// Plan 5 will add a Resend-backed implementation; this plan ships only the
// LogSender, which writes the payload to slog without delivering anything.
package email

import (
	"context"
	"log/slog"
)

type Sender interface {
	Send(ctx context.Context, to, template string, data map[string]any) error
}

type LogSender struct {
	Log *slog.Logger
}

func (s LogSender) Send(ctx context.Context, to, template string, data map[string]any) error {
	s.Log.InfoContext(ctx, "email (stubbed)", "to", to, "template", template, "data", data)
	return nil
}
