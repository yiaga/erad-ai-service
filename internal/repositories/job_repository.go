package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/yiaga/erad-ai-service/internal/models"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *models.ExtractionJob) error
	GetJobByID(ctx context.Context, id string) (*models.ExtractionJob, error)
	UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, retryCount int) error
	UpdateJobHash(ctx context.Context, id string, hash string) error
	GetJobByHash(ctx context.Context, hash string) (*models.ExtractionJob, error)
	
	SaveResult(ctx context.Context, result *models.ExtractionResult) error
	GetResultByJobID(ctx context.Context, jobID string) (*models.ExtractionResult, error)
	
	AddFlag(ctx context.Context, flag *models.ExtractionFlag) error
	GetFlagsByJobID(ctx context.Context, jobID string) ([]models.ExtractionFlag, error)
	
	SaveError(ctx context.Context, errRecord *models.ExtractionError) error
}

type postgresJobRepository struct {
	db *sqlx.DB
}

func NewPostgresJobRepository(db *sqlx.DB) JobRepository {
	return &postgresJobRepository{db: db}
}

func (r *postgresJobRepository) CreateJob(ctx context.Context, job *models.ExtractionJob) error {
	query := `INSERT INTO extraction_jobs (id, local_image_path, image_hash, provider, status, retry_count, created_at, updated_at)
			  VALUES (:id, :local_image_path, :image_hash, :provider, :status, :retry_count, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, job)
	return err
}

func (r *postgresJobRepository) GetJobByID(ctx context.Context, id string) (*models.ExtractionJob, error) {
	job := &models.ExtractionJob{}
	err := r.db.GetContext(ctx, job, "SELECT * FROM extraction_jobs WHERE id = $1", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return job, err
}

func (r *postgresJobRepository) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, retryCount int) error {
	query := `UPDATE extraction_jobs SET status = $1, retry_count = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, retryCount, id)
	return err
}

func (r *postgresJobRepository) UpdateJobHash(ctx context.Context, id string, hash string) error {
	query := `UPDATE extraction_jobs SET image_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, hash, id)
	return err
}

func (r *postgresJobRepository) GetJobByHash(ctx context.Context, hash string) (*models.ExtractionJob, error) {
	job := &models.ExtractionJob{}
	err := r.db.GetContext(ctx, job, "SELECT * FROM extraction_jobs WHERE image_hash = $1 LIMIT 1", hash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return job, err
}

func (r *postgresJobRepository) SaveResult(ctx context.Context, result *models.ExtractionResult) error {
	query := `INSERT INTO extraction_results (id, job_id, extracted_text, structured_json, confidence_score, provider_used, processing_duration_ms, created_at)
			  VALUES (:id, :job_id, :extracted_text, :structured_json, :confidence_score, :provider_used, :processing_duration_ms, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, result)
	return err
}

func (r *postgresJobRepository) GetResultByJobID(ctx context.Context, jobID string) (*models.ExtractionResult, error) {
	result := &models.ExtractionResult{}
	err := r.db.GetContext(ctx, result, "SELECT * FROM extraction_results WHERE job_id = $1", jobID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return result, err
}

func (r *postgresJobRepository) AddFlag(ctx context.Context, flag *models.ExtractionFlag) error {
	query := `INSERT INTO extraction_flags (id, job_id, flag_type, description, created_at)
			  VALUES (:id, :job_id, :flag_type, :description, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, flag)
	return err
}

func (r *postgresJobRepository) GetFlagsByJobID(ctx context.Context, jobID string) ([]models.ExtractionFlag, error) {
	var flags []models.ExtractionFlag
	err := r.db.SelectContext(ctx, &flags, "SELECT * FROM extraction_flags WHERE job_id = $1", jobID)
	return flags, err
}

func (r *postgresJobRepository) SaveError(ctx context.Context, errRecord *models.ExtractionError) error {
	query := `INSERT INTO extraction_errors (id, job_id, error_message, retry_count, provider, created_at)
			  VALUES (:id, :job_id, :error_message, :retry_count, :provider, :created_at)`
	_, err := r.db.NamedExecContext(ctx, query, errRecord)
	return err
}
