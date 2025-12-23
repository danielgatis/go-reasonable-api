package taskqueue

import (
	"context"
	"encoding/json"

	"go-reasonable-api/support/logger"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

// TaskMetadata contains tracing information for background jobs.
// JobID is always generated, RequestID and UserID are optional and
// carried over from the original HTTP request when available.
type TaskMetadata struct {
	JobID     string `json:"job_id"`
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

// TaskEnvelope wraps any task payload with metadata for tracing.
type TaskEnvelope struct {
	Metadata TaskMetadata    `json:"metadata"`
	Payload  json.RawMessage `json:"payload"`
}

// NewTaskMetadataWithValues creates a TaskMetadata with explicit values.
func NewTaskMetadataWithValues(requestID, userID string) TaskMetadata {
	return TaskMetadata{
		JobID:     uuid.New().String(),
		RequestID: requestID,
		UserID:    userID,
	}
}

// Logger creates a zerolog.Logger enriched with the task metadata.
func (m TaskMetadata) Logger(base *zerolog.Logger) *zerolog.Logger {
	l := base.With().Str("job_id", m.JobID)

	if m.RequestID != "" {
		l = l.Str("request_id", m.RequestID)
	}
	if m.UserID != "" {
		l = l.Str("user_id", m.UserID)
	}

	result := l.Logger()
	return &result
}

// LoggerContext creates a context with an enriched logger.
func (m TaskMetadata) LoggerContext(ctx context.Context, base *zerolog.Logger) context.Context {
	return logger.WithContext(ctx, m.Logger(base))
}

// WrapPayload wraps a payload with metadata into a TaskEnvelope.
func WrapPayload(metadata TaskMetadata, payload any) ([]byte, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, eris.Wrap(err, "failed to marshal payload")
	}

	envelope := TaskEnvelope{
		Metadata: metadata,
		Payload:  payloadBytes,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, eris.Wrap(err, "failed to marshal envelope")
	}

	return data, nil
}

// UnwrapPayload extracts the metadata and unmarshals the payload from a TaskEnvelope.
func UnwrapPayload(data []byte, payload any) (TaskMetadata, error) {
	var envelope TaskEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return TaskMetadata{}, eris.Wrap(err, "failed to unmarshal envelope")
	}

	if err := json.Unmarshal(envelope.Payload, payload); err != nil {
		return envelope.Metadata, eris.Wrap(err, "failed to unmarshal payload")
	}

	return envelope.Metadata, nil
}
