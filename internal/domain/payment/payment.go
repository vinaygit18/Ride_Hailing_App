package payment

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusRefunded   Status = "refunded"
)

type Method string

const (
	MethodCard   Method = "card"
	MethodWallet Method = "wallet"
	MethodCash   Method = "cash"
	MethodUPI    Method = "upi"
)

type Payment struct {
	ID                      uuid.UUID   `json:"id"`
	TripID                  uuid.UUID   `json:"trip_id"`
	Amount                  float64     `json:"amount"`
	Status                  Status      `json:"status"`
	PaymentMethod           Method      `json:"payment_method"`
	ExternalTransactionID   string      `json:"external_transaction_id,omitempty"`
	PaymentGatewayResponse  interface{} `json:"payment_gateway_response,omitempty"`
	FailureReason           string      `json:"failure_reason,omitempty"`
	IdempotencyKey          string      `json:"-"`
	ProcessedAt             *time.Time  `json:"processed_at,omitempty"`
	CreatedAt               time.Time   `json:"created_at"`
	UpdatedAt               time.Time   `json:"updated_at"`
}

type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id uuid.UUID) (*Payment, error)
	GetByTripID(ctx context.Context, tripID uuid.UUID) (*Payment, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*Payment, error)
	Update(ctx context.Context, payment *Payment) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status Status) error
}

var (
	ErrPaymentNotFound = errors.New("payment not found")
	ErrPaymentFailed   = errors.New("payment failed")
)
