package driver

import (
	"time"

	"github.com/google/uuid"
)

// Status represents driver availability status
type Status string

const (
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
	StatusBusy    Status = "busy"
)

// VehicleType represents the type of vehicle
type VehicleType string

const (
	VehicleEconomy VehicleType = "economy"
	VehiclePremium VehicleType = "premium"
	VehicleLuxury  VehicleType = "luxury"
)

// Driver represents a driver entity
type Driver struct {
	ID               uuid.UUID   `json:"id"`
	Name             string      `json:"name"`
	Email            string      `json:"email"`
	Phone            string      `json:"phone"`
	Status           Status      `json:"status"`
	VehicleType      VehicleType `json:"vehicle_type"`
	CurrentLatitude  *float64    `json:"current_latitude,omitempty"`
	CurrentLongitude *float64    `json:"current_longitude,omitempty"`
	Rating           float64     `json:"rating"`
	TotalRides       int         `json:"total_rides"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

// Location represents a geographic location
type Location struct {
	Latitude  float64
	Longitude float64
}

// IsValid validates the driver entity
func (d *Driver) IsValid() error {
	if d.Name == "" {
		return ErrInvalidDriverName
	}
	if d.Email == "" {
		return ErrInvalidDriverEmail
	}
	if d.Phone == "" {
		return ErrInvalidDriverPhone
	}
	if !d.Status.IsValid() {
		return ErrInvalidDriverStatus
	}
	if !d.VehicleType.IsValid() {
		return ErrInvalidVehicleType
	}
	return nil
}

// IsValid validates the status
func (s Status) IsValid() bool {
	switch s {
	case StatusOnline, StatusOffline, StatusBusy:
		return true
	}
	return false
}

// IsValid validates the vehicle type
func (v VehicleType) IsValid() bool {
	switch v {
	case VehicleEconomy, VehiclePremium, VehicleLuxury:
		return true
	}
	return false
}

// CanAcceptRides returns true if driver can accept new rides
func (d *Driver) CanAcceptRides() bool {
	return d.Status == StatusOnline
}

// SetLocation updates the driver's current location
func (d *Driver) SetLocation(lat, lng float64) {
	d.CurrentLatitude = &lat
	d.CurrentLongitude = &lng
	d.UpdatedAt = time.Now()
}

// SetStatus updates the driver's status
func (d *Driver) SetStatus(status Status) error {
	if !status.IsValid() {
		return ErrInvalidDriverStatus
	}
	d.Status = status
	d.UpdatedAt = time.Now()
	return nil
}

// GetLocation returns the driver's current location
func (d *Driver) GetLocation() *Location {
	if d.CurrentLatitude == nil || d.CurrentLongitude == nil {
		return nil
	}
	return &Location{
		Latitude:  *d.CurrentLatitude,
		Longitude: *d.CurrentLongitude,
	}
}
