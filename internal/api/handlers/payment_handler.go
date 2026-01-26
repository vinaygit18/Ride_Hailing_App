package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gocomet/ride-hailing/internal/api/dto"
	"github.com/gocomet/ride-hailing/pkg/logger"
)

// ProcessPayment handles POST /v1/payments
func (h *Handlers) ProcessPayment(c *gin.Context) {
	ctx := context.Background()

	var req dto.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload", "details": err.Error()})
		return
	}

	// Check idempotency
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header required"})
		return
	}

	// Check if payment already processed
	cacheKey := fmt.Sprintf("payment:idempotency:%s", idempotencyKey)
	cachedResponse, err := h.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		h.Logger.Info("Returning cached payment response", logger.String("idempotency_key", idempotencyKey))
		var response map[string]interface{}
		if err := json.Unmarshal([]byte(cachedResponse), &response); err == nil {
			c.JSON(http.StatusOK, response)
			return
		}
	}

	h.Logger.Info("Processing payment",
		logger.String("trip_id", req.TripID),
		logger.Float64("amount", req.Amount),
		logger.String("payment_method", req.PaymentMethod),
	)

	// Validate trip exists and amount matches
	// req.TripID is actually the ride_id, get the actual trip UUID
	var tripAmount float64
	var tripUUID string
	err = h.DB.QueryRowContext(ctx, `
		SELECT id, total_fare
		FROM trips
		WHERE ride_id = $1 AND status = 'completed'
	`, req.TripID).Scan(&tripUUID, &tripAmount)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Trip not found or not completed"})
		return
	}

	if err != nil {
		h.Logger.Error("Failed to validate trip", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	if tripAmount != req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Amount mismatch",
			"expected": tripAmount,
			"provided": req.Amount,
		})
		return
	}

	// Generate external transaction ID (mock PSP)
	externalTransactionID := fmt.Sprintf("txn_%d_%s", time.Now().Unix(), generateRideID())

	// Mock PSP processing (simulate delay)
	time.Sleep(100 * time.Millisecond)

	// Insert payment record
	paymentID := uuid.New().String()
	_, err = h.DB.ExecContext(ctx, `
		INSERT INTO payments (
			id, trip_id, amount, status, payment_method,
			external_transaction_id, idempotency_key, created_at
		) VALUES ($1, $2, $3, 'completed', $4, $5, $6, NOW())
		ON CONFLICT (idempotency_key) DO UPDATE SET
			updated_at = NOW()
		RETURNING id
	`, paymentID, tripUUID, req.Amount, req.PaymentMethod, externalTransactionID, idempotencyKey)

	if err != nil {
		h.Logger.Error("Failed to create payment record", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process payment"})
		return
	}

	response := gin.H{
		"payment_id":     paymentID,
		"trip_id":        req.TripID,
		"amount":         req.Amount,
		"status":         "completed",
		"payment_method": req.PaymentMethod,
		"transaction_id": externalTransactionID,
		"processed_at":   time.Now(),
	}

	// Cache response for idempotency
	responseJSON, _ := json.Marshal(response)
	h.Redis.Set(ctx, cacheKey, responseJSON, 24*time.Hour)

	h.Logger.Info("Payment processed successfully",
		logger.String("payment_id", paymentID),
		logger.String("trip_id", req.TripID),
		logger.Float64("amount", req.Amount),
	)

	c.JSON(http.StatusOK, response)
}
