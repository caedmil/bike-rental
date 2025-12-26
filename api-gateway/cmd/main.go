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

	"github.com/Domenick1991/students/api-gateway/internal/client"
	"github.com/Domenick1991/students/api-gateway/internal/handlers"
	"github.com/Domenick1991/students/config"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Bike Rental API
// @version 1.0
// @description API for bike rental system
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@bikerental.com

// @host localhost:8080
// @BasePath /
func main() {
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize gRPC client for Rent Service
	rentClient, err := client.NewRentClient(cfg.Services.RentService)
	if err != nil {
		log.Fatalf("Failed to create rent client: %v", err)
	}
	defer rentClient.Close()

	// Initialize HTTP client for Stats Service
	statsClient := client.NewStatsClient(fmt.Sprintf("http://stats-service:%d", cfg.Server.StatsServicePort))

	// Setup handlers
	h := handlers.NewHandlers(rentClient, statsClient)

	// Setup router
	r := chi.NewRouter()

	// Swagger - serve YAML as JSON for compatibility
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		// For now, return empty JSON - in production, convert YAML to JSON
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"openapi":"3.0.0","info":{"title":"Bike Rental API","version":"1.0.0"},"paths":{}}`))
	})

	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger.json"),
	))

	// Register API routes
	h.RegisterRoutes(r)

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.APIGatewayPort),
		Handler: r,
	}

	log.Printf("API Gateway starting on port %d", cfg.Server.APIGatewayPort)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
}

