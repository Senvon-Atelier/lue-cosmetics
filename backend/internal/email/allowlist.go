package email

import (
	"context"
	"strings"

	"go.uber.org/zap"
)

// AllowlistSender wraps an inner Sender and decides at Send-time whether
// the recipient receives a real delivery. Non-allowlisted recipients are
// logged at Info and the call returns nil (so the caller's flow continues
// as if mail had been sent).
//
// The allowlist is a slice of lowercase email addresses; a single entry "*"
// allowlists every address.
type AllowlistSender struct {
	Inner     Sender
	Allowlist []string
	Log       *zap.Logger
}

func NewAllowlistSender(inner Sender, raw []string, log *zap.Logger) AllowlistSender {
	out := make([]string, 0, len(raw))
	for _, s := range raw {
		if s = strings.TrimSpace(strings.ToLower(s)); s != "" {
			out = append(out, s)
		}
	}
	return AllowlistSender{Inner: inner, Allowlist: out, Log: log}
}

func (s AllowlistSender) Send(ctx context.Context, to, template string, data map[string]any) error {
	addr := strings.ToLower(strings.TrimSpace(to))
	for _, a := range s.Allowlist {
		if a == "*" || a == addr {
			return s.Inner.Send(ctx, to, template, data)
		}
	}
	s.Log.Info("email suppressed (not allowlisted)",
		zap.String("to", addr),
		zap.String("template", template))
	return nil
}
