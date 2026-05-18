package queue

import (
	"context"
)

type JobPayload struct {
	JobID string `json:"job_id"`
}

type Queue interface {
	Publish(ctx context.Context, payload JobPayload) error
	Subscribe(ctx context.Context, handler func(ctx context.Context, payload JobPayload) error) error
	Close() error
}
