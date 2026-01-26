package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocomet/ride-hailing/internal/api/dto"
	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/gocomet/ride-hailing/pkg/websocket"
	"github.com/redis/go-redis/v9"
)

// UpdateDriverLocation handles POST /v1/drivers/:id/location
func (h *Handlers) UpdateDriverLocation(c *gin.Context) {
	driverID := c.Param("id")
	ctx := context.Background()

	var req dto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	h.Logger.Info("Driver location update",
		logger.String("driver_id", driverID),
		logger.Float64("latitude", req.Latitude),
		logger.Float64("longitude", req.Longitude),
	)

	// Update Redis geo-spatial index for fast lookups
	_, err := h.Redis.GeoAdd(ctx, "drivers:locations", &redis.GeoLocation{
		Name:      driverID,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
	}).Result()

	if err != nil {
		h.Logger.Error("Failed to update Redis location", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update location"})
		return
	}

	// Also update PostgreSQL (debounced in production)
	_, err = h.DB.ExecContext(ctx, `
		UPDATE drivers
		SET current_latitude = $1,
		    current_longitude = $2,
		    updated_at = NOW()
		WHERE id = $3
	`, req.Latitude, req.Longitude, driverID)

	if err != nil {
		h.Logger.Warn("Failed to update PostgreSQL location", logger.Err(err))
		// Don't fail the request - Redis is more critical
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"driver_id": driverID,
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
		"timestamp": time.Now().UTC(),
	})
}

// AcceptRide handles POST /v1/drivers/:id/accept
func (h *Handlers) AcceptRide(c *gin.Context) {
	driverID := c.Param("id")

	var req dto.AcceptRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	h.Logger.Info("Driver accepting ride",
		logger.String("driver_id", driverID),
		logger.String("ride_id", req.RideID),
	)

	// Store current ride in Redis
	ctx := context.Background()
	currentRideKey := fmt.Sprintf("driver:%s:current_ride", driverID)
	// Store with 24 hour expiry (in case trip never completes, auto-cleanup)
	h.Redis.Set(ctx, currentRideKey, req.RideID, 24*time.Hour)
	h.Logger.Info("Stored current ride for driver", logger.String("driver_id", driverID), logger.String("ride_id", req.RideID))

	// Send notification to rider
	riderNotification := map[string]interface{}{
		"type": "ride_accepted",
		"data": map[string]interface{}{
			"ride_id":   req.RideID,
			"driver_id": driverID,
			"status":    "accepted",
			"message":   "Driver is on the way!",
			"eta":       "5 mins",
		},
	}

	// Broadcast to all riders (in production, send to specific rider)
	if wsHub, ok := h.Hub.(*websocket.Hub); ok {
		wsHub.BroadcastToType("rider", riderNotification)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "accepted",
		"ride_id": req.RideID,
		"message": "Ride accepted successfully",
	})
}

// GetRandomDriver handles GET /v1/drivers/random (for testing)
func (h *Handlers) GetRandomDriver(c *gin.Context) {
	ctx := context.Background()

	// Get a random online driver
	var driverID, name string
	var rating float64
	var latitude, longitude *float64

	err := h.DB.QueryRowContext(ctx, `
		SELECT id, name, rating, current_latitude, current_longitude
		FROM drivers
		WHERE status = 'online'
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&driverID, &name, &rating, &latitude, &longitude)

	if err != nil {
		h.Logger.Error("Failed to get random driver", logger.Err(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "No drivers available"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        driverID,
		"name":      name,
		"rating":    rating,
		"latitude":  latitude,
		"longitude": longitude,
	})
}

// GetAllDrivers handles GET /v1/drivers/all
func (h *Handlers) GetAllDrivers(c *gin.Context) {
	ctx := context.Background()

	// Query all drivers with earnings
	rows, err := h.DB.QueryContext(ctx, `
		SELECT
			d.id,
			d.name,
			d.phone,
			d.status,
			d.vehicle_type,
			d.rating,
			d.current_latitude,
			d.current_longitude,
			COALESCE(SUM(de.total_earnings), 0) as total_earnings,
			COUNT(r.id) as total_rides
		FROM drivers d
		LEFT JOIN driver_earnings de ON d.id = de.driver_id
		LEFT JOIN rides r ON d.id = r.driver_id AND r.status = 'completed'
		GROUP BY d.id, d.name, d.phone, d.status, d.vehicle_type, d.rating, d.current_latitude, d.current_longitude
		ORDER BY d.name
	`)

	if err != nil {
		h.Logger.Error("Failed to query drivers", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get drivers"})
		return
	}
	defer rows.Close()

	var drivers []gin.H
	for rows.Next() {
		var (
			id, name, phone, status, vehicleType string
			rating, totalEarnings                float64
			latitude, longitude                  *float64
			totalRides                           int
		)

		if err := rows.Scan(&id, &name, &phone, &status, &vehicleType, &rating,
			&latitude, &longitude, &totalEarnings, &totalRides); err != nil {
			h.Logger.Error("Failed to scan driver row", logger.Err(err))
			continue
		}

		// Get current ride from Redis
		currentRideKey := fmt.Sprintf("driver:%s:current_ride", id)
		currentRide, _ := h.Redis.Get(ctx, currentRideKey).Result()

		driver := gin.H{
			"id":             id,
			"name":           name,
			"phone":          phone,
			"status":         status,
			"vehicle_type":   vehicleType,
			"rating":         rating,
			"latitude":       latitude,
			"longitude":      longitude,
			"total_earnings": totalEarnings,
			"total_rides":    totalRides,
			"current_ride":   currentRide,
		}

		drivers = append(drivers, driver)
	}

	// Get fleet statistics
	var onlineCount, busyCount, offlineCount int
	h.DB.QueryRowContext(ctx, `
		SELECT
			COUNT(CASE WHEN status = 'online' THEN 1 END) as online,
			COUNT(CASE WHEN status = 'busy' THEN 1 END) as busy,
			COUNT(CASE WHEN status = 'offline' THEN 1 END) as offline
		FROM drivers
	`).Scan(&onlineCount, &busyCount, &offlineCount)

	// Get active rides count
	var activeRides int
	h.DB.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM rides WHERE status IN ('requested', 'assigned', 'accepted', 'started')
	`).Scan(&activeRides)

	// Get total earnings today
	var todayEarnings float64
	h.DB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_earnings), 0)
		FROM driver_earnings
		WHERE date = CURRENT_DATE
	`).Scan(&todayEarnings)

	c.JSON(http.StatusOK, gin.H{
		"drivers": drivers,
		"overview": gin.H{
			"total_drivers":  len(drivers),
			"online":         onlineCount,
			"busy":           busyCount,
			"offline":        offlineCount,
			"active_rides":   activeRides,
			"today_earnings": todayEarnings,
		},
	})
}
