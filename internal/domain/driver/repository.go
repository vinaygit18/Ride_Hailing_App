package driver

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for driver data access
type Repository interface {
	// Create creates a new driver
	Create(ctx context.Context, driver *Driver) error

	// GetByID retrieves a driver by ID
	GetByID(ctx context.Context, id uuid.UUID) (*Driver, error)

	// GetByEmail retrieves a driver by email
	GetByEmail(ctx context.Context, email string) (*Driver, error)

	// Update updates a driver
	Update(ctx context.Context, driver *Driver) error

	// UpdateStatus updates driver status
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error

	// UpdateLocation updates driver location
	UpdateLocation(ctx context.Context, id uuid.UUID, lat, lng float64) error

	// GetNearbyDrivers finds drivers within a radius
	GetNearbyDrivers(ctx context.Context, lat, lng, radiusKM float64, vehicleType VehicleType, limit int) ([]*Driver, error)

	// GetAvailableDrivers retrieves all online drivers
	GetAvailableDrivers(ctx context.Context, vehicleType VehicleType) ([]*Driver, error)

	// Delete deletes a driver
	Delete(ctx context.Context, id uuid.UUID) error
}
