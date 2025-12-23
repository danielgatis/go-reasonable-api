package middlewares

import (
	"go-reasonable-api/support/http/reqctx"
	"go-reasonable-api/support/logger"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const HeaderXRequestID = "X-Request-ID"

// RequestIDMiddleware generates a unique request ID for each request,
// stores it in the context, sets it in the response header, and creates
// a request-scoped logger with request_id included.
// The logger is stored both in echo.Context and in the request's context.Context,
// allowing services to access it via logger.Ctx(ctx).
func RequestIDMiddleware(baseLogger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check if request ID was provided in header, otherwise generate one
			requestID := c.Request().Header.Get(HeaderXRequestID)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Store request ID in echo context
			reqctx.SetRequestID(c, requestID)

			// Set request ID in response header
			c.Response().Header().Set(HeaderXRequestID, requestID)

			// Create a request-scoped logger with request_id
			reqLogger := baseLogger.With().Str("request_id", requestID).Logger()
			reqctx.SetLogger(c, &reqLogger)

			// Inject both the logger and request_id into the request's context.Context
			// This allows services to access the logger via logger.Ctx(ctx)
			// and extract request_id via logger.RequestIDFromContext(ctx)
			ctx := c.Request().Context()
			ctx = logger.WithContext(ctx, &reqLogger)
			ctx = logger.WithRequestID(ctx, requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
