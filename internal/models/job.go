package models

import (
	"encoding/json"
	"time"
)

type JobStatus string

const (
	StatusPending        JobStatus = "pending"
	StatusValidating     JobStatus = "validating_image"
	StatusQueued         JobStatus = "queued"
	StatusProcessing     JobStatus = "processing"
	StatusRetrying       JobStatus = "retrying"
	StatusCompleted      JobStatus = "completed"
	StatusFailed         JobStatus = "failed"
)

type ExtractionJob struct {
	ID             string    `db:"id" json:"id"`
	LocalImagePath string    `db:"local_image_path" json:"local_image_path"`
	ImageHash      string    `db:"image_hash" json:"image_hash"`
	Provider       string    `db:"provider" json:"provider"`
	Status         JobStatus `db:"status" json:"status"`
	RetryCount     int       `db:"retry_count" json:"retry_count"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type ExtractionResult struct {
	ID                    string          `db:"id" json:"id"`
	JobID                 string          `db:"job_id" json:"job_id"`
	ExtractedText         string          `db:"extracted_text" json:"extracted_text"`
	StructuredJSON        json.RawMessage `db:"structured_json" json:"structured_json"`
	ConfidenceScore       float64         `db:"confidence_score" json:"confidence_score"`
	ProviderUsed          string          `db:"provider_used" json:"provider_used"`
	ProcessingDurationMS  int64           `db:"processing_duration_ms" json:"processing_duration_ms"`
	CreatedAt             time.Time       `db:"created_at" json:"created_at"`
}

type ExtractionFlag struct {
	ID          string    `db:"id" json:"id"`
	JobID       string    `db:"job_id" json:"job_id"`
	FlagType    string    `db:"flag_type" json:"flag_type"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type ExtractionError struct {
	ID           string    `db:"id" json:"id"`
	JobID        string    `db:"job_id" json:"job_id"`
	ErrorMessage string    `db:"error_message" json:"error_message"`
	RetryCount   int       `db:"retry_count" json:"retry_count"`
	Provider     string    `db:"provider" json:"provider"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}
