package handlers_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-reasonable-api/api/handlers"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRedisPinger is a mock implementation of handlers.RedisPinger
type mockRedisPinger struct {
	err error
}

func (m *mockRedisPinger) Ping() error {
	return m.err
}

func TestHealthHandler_Health(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(sqlmock.Sqlmock)
		redisError     error
		expectedStatus int
		expectedHealth string
	}{
		{
			name: "returns healthy when all dependencies are up",
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			redisError:     nil,
			expectedStatus: http.StatusOK,
			expectedHealth: "healthy",
		},
		{
			name: "returns unhealthy when database is down",
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			redisError:     nil,
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
		{
			name: "returns unhealthy when redis is down",
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			redisError:     eris.New("redis connection refused"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
		{
			name: "returns unhealthy when both are down",
			setupDB: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(sql.ErrConnDone)
			},
			redisError:     eris.New("redis connection refused"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := echo.New()
			db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			require.NoError(t, err)
			defer func() { _ = db.Close() }()

			tt.setupDB(mock)

			redisPinger := &mockRedisPinger{err: tt.redisError}
			handler := handlers.NewHealthHandler(db, redisPinger)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Execute
			err = handler.Health(c)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedHealth)

			// Verify all DB expectations were met
			err = mock.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
