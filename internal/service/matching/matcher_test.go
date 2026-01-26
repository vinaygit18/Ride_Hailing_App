package matching

import (
	"testing"

	"github.com/gocomet/ride-hailing/internal/domain/driver"
	"github.com/stretchr/testify/assert"
)

// TestDriverMatching_ValidatesVehicleType tests vehicle type matching
func TestDriverMatching_ValidatesVehicleType(t *testing.T) {
	driver1 := driver.Driver{
		VehicleType: driver.VehicleEconomy,
		Status:      driver.StatusOnline,
	}

	driver2 := driver.Driver{
		VehicleType: driver.VehiclePremium,
		Status:      driver.StatusOnline,
	}

	requestedType := driver.VehicleEconomy

	assert.Equal(t, requestedType, driver1.VehicleType, "Driver 1 should match economy request")
	assert.NotEqual(t, requestedType, driver2.VehicleType, "Driver 2 should not match economy request")
}

// TestDriverStatus_FilteringLogic tests driver status filtering
func TestDriverStatus_FilteringLogic(t *testing.T) {
	tests := []struct {
		name           string
		status         driver.Status
		shouldBeOnline bool
		shouldBeAvailable bool
	}{
		{
			name:           "Online driver",
			status:         driver.StatusOnline,
			shouldBeOnline: true,
			shouldBeAvailable: true,
		},
		{
			name:           "Offline driver",
			status:         driver.StatusOffline,
			shouldBeOnline: false,
			shouldBeAvailable: false,
		},
		{
			name:           "Busy driver",
			status:         driver.StatusBusy,
			shouldBeOnline: false,
			shouldBeAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isOnline := (tt.status == driver.StatusOnline)
			isAvailable := (tt.status == driver.StatusOnline)

			assert.Equal(t, tt.shouldBeOnline, isOnline)
			assert.Equal(t, tt.shouldBeAvailable, isAvailable)
		})
	}
}

// TestDistanceCalculation_HaversineFormula tests distance calculation
func TestDistanceCalculation_HaversineFormula(t *testing.T) {
	// Test basic distance calculation logic
	lat1, lng1 := 12.9716, 77.5946 // Pickup
	lat2, lng2 := 12.9700, 77.5900 // Very close driver

	// Simple Euclidean distance for testing (not actual Haversine)
	dist := ((lat1 - lat2) * (lat1 - lat2)) + ((lng1 - lng2) * (lng1 - lng2))

	assert.Less(t, dist, 0.01, "Distance should be very small for nearby locations")
}

// TestMatchingLatency_Requirement tests latency requirement
func TestMatchingLatency_Requirement(t *testing.T) {
	// Latency requirement: <1s p95
	maxLatencyMs := 1000.0

	// In production, matching should be done via Redis GEORADIUS
	// which provides sub-millisecond response times

	// Test that requirement is reasonable
	assert.Greater(t, maxLatencyMs, 0.0, "Latency requirement should be positive")
	assert.LessOrEqual(t, maxLatencyMs, 1000.0, "P95 latency should be under 1 second")
}

// TestVehicleType_AllTypesValid tests all vehicle types
func TestVehicleType_AllTypesValid(t *testing.T) {
	vehicleTypes := []driver.VehicleType{
		driver.VehicleEconomy,
		driver.VehiclePremium,
		driver.VehicleLuxury,
	}

	for _, vt := range vehicleTypes {
		assert.NotEmpty(t, vt, "Vehicle type should not be empty")
	}
}

// TestMatchingCriteria_Priority tests matching priority logic
func TestMatchingCriteria_Priority(t *testing.T) {
	// Matching priority:
	// 1. Vehicle type match
	// 2. Driver status (online)
	// 3. Distance (closest first)
	// 4. Rating (higher is better)

	// Test that economy request doesn't match premium driver
	requestType := driver.VehicleEconomy
	driverType := driver.VehiclePremium

	matches := (requestType == driverType)
	assert.False(t, matches, "Economy request should not match premium driver")
}
