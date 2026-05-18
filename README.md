# ERAD AI Service

A high-performance, asynchronous AI extraction microservice for processing election result documents at national scale.

---

## Overview

ERAD AI Service acts as a bridge between the ERAD platform and AI providers (Azure AI Document Intelligence, Google Cloud Vision/Document AI). Its sole responsibility is:

1. Accept extraction jobs for locally stored images
2. Validate and hash images (for integrity + duplicate detection)
3. Send images to AI providers for OCR/document extraction
4. Normalise provider output into a unified format
5. Persist results and operational flags to PostgreSQL

> ⚠️ **This service is strictly an extraction engine.** It does NOT validate election data, detect over-voting, or perform any electoral analysis.

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      ERAD Platform                       │
└────────────────────────┬────────────────────────────────┘
                         │  POST /jobs
                         ▼
              ┌──────────────────────┐
              │   erad-ai-api (Chi)   │
              └──────────┬───────────┘
                         │ Publishes JobPayload
                         ▼
              ┌──────────────────────┐
              │       Kafka          │
              └──────────┬───────────┘
                         │ Consumed by N workers
                         ▼
        ┌────────────────────────────────┐
        │         Worker Pool            │
        │  ┌──────────────────────────┐  │
        │  │ 1. Validate image path   │  │
        │  │ 2. SHA-256 hash + dedup  │  │
        │  │ 3. Call AI Provider      │  │
        │  │ 4. Normalise output      │  │
        │  │ 5. Save to PostgreSQL    │  │
        │  └──────────────────────────┘  │
        └────────────────────────────────┘
                         │
              ┌──────────▼───────────┐
              │     PostgreSQL        │
              │  extraction_jobs      │
              │  extraction_results   │
              │  extraction_flags     │
              │  extraction_errors    │
              └──────────────────────┘
```

---

## Project Structure

```
erad-ai-service/
├── cmd/
│   ├── api/            # HTTP API server entry point
│   └── worker/         # Background worker entry point
├── internal/
│   ├── api/            # Chi router setup
│   ├── config/         # Viper config loader
│   ├── handlers/       # HTTP request handlers
│   ├── logging/        # Zap structured logger
│   ├── models/         # Database models & job status enums
│   ├── providers/      # AI provider abstraction + implementations
│   ├── queue/          # Kafka producer/consumer
│   ├── repositories/   # PostgreSQL data access layer (sqlx)
│   └── workers/        # Worker pool & orchestration
├── pkg/
│   └── utils/          # Shared utilities (SHA-256 hashing)
├── scripts/
│   └── migrations/     # SQL migration files
├── k8s/                # Kubernetes manifests
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

---

## Tech Stack

| Component        | Technology                            |
|------------------|---------------------------------------|
| Language         | Go 1.22                               |
| HTTP Framework   | Chi v5                                |
| Database         | PostgreSQL 15 + sqlx                  |
| Queue / Broker   | Apache Kafka (segmentio/kafka-go)     |
| Configuration    | Viper                                 |
| Logging          | Uber Zap (structured JSON)            |
| Container        | Docker + docker-compose               |
| Orchestration    | Kubernetes (AKS / GKE compatible)     |

---

## Prerequisites

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 15
- Apache Kafka
- Azure AI or GCP credentials (at least one provider required)

---

## Quick Start — Local Development

### 1. Clone & configure

```bash
git clone https://github.com/yiaga/erad-ai-service.git
cd erad-ai-service
cp .env.example .env
# Edit .env and fill in DATABASE_URL, KAFKA_BROKERS, provider keys
```

### 2. Start infrastructure (Postgres + Kafka)

```bash
docker-compose up -d postgres zookeeper kafka
```

### 3. Apply database migrations

```bash
psql "$DATABASE_URL" -f scripts/migrations/000001_init_schema.up.sql
```

### 4. Run the API server

```bash
go run ./cmd/api
```

### 5. Run the worker process (separate terminal)

```bash
go run ./cmd/worker
```

---

## Quick Start — Docker Compose (Full Stack)

```bash
# Copy and fill in provider secrets
cp .env.example .env

docker-compose up --build
```

This starts: `postgres`, `zookeeper`, `kafka`, `api` (port 8080), `worker`.

---

## API Reference

### `POST /jobs` — Submit extraction job

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "local_image_path": "/storage/elections/ekiti2026/pus/001/result_sheet.jpg",
    "provider": "azure"
  }'
```

Response `202 Accepted`:
```json
{
  "id": "uuid",
  "local_image_path": "...",
  "provider": "azure",
  "status": "pending",
  "retry_count": 0,
  "created_at": "..."
}
```

---

### `GET /jobs/{id}` — Get job status

```bash
curl http://localhost:8080/jobs/{id}
```

**Status values:** `pending` → `validating_image` → `queued` → `processing` → `completed` / `failed` / `retrying`

---

### `GET /jobs/{id}/result` — Get extraction result

```bash
curl http://localhost:8080/jobs/{id}/result
```

Response:
```json
{
  "id": "uuid",
  "job_id": "uuid",
  "extracted_text": "...",
  "structured_json": "{...}",
  "confidence_score": 0.98,
  "provider_used": "azure",
  "processing_duration_ms": 1240
}
```

---

### `GET /jobs/{id}/flags` — Get operational flags

```bash
curl http://localhost:8080/jobs/{id}/flags
```

**Flag types:** `duplicate_image`, `file_not_found`, `extraction_failed`, `low_quality_image`, `retry_required`

---

### `POST /retry/{id}` — Retry a failed job

```bash
curl -X POST http://localhost:8080/retry/{id}
```

---

### `GET /health` — Health check

```bash
curl http://localhost:8080/health
# → 200 OK
```

---

## Database Schema

```sql
-- extraction_jobs: core job tracking
-- extraction_results: raw + structured OCR output
-- extraction_flags: operational workflow flags (NOT election flags)
-- extraction_errors: error audit trail per job
```

See: `scripts/migrations/000001_init_schema.up.sql`

---

## Job Processing Flow

```
POST /jobs
    → CreateJob (status: pending)
    → Publish to Kafka
    → Worker picks up job
    → Validate file path (status: validating_image)
    → Generate SHA-256 hash
    → Check for duplicate hash → flag: duplicate_image (if found)
    → Call AI Provider (status: processing)
    → On failure → SaveError → retry up to 3x (status: retrying)
    → On 3rd failure → status: failed
    → On success → SaveResult → status: completed
```

---

## Kubernetes Deployment

```bash
# Apply all manifests
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/hpa.yaml
```

- API scales from **2–10 replicas** based on CPU
- Worker scales from **3–30 replicas** based on CPU (handles burst election loads)

---

## Running Tests

```bash
go test ./...
```

| Package              | Coverage                                |
|----------------------|-----------------------------------------|
| `pkg/utils`          | SHA-256 hashing (same/diff/missing)     |
| `internal/providers` | Mock provider success + error paths     |

---

## Adding a New AI Provider

1. Create `internal/providers/myprovider.go`
2. Implement the `AIProvider` interface:
   ```go
   type AIProvider interface {
       ProcessDocument(ctx context.Context, input DocumentInput) (*ExtractionResult, error)
       GetName() string
   }
   ```
3. Register in `cmd/worker/main.go`:
   ```go
   aiProviders["myprovider"] = providers.NewMyProvider(...)
   ```
4. Submit jobs with `"provider": "myprovider"`

---

## Security Notes

- AI provider API keys are injected via environment variables / Kubernetes Secrets — never hardcoded
- File paths are validated for existence before processing
- Directory traversal prevention: validate that `local_image_path` is within an allowed base directory before reading (recommended addition for production)

---

## Scaling Strategy

For nationwide election workloads:

1. **Horizontal**: Scale the `erad-ai-worker` deployment (HPA handles this automatically)
2. **Kafka Partitions**: Increase partition count on the `extraction_jobs` topic to match worker replicas
3. **DB Connections**: Use PgBouncer connection pooling in front of PostgreSQL
4. **Storage**: Ensure local image mounts are consistent across worker nodes (e.g., shared NFS or distributed filesystem)
