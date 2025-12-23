package reqctx

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const (
	contextKeyUserID    = "user_id"
	contextKeyToken     = "token"
	contextKeyRequestID = "request_id"
	contextKeyLogger    = "logger"
)

func SetUserID(c echo.Context, userID uuid.UUID) {
	c.Set(contextKeyUserID, userID)
}

func GetUserID(c echo.Context) (uuid.UUID, bool) {
	userID, ok := c.Get(contextKeyUserID).(uuid.UUID)
	return userID, ok
}

func SetToken(c echo.Context, token string) {
	c.Set(contextKeyToken, token)
}

func GetToken(c echo.Context) (string, bool) {
	token, ok := c.Get(contextKeyToken).(string)
	return token, ok
}

func SetRequestID(c echo.Context, requestID string) {
	c.Set(contextKeyRequestID, requestID)
}

func GetRequestID(c echo.Context) string {
	requestID, _ := c.Get(contextKeyRequestID).(string)
	return requestID
}

func SetLogger(c echo.Context, logger *zerolog.Logger) {
	c.Set(contextKeyLogger, logger)
}

// Logger returns the request-scoped logger with request_id and user_id already set.
// If no logger is set in the context, returns nil.
func Logger(c echo.Context) *zerolog.Logger {
	logger, _ := c.Get(contextKeyLogger).(*zerolog.Logger)
	return logger
}
