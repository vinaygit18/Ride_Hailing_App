package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocomet/ride-hailing/internal/api/dto"
	"github.com/gocomet/ride-hailing/internal/domain/driver"
	"github.com/gocomet/ride-hailing/internal/service/matching"
	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/gocomet/ride-hailing/pkg/websocket"
)

// CreateRide handles POST /v1/rides
func (h *Handlers) CreateRide(c *gin.Context) {
	var req dto.CreateRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	// Generate ride ID
	rideID := generateRideID()

	h.Logger.Info("Ride request received",
		logger.String("ride_id", rideID),
		logger.String("rider_id", req.RiderID),
		logger.Float64("pickup_lat", req.PickupLatitude),
		logger.Float64("pickup_lng", req.PickupLongitude),
	)

	// Parse vehicle type
	var vehicleType driver.VehicleType
	switch req.VehicleType {
	case "economy":
		vehicleType = driver.VehicleEconomy
	case "premium":
		vehicleType = driver.VehiclePremium
	case "luxury":
		vehicleType = driver.VehicleLuxury
	default:
		vehicleType = driver.VehicleEconomy
	}

	// Create matching service with progressive radius expansion
	// Starts at 5km, expands to 10km, 20km, up to 50km if no drivers found
	matchingService := matching.NewService(h.Redis, h.Logger, matching.Config{
		MaxRadiusKM:       5.0,  // Initial search radius
		MaxExpandedRadius: 50.0, // Maximum expanded radius
		MaxTimeout:        30,
		MaxCandidates:     50,   // Check up to 50 candidates to handle concurrent requests
	})

	// Find nearest driver
	ctx := context.Background()
	foundDriver, err := matchingService.FindNearestDriver(ctx, req.PickupLatitude, req.PickupLongitude, vehicleType)
	if err != nil {
		h.Logger.Warn("No drivers available", logger.Err(err))
		c.JSON(http.StatusOK, gin.H{
			"id":       rideID,
			"rider_id": req.RiderID,
			"status":   "requested",
			"message":  "Searching for drivers...",
			"driver":   nil,
		})
		return
	}

	// Save ride to PostgreSQL
	_, err = h.DB.ExecContext(ctx, `
		INSERT INTO rides (
			id, rider_id, driver_id, status, vehicle_type,
			pickup_latitude, pickup_longitude,
			dropoff_latitude, dropoff_longitude,
			estimated_fare, requested_at, assigned_at
		) VALUES ($1, $2, $3, 'assigned', $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`, rideID, req.RiderID, foundDriver.ID.String(), req.VehicleType,
		req.PickupLatitude, req.PickupLongitude,
		req.DropoffLatitude, req.DropoffLongitude, 250.00)

	if err != nil {
		h.Logger.Error("Failed to save ride to PostgreSQL", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ride"})
		return
	}

	h.Logger.Info("Ride saved to PostgreSQL",
		logger.String("ride_id", rideID),
		logger.String("driver_id", foundDriver.ID.String()),
	)

	// Set actual ride ID for driver (matching service already removed from available set)
	driverIDStr := foundDriver.ID.String()
	h.Redis.Set(ctx, fmt.Sprintf("driver:%s:current_ride", driverIDStr), rideID, 0)

	h.Logger.Info("Driver marked as busy",
		logger.String("driver_id", driverIDStr),
		logger.String("ride_id", rideID),
	)

	// Send WebSocket notification to dashboard
	driverNotification := map[string]interface{}{
		"type": "ride_request",
		"data": map[string]interface{}{
			"ride_id":           rideID,
			"driver_id":         foundDriver.ID.String(),
			"rider_id":          req.RiderID,
			"pickup_latitude":   req.PickupLatitude,
			"pickup_longitude":  req.PickupLongitude,
			"dropoff_latitude":  req.DropoffLatitude,
			"dropoff_longitude": req.DropoffLongitude,
			"vehicle_type":      req.VehicleType,
			"distance":          "2.5 km",
			"estimated_fare":    250.00,
		},
	}
	// Broadcast to all dashboard users
	if wsHub, ok := h.Hub.(*websocket.Hub); ok {
		wsHub.BroadcastToType("dashboard", driverNotification)
	}

	h.Logger.Info("Driver matched and dashboard notified",
		logger.String("ride_id", rideID),
		logger.String("driver_id", foundDriver.ID.String()),
	)

	// Return response to rider
	c.JSON(http.StatusOK, gin.H{
		"id":        rideID,
		"rider_id":  req.RiderID,
		"status":    "assigned",
		"driver_id": foundDriver.ID.String(),
		"driver_name": foundDriver.Name,
		"driver": map[string]interface{}{
			"id":        foundDriver.ID.String(),
			"name":      foundDriver.Name,
			"rating":    foundDriver.Rating,
			"vehicle":   req.VehicleType,
			"latitude":  foundDriver.CurrentLatitude,
			"longitude": foundDriver.CurrentLongitude,
		},
		"estimated_arrival": "5 mins",
		"estimated_fare":    250.00,
	})
}

// GetRide handles GET /v1/rides/:id
func (h *Handlers) GetRide(c *gin.Context) {
	rideID := c.Param("id")
	ctx := context.Background()

	// Query PostgreSQL with LEFT JOIN to drivers
	query := `
		SELECT r.id, r.rider_id, r.driver_id, r.status, r.vehicle_type,
		       r.pickup_latitude, r.pickup_longitude,
		       r.dropoff_latitude, r.dropoff_longitude,
		       r.estimated_fare, r.requested_at, r.assigned_at,
		       r.accepted_at, r.completed_at,
		       d.name as driver_name, d.rating as driver_rating,
		       d.phone as driver_phone
		FROM rides r
		LEFT JOIN drivers d ON r.driver_id = d.id
		WHERE r.id = $1
	`

	var ride struct {
		ID                string
		RiderID           string
		DriverID          sql.NullString
		Status            string
		VehicleType       string
		PickupLatitude    float64
		PickupLongitude   float64
		DropoffLatitude   float64
		DropoffLongitude  float64
		EstimatedFare     sql.NullFloat64
		RequestedAt       time.Time
		AssignedAt        sql.NullTime
		AcceptedAt        sql.NullTime
		CompletedAt       sql.NullTime
		DriverName        sql.NullString
		DriverRating      sql.NullFloat64
		DriverPhone       sql.NullString
	}

	err := h.DB.QueryRowContext(ctx, query, rideID).Scan(
		&ride.ID, &ride.RiderID, &ride.DriverID, &ride.Status, &ride.VehicleType,
		&ride.PickupLatitude, &ride.PickupLongitude,
		&ride.DropoffLatitude, &ride.DropoffLongitude,
		&ride.EstimatedFare, &ride.RequestedAt, &ride.AssignedAt,
		&ride.AcceptedAt, &ride.CompletedAt,
		&ride.DriverName, &ride.DriverRating, &ride.DriverPhone,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ride not found"})
		return
	}

	if err != nil {
		h.Logger.Error("Failed to get ride", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ride"})
		return
	}

	// Build response
	response := gin.H{
		"id":                 ride.ID,
		"rider_id":           ride.RiderID,
		"status":             ride.Status,
		"vehicle_type":       ride.VehicleType,
		"pickup_latitude":    ride.PickupLatitude,
		"pickup_longitude":   ride.PickupLongitude,
		"dropoff_latitude":   ride.DropoffLatitude,
		"dropoff_longitude":  ride.DropoffLongitude,
		"requested_at":       ride.RequestedAt,
	}

	if ride.EstimatedFare.Valid {
		response["estimated_fare"] = ride.EstimatedFare.Float64
	}

	if ride.DriverID.Valid {
		response["driver_id"] = ride.DriverID.String
		response["driver"] = gin.H{
			"name":   ride.DriverName.String,
			"rating": ride.DriverRating.Float64,
			"phone":  ride.DriverPhone.String,
		}
	}

	if ride.AssignedAt.Valid {
		response["assigned_at"] = ride.AssignedAt.Time
	}

	if ride.AcceptedAt.Valid {
		response["accepted_at"] = ride.AcceptedAt.Time
	}

	if ride.CompletedAt.Valid {
		response["completed_at"] = ride.CompletedAt.Time

		// If completed, also get trip details
		var trip struct {
			ID              string
			DistanceKm      float64
			DurationMinutes int
			TotalFare       float64
		}

		err = h.DB.QueryRowContext(ctx, `
			SELECT id, distance_km, duration_minutes, total_fare
			FROM trips
			WHERE ride_id = $1 AND status = 'completed'
		`, rideID).Scan(&trip.ID, &trip.DistanceKm, &trip.DurationMinutes, &trip.TotalFare)

		if err == nil {
			response["trip"] = gin.H{
				"id":               trip.ID,
				"distance_km":      trip.DistanceKm,
				"duration_minutes": trip.DurationMinutes,
				"total_fare":       trip.TotalFare,
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

// Helper function to generate ride ID
func generateRideID() string {
	return fmt.Sprintf("ride-%d", time.Now().UnixNano())
}
