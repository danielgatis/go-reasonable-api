package middlewares

import (
	"strings"

	"go-reasonable-api/app/errors"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/support/http/reqctx"
	"go-reasonable-api/support/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(sessionService services.SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return errors.ErrMissingAuthHeader
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return errors.ErrInvalidAuthFormat
			}

			token := parts[1]
			authToken, err := sessionService.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return err
			}

			setAuthContext(c, authToken.UserID, token)

			return next(c)
		}
	}
}

func OptionalAuthMiddleware(sessionService services.SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return next(c)
			}

			token := parts[1]
			authToken, err := sessionService.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return next(c)
			}

			setAuthContext(c, authToken.UserID, token)

			return next(c)
		}
	}
}

// setAuthContext sets the authenticated user's ID and token in the request context.
// It also enriches the logger with the user_id for request tracing.
func setAuthContext(c echo.Context, userID uuid.UUID, token string) {
	reqctx.SetUserID(c, userID)
	reqctx.SetToken(c, token)

	userIDStr := userID.String()
	if reqLogger := reqctx.Logger(c); reqLogger != nil {
		enrichedLogger := reqLogger.With().Str("user_id", userIDStr).Logger()
		reqctx.SetLogger(c, &enrichedLogger)

		ctx := c.Request().Context()
		ctx = logger.WithContext(ctx, &enrichedLogger)
		ctx = logger.WithUserID(ctx, userIDStr)
		c.SetRequest(c.Request().WithContext(ctx))
	}
}
