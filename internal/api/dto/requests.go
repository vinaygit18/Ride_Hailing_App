package dto

import "github.com/google/uuid"

// CreateRideRequest represents a request to create a new ride
type CreateRideRequest struct {
	RiderID          string  `json:"rider_id" binding:"required"`
	PickupLatitude   float64 `json:"pickup_latitude" binding:"required"`
	PickupLongitude  float64 `json:"pickup_longitude" binding:"required"`
	DropoffLatitude  float64 `json:"dropoff_latitude" binding:"required"`
	DropoffLongitude float64 `json:"dropoff_longitude" binding:"required"`
	VehicleType      string  `json:"vehicle_type" binding:"required,oneof=economy premium luxury"`
}

// UpdateLocationRequest represents a driver location update
type UpdateLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

// AcceptRideRequest represents a driver accepting a ride
type AcceptRideRequest struct {
	RideID string `json:"ride_id" binding:"required"`
}

// EndTripRequest represents ending a trip
type EndTripRequest struct {
	DriverID        string  `json:"driver_id" binding:"required"`
	DistanceKm      float64 `json:"distance_km" binding:"required"`
	DurationMinutes int     `json:"duration_minutes" binding:"required"`
}

// CreatePaymentRequest represents a payment request
type CreatePaymentRequest struct {
	TripID        string  `json:"trip_id" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=card wallet cash upi"`
	Amount        float64 `json:"amount" binding:"required"`
}

// Ride response
type RideResponse struct {
	ID                  uuid.UUID        `json:"id"`
	RiderID             uuid.UUID        `json:"rider_id"`
	DriverID            *uuid.UUID       `json:"driver_id,omitempty"`
	Status              string           `json:"status"`
	VehicleType         string           `json:"vehicle_type"`
	PickupLocation      LocationResponse `json:"pickup_location"`
	DropoffLocation     LocationResponse `json:"dropoff_location"`
	EstimatedFare       *float64         `json:"estimated_fare,omitempty"`
	EstimatedArrival    string           `json:"estimated_arrival,omitempty"`
	Driver              *DriverResponse  `json:"driver,omitempty"`
}

type LocationResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type DriverResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	VehicleType string    `json:"vehicle_type"`
	Rating      float64   `json:"rating"`
}

// Error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
