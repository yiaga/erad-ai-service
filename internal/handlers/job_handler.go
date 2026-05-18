package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/yiaga/erad-ai-service/internal/models"
	"github.com/yiaga/erad-ai-service/internal/queue"
	"github.com/yiaga/erad-ai-service/internal/repositories"
)

type JobHandler struct {
	repo repositories.JobRepository
	q    queue.Queue
}

func NewJobHandler(repo repositories.JobRepository, q queue.Queue) *JobHandler {
	return &JobHandler{repo: repo, q: q}
}

type CreateJobRequest struct {
	LocalImagePath string `json:"local_image_path"`
	Provider       string `json:"provider"`
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	job := &models.ExtractionJob{
		ID:             uuid.New().String(),
		LocalImagePath: req.LocalImagePath,
		Provider:       req.Provider,
		Status:         models.StatusPending,
		RetryCount:     0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.repo.CreateJob(r.Context(), job); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Publish to queue
	if err := h.q.Publish(r.Context(), queue.JobPayload{JobID: job.ID}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.repo.GetJobByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

func (h *JobHandler) GetJobResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.repo.GetResultByJobID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		http.Error(w, "Result not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *JobHandler) GetJobFlags(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	flags, err := h.repo.GetFlagsByJobID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flags)
}

func (h *JobHandler) RetryJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.repo.GetJobByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Reset job and re-publish
	h.repo.UpdateJobStatus(r.Context(), job.ID, models.StatusPending, 0)
	h.q.Publish(r.Context(), queue.JobPayload{JobID: job.ID})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Job retried"})
}

func (h *JobHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
