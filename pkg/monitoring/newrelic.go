package monitoring

import (
	"fmt"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// Config holds New Relic configuration
type Config struct {
	LicenseKey string
	AppName    string
	Enabled    bool
	LogLevel   string
}

// NewRelicApp wraps the New Relic application
type NewRelicApp struct {
	*newrelic.Application
	enabled bool
}

// New creates a new New Relic application
func New(cfg Config) (*NewRelicApp, error) {
	if !cfg.Enabled || cfg.LicenseKey == "" {
		// Return disabled app
		return &NewRelicApp{nil, false}, nil
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.AppName),
		newrelic.ConfigLicense(cfg.LicenseKey),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigDistributedTracerEnabled(true),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create New Relic application: %w", err)
	}

	return &NewRelicApp{app, true}, nil
}

// StartTransaction starts a new transaction
func (nr *NewRelicApp) StartTransaction(name string) *newrelic.Transaction {
	if !nr.enabled || nr.Application == nil {
		return nil
	}
	return nr.Application.StartTransaction(name)
}

// RecordCustomEvent records a custom event
func (nr *NewRelicApp) RecordCustomEvent(eventType string, params map[string]interface{}) {
	if !nr.enabled || nr.Application == nil {
		return
	}
	nr.Application.RecordCustomEvent(eventType, params)
}

// RecordCustomMetric records a custom metric
func (nr *NewRelicApp) RecordCustomMetric(name string, value float64) {
	if !nr.enabled || nr.Application == nil {
		return
	}
	nr.Application.RecordCustomMetric(name, value)
}

// Shutdown gracefully shuts down the New Relic application
func (nr *NewRelicApp) Shutdown(timeout time.Duration) {
	if !nr.enabled || nr.Application == nil {
		return
	}
	nr.Application.Shutdown(timeout)
}

// Custom metric helpers

// RecordMatchingLatency records driver matching latency
func (nr *NewRelicApp) RecordMatchingLatency(latencyMs float64) {
	nr.RecordCustomMetric("custom/ride/matching_latency_ms", latencyMs)
}

// RecordLocationUpdate records driver location update
func (nr *NewRelicApp) RecordLocationUpdate() {
	nr.RecordCustomMetric("custom/driver/location_update", 1)
}

// RecordRideCreated records ride creation
func (nr *NewRelicApp) RecordRideCreated(vehicleType string) {
	nr.RecordCustomEvent("RideCreated", map[string]interface{}{
		"vehicle_type": vehicleType,
		"timestamp":    time.Now().Unix(),
	})
}

// RecordRideCompleted records ride completion
func (nr *NewRelicApp) RecordRideCompleted(rideID string, fare float64, distance float64, duration int) {
	nr.RecordCustomEvent("RideCompleted", map[string]interface{}{
		"ride_id":  rideID,
		"fare":     fare,
		"distance": distance,
		"duration": duration,
	})
}

// RecordPaymentProcessed records payment processing
func (nr *NewRelicApp) RecordPaymentProcessed(amount float64, method string, status string) {
	nr.RecordCustomEvent("PaymentProcessed", map[string]interface{}{
		"amount": amount,
		"method": method,
		"status": status,
	})
}

// RecordSurgeMultiplier records surge pricing multiplier
func (nr *NewRelicApp) RecordSurgeMultiplier(region string, multiplier float64) {
	nr.RecordCustomMetric(fmt.Sprintf("custom/pricing/surge_multiplier/%s", region), multiplier)
}

// RecordDatabasePoolStats records database connection pool statistics
func (nr *NewRelicApp) RecordDatabasePoolStats(stats map[string]interface{}) {
	if totalConns, ok := stats["total_connections"].(int32); ok {
		nr.RecordCustomMetric("custom/db/total_connections", float64(totalConns))
	}
	if idleConns, ok := stats["idle_connections"].(int32); ok {
		nr.RecordCustomMetric("custom/db/idle_connections", float64(idleConns))
	}
	if acquiredConns, ok := stats["acquired_connections"].(int32); ok {
		nr.RecordCustomMetric("custom/db/acquired_connections", float64(acquiredConns))
	}
}

// RecordRedisPoolStats records Redis pool statistics
func (nr *NewRelicApp) RecordRedisPoolStats(stats map[string]interface{}) {
	if hits, ok := stats["hits"].(uint32); ok {
		nr.RecordCustomMetric("custom/redis/cache_hits", float64(hits))
	}
	if misses, ok := stats["misses"].(uint32); ok {
		nr.RecordCustomMetric("custom/redis/cache_misses", float64(misses))
	}
	if timeouts, ok := stats["timeouts"].(uint32); ok {
		nr.RecordCustomMetric("custom/redis/timeouts", float64(timeouts))
	}
}

// IsEnabled returns whether New Relic is enabled
func (nr *NewRelicApp) IsEnabled() bool {
	return nr.enabled
}
