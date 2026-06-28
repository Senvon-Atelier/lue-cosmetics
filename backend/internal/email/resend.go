package email

import (
	"context"
	"errors"

	"github.com/resendlabs/resend-go"
	"go.uber.org/zap"
)

var ErrResendNotConfigured = errors.New("email: resend not configured")

type ResendSender struct {
	Client    *resend.Client
	FromEmail string
	Renderer  *Renderer
	Log       *zap.Logger
}

// NewResendSender returns a sender bound to a real Resend account. If apiKey
// or fromEmail is empty, returns (nil, ErrResendNotConfigured) so app.New
// can fall back to a LogSender.
func NewResendSender(apiKey, fromEmail string, r *Renderer, log *zap.Logger) (*ResendSender, error) {
	if apiKey == "" || fromEmail == "" {
		return nil, ErrResendNotConfigured
	}
	return &ResendSender{
		Client:    resend.NewClient(apiKey),
		FromEmail: fromEmail,
		Renderer:  r,
		Log:       log,
	}, nil
}

func (s *ResendSender) Send(ctx context.Context, to, template string, data map[string]any) error {
	html, text, err := s.Renderer.Render(template, data)
	if err != nil {
		return err
	}
	subject := subjectFor(template, data)
	_, err = s.Client.Emails.Send(&resend.SendEmailRequest{
		From:    s.FromEmail,
		To:      []string{to},
		Subject: subject,
		Html:    html,
		Text:    text,
	})
	if err != nil {
		return err
	}
	return nil
}

// subjectFor maps a template name + data to a subject line. Keep this tiny;
// proliferate as templates grow.
func subjectFor(template string, data map[string]any) string {
	switch template {
	case "verify_email":
		return "Verify your email for Rue Cosmetics"
	case "password_reset":
		return "Reset your password"
	case "welcome":
		return "Welcome to Rue Cosmetics"
	case "order_confirmation":
		if ref, ok := data["paystack_reference"].(string); ok {
			return "Your Rue Cosmetics order " + ref
		}
		return "Your Rue Cosmetics order"
	}
	return "Rue Cosmetics"
}
