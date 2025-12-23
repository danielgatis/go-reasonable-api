package http

import (
	"fmt"

	"go-reasonable-api/support/http/middlewares"
	"go-reasonable-api/support/http/reqctx"
	"go-reasonable-api/support/logger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

func (r *Router) setupMiddlewares() {
	// RequestIDMiddleware must come first to generate request_id and set up the request-scoped logger
	r.echo.Use(middlewares.RequestIDMiddleware(r.logger))
	r.echo.Use(middlewares.LoggerMiddleware(r.logger))

	// Sentry middleware for context setup and panic recovery
	if r.config.Sentry.DSN != "" {
		r.echo.Use(middlewares.SentryMiddleware())
		if r.config.Sentry.EnableTracing {
			r.echo.Use(middlewares.SentryTracingMiddleware())
		}
	}

	r.echo.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll: true,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			l := reqctx.Logger(c)
			if l == nil {
				l = r.logger
			}
			l.Error().Err(err).Msg("panic recovered")
			// Print readable stack trace in DEV
			if logger.IsDev() && len(stack) > 0 {
				fmt.Printf("\n%s\n", stack)
			}
			return err
		},
	}))
	r.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     r.config.Server.CORS.AllowOrigins,
		AllowMethods:     r.config.Server.CORS.AllowMethods,
		AllowHeaders:     r.config.Server.CORS.AllowHeaders,
		AllowCredentials: r.config.Server.CORS.AllowCredentials,
		MaxAge:           r.config.Server.CORS.MaxAge,
	}))
	r.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(r.config.Server.RateLimitPerSecond))))
	r.echo.Use(middlewares.SecurityHeaders())
}
