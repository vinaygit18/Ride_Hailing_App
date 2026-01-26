package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocomet/ride-hailing/pkg/logger"
)

// GetRandomRider handles GET /v1/riders/random (for testing)
func (h *Handlers) GetRandomRider(c *gin.Context) {
	ctx := context.Background()

	// Get a random rider
	var riderID, name, email string
	var rating float64

	err := h.DB.QueryRowContext(ctx, `
		SELECT id, name, email, rating
		FROM riders
		ORDER BY RANDOM()
		LIMIT 1
	`).Scan(&riderID, &name, &email, &rating)

	if err != nil {
		h.Logger.Error("Failed to get random rider", logger.Err(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "No riders available"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     riderID,
		"name":   name,
		"email":  email,
		"rating": rating,
	})
}
