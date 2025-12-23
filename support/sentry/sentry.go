package sentry

import (
	"time"

	"go-reasonable-api/support/config"
	"go-reasonable-api/support/version"

	"github.com/getsentry/sentry-go"
	"github.com/rotisserie/eris"
)

// Init initializes Sentry with the given configuration.
// Returns a cleanup function that should be deferred.
// If DSN is empty, Sentry is disabled and a no-op cleanup is returned.
func Init(cfg *config.Config, serverName string) (func(), error) {
	if cfg.Sentry.DSN == "" {
		return func() {}, nil
	}

	opts := sentry.ClientOptions{
		Dsn:              cfg.Sentry.DSN,
		Environment:      string(cfg.Environment),
		Release:          version.Short(),
		ServerName:       serverName,
		AttachStacktrace: true,
	}

	if cfg.Sentry.EnableTracing {
		opts.EnableTracing = true
		opts.TracesSampleRate = cfg.Sentry.TracesSampleRate
	}

	if err := sentry.Init(opts); err != nil {
		return nil, eris.Wrap(err, "failed to initialize sentry")
	}

	cleanup := func() {
		sentry.Flush(2 * time.Second)
	}

	return cleanup, nil
}

// CaptureError captures an error to Sentry with optional extra context.
// eris errors are automatically handled with their full stack trace.
func CaptureError(err error, extras map[string]any) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		// Add eris formatted error as extra context for better debugging
		scope.SetExtra("eris_formatted", eris.ToString(err, true))

		for k, v := range extras {
			scope.SetExtra(k, v)
		}

		sentry.CaptureException(err)
	})
}

// CaptureMessage captures a message to Sentry.
func CaptureMessage(message string) {
	sentry.CaptureMessage(message)
}

// SetUser sets the user context for Sentry events.
func SetUser(id, email, username string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       id,
			Email:    email,
			Username: username,
		})
	})
}

// SetTag sets a tag for all future Sentry events.
func SetTag(key, value string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}
