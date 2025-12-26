package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"bike-rental/config"
	kafkawriter "bike-rental/rent-service/internal/kafka"
	"bike-rental/rent-service/internal/repository"
	"bike-rental/rent-service/internal/server"
	"bike-rental/rent-service/internal/service"
	"bike-rental/rent-service/proto/rent"

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
	// Format: postgres://user:password@host:port/dbname?sslmode=...
	// IMPORTANT: Order is user, password, host, port, dbname, sslmode
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)
	log.Printf("Connecting to database: host=%s port=%d user=%s dbname=%s",
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.DBName)

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create Kafka writer
	kafkaWriterImpl := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Kafka.Brokers...),
		Topic:                  cfg.Kafka.Topics.RentEvents,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer kafkaWriterImpl.Close()

	// Test Kafka connection by creating topic if it doesn't exist
	log.Printf("Kafka writer configured: brokers=%v, topic=%s", cfg.Kafka.Brokers, cfg.Kafka.Topics.RentEvents)

	// Initialize repository and service
	repo := repository.NewRepository(db)
	kafkaWriter := kafkawriter.NewKafkaWriter(kafkaWriterImpl)
	svc := service.NewService(repo, kafkaWriter, cfg.Kafka.Topics.RentEvents)

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
