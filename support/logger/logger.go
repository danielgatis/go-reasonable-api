package logger

import (
	"context"
	"io"
	"os"
	"time"

	"go-reasonable-api/support/config"

	"github.com/rs/zerolog"
)

type ctxKey struct{}
type requestIDKey struct{}
type userIDKey struct{}
type envKey struct{}

var log zerolog.Logger
var env config.Environment

func Init(cfg *config.Config) {
	env = cfg.Environment

	var output io.Writer = os.Stdout

	// DEV: pretty log with colors (human-readable)
	// PROD: structured JSON log
	if cfg.Environment.IsDev() {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	lvl, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	log = zerolog.New(output).
		Level(lvl).
		With().
		Timestamp().
		Logger()
}

func IsDev() bool {
	return env.IsDev()
}

func Get() *zerolog.Logger {
	return &log
}

func Debug() *zerolog.Event {
	return log.Debug()
}

func Info() *zerolog.Event {
	return log.Info()
}

func Warn() *zerolog.Event {
	return log.Warn()
}

func Error() *zerolog.Event {
	return log.Error()
}

func Fatal() *zerolog.Event {
	return log.Fatal()
}

func With() zerolog.Context {
	return log.With()
}

// WithContext returns a new context with the logger attached.
func WithContext(ctx context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

// FromContext returns the logger from the context.
// If no logger is found, returns the global logger.
func FromContext(ctx context.Context) *zerolog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(*zerolog.Logger); ok {
		return l
	}
	return &log
}

// Ctx is a shorthand for FromContext.
// Usage: logger.Ctx(ctx).Info().Msg("message")
func Ctx(ctx context.Context) *zerolog.Logger {
	return FromContext(ctx)
}

// WithRequestID adds request_id to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestIDFromContext extracts request_id from the context.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

// WithUserID adds user_id to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserIDFromContext extracts user_id from the context.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey{}).(string); ok {
		return v
	}
	return ""
}

// WithEnv adds environment to the context.
func WithEnv(ctx context.Context, environment config.Environment) context.Context {
	return context.WithValue(ctx, envKey{}, environment)
}

// EnvFromContext extracts environment from the context.
func EnvFromContext(ctx context.Context) config.Environment {
	if v, ok := ctx.Value(envKey{}).(config.Environment); ok {
		return v
	}
	return env
}
