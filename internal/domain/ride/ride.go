package ride

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status represents ride status
type Status string

const (
	StatusRequested Status = "requested"
	StatusAssigned  Status = "assigned"
	StatusAccepted  Status = "accepted"
	StatusStarted   Status = "started"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

// VehicleType matches driver vehicle types
type VehicleType string

const (
	VehicleEconomy VehicleType = "economy"
	VehiclePremium VehicleType = "premium"
	VehicleLuxury  VehicleType = "luxury"
)

// Ride represents a ride request/assignment
type Ride struct {
	ID                       uuid.UUID    `json:"id"`
	RiderID                  uuid.UUID    `json:"rider_id"`
	DriverID                 *uuid.UUID   `json:"driver_id,omitempty"`
	Status                   Status       `json:"status"`
	VehicleType              VehicleType  `json:"vehicle_type"`
	PickupLatitude           float64      `json:"pickup_latitude"`
	PickupLongitude          float64      `json:"pickup_longitude"`
	DropoffLatitude          float64      `json:"dropoff_latitude"`
	DropoffLongitude         float64      `json:"dropoff_longitude"`
	PickupAddress            string       `json:"pickup_address,omitempty"`
	DropoffAddress           string       `json:"dropoff_address,omitempty"`
	EstimatedFare            *float64     `json:"estimated_fare,omitempty"`
	EstimatedDistanceKM      *float64     `json:"estimated_distance_km,omitempty"`
	EstimatedDurationMinutes *int         `json:"estimated_duration_minutes,omitempty"`
	RequestedAt              time.Time    `json:"requested_at"`
	AssignedAt               *time.Time   `json:"assigned_at,omitempty"`
	AcceptedAt               *time.Time   `json:"accepted_at,omitempty"`
	StartedAt                *time.Time   `json:"started_at,omitempty"`
	CompletedAt              *time.Time   `json:"completed_at,omitempty"`
	CancelledAt              *time.Time   `json:"cancelled_at,omitempty"`
	CancellationReason       string       `json:"cancellation_reason,omitempty"`
	IdempotencyKey           string       `json:"-"`
	CreatedAt                time.Time    `json:"created_at"`
	UpdatedAt                time.Time    `json:"updated_at"`
}

// Repository interface
type Repository interface {
	Create(ctx context.Context, ride *Ride) error
	GetByID(ctx context.Context, id uuid.UUID) (*Ride, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*Ride, error)
	Update(ctx context.Context, ride *Ride) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
	AssignDriver(ctx context.Context, rideID, driverID uuid.UUID) error
	GetActiveRideByDriver(ctx context.Context, driverID uuid.UUID) (*Ride, error)
	GetActiveRideByRider(ctx context.Context, riderID uuid.UUID) (*Ride, error)
}

// Errors
var (
	ErrRideNotFound        = errors.New("ride not found")
	ErrInvalidStatus       = errors.New("invalid status transition")
	ErrRideAlreadyAssigned = errors.New("ride already assigned")
)

// CanAssignDriver checks if a driver can be assigned to this ride
func (r *Ride) CanAssignDriver() bool {
	return r.Status == StatusRequested
}

// CanAccept checks if ride can be accepted by driver
func (r *Ride) CanAccept() bool {
	return r.Status == StatusAssigned
}

// CanStart checks if ride can be started
func (r *Ride) CanStart() bool {
	return r.Status == StatusAccepted
}

// CanComplete checks if ride can be completed
func (r *Ride) CanComplete() bool {
	return r.Status == StatusStarted
}
