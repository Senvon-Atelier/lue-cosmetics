package email

import (
	"errors"

	"go.uber.org/zap"
)

// Select picks the outgoing email implementation from config, loudly.
// Production without a configured provider is a startup error, not a
// silent fallback to log-only delivery.
func Select(env, resendAPIKey, resendFrom string, allowlist []string, renderer *Renderer, log *zap.Logger) (Sender, error) {
	var inner Sender
	switch {
	case resendAPIKey != "":
		rs, err := NewResendSender(resendAPIKey, resendFrom, renderer, log)
		if err != nil {
			return nil, err
		}
		log.Info("email: using Resend sender", zap.String("from", resendFrom))
		inner = rs
	case env == "production":
		return nil, errors.New("email: RESEND_API_KEY is required in production")
	default:
		log.Warn("email: no provider configured, using log-only sender")
		inner = LogSender{Log: log}
	}
	return NewAllowlistSender(inner, allowlist, log), nil
}
