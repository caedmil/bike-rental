package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Domenick1991/students/rent-service/internal/models"
	"github.com/Domenick1991/students/rent-service/internal/repository"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type Service interface {
	StartRent(ctx context.Context, userID string, bikeID string) (*models.Rent, error)
	EndRent(ctx context.Context, rentID string, userID string) (*models.Rent, error)
	GetAvailableBikes(ctx context.Context, location string) ([]models.Bike, error)
}

type service struct {
	repo   repository.Repository
	writer *kafka.Writer
}

func NewService(repo repository.Repository, writer *kafka.Writer) Service {
	return &service{
		repo:   repo,
		writer: writer,
	}
}

func (s *service) StartRent(ctx context.Context, userID string, bikeID string) (*models.Rent, error) {
	bikeUUID, err := uuid.Parse(bikeID)
	if err != nil {
		return nil, fmt.Errorf("invalid bike_id: %w", err)
	}

	rent, err := s.repo.StartRent(ctx, userID, bikeUUID)
	if err != nil {
		return nil, err
	}

	// Send event to Kafka
	event := models.RentEvent{
		RentID:    rent.ID.String(),
		UserID:    userID,
		BikeID:    bikeID,
		EventType: "start",
		Timestamp: time.Now(),
	}

	if err := s.publishRentEvent(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish rent event: %v\n", err)
	}

	return rent, nil
}

func (s *service) EndRent(ctx context.Context, rentID string, userID string) (*models.Rent, error) {
	rentUUID, err := uuid.Parse(rentID)
	if err != nil {
		return nil, fmt.Errorf("invalid rent_id: %w", err)
	}

	rent, err := s.repo.EndRent(ctx, rentUUID, userID)
	if err != nil {
		return nil, err
	}

	// Send event to Kafka
	event := models.RentEvent{
		RentID:    rent.ID.String(),
		UserID:    userID,
		BikeID:    rent.BikeID.String(),
		EventType: "end",
		Timestamp: time.Now(),
	}

	if err := s.publishRentEvent(ctx, event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Failed to publish rent event: %v\n", err)
	}

	return rent, nil
}

func (s *service) GetAvailableBikes(ctx context.Context, location string) ([]models.Bike, error) {
	return s.repo.GetAvailableBikes(ctx, location)
}

func (s *service) publishRentEvent(ctx context.Context, event models.RentEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = s.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.RentID),
		Value: eventJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

