package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yiaga/erad-ai-service/internal/config"
	"github.com/yiaga/erad-ai-service/internal/logging"
	"github.com/yiaga/erad-ai-service/internal/providers"
	"github.com/yiaga/erad-ai-service/internal/queue"
	"github.com/yiaga/erad-ai-service/internal/repositories"
	"github.com/yiaga/erad-ai-service/internal/services"
	"github.com/yiaga/erad-ai-service/internal/workers"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := logging.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	db, err := sqlx.Connect("postgres", cfg.DBURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	repo := repositories.NewPostgresJobRepository(db)
	
	if len(cfg.KafkaBrokers) == 0 {
		logger.Fatal("KAFKA_BROKERS environment variable is required")
	}
	
	q := queue.NewKafkaQueue(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID)
	defer q.Close()

	// Initialize AI providers
	aiProviders := map[string]providers.AIProvider{
		"azure": providers.NewAzureProvider(cfg.AzureAPIKey, cfg.AzureEndpoint),
		"gcp":   providers.NewGCPProvider(cfg.GCPProjectID, cfg.GCPLocation, cfg.GCPProcessorID),
		"mock":  &providers.MockProvider{Name: "mock"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var refiner *services.ResultRefiner
	if cfg.GeminiAPIKey != "" {
		refiner, err = services.NewResultRefiner(ctx, cfg.GeminiAPIKey)
		if err != nil {
			logger.Error("Failed to initialize Gemini refiner", zap.Error(err))
		} else {
			defer refiner.Close()
		}
	}

	manager := workers.NewWorkerManager(repo, q, aiProviders, refiner, logger)

	logger.Info("Starting extraction workers", zap.Int("count", cfg.WorkerCount))
	manager.Start(ctx, cfg.WorkerCount)

	// Keep alive until signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Workers shutting down")
}
