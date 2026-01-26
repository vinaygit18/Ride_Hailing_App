package pricing

import (
	"context"
	"fmt"

	"github.com/gocomet/ride-hailing/internal/domain/driver"
	"github.com/redis/go-redis/v9"
)

// Service handles fare calculation
type Service struct {
	redis  *redis.Client
	config Config
}

// Config holds pricing configuration
type Config struct {
	BaseFare map[driver.VehicleType]float64
	PerKMRate map[driver.VehicleType]float64
	PerMinuteRate map[driver.VehicleType]float64
	MaxSurgeMultiplier float64
	MinSurgeMultiplier float64
}

// FareBreakdown represents the breakdown of a fare
type FareBreakdown struct {
	BaseFare        float64 `json:"base_fare"`
	DistanceFare    float64 `json:"distance_fare"`
	TimeFare        float64 `json:"time_fare"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	Subtotal        float64 `json:"subtotal"`
	Total           float64 `json:"total"`
}

// NewService creates a new pricing service
func NewService(redis *redis.Client, config Config) *Service {
	return &Service{
		redis:  redis,
		config: config,
	}
}

// CalculateFare calculates the total fare for a trip
func (s *Service) CalculateFare(ctx context.Context, vehicleType driver.VehicleType, distanceKM float64, durationMinutes int, region string) (*FareBreakdown, error) {
	baseFare := s.config.BaseFare[vehicleType]
	perKM := s.config.PerKMRate[vehicleType]
	perMinute := s.config.PerMinuteRate[vehicleType]

	distanceFare := distanceKM * perKM
	timeFare := float64(durationMinutes) * perMinute
	subtotal := baseFare + distanceFare + timeFare

	// Get surge multiplier
	surgeMultiplier := s.GetSurgeMultiplier(ctx, region)

	total := subtotal * surgeMultiplier

	return &FareBreakdown{
		BaseFare:        baseFare,
		DistanceFare:    distanceFare,
		TimeFare:        timeFare,
		SurgeMultiplier: surgeMultiplier,
		Subtotal:        subtotal,
		Total:           total,
	}, nil
}

// EstimateFare estimates fare before trip starts
func (s *Service) EstimateFare(vehicleType driver.VehicleType, distanceKM float64, estimatedMinutes int) float64 {
	baseFare := s.config.BaseFare[vehicleType]
	perKM := s.config.PerKMRate[vehicleType]
	perMinute := s.config.PerMinuteRate[vehicleType]

	return baseFare + (distanceKM * perKM) + (float64(estimatedMinutes) * perMinute)
}

// GetSurgeMultiplier gets the current surge multiplier for a region
func (s *Service) GetSurgeMultiplier(ctx context.Context, region string) float64 {
	key := fmt.Sprintf("surge:%s", region)
	val, err := s.redis.Get(ctx, key).Float64()
	if err != nil {
		return 1.0 // Default no surge
	}

	if val > s.config.MaxSurgeMultiplier {
		return s.config.MaxSurgeMultiplier
	}
	if val < s.config.MinSurgeMultiplier {
		return s.config.MinSurgeMultiplier
	}

	return val
}

// SetSurgeMultiplier sets the surge multiplier for a region
func (s *Service) SetSurgeMultiplier(ctx context.Context, region string, multiplier float64) error {
	if multiplier > s.config.MaxSurgeMultiplier {
		multiplier = s.config.MaxSurgeMultiplier
	}
	if multiplier < s.config.MinSurgeMultiplier {
		multiplier = s.config.MinSurgeMultiplier
	}

	key := fmt.Sprintf("surge:%s", region)
	return s.redis.Set(ctx, key, multiplier, 0).Err()
}

// CalculateSurgeBasedOnDemand calculates surge based on demand/supply ratio
func (s *Service) CalculateSurgeBasedOnDemand(activeRides, availableDrivers int) float64 {
	if availableDrivers == 0 {
		return s.config.MaxSurgeMultiplier
	}

	ratio := float64(activeRides) / float64(availableDrivers)

	// Simple surge calculation
	// ratio < 0.5 -> 1.0x
	// ratio 0.5-1.0 -> 1.0-1.5x
	// ratio 1.0-2.0 -> 1.5-2.5x
	// ratio > 2.0 -> 2.5-3.0x

	if ratio < 0.5 {
		return 1.0
	} else if ratio < 1.0 {
		return 1.0 + (ratio * 0.5)
	} else if ratio < 2.0 {
		return 1.5 + ((ratio - 1.0) * 1.0)
	} else {
		multiplier := 2.5 + ((ratio - 2.0) * 0.25)
		if multiplier > s.config.MaxSurgeMultiplier {
			return s.config.MaxSurgeMultiplier
		}
		return multiplier
	}
}
