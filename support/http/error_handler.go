package http

import (
	"net/http"

	"go-reasonable-api/support/errors"
	"go-reasonable-api/support/http/reqctx"
	"go-reasonable-api/support/sentry"

	"github.com/labstack/echo/v4"
)

// ErrorResponse is the JSON structure returned for all errors.
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details"`
}

// ErrorHandler is Echo's custom error handler. It transforms errors into
// consistent JSON responses and reports 5xx errors to Sentry.
func ErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	var statusCode int
	var response ErrorResponse

	if ae, ok := errors.Is(err); ok {
		statusCode = ae.StatusCode
		response = ErrorResponse{
			Code:    ae.Code,
			Message: ae.Message,
			Details: ae.Details,
		}
	} else if he, ok := err.(*echo.HTTPError); ok {
		statusCode = he.Code
		msg := "internal server error"
		if m, ok := he.Message.(string); ok {
			msg = m
		}
		response = ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: msg,
		}
	} else {
		statusCode = http.StatusInternalServerError
		response = ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		}
	}

	// Capture 5xx errors to Sentry
	if statusCode >= 500 {
		extras := map[string]any{
			"request_id": reqctx.GetRequestID(c),
			"method":     c.Request().Method,
			"path":       c.Path(),
			"uri":        c.Request().RequestURI,
		}
		sentry.CaptureError(err, extras)
	}

	_ = c.JSON(statusCode, response)
}
