-- Create extraction_jobs table
CREATE TABLE IF NOT EXISTS extraction_jobs (
    id UUID PRIMARY KEY,
    local_image_path TEXT NOT NULL,
    image_hash TEXT,
    provider TEXT NOT NULL,
    status TEXT NOT NULL,
    retry_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create extraction_results table
CREATE TABLE IF NOT EXISTS extraction_results (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES extraction_jobs(id) ON DELETE CASCADE,
    extracted_text TEXT,
    structured_json JSONB,
    confidence_score DECIMAL(5,2),
    provider_used TEXT NOT NULL,
    processing_duration_ms BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create extraction_flags table
CREATE TABLE IF NOT EXISTS extraction_flags (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES extraction_jobs(id) ON DELETE CASCADE,
    flag_type TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create extraction_errors table
CREATE TABLE IF NOT EXISTS extraction_errors (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES extraction_jobs(id) ON DELETE CASCADE,
    error_message TEXT NOT NULL,
    retry_count INT DEFAULT 0,
    provider TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_extraction_jobs_status ON extraction_jobs(status);
CREATE INDEX IF NOT EXISTS idx_extraction_jobs_image_hash ON extraction_jobs(image_hash);
CREATE INDEX IF NOT EXISTS idx_extraction_results_job_id ON extraction_results(job_id);
CREATE INDEX IF NOT EXISTS idx_extraction_flags_job_id ON extraction_flags(job_id);
CREATE INDEX IF NOT EXISTS idx_extraction_errors_job_id ON extraction_errors(job_id);
