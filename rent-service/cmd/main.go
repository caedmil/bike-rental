package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Domenick1991/students/config"
	"github.com/Domenick1991/students/rent-service/internal/repository"
	"github.com/Domenick1991/students/rent-service/internal/server"
	"github.com/Domenick1991/students/rent-service/internal/service"
	"github.com/Domenick1991/students/rent-service/proto/rent"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to PostgreSQL
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create Kafka writer
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Kafka.Brokers...),
		Topic:    cfg.Kafka.Topics.RentEvents,
		Balancer: &kafka.LeastBytes{},
	}
	defer writer.Close()

	// Initialize repository and service
	repo := repository.NewRepository(db)
	svc := service.NewService(repo, writer)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	rentServer := server.NewRentServer(svc)
	rent.RegisterRentServiceServer(grpcServer, rentServer)

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.RentServicePort))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Rent Service starting on port %d", cfg.Server.RentServicePort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Rent Service...")
	grpcServer.GracefulStop()
}

