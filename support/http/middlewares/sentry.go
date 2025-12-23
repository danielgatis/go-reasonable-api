package middlewares

import (
	"net/http"

	"go-reasonable-api/support/http/reqctx"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v4"
)

// SentryMiddleware sets up Sentry context for each request and captures panics.
// Error capturing is handled by the error handler.
func SentryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			hub := sentry.GetHubFromContext(ctx)
			if hub == nil {
				hub = sentry.CurrentHub().Clone()
				ctx = sentry.SetHubOnContext(ctx, hub)
				c.SetRequest(c.Request().WithContext(ctx))
			}

			hub.Scope().SetRequest(c.Request())
			hub.Scope().SetTag("request_id", reqctx.GetRequestID(c))

			defer func() {
				if r := recover(); r != nil {
					hub.RecoverWithContext(ctx, r)
					c.Error(echo.NewHTTPError(http.StatusInternalServerError, "internal server error"))
				}
			}()

			return next(c)
		}
	}
}

// SentryTracingMiddleware adds performance tracing to requests.
// Only use this if tracing is enabled in Sentry config.
func SentryTracingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()

			options := []sentry.SpanOption{
				sentry.WithOpName("http.server"),
				sentry.ContinueFromRequest(c.Request()),
				sentry.WithTransactionSource(sentry.SourceURL),
			}

			transactionName := c.Request().Method + " " + c.Path()
			span := sentry.StartSpan(ctx, "http.server", options...)
			span.Name = transactionName
			span.SetTag("http.method", c.Request().Method)
			span.SetTag("http.url", c.Request().URL.String())

			defer func() {
				span.Status = httpStatusToSpanStatus(c.Response().Status)
				span.Finish()
			}()

			c.SetRequest(c.Request().WithContext(span.Context()))
			return next(c)
		}
	}
}

func httpStatusToSpanStatus(status int) sentry.SpanStatus {
	switch {
	case status >= 200 && status < 300:
		return sentry.SpanStatusOK
	case status == 400:
		return sentry.SpanStatusInvalidArgument
	case status == 401:
		return sentry.SpanStatusUnauthenticated
	case status == 403:
		return sentry.SpanStatusPermissionDenied
	case status == 404:
		return sentry.SpanStatusNotFound
	case status == 409:
		return sentry.SpanStatusAlreadyExists
	case status == 429:
		return sentry.SpanStatusResourceExhausted
	case status >= 500:
		return sentry.SpanStatusInternalError
	default:
		return sentry.SpanStatusUnknown
	}
}
