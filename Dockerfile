# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o erad-ai-api ./cmd/api/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o erad-ai-worker ./cmd/worker/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/erad-ai-api .
COPY --from=builder /app/erad-ai-worker .

EXPOSE 8080

CMD ["./erad-ai-api"]
