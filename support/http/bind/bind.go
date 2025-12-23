package bind

import (
	"fmt"
	"strings"

	"go-reasonable-api/support/errors"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// AndValidate binds the request body to the given struct and validates it.
// Returns an appropriate error if binding or validation fails.
// Binding errors return a wrapped error for debugging while the error handler
// will return a generic message to the user.
func AndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return errors.Wrap(err, "failed to bind request")
	}

	if err := c.Validate(req); err != nil {
		return eris.Wrap(err, "failed to validate request")
	}

	return nil
}

// RequiredParam extracts a path parameter and returns an error if it's empty.
func RequiredParam(c echo.Context, name string) (string, error) {
	value := c.Param(name)
	if value == "" {
		code := fmt.Sprintf("MISSING_%s", strings.ToUpper(name))
		msg := fmt.Sprintf("%s is required", name)
		return "", errors.BadRequest(code, msg)
	}
	return value, nil
}
