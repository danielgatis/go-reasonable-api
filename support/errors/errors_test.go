package errors_test

import (
	"net/http"
	"testing"

	apperrors "go-reasonable-api/support/errors"

	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
)

func TestIs_unwrapsErisWrappedAppError(t *testing.T) {
	t.Parallel()

	sentinel := apperrors.NotFound("USER_NOT_FOUND", "user not found")
	wrapped := eris.Wrap(sentinel, "failed to get user")

	ae, ok := apperrors.Is(wrapped)

	assert.True(t, ok, "Is must unwrap through eris.Wrap")
	assert.Equal(t, sentinel.Code, ae.Code)
	assert.Equal(t, http.StatusNotFound, ae.StatusCode)
}

func TestIs_returnsFalseForPlainError(t *testing.T) {
	t.Parallel()

	ae, ok := apperrors.Is(eris.New("boom"))

	assert.False(t, ok)
	assert.Nil(t, ae)
}

func TestIs_returnsFalseForNil(t *testing.T) {
	t.Parallel()

	ae, ok := apperrors.Is(nil)

	assert.False(t, ok)
	assert.Nil(t, ae)
}
