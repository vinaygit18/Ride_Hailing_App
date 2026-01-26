package rider

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRiderNotFound = errors.New("rider not found")
	ErrInvalidRider  = errors.New("invalid rider data")
)

// Rider represents a rider entity
type Rider struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Rating     float64   `json:"rating"`
	TotalRides int       `json:"total_rides"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Repository defines the interface for rider data access
type Repository interface {
	Create(ctx context.Context, rider *Rider) error
	GetByID(ctx context.Context, id uuid.UUID) (*Rider, error)
	GetByEmail(ctx context.Context, email string) (*Rider, error)
	Update(ctx context.Context, rider *Rider) error
	Delete(ctx context.Context, id uuid.UUID) error
}
