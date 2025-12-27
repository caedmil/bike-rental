package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bike-rental/rent-service/internal/kafka"
	"bike-rental/rent-service/internal/models"
	"bike-rental/rent-service/internal/repository"
	"github.com/google/uuid"
	kafkago "github.com/segmentio/kafka-go"
)

type Service interface {
	StartRent(ctx context.Context, userID string, bikeID string) (*models.Rent, error)
	EndRent(ctx context.Context, rentID string, userID string) (*models.Rent, error)
	GetAvailableBikes(ctx context.Context, location string) ([]models.Bike, error)
	AddBike(ctx context.Context, name, location string) (*models.Bike, error)
	DeleteBike(ctx context.Context, bikeID string) error
}

type service struct {
	repo   repository.Repository
	writer kafka.Writer
	topic  string
}

func NewService(repo repository.Repository, writer kafka.Writer, topic string) Service {
	return &service{
		repo:   repo,
		writer: writer,
		topic:  topic,
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
		log.Printf("Failed to publish rent event: %v", err)
	} else {
		log.Printf("Successfully published rent event: rent_id=%s, event_type=%s", event.RentID, event.EventType)
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
		log.Printf("Failed to publish rent event: %v", err)
	} else {
		log.Printf("Successfully published rent event: rent_id=%s, event_type=%s", event.RentID, event.EventType)
	}

	return rent, nil
}

func (s *service) GetAvailableBikes(ctx context.Context, location string) ([]models.Bike, error) {
	return s.repo.GetAvailableBikes(ctx, location)
}

func (s *service) AddBike(ctx context.Context, name, location string) (*models.Bike, error) {
	if name == "" {
		return nil, fmt.Errorf("bike name is required")
	}
	if location == "" {
		return nil, fmt.Errorf("bike location is required")
	}
	
	bike, err := s.repo.AddBike(ctx, name, location)
	if err != nil {
		return nil, err
	}
	
	log.Printf("Successfully added new bike: id=%s, name=%s, location=%s", bike.ID, bike.Name, bike.Location)
	return bike, nil
}

func (s *service) DeleteBike(ctx context.Context, bikeID string) error {
	bikeUUID, err := uuid.Parse(bikeID)
	if err != nil {
		return fmt.Errorf("invalid bike_id: %w", err)
	}
	
	// Check if bike has active rents
	hasActiveRent, err := s.repo.HasActiveRent(ctx, bikeUUID)
	if err != nil {
		return fmt.Errorf("failed to check active rents: %w", err)
	}
	
	if hasActiveRent {
		return fmt.Errorf("cannot delete bike: bike has active rent")
	}
	
	// Get bike info for logging before deletion
	bike, err := s.repo.GetBikeByID(ctx, bikeUUID)
	if err != nil {
		return fmt.Errorf("bike not found: %w", err)
	}
	
	// Delete the bike
	err = s.repo.DeleteBike(ctx, bikeUUID)
	if err != nil {
		return err
	}
	
	log.Printf("Successfully deleted bike: id=%s, name=%s, location=%s", bike.ID, bike.Name, bike.Location)
	return nil
}

func (s *service) publishRentEvent(ctx context.Context, event models.RentEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	log.Printf("Publishing event to Kafka: topic=%s, rent_id=%s, event_type=%s", s.topic, event.RentID, event.EventType)
	
	err = s.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(event.RentID),
		Value: eventJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	log.Printf("Event published successfully to Kafka: rent_id=%s", event.RentID)
	return nil
}

