package senders

import (
	"context"

	"go-reasonable-api/support/config"
	"go-reasonable-api/support/logger"

	"github.com/rotisserie/eris"
	"github.com/wneessen/go-mail"
)

// SMTPSender implements EmailSender using SMTP (for MailHog in development)
type SMTPSender struct {
	config *config.Config
}

// NewSMTPSender creates a new SMTP email sender
func NewSMTPSender(cfg *config.Config) *SMTPSender {
	return &SMTPSender{
		config: cfg,
	}
}

// Send sends an email via SMTP
func (s *SMTPSender) Send(ctx context.Context, msg Message) error {
	log := logger.Ctx(ctx)

	m := mail.NewMsg()

	if err := m.From(s.config.Email.From); err != nil {
		return eris.Wrap(err, "failed to set from address")
	}

	if err := m.To(msg.To); err != nil {
		return eris.Wrap(err, "failed to set to address")
	}

	m.Subject(msg.Subject)

	if msg.IsHTML {
		m.SetBodyString(mail.TypeTextHTML, msg.Body)
	} else {
		m.SetBodyString(mail.TypeTextPlain, msg.Body)
	}

	// Configure SMTP client options
	opts := []mail.Option{
		mail.WithPort(s.config.Email.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthNoAuth),
		mail.WithTLSPolicy(mail.NoTLS),
	}

	// Add authentication if credentials are provided
	if s.config.Email.SMTPUser != "" && s.config.Email.SMTPPassword != "" {
		opts = []mail.Option{
			mail.WithPort(s.config.Email.SMTPPort),
			mail.WithSMTPAuth(mail.SMTPAuthPlain),
			mail.WithUsername(s.config.Email.SMTPUser),
			mail.WithPassword(s.config.Email.SMTPPassword),
			mail.WithTLSPolicy(mail.TLSOpportunistic),
		}
	}

	client, err := mail.NewClient(s.config.Email.SMTPHost, opts...)
	if err != nil {
		return eris.Wrap(err, "failed to create SMTP client")
	}

	if err := client.DialAndSend(m); err != nil {
		return eris.Wrap(err, "failed to send email")
	}

	log.Info().
		Str("to", msg.To).
		Str("subject", msg.Subject).
		Msg("email sent via SMTP")

	return nil
}
