package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/gocomet/ride-hailing/internal/api/handlers"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine, h *handlers.Handlers, nrApp *newrelic.Application) {
	// Add New Relic middleware if enabled
	if nrApp != nil {
		r.Use(nrgin.Middleware(nrApp))
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API v1 routes
	v1 := r.Group("/v1")
	{
		// WebSocket connection
		v1.GET("/ws", h.HandleWebSocket)

		// Ride endpoints
		rides := v1.Group("/rides")
		{
			rides.POST("", h.CreateRide)
			rides.GET("/:id", h.GetRide)
		}

		// Driver endpoints
		drivers := v1.Group("/drivers")
		{
			drivers.GET("/all", h.GetAllDrivers)
			drivers.GET("/random", h.GetRandomDriver)
			drivers.POST("/:id/location", h.UpdateDriverLocation)
			drivers.POST("/:id/accept", h.AcceptRide)
		}

		// Trip endpoints
		trips := v1.Group("/trips")
		{
			trips.POST("/:id/end", h.EndTrip)
		}

		// Payment endpoints
		v1.POST("/payments", h.ProcessPayment)

		// Rider endpoints (testing)
		riders := v1.Group("/riders")
		{
			riders.GET("/random", h.GetRandomRider)
		}
	}
}
