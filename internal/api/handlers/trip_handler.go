package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocomet/ride-hailing/internal/api/dto"
	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/gocomet/ride-hailing/pkg/websocket"
)

// EndTrip handles POST /v1/trips/:id/end
func (h *Handlers) EndTrip(c *gin.Context) {
	rideID := c.Param("id")

	var req dto.EndTripRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	h.Logger.Info("Ending trip",
		logger.String("ride_id", rideID),
		logger.String("driver_id", req.DriverID),
		logger.Float64("distance_km", req.DistanceKm),
		logger.Int("duration_minutes", req.DurationMinutes),
	)

	// Calculate fare (simplified pricing)
	baseFare := 50.0
	perKmFare := 10.0
	perMinuteFare := 2.0

	distanceFare := req.DistanceKm * perKmFare
	timeFare := float64(req.DurationMinutes) * perMinuteFare
	totalFare := baseFare + distanceFare + timeFare

	h.Logger.Info("Fare calculated",
		logger.Float64("total_fare", totalFare),
		logger.Float64("base_fare", baseFare),
		logger.Float64("distance_fare", distanceFare),
		logger.Float64("time_fare", timeFare),
	)

	ctx := context.Background()

	// Start PostgreSQL transaction
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		h.Logger.Error("Failed to begin transaction", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	// Update ride status to completed
	_, err = tx.ExecContext(ctx, `
		UPDATE rides
		SET status = 'completed', completed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, rideID)
	if err != nil {
		h.Logger.Error("Failed to update ride", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ride"})
		return
	}

	// Create or update trip record
	_, err = tx.ExecContext(ctx, `
		INSERT INTO trips (
			ride_id, distance_km, duration_minutes,
			base_fare, distance_fare, time_fare, total_fare,
			status, ended_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'completed', NOW())
		ON CONFLICT (ride_id) DO UPDATE SET
			distance_km = EXCLUDED.distance_km,
			duration_minutes = EXCLUDED.duration_minutes,
			base_fare = EXCLUDED.base_fare,
			distance_fare = EXCLUDED.distance_fare,
			time_fare = EXCLUDED.time_fare,
			total_fare = EXCLUDED.total_fare,
			status = EXCLUDED.status,
			ended_at = EXCLUDED.ended_at,
			updated_at = NOW()
	`, rideID, req.DistanceKm, req.DurationMinutes, baseFare, distanceFare, timeFare, totalFare)
	if err != nil {
		h.Logger.Error("Failed to create/update trip", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save trip"})
		return
	}

	// Update driver earnings (UPSERT into driver_earnings table)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO driver_earnings (driver_id, date, total_rides, total_earnings)
		VALUES ($1, CURRENT_DATE, 1, $2)
		ON CONFLICT (driver_id, date) DO UPDATE SET
			total_rides = driver_earnings.total_rides + 1,
			total_earnings = driver_earnings.total_earnings + $2,
			updated_at = NOW()
	`, req.DriverID, totalFare)
	if err != nil {
		h.Logger.Error("Failed to update driver earnings", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update earnings"})
		return
	}

	// Update driver status back to online (no longer busy)
	_, err = tx.ExecContext(ctx, `
		UPDATE drivers
		SET status = 'online', updated_at = NOW()
		WHERE id = $1
	`, req.DriverID)
	if err != nil {
		h.Logger.Warn("Failed to update driver status", logger.Err(err))
		// Don't fail the request, just log
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		h.Logger.Error("Failed to commit transaction", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete trip"})
		return
	}

	h.Logger.Info("Trip completed in PostgreSQL",
		logger.String("ride_id", rideID),
		logger.String("driver_id", req.DriverID),
		logger.Float64("fare", totalFare),
	)

	// Clear current ride from Redis (temporary state)
	currentRideKey := fmt.Sprintf("driver:%s:current_ride", req.DriverID)
	h.Redis.Del(ctx, currentRideKey)
	h.Logger.Info("Cleared current ride from Redis", logger.String("driver_id", req.DriverID))

	// Get driver name from PostgreSQL
	var driverName string
	err = h.DB.QueryRowContext(ctx, "SELECT name FROM drivers WHERE id = $1", req.DriverID).Scan(&driverName)
	if err != nil {
		driverName = fmt.Sprintf("Driver %s", req.DriverID[:8])
	}

	// Send notification to dashboard
	tripCompletedNotification := map[string]interface{}{
		"type": "trip_completed",
		"data": map[string]interface{}{
			"ride_id":          rideID,
			"driver_id":        req.DriverID,
			"driver_name":      driverName,
			"distance_km":      req.DistanceKm,
			"duration_minutes": req.DurationMinutes,
			"total_fare":       totalFare,
			"fare":             totalFare,
		},
	}
	if wsHub, ok := h.Hub.(*websocket.Hub); ok {
		wsHub.BroadcastToType("dashboard", tripCompletedNotification)
	}

	// Also notify rider
	riderNotification := map[string]interface{}{
		"type": "trip_completed",
		"data": map[string]interface{}{
			"ride_id":     rideID,
			"status":      "completed",
			"total_fare":  totalFare,
			"distance_km": req.DistanceKm,
			"duration":    req.DurationMinutes,
		},
	}
	if wsHub, ok := h.Hub.(*websocket.Hub); ok {
		wsHub.BroadcastToType("rider", riderNotification)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "completed",
		"ride_id":          rideID,
		"total_fare":       totalFare,
		"fare":             totalFare,
		"distance_km":      req.DistanceKm,
		"duration_minutes": req.DurationMinutes,
		"fare_breakdown": map[string]interface{}{
			"base_fare":     baseFare,
			"distance_fare": req.DistanceKm * perKmFare,
			"time_fare":     float64(req.DurationMinutes) * perMinuteFare,
		},
	})
}
