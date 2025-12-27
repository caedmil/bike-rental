package repository

import (
	"context"
	"database/sql"
	"fmt"

	"bike-rental/rent-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	GetAvailableBikes(ctx context.Context, location string) ([]models.Bike, error)
	GetBikeByID(ctx context.Context, bikeID uuid.UUID) (*models.Bike, error)
	StartRent(ctx context.Context, userID string, bikeID uuid.UUID) (*models.Rent, error)
	EndRent(ctx context.Context, rentID uuid.UUID, userID string) (*models.Rent, error)
	GetRentByID(ctx context.Context, rentID uuid.UUID) (*models.Rent, error)
	UpdateBikeStatus(ctx context.Context, bikeID uuid.UUID, status string) error
	AddBike(ctx context.Context, name, location string) (*models.Bike, error)
	DeleteBike(ctx context.Context, bikeID uuid.UUID) error
	HasActiveRent(ctx context.Context, bikeID uuid.UUID) (bool, error)
}

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) GetAvailableBikes(ctx context.Context, location string) ([]models.Bike, error) {
	query := `
		SELECT id, name, status, location, created_at
		FROM bikes
		WHERE status = 'available'
	`
	args := []interface{}{}
	
	if location != "" {
		query += " AND location = $1"
		args = append(args, location)
	}
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query bikes: %w", err)
	}
	defer rows.Close()

	var bikes []models.Bike
	for rows.Next() {
		var bike models.Bike
		err := rows.Scan(&bike.ID, &bike.Name, &bike.Status, &bike.Location, &bike.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bike: %w", err)
		}
		bikes = append(bikes, bike)
	}

	return bikes, nil
}

func (r *repository) GetBikeByID(ctx context.Context, bikeID uuid.UUID) (*models.Bike, error) {
	var bike models.Bike
	err := r.db.QueryRow(ctx,
		"SELECT id, name, status, location, created_at FROM bikes WHERE id = $1",
		bikeID,
	).Scan(&bike.ID, &bike.Name, &bike.Status, &bike.Location, &bike.CreatedAt)
	
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("bike not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bike: %w", err)
	}
	
	return &bike, nil
}

func (r *repository) StartRent(ctx context.Context, userID string, bikeID uuid.UUID) (*models.Rent, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Check if bike is available
	var bikeStatus string
	err = tx.QueryRow(ctx, "SELECT status FROM bikes WHERE id = $1", bikeID).Scan(&bikeStatus)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("bike not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check bike status: %w", err)
	}
	if bikeStatus != "available" {
		return nil, fmt.Errorf("bike is not available")
	}

	// Update bike status
	_, err = tx.Exec(ctx, "UPDATE bikes SET status = 'rented' WHERE id = $1", bikeID)
	if err != nil {
		return nil, fmt.Errorf("failed to update bike status: %w", err)
	}

	// Create rent record
	var rent models.Rent
	err = tx.QueryRow(ctx,
		`INSERT INTO rents (user_id, bike_id, start_time, status)
		 VALUES ($1, $2, NOW(), 'active')
		 RETURNING id, user_id, bike_id, start_time, end_time, status`,
		userID, bikeID,
	).Scan(&rent.ID, &rent.UserID, &rent.BikeID, &rent.StartTime, &rent.EndTime, &rent.Status)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create rent: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &rent, nil
}

func (r *repository) EndRent(ctx context.Context, rentID uuid.UUID, userID string) (*models.Rent, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get rent
	var rent models.Rent
	err = tx.QueryRow(ctx,
		"SELECT id, user_id, bike_id, start_time, end_time, status FROM rents WHERE id = $1 AND user_id = $2",
		rentID, userID,
	).Scan(&rent.ID, &rent.UserID, &rent.BikeID, &rent.StartTime, &rent.EndTime, &rent.Status)
	
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("rent not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rent: %w", err)
	}
	
	if rent.Status != "active" {
		return nil, fmt.Errorf("rent is not active")
	}

	// Update rent
	endTime := sql.NullTime{}
	err = tx.QueryRow(ctx,
		`UPDATE rents SET end_time = NOW(), status = 'completed'
		 WHERE id = $1
		 RETURNING end_time`,
		rentID,
	).Scan(&endTime)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update rent: %w", err)
	}
	
	if endTime.Valid {
		rent.EndTime = &endTime.Time
	}
	rent.Status = "completed"

	// Update bike status
	_, err = tx.Exec(ctx, "UPDATE bikes SET status = 'available' WHERE id = $1", rent.BikeID)
	if err != nil {
		return nil, fmt.Errorf("failed to update bike status: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &rent, nil
}

func (r *repository) GetRentByID(ctx context.Context, rentID uuid.UUID) (*models.Rent, error) {
	var rent models.Rent
	err := r.db.QueryRow(ctx,
		"SELECT id, user_id, bike_id, start_time, end_time, status FROM rents WHERE id = $1",
		rentID,
	).Scan(&rent.ID, &rent.UserID, &rent.BikeID, &rent.StartTime, &rent.EndTime, &rent.Status)
	
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("rent not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rent: %w", err)
	}
	
	return &rent, nil
}

func (r *repository) UpdateBikeStatus(ctx context.Context, bikeID uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, "UPDATE bikes SET status = $1 WHERE id = $2", status, bikeID)
	if err != nil {
		return fmt.Errorf("failed to update bike status: %w", err)
	}
	return nil
}

func (r *repository) AddBike(ctx context.Context, name, location string) (*models.Bike, error) {
	var bike models.Bike
	err := r.db.QueryRow(ctx,
		`INSERT INTO bikes (name, status, location, created_at)
		 VALUES ($1, 'available', $2, NOW())
		 RETURNING id, name, status, location, created_at`,
		name, location,
	).Scan(&bike.ID, &bike.Name, &bike.Status, &bike.Location, &bike.CreatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to add bike: %w", err)
	}
	
	return &bike, nil
}

func (r *repository) HasActiveRent(ctx context.Context, bikeID uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM rents WHERE bike_id = $1 AND status = 'active'`,
		bikeID,
	).Scan(&count)
	
	if err != nil {
		return false, fmt.Errorf("failed to check active rents: %w", err)
	}
	
	return count > 0, nil
}

func (r *repository) DeleteBike(ctx context.Context, bikeID uuid.UUID) error {
	// First check if bike exists
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM bikes WHERE id = $1)", bikeID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check bike existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("bike not found")
	}
	
	// Start transaction to delete bike and related rents
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	
	// First delete all rents for this bike (CASCADE manually)
	_, err = tx.Exec(ctx, "DELETE FROM rents WHERE bike_id = $1", bikeID)
	if err != nil {
		return fmt.Errorf("failed to delete related rents: %w", err)
	}
	
	// Then delete the bike
	result, err := tx.Exec(ctx, "DELETE FROM bikes WHERE id = $1", bikeID)
	if err != nil {
		return fmt.Errorf("failed to delete bike: %w", err)
	}
	
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("bike not found")
	}
	
	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

