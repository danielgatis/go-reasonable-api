package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// RedisPinger is an interface for checking Redis connectivity.
// This is implemented by *asynq.Client and allows for mocking in tests.
type RedisPinger interface {
	Ping() error
}

// HealthHandler provides liveness and readiness probes.
type HealthHandler struct {
	db          *sql.DB
	redisPinger RedisPinger
}

func NewHealthHandler(db *sql.DB, redisPinger RedisPinger) *HealthHandler {
	return &HealthHandler{
		db:          db,
		redisPinger: redisPinger,
	}
}

type HealthResponse struct {
	Status       string                  `json:"status"`
	Dependencies map[string]HealthStatus `json:"dependencies"`
}

type HealthStatus struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Health returns the health status of the API and its dependencies
// @Summary Health check
// @Description Check the health status of the API and its dependencies (database, redis)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:       "healthy",
		Dependencies: make(map[string]HealthStatus),
	}

	// Check database
	dbStatus := h.checkDatabase(ctx)
	response.Dependencies["database"] = dbStatus
	if dbStatus.Status != "ok" {
		response.Status = "unhealthy"
	}

	// Check Redis (via Asynq)
	redisStatus := h.checkRedis(ctx)
	response.Dependencies["redis"] = redisStatus
	if redisStatus.Status != "ok" {
		response.Status = "unhealthy"
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, response)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) HealthStatus {
	start := time.Now()

	if err := h.db.PingContext(ctx); err != nil {
		return HealthStatus{
			Status: "error",
			Error:  err.Error(),
		}
	}

	return HealthStatus{
		Status:  "ok",
		Latency: time.Since(start).String(),
	}
}

func (h *HealthHandler) checkRedis(ctx context.Context) HealthStatus {
	start := time.Now()

	// Run ping with context timeout since asynq.Client.Ping() doesn't accept context
	errCh := make(chan error, 1)
	go func() {
		errCh <- h.redisPinger.Ping()
	}()

	select {
	case <-ctx.Done():
		return HealthStatus{
			Status: "error",
			Error:  "redis health check timed out",
		}
	case err := <-errCh:
		if err != nil {
			return HealthStatus{
				Status: "error",
				Error:  err.Error(),
			}
		}
	}

	return HealthStatus{
		Status:  "ok",
		Latency: time.Since(start).String(),
	}
}
