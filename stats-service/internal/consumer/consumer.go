package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/Domenick1991/students/stats-service/internal/repository"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	repo    repository.Repository
	stopCh  chan struct{}
}

type RentEvent struct {
	RentID    string    `json:"rent_id"`
	UserID    string    `json:"user_id"`
	BikeID    string    `json:"bike_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

func NewConsumer(brokers []string, topic string, repo repository.Repository) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		GroupID:  "stats-service-group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader: reader,
		repo:   repo,
		stopCh: make(chan struct{}),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	log.Println("Starting Kafka consumer...")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Context cancelled, stopping consumer")
			return
		case <-c.stopCh:
			log.Println("Stop signal received, stopping consumer")
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				time.Sleep(time.Second)
				continue
			}

			if err := c.processMessage(ctx, msg.Value); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}
	}
}

func (c *Consumer) Stop() {
	close(c.stopCh)
	if err := c.reader.Close(); err != nil {
		log.Printf("Error closing reader: %v", err)
	}
}

func (c *Consumer) processMessage(ctx context.Context, data []byte) error {
	var event RentEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	date := event.Timestamp.Format("2006-01-02")

	switch event.EventType {
	case "start":
		if err := c.repo.IncrementDailyRent(ctx, date); err != nil {
			return fmt.Errorf("failed to increment daily rent: %w", err)
		}
		if err := c.repo.IncrementActiveRents(ctx); err != nil {
			return fmt.Errorf("failed to increment active rents: %w", err)
		}
		// Note: location would need to come from bike data, for now we'll skip it
		// In a real implementation, you'd fetch bike location from rent service
		
	case "end":
		if err := c.repo.DecrementActiveRents(ctx); err != nil {
			return fmt.Errorf("failed to decrement active rents: %w", err)
		}
	}

	return nil
}

