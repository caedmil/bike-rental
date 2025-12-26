package models

import (
	"time"

	"github.com/google/uuid"
)

type Bike struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Status    string    `db:"status"`
	Location  string    `db:"location"`
	CreatedAt time.Time `db:"created_at"`
}

type Rent struct {
	ID        uuid.UUID  `db:"id"`
	UserID    string     `db:"user_id"`
	BikeID    uuid.UUID  `db:"bike_id"`
	StartTime time.Time  `db:"start_time"`
	EndTime   *time.Time `db:"end_time"`
	Status    string     `db:"status"`
}

type RentEvent struct {
	RentID    string    `json:"rent_id"`
	UserID    string    `json:"user_id"`
	BikeID    string    `json:"bike_id"`
	EventType string    `json:"event_type"` // "start" or "end"
	Timestamp time.Time `json:"timestamp"`
}

type StatusEvent struct {
	BikeID    string    `json:"bike_id"`
	Status    string    `json:"status"`
	Location  string    `json:"location"`
	Timestamp time.Time `json:"timestamp"`
}

