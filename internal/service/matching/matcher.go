package matching

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/gocomet/ride-hailing/internal/domain/driver"
	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Service handles driver-rider matching
type Service struct {
	redis  *redis.Client
	logger *logger.Logger
	config Config
}

// Config holds matching configuration
type Config struct {
	MaxRadiusKM   float64
	MaxTimeout    time.Duration
	MaxCandidates int
}

// DriverCandidate represents a nearby driver
type DriverCandidate struct {
	Driver   *driver.Driver
	Distance float64
}

// NewService creates a new matching service
func NewService(redis *redis.Client, logger *logger.Logger, config Config) *Service {
	return &Service{
		redis:  redis,
		logger: logger,
		config: config,
	}
}

// FindNearestDriver finds the nearest available driver
func (s *Service) FindNearestDriver(ctx context.Context, pickupLat, pickupLng float64, vehicleType driver.VehicleType) (*driver.Driver, error) {
	startTime := time.Now()

	// Use Redis GEORADIUS to find nearby drivers
	key := "drivers:locations"

	// Search for drivers within radius
	results, err := s.redis.GeoRadius(ctx, key, pickupLng, pickupLat, &redis.GeoRadiusQuery{
		Radius:      s.config.MaxRadiusKM,
		Unit:        "km",
		WithCoord:   true,
		WithDist:    true,
		Count:       s.config.MaxCandidates,
		Sort:        "ASC",
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to search nearby drivers: %w", err)
	}

	if len(results) == 0 {
		return nil, driver.ErrDriverNotAvailable
	}

	// Filter by vehicle type and availability (would normally query DB here)
	// For MVP, we'll use the first available driver
	for _, result := range results {
		driverID := result.Name

		// Check if driver is available in the available set
		isAvailable, err := s.redis.SIsMember(ctx, "drivers:available", driverID).Result()
		if err != nil || !isAvailable {
			continue
		}

		// Check if driver is already on a ride
		currentRideKey := fmt.Sprintf("driver:%s:current_ride", driverID)
		currentRide, err := s.redis.Get(ctx, currentRideKey).Result()
		if err == nil && currentRide != "" {
			// Driver is already on a ride, skip to next nearest driver
			s.logger.Info("Driver skipped - already on ride",
				logger.String("driver_id", driverID),
				logger.String("current_ride", currentRide),
				logger.Float64("distance_km", result.Dist),
			)
			continue
		}

		// Check vehicle type matches (simplified - would query DB in production)
		// For MVP, we'll create a mock driver
		lat := result.Latitude
		lng := result.Longitude

		// Parse or generate UUID for the driver
		driverUUID, err := uuid.Parse(driverID)
		if err != nil {
			// If not a valid UUID, generate a new one
			driverUUID = uuid.New()
		}

		foundDriver := &driver.Driver{
			ID:               driverUUID,
			Name:             "Driver " + driverID[:8],
			Status:           driver.StatusOnline,
			VehicleType:      vehicleType,
			CurrentLatitude:  &lat,
			CurrentLongitude: &lng,
			Rating:           4.8,
		}

		elapsed := time.Since(startTime).Milliseconds()
		s.logger.Info("Driver matched",
			logger.String("driver_id", driverID),
			logger.Float64("distance_km", result.Dist),
			logger.Int64("latency_ms", elapsed),
		)

		return foundDriver, nil
	}

	return nil, driver.ErrDriverNotAvailable
}

// CalculateDistance calculates haversine distance between two points
func CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // kilometers

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}
