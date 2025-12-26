package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"bike-rental/api-gateway/internal/client"
	"bike-rental/api-gateway/internal/handlers"
	"bike-rental/config"
	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"gopkg.in/yaml.v3"
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
	r.Get("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		// Read swagger.yaml file - try multiple possible paths
		possiblePaths := []string{
			filepath.Join("docs", "swagger.yaml"),
			filepath.Join("api-gateway", "docs", "swagger.yaml"),
			filepath.Join("/root", "docs", "swagger.yaml"),
		}
		
		var yamlData []byte
		var err error
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				yamlData, err = os.ReadFile(path)
				if err == nil {
					break
				}
			}
		}
		
		if yamlData == nil {
			log.Printf("Failed to find swagger.yaml in any of the paths: %v", possiblePaths)
			http.Error(w, "Failed to load Swagger documentation", http.StatusInternalServerError)
			return
		}

		// Parse YAML
		var swaggerData map[string]interface{}
		if err := yaml.Unmarshal(yamlData, &swaggerData); err != nil {
			log.Printf("Failed to parse swagger.yaml: %v", err)
			http.Error(w, "Failed to parse Swagger documentation", http.StatusInternalServerError)
			return
		}

		// Convert to JSON
		jsonData, err := json.Marshal(swaggerData)
		if err != nil {
			log.Printf("Failed to convert to JSON: %v", err)
			http.Error(w, "Failed to convert Swagger documentation", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
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

