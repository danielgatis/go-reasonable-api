package http

import (
	"go-reasonable-api/support/http/middlewares"

	"github.com/labstack/echo/v5/middleware"
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
	}))
	r.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     r.config.Server.CORS.AllowOrigins,
		AllowMethods:     r.config.Server.CORS.AllowMethods,
		AllowHeaders:     r.config.Server.CORS.AllowHeaders,
		AllowCredentials: r.config.Server.CORS.AllowCredentials,
		MaxAge:           r.config.Server.CORS.MaxAge,
	}))
	r.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(float64(r.config.Server.RateLimitPerSecond))))
	r.echo.Use(middlewares.SecurityHeaders())
}
