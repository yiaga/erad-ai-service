package queue

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/yiaga/erad-ai-service/internal/models"
)

// PostgresQueue is a simple database-backed queue for testing purposes.
type PostgresQueue struct {
	db *sqlx.DB
}

func NewPostgresQueue(db *sqlx.DB) *PostgresQueue {
	return &PostgresQueue{db: db}
}

func (q *PostgresQueue) Publish(ctx context.Context, payload JobPayload) error {
	// In this implementation, the job is already persisted in the DB
	// by the time Publish is called in the handler.
	return nil
}

func (q *PostgresQueue) Subscribe(ctx context.Context, handler func(ctx context.Context, payload JobPayload) error) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			var jobs []models.ExtractionJob
			// Atomically pick up a batch of pending jobs
			err := q.db.SelectContext(ctx, &jobs, 
				"SELECT id FROM extraction_jobs WHERE status = $1 LIMIT 5", 
				models.StatusPending)
			if err != nil {
				continue
			}

			for _, job := range jobs {
				// Try to "claim" the job by moving it to StatusValidating
				// This prevents multiple workers from picking up the same job
				res, err := q.db.ExecContext(ctx, 
					"UPDATE extraction_jobs SET status = $1, updated_at = NOW() WHERE id = $2 AND status = $3", 
					models.StatusValidating, job.ID, models.StatusPending)
				
				if err != nil {
					continue
				}

				rows, _ := res.RowsAffected()
				if rows == 0 {
					continue // Already claimed by another worker
				}

				// Trigger the handler
				_ = handler(ctx, JobPayload{JobID: job.ID})
			}
		}
	}
}

func (q *PostgresQueue) Close() error {
	return nil
}
