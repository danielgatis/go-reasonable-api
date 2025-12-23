package providers

import (
	"go-reasonable-api/app/interfaces/support"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/email/senders"
)

// ProvideEmailSender creates the appropriate EmailSender based on configuration
// In development (default), it uses SMTP (MailHog)
// In production, it should use SendGrid (configured via email.provider)
func ProvideEmailSender(cfg *config.Config) support.EmailSender {
	switch cfg.Email.Provider {
	case "sendgrid":
		return senders.NewSendGridSender(cfg)
	case "smtp":
		fallthrough
	default:
		return senders.NewSMTPSender(cfg)
	}
}
