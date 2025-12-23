package middlewares

import (
	"fmt"
	"os"
	"time"

	"go-reasonable-api/support/errors"
	"go-reasonable-api/support/http/reqctx"
	"go-reasonable-api/support/logger"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

func LoggerMiddleware(fallbackLogger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			l := reqctx.Logger(c)
			if l == nil {
				l = fallbackLogger
			}

			// Choose log level based on status
			var evt *zerolog.Event
			status := res.Status
			if err != nil {
				// For errors, get status from error if available
				if ae, ok := errors.Is(err); ok {
					status = ae.StatusCode
				} else if he, ok := err.(*echo.HTTPError); ok {
					status = he.Code
				} else {
					status = 500
				}
			}

			switch {
			case status >= 500:
				evt = l.Error()
			case status >= 400:
				evt = l.Warn()
			default:
				evt = l.Info()
			}

			if err != nil {
				evt = evt.Err(err)
			}

			evt.
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("remote_ip", c.RealIP()).
				Int("status", status).
				Dur("latency", latency).
				Str("user_agent", req.UserAgent()).
				Msg("request")

			// In DEV, print readable stacktrace for 500+ errors
			if err != nil && status >= 500 && logger.IsDev() {
				if st := errors.StackTrace(err); st != "" {
					fmt.Fprintln(os.Stderr, st)
				}
			}

			return err
		}
	}
}
