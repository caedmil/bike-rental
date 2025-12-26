package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bike-rental/config"
	"bike-rental/stats-service/internal/consumer"
	"bike-rental/stats-service/internal/handlers"
	"bike-rental/stats-service/internal/repository"
	"bike-rental/stats-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Database.Redis.Address,
		Password: cfg.Database.Redis.Password,
		DB:       cfg.Database.Redis.DB,
	})
	defer rdb.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repository and service
	repo := repository.NewRepository(rdb)
	svc := service.NewService(repo)

	// Start Kafka consumer
	kafkaConsumer := consumer.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.Topics.RentEvents,
		repo,
	)

	go kafkaConsumer.Start(context.Background())
	defer kafkaConsumer.Stop()

	// Setup HTTP server
	r := chi.NewRouter()
	handlers := handlers.NewHandlers(svc)
	handlers.RegisterRoutes(r)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.StatsServicePort),
		Handler: r,
	}

	log.Printf("Stats Service starting on port %d", cfg.Server.StatsServicePort)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Stats Service...")
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

