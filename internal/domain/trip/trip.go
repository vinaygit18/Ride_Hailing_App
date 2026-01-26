package trip

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type Trip struct {
	ID              uuid.UUID  `json:"id"`
	RideID          uuid.UUID  `json:"ride_id"`
	StartedAt       time.Time  `json:"started_at"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
	DistanceKM      *float64   `json:"distance_km,omitempty"`
	DurationMinutes *int       `json:"duration_minutes,omitempty"`
	BaseFare        float64    `json:"base_fare"`
	DistanceFare    float64    `json:"distance_fare"`
	TimeFare        float64    `json:"time_fare"`
	SurgeMultiplier float64    `json:"surge_multiplier"`
	TotalFare       *float64   `json:"total_fare,omitempty"`
	Status          Status     `json:"status"`
	RoutePolyline   string     `json:"route_polyline,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, trip *Trip) error
	GetByID(ctx context.Context, id uuid.UUID) (*Trip, error)
	GetByRideID(ctx context.Context, rideID uuid.UUID) (*Trip, error)
	Update(ctx context.Context, trip *Trip) error
	Complete(ctx context.Context, id uuid.UUID, distanceKM float64, durationMinutes int, totalFare float64) error
}

var (
	ErrTripNotFound        = errors.New("trip not found")
	ErrTripAlreadyCompleted = errors.New("trip already completed")
)
