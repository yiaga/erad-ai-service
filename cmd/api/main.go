package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/yiaga/erad-ai-service/internal/api"
	"github.com/yiaga/erad-ai-service/internal/config"
	"github.com/yiaga/erad-ai-service/internal/handlers"
	"github.com/yiaga/erad-ai-service/internal/logging"
	"github.com/yiaga/erad-ai-service/internal/queue"
	"github.com/yiaga/erad-ai-service/internal/repositories"
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

	jobHandler := handlers.NewJobHandler(repo, q)
	router := api.NewRouter(jobHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		logger.Info("Starting API server", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
}
