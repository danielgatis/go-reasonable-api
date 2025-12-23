package http

import (
	routes "go-reasonable-api/api"
	"go-reasonable-api/api/handlers"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/support/config"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Router wraps Echo with application-specific configuration.
// Created via Wire with all handlers and services injected.
type Router struct {
	echo                     *echo.Echo
	config                   *config.Config
	logger                   *zerolog.Logger
	userHandler              *handlers.UserHandler
	sessionHandler           *handlers.SessionHandler
	passwordResetHandler     *handlers.PasswordResetHandler
	emailVerificationHandler *handlers.EmailVerificationHandler
	healthHandler            *handlers.HealthHandler
	sessionService           services.SessionService
}

func NewRouter(
	cfg *config.Config,
	logger *zerolog.Logger,
	userHandler *handlers.UserHandler,
	sessionHandler *handlers.SessionHandler,
	passwordResetHandler *handlers.PasswordResetHandler,
	emailVerificationHandler *handlers.EmailVerificationHandler,
	healthHandler *handlers.HealthHandler,
	sessionService services.SessionService,
) *Router {
	return &Router{
		echo:                     echo.New(),
		config:                   cfg,
		logger:                   logger,
		userHandler:              userHandler,
		sessionHandler:           sessionHandler,
		passwordResetHandler:     passwordResetHandler,
		emailVerificationHandler: emailVerificationHandler,
		healthHandler:            healthHandler,
		sessionService:           sessionService,
	}
}

func (r *Router) Port() string {
	return r.config.Server.Port
}

// Setup configures middleware, routes, and returns the Echo instance.
// Call this once before starting the server.
func (r *Router) Setup() *echo.Echo {
	r.echo.HideBanner = true
	r.echo.Validator = NewValidator()
	r.echo.HTTPErrorHandler = ErrorHandler
	r.setupMiddlewares()
	routes.SetupRoutes(
		r.echo,
		r.sessionService,
		r.userHandler,
		r.sessionHandler,
		r.passwordResetHandler,
		r.emailVerificationHandler,
		r.healthHandler,
	)
	return r.echo
}
