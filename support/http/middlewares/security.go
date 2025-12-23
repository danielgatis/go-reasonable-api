package middlewares

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds common security headers to all responses.
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()

			// Prevent MIME type sniffing
			h.Set("X-Content-Type-Options", "nosniff")

			// Prevent clickjacking
			h.Set("X-Frame-Options", "DENY")

			// Enable XSS filtering (legacy, but still useful for older browsers)
			h.Set("X-XSS-Protection", "1; mode=block")

			// Disable caching of sensitive responses by default
			h.Set("Cache-Control", "no-store")

			// Referrer policy - don't leak referrer info
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Content Security Policy - basic restrictive policy
			h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

			return next(c)
		}
	}
}

// StrictTransportSecurity adds HSTS header for HTTPS enforcement.
// Only use this when serving over HTTPS.
func StrictTransportSecurity(maxAge int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if maxAge > 0 {
				c.Response().Header().Set(
					"Strict-Transport-Security",
					fmt.Sprintf("max-age=%d; includeSubDomains", maxAge),
				)
			}
			return next(c)
		}
	}
}
