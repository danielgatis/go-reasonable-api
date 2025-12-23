package support

import (
	"context"

	"go-reasonable-api/support/email/senders"
)

// EmailSender abstracts email delivery.
//
// Implementations include SMTP (for development with MailHog)
// and SendGrid (for production). The provider is selected via config.
type EmailSender interface {
	Send(ctx context.Context, msg senders.Message) error
}
