package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-reasonable-api/api/handlers"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"github.com/pashagolub/pgxmock/v4"
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
		setupDB        func(pgxmock.PgxPoolIface)
		redisError     error
		expectedStatus int
		expectedHealth string
	}{
		{
			name: "returns healthy when all dependencies are up",
			setupDB: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectPing()
			},
			redisError:     nil,
			expectedStatus: http.StatusOK,
			expectedHealth: "healthy",
		},
		{
			name: "returns unhealthy when database is down",
			setupDB: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectPing().WillReturnError(pgx.ErrTxClosed)
			},
			redisError:     nil,
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
		{
			name: "returns unhealthy when redis is down",
			setupDB: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectPing()
			},
			redisError:     eris.New("redis connection refused"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
		{
			name: "returns unhealthy when both are down",
			setupDB: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectPing().WillReturnError(pgx.ErrTxClosed)
			},
			redisError:     eris.New("redis connection refused"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupDB(mockPool)

			redisPinger := &mockRedisPinger{err: tt.redisError}
			handler := handlers.NewHealthHandler(mockPool, redisPinger)

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = handler.Health(c)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedHealth)

			err = mockPool.ExpectationsWereMet()
			assert.NoError(t, err)
		})
	}
}
