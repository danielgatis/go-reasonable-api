package api

import (
	_ "go-reasonable-api/api/docs"
	"go-reasonable-api/api/handlers"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/support/http/middlewares"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupRoutes configures all API routes on the Echo instance.
// Routes are organized by resource with consistent middleware application.
//
// @title ZapAgenda API
// @version 1.0
// @description API para gerenciamento de agendamentos
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication

func SetupRoutes(
	e *echo.Echo,
	sessionService services.SessionService,
	userHandler *handlers.UserHandler,
	sessionHandler *handlers.SessionHandler,
	passwordResetHandler *handlers.PasswordResetHandler,
	emailVerificationHandler *handlers.EmailVerificationHandler,
	healthHandler *handlers.HealthHandler,
) {
	e.GET("/health", healthHandler.Health)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	authMiddleware := middlewares.AuthMiddleware(sessionService)
	optionalAuthMiddleware := middlewares.OptionalAuthMiddleware(sessionService)

	// Users
	e.POST("/users", userHandler.Create)
	e.GET("/users/me", userHandler.Me, authMiddleware)
	e.DELETE("/users/me", userHandler.Delete, authMiddleware)

	// Sessions
	e.POST("/sessions", sessionHandler.Create)
	e.DELETE("/sessions/current", sessionHandler.DeleteCurrent, authMiddleware)

	// Password Resets
	e.POST("/password-resets", passwordResetHandler.Create)
	e.PUT("/password-resets/:token", passwordResetHandler.Update)

	// Email Verifications
	e.POST("/email-verifications", emailVerificationHandler.Create, optionalAuthMiddleware)
	e.PUT("/email-verifications/:token", emailVerificationHandler.Update)
}
