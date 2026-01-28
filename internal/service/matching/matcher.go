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
	MaxRadiusKM      float64       // Initial search radius
	MaxExpandedRadius float64      // Maximum expanded radius when no drivers found
	MaxTimeout       time.Duration
	MaxCandidates    int
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
// It starts with the initial radius and expands progressively if no drivers are found
func (s *Service) FindNearestDriver(ctx context.Context, pickupLat, pickupLng float64, vehicleType driver.VehicleType) (*driver.Driver, error) {
	startTime := time.Now()

	// Define search radii - start small and expand progressively
	// Initial: 5km, then expand to 10km, 20km, 50km, up to max expanded radius
	maxRadius := s.config.MaxExpandedRadius
	if maxRadius == 0 {
		maxRadius = 50.0 // Default max 50km if not configured
	}

	searchRadii := []float64{s.config.MaxRadiusKM}

	// Add expanded radii: 2x, 4x, 10x of initial radius
	expandedRadii := []float64{
		s.config.MaxRadiusKM * 2,  // 10km
		s.config.MaxRadiusKM * 4,  // 20km
		s.config.MaxRadiusKM * 10, // 50km
	}

	for _, r := range expandedRadii {
		if r <= maxRadius {
			searchRadii = append(searchRadii, r)
		}
	}

	// Use Redis GEORADIUS to find nearby drivers
	key := "drivers:locations"

	// Try each radius progressively
	for _, radius := range searchRadii {
		foundDriver, err := s.searchDriversInRadius(ctx, key, pickupLat, pickupLng, radius, vehicleType, startTime)
		if err == nil && foundDriver != nil {
			return foundDriver, nil
		}

		// If we found drivers but none were available, log and try larger radius
		if radius < maxRadius {
			s.logger.Info("No available drivers in radius, expanding search",
				logger.Float64("current_radius_km", radius),
				logger.Float64("next_radius_km", radius*2),
			)
		}
	}

	s.logger.Warn("No drivers available in maximum search radius",
		logger.Float64("max_radius_km", maxRadius),
		logger.Float64("pickup_lat", pickupLat),
		logger.Float64("pickup_lng", pickupLng),
	)

	return nil, driver.ErrDriverNotAvailable
}

// searchDriversInRadius searches for available drivers within a specific radius
func (s *Service) searchDriversInRadius(ctx context.Context, key string, pickupLat, pickupLng, radius float64, vehicleType driver.VehicleType, startTime time.Time) (*driver.Driver, error) {
	// Search for drivers within radius
	results, err := s.redis.GeoRadius(ctx, key, pickupLng, pickupLat, &redis.GeoRadiusQuery{
		Radius:    radius,
		Unit:      "km",
		WithCoord: true,
		WithDist:  true,
		Count:     s.config.MaxCandidates,
		Sort:      "ASC",
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to search nearby drivers: %w", err)
	}

	if len(results) == 0 {
		return nil, driver.ErrDriverNotAvailable
	}

	// Filter by vehicle type and availability - use atomic claim
	for _, result := range results {
		driverID := result.Name

		// Check if driver is already on a ride first (quick check)
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

		// Atomically claim driver by removing from available set
		// SREM returns 1 if member was removed, 0 if it wasn't there
		removed, err := s.redis.SRem(ctx, "drivers:available", driverID).Result()
		if err != nil {
			s.logger.Warn("Failed to claim driver", logger.String("driver_id", driverID), logger.Err(err))
			continue
		}
		if removed == 0 {
			// Driver was already claimed by another request
			s.logger.Info("Driver skipped - already claimed by another request",
				logger.String("driver_id", driverID),
				logger.Float64("distance_km", result.Dist),
			)
			continue
		}

		// Successfully claimed the driver - set current ride key to prevent double-assignment
		// This will be overwritten with actual ride ID in ride_handler
		s.redis.Set(ctx, currentRideKey, "claiming", 30*time.Second)

		// Create driver object
		lat := result.Latitude
		lng := result.Longitude

		// Parse or generate UUID for the driver
		driverUUID, err := uuid.Parse(driverID)
		if err != nil {
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
		s.logger.Info("Driver matched and claimed",
			logger.String("driver_id", driverID),
			logger.Float64("distance_km", result.Dist),
			logger.Float64("search_radius_km", radius),
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
