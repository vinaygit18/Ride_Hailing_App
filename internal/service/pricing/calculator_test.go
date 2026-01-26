package pricing

import (
	"testing"

	"github.com/gocomet/ride-hailing/internal/domain/driver"
	"github.com/stretchr/testify/assert"
)

// getTestConfig returns a test configuration
func getTestConfig() Config {
	return Config{
		BaseFare: map[driver.VehicleType]float64{
			driver.VehicleEconomy: 50.0,
			driver.VehiclePremium: 100.0,
			driver.VehicleLuxury:  200.0,
		},
		PerKMRate: map[driver.VehicleType]float64{
			driver.VehicleEconomy: 10.0,
			driver.VehiclePremium: 15.0,
			driver.VehicleLuxury:  25.0,
		},
		PerMinuteRate: map[driver.VehicleType]float64{
			driver.VehicleEconomy: 2.0,
			driver.VehiclePremium: 3.0,
			driver.VehicleLuxury:  5.0,
		},
		MaxSurgeMultiplier: 3.0,
		MinSurgeMultiplier: 1.0,
	}
}

// TestEstimateFare_BaseCalculation tests basic fare estimation
func TestEstimateFare_BaseCalculation(t *testing.T) {
	service := &Service{config: getTestConfig()}

	tests := []struct {
		name        string
		vehicleType driver.VehicleType
		distanceKm  float64
		durationMin int
		expected    float64
	}{
		{
			name:        "Economy 10km 20min",
			vehicleType: driver.VehicleEconomy,
			distanceKm:  10.0,
			durationMin: 20,
			expected:    190.0, // 50 + (10*10) + (20*2)
		},
		{
			name:        "Premium 15km 30min",
			vehicleType: driver.VehiclePremium,
			distanceKm:  15.0,
			durationMin: 30,
			expected:    415.0, // 100 + (15*15) + (30*3)
		},
		{
			name:        "Luxury 20km 45min",
			vehicleType: driver.VehicleLuxury,
			distanceKm:  20.0,
			durationMin: 45,
			expected:    925.0, // 200 + (20*25) + (45*5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fare := service.EstimateFare(tt.vehicleType, tt.distanceKm, tt.durationMin)
			assert.Equal(t, tt.expected, fare, "Fare should match expected value")
		})
	}
}

// TestEstimateFare_MinimumFare tests minimum fare is enforced
func TestEstimateFare_MinimumFare(t *testing.T) {
	service := &Service{config: getTestConfig()}

	// Very short trip - should still have base fare
	fare := service.EstimateFare(driver.VehicleEconomy, 0.5, 2)

	assert.GreaterOrEqual(t, fare, 50.0, "Fare should be at least the base fare")
}

// TestEstimateFare_ZeroDistance tests edge case of zero distance
func TestEstimateFare_ZeroDistance(t *testing.T) {
	service := &Service{config: getTestConfig()}

	fare := service.EstimateFare(driver.VehicleEconomy, 0, 10)

	expected := 70.0 // 50 + (10*2)
	assert.Equal(t, expected, fare, "Zero distance should charge base + time")
}

// TestEstimateFare_DifferentVehicleTypes tests all vehicle types
func TestEstimateFare_DifferentVehicleTypes(t *testing.T) {
	service := &Service{config: getTestConfig()}

	distanceKm := 10.0
	durationMin := 20

	economyFare := service.EstimateFare(driver.VehicleEconomy, distanceKm, durationMin)
	premiumFare := service.EstimateFare(driver.VehiclePremium, distanceKm, durationMin)
	luxuryFare := service.EstimateFare(driver.VehicleLuxury, distanceKm, durationMin)

	assert.Less(t, economyFare, premiumFare, "Economy should be cheaper than Premium")
	assert.Less(t, premiumFare, luxuryFare, "Premium should be cheaper than Luxury")
}

// TestSurgeCalculation_DemandSupplyRatio tests surge calculation
func TestSurgeCalculation_DemandSupplyRatio(t *testing.T) {
	service := &Service{config: getTestConfig()}

	tests := []struct {
		name             string
		activeRides      int
		availableDrivers int
		expectedMin      float64
		expectedMax      float64
	}{
		{
			name:             "Low demand",
			activeRides:      5,
			availableDrivers: 20,
			expectedMin:      1.0,
			expectedMax:      1.0,
		},
		{
			name:             "Moderate demand",
			activeRides:      15,
			availableDrivers: 20,
			expectedMin:      1.0,
			expectedMax:      1.5,
		},
		{
			name:             "High demand",
			activeRides:      40,
			availableDrivers: 20,
			expectedMin:      1.5,
			expectedMax:      2.5,
		},
		{
			name:             "Very high demand",
			activeRides:      100,
			availableDrivers: 10,
			expectedMin:      2.5,
			expectedMax:      3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			surge := service.CalculateSurgeBasedOnDemand(tt.activeRides, tt.availableDrivers)

			assert.GreaterOrEqual(t, surge, tt.expectedMin)
			assert.LessOrEqual(t, surge, tt.expectedMax)
			assert.LessOrEqual(t, surge, 3.0, "Surge should never exceed max")
		})
	}
}

// TestSurgeCalculation_NoDrivers tests surge when no drivers
func TestSurgeCalculation_NoDrivers(t *testing.T) {
	service := &Service{config: getTestConfig()}

	surge := service.CalculateSurgeBasedOnDemand(50, 0)

	assert.Equal(t, 3.0, surge, "Surge should be max when no drivers")
}

// BenchmarkEstimateFare benchmarks fare calculation
func BenchmarkEstimateFare(b *testing.B) {
	service := &Service{config: getTestConfig()}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.EstimateFare(driver.VehicleEconomy, 10.0, 20)
	}
}
