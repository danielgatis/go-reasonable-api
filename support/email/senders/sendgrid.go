package senders

import (
	"context"
	"net/http"

	"go-reasonable-api/support/config"
	"go-reasonable-api/support/logger"

	"github.com/rotisserie/eris"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGridSender implements EmailSender using SendGrid API (for production)
type SendGridSender struct {
	config *config.Config
}

// NewSendGridSender creates a new SendGrid email sender
func NewSendGridSender(cfg *config.Config) *SendGridSender {
	return &SendGridSender{
		config: cfg,
	}
}

// Send sends an email via SendGrid API
func (s *SendGridSender) Send(ctx context.Context, msg Message) error {
	log := logger.Ctx(ctx)

	if s.config.Email.SendGridAPIKey == "" {
		return eris.New("sendgrid API key not configured")
	}

	from := mail.NewEmail(s.config.Email.FromName, s.config.Email.From)
	to := mail.NewEmail("", msg.To)

	var message *mail.SGMailV3
	if msg.IsHTML {
		message = mail.NewSingleEmail(from, msg.Subject, to, "", msg.Body)
	} else {
		message = mail.NewSingleEmail(from, msg.Subject, to, msg.Body, "")
	}

	client := sendgrid.NewSendClient(s.config.Email.SendGridAPIKey)
	response, err := client.SendWithContext(ctx, message)
	if err != nil {
		return eris.Wrap(err, "failed to send email via SendGrid")
	}

	if response.StatusCode >= http.StatusBadRequest {
		return eris.Errorf("sendgrid returned error status %d: %s", response.StatusCode, response.Body)
	}

	log.Info().
		Str("to", msg.To).
		Str("subject", msg.Subject).
		Int("status_code", response.StatusCode).
		Msg("email sent via SendGrid")

	return nil
}
