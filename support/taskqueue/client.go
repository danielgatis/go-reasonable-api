package taskqueue

import (
	"context"

	"go-reasonable-api/support/http/reqctx"
	"go-reasonable-api/support/logger"
	"go-reasonable-api/support/sentry"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// Client wraps asynq.Client to automatically inject metadata into tasks.
type Client struct {
	asynq *asynq.Client
}

// NewClient creates a new task Client.
func NewClient(asynqClient *asynq.Client) *Client {
	return &Client{asynq: asynqClient}
}

// Enqueue creates and enqueues a task with metadata extracted from echo.Context.
// Use this when you have an echo.Context (in handlers).
func (c *Client) Enqueue(ec echo.Context, taskType string, payload any, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	requestID := reqctx.GetRequestID(ec)
	userID := ""
	if uid, ok := reqctx.GetUserID(ec); ok {
		userID = uid.String()
	}

	return c.EnqueueWithMeta(taskType, payload, requestID, userID, opts...)
}

// EnqueueCtx creates and enqueues a task with metadata extracted from context.Context.
// Use this when you only have a context.Context (in services).
// request_id and user_id are automatically extracted from the context.
// Errors are sent to Sentry but not returned - fire-and-forget.
func (c *Client) EnqueueCtx(ctx context.Context, taskType string, payload any, opts ...asynq.Option) {
	requestID := logger.RequestIDFromContext(ctx)
	userID := logger.UserIDFromContext(ctx)
	_, err := c.EnqueueWithMeta(taskType, payload, requestID, userID, opts...)
	if err != nil {
		sentry.CaptureError(err, map[string]any{
			"task_type":  taskType,
			"request_id": requestID,
			"user_id":    userID,
		})
	}
}

// EnqueueWithMeta creates and enqueues a task with explicit metadata values.
func (c *Client) EnqueueWithMeta(taskType string, payload any, requestID, userID string, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	meta := NewTaskMetadataWithValues(requestID, userID)

	data, err := WrapPayload(meta, payload)
	if err != nil {
		return nil, eris.Wrap(err, "failed to wrap payload")
	}

	task := asynq.NewTask(taskType, data)
	info, err := c.asynq.Enqueue(task, opts...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to enqueue task")
	}

	return info, nil
}

// EnqueueRaw enqueues a raw asynq.Task without metadata wrapping.
// Use this for backward compatibility or when you don't need tracing.
func (c *Client) EnqueueRaw(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	info, err := c.asynq.Enqueue(task, opts...)
	if err != nil {
		return nil, eris.Wrap(err, "failed to enqueue raw task")
	}
	return info, nil
}

// Close closes the underlying asynq client.
func (c *Client) Close() error {
	if err := c.asynq.Close(); err != nil {
		return eris.Wrap(err, "failed to close asynq client")
	}
	return nil
}
