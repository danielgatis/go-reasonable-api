package email

import (
	"context"
	"time"

	"go-reasonable-api/app/interfaces/support"
	"go-reasonable-api/support/email/senders"

	"github.com/rotisserie/eris"
)

// BaseData contains common fields for all email templates
type BaseData struct {
	Subject string
	Year    int
}

// NewBaseData creates base data with current year
func NewBaseData(subject string) BaseData {
	return BaseData{
		Subject: subject,
		Year:    time.Now().Year(),
	}
}

// Email is the base struct for all email types
type Email struct {
	templates *Templates
	sender    support.EmailSender
}

// NewEmail creates a new email with templates and sender
func NewEmail(sender support.EmailSender) (*Email, error) {
	tmpl, err := NewTemplates()
	if err != nil {
		return nil, eris.Wrap(err, "failed to load email templates")
	}

	return &Email{
		templates: tmpl,
		sender:    sender,
	}, nil
}

// Send renders and sends an email
func (e *Email) Send(ctx context.Context, templateName, to, subject string, data any) error {
	html, err := e.templates.Render(templateName, data)
	if err != nil {
		return eris.Wrapf(err, "failed to render %s email", templateName)
	}

	msg := senders.Message{
		To:      to,
		Subject: subject,
		Body:    html,
		IsHTML:  true,
	}

	if err := e.sender.Send(ctx, msg); err != nil {
		return eris.Wrapf(err, "failed to send %s email to %s", templateName, to)
	}

	return nil
}
