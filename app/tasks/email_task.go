package tasks

import (
	"context"

	"go-reasonable-api/app/interfaces/support"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/email"
	"go-reasonable-api/support/logger"
	"go-reasonable-api/support/taskqueue"

	"github.com/hibiken/asynq"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

const TypeEmail = "email:send"

// EmailTaskOptions returns the asynq options for email tasks from config.
func EmailTaskOptions(cfg *config.Config) []asynq.Option {
	return []asynq.Option{
		asynq.MaxRetry(cfg.Worker.EmailMaxRetry),
		asynq.Timeout(cfg.Worker.EmailTimeout),
		asynq.Retention(cfg.Worker.EmailRetention),
	}
}

// EmailPayload is the generic payload for sending any email
type EmailPayload struct {
	To       string         `json:"to"`
	Subject  string         `json:"subject"`
	Template string         `json:"template"`
	Data     map[string]any `json:"data"`
}

// EmailTask is a generic task handler that can send any email template
type EmailTask struct {
	logger *zerolog.Logger
	email  *email.Email
}

func NewEmailTask(logger *zerolog.Logger, emailSender support.EmailSender) (*EmailTask, error) {
	e, err := email.NewEmail(emailSender)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create email")
	}

	return &EmailTask{
		logger: logger,
		email:  e,
	}, nil
}

func (t *EmailTask) Handle(ctx context.Context, task *asynq.Task) error {
	var payload EmailPayload
	meta, err := taskqueue.UnwrapPayload(task.Payload(), &payload)
	if err != nil {
		return eris.Wrap(err, "failed to unmarshal payload")
	}

	ctx = meta.LoggerContext(ctx, t.logger)
	log := logger.Ctx(ctx)

	log.Info().
		Str("task", TypeEmail).
		Str("to", payload.To).
		Str("template", payload.Template).
		Msg("sending email")

	if err := t.email.Send(ctx, payload.Template, payload.To, payload.Subject, payload.Data); err != nil {
		log.Error().Err(err).
			Str("template", payload.Template).
			Str("to", payload.To).
			Msg("failed to send email")
		return eris.Wrap(err, "failed to send email")
	}

	log.Info().
		Str("to", payload.To).
		Str("template", payload.Template).
		Msg("email sent successfully")

	return nil
}
