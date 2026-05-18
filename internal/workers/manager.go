package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/yiaga/erad-ai-service/internal/models"
	"github.com/yiaga/erad-ai-service/internal/providers"
	"github.com/yiaga/erad-ai-service/internal/queue"
	"github.com/yiaga/erad-ai-service/internal/repositories"
	"github.com/yiaga/erad-ai-service/internal/services"
	"github.com/yiaga/erad-ai-service/pkg/utils"
	"go.uber.org/zap"
)

type WorkerManager struct {
	repo      repositories.JobRepository
	q         queue.Queue
	providers map[string]providers.AIProvider
	refiner   *services.ResultRefiner
	logger    *zap.Logger
}

func NewWorkerManager(repo repositories.JobRepository, q queue.Queue, providers map[string]providers.AIProvider, refiner *services.ResultRefiner, logger *zap.Logger) *WorkerManager {
	return &WorkerManager{
		repo:      repo,
		q:         q,
		providers: providers,
		refiner:   refiner,
		logger:    logger,
	}
}

func (m *WorkerManager) Start(ctx context.Context, workerCount int) {
	for i := 0; i < workerCount; i++ {
		go m.worker(ctx, i)
	}
}

func (m *WorkerManager) worker(ctx context.Context, id int) {
	m.logger.Info("Starting worker", zap.Int("worker_id", id))
	err := m.q.Subscribe(ctx, m.handleJob)
	if err != nil {
		m.logger.Error("Queue subscription failed", zap.Error(err))
	}
}

func (m *WorkerManager) handleJob(ctx context.Context, payload queue.JobPayload) error {
	m.logger.Info("Processing job", zap.String("job_id", payload.JobID))
	
	job, err := m.repo.GetJobByID(ctx, payload.JobID)
	if err != nil || job == nil {
		return fmt.Errorf("job not found: %w", err)
	}

	// Update status to processing
	m.repo.UpdateJobStatus(ctx, job.ID, models.StatusProcessing, job.RetryCount)

	// Step 1: Image Validation & Hashing
	if _, err := os.Stat(job.LocalImagePath); os.IsNotExist(err) {
		m.repo.UpdateJobStatus(ctx, job.ID, models.StatusFailed, job.RetryCount)
		m.repo.AddFlag(ctx, &models.ExtractionFlag{
			ID:          uuid.New().String(),
			JobID:       job.ID,
			FlagType:    "file_not_found",
			Description: fmt.Sprintf("Image not found at path: %s", job.LocalImagePath),
			CreatedAt:   time.Now(),
		})
		return nil
	}

	hash, err := utils.GenerateFileHash(job.LocalImagePath)
	if err != nil {
		m.logger.Error("Failed to generate file hash", zap.Error(err))
		return err
	}

	// Duplicate detection
	existingJob, _ := m.repo.GetJobByHash(ctx, hash)
	if existingJob != nil && existingJob.ID != job.ID {
		m.repo.UpdateJobStatus(ctx, job.ID, models.StatusFailed, job.RetryCount)
		m.repo.AddFlag(ctx, &models.ExtractionFlag{
			ID:          uuid.New().String(),
			JobID:       job.ID,
			FlagType:    "duplicate_image",
			Description: fmt.Sprintf("Image hash matches existing job: %s", existingJob.ID),
			CreatedAt:   time.Now(),
		})
		return nil
	}

	// Update job hash
	m.repo.UpdateJobHash(ctx, job.ID, hash)

	// Step 2: Extraction
	provider, ok := m.providers[job.Provider]
	if !ok {
		return fmt.Errorf("provider not found: %s", job.Provider)
	}

	startTime := time.Now()
	result, err := provider.ProcessDocument(ctx, providers.DocumentInput{
		LocalPath: job.LocalImagePath,
	})
	duration := time.Since(startTime)

	if err != nil {
		m.handleFailure(ctx, job, err)
		return nil
	}

	// Step 3: Refine Results
	var electionData *models.ElectionResultData
	if result.ElectionData != nil {
		electionData = result.ElectionData
	} else if m.refiner != nil {
		m.logger.Info("Refining results with LLM", zap.String("job_id", job.ID))
		refined, err := m.refiner.Refine(ctx, result.RawText)
		if err != nil {
			m.logger.Warn("Failed to refine results", zap.Error(err))
			// We continue with raw results if refinement fails
		} else {
			electionData = refined
		}
	}

	// Step 4: Save Results
	var structuredJSON []byte
	if electionData != nil {
		structuredJSON, _ = json.Marshal(electionData)
	} else {
		structuredJSON, _ = json.Marshal(result.StructuredData)
	}
	err = m.repo.SaveResult(ctx, &models.ExtractionResult{
		ID:                   uuid.New().String(),
		JobID:                job.ID,
		ExtractedText:        result.RawText,
		StructuredJSON:       structuredJSON,
		ConfidenceScore:      result.ConfidenceScore,
		ProviderUsed:         result.ProviderName,
		ProcessingDurationMS: duration.Milliseconds(),
		CreatedAt:            time.Now(),
	})

	if err != nil {
		m.logger.Error("Failed to save result to database", zap.Error(err), zap.String("job_id", job.ID))
		m.handleFailure(ctx, job, err)
		return nil
	}

	m.logger.Info("Result saved successfully", zap.String("job_id", job.ID))
	m.repo.UpdateJobStatus(ctx, job.ID, models.StatusCompleted, job.RetryCount)
	m.logger.Info("Job completed successfully", zap.String("job_id", job.ID))
	return nil
}

func (m *WorkerManager) handleFailure(ctx context.Context, job *models.ExtractionJob, err error) {
	m.logger.Warn("Job failed", zap.String("job_id", job.ID), zap.Error(err))
	
	m.repo.SaveError(ctx, &models.ExtractionError{
		ID:           uuid.New().String(),
		JobID:        job.ID,
		ErrorMessage: err.Error(),
		RetryCount:   job.RetryCount,
		Provider:     job.Provider,
		CreatedAt:    time.Now(),
	})

	if job.RetryCount < 3 {
		m.repo.UpdateJobStatus(ctx, job.ID, models.StatusRetrying, job.RetryCount+1)
		// Re-publish to queue with backoff or immediately
		// For simplicity, we republish immediately. In production, use backoff.
		m.q.Publish(ctx, queue.JobPayload{JobID: job.ID})
	} else {
		m.repo.UpdateJobStatus(ctx, job.ID, models.StatusFailed, job.RetryCount)
	}
}
