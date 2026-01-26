package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError
func NewAppError(code, message string, status int, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
	}
}

// Common error constructors

// BadRequest creates a 400 error
func BadRequest(message string, err error) *AppError {
	return &AppError{
		Code:    "BAD_REQUEST",
		Message: message,
		Status:  http.StatusBadRequest,
		Err:     err,
	}
}

// Unauthorized creates a 401 error
func Unauthorized(message string, err error) *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Status:  http.StatusUnauthorized,
		Err:     err,
	}
}

// Forbidden creates a 403 error
func Forbidden(message string, err error) *AppError {
	return &AppError{
		Code:    "FORBIDDEN",
		Message: message,
		Status:  http.StatusForbidden,
		Err:     err,
	}
}

// NotFound creates a 404 error
func NotFound(message string, err error) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: message,
		Status:  http.StatusNotFound,
		Err:     err,
	}
}

// Conflict creates a 409 error
func Conflict(message string, err error) *AppError {
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Status:  http.StatusConflict,
		Err:     err,
	}
}

// Internal creates a 500 error
func Internal(message string, err error) *AppError {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

// ServiceUnavailable creates a 503 error
func ServiceUnavailable(message string, err error) *AppError {
	return &AppError{
		Code:    "SERVICE_UNAVAILABLE",
		Message: message,
		Status:  http.StatusServiceUnavailable,
		Err:     err,
	}
}

// Domain-specific errors

var (
	ErrDriverNotFound      = NotFound("Driver not found", nil)
	ErrRiderNotFound       = NotFound("Rider not found", nil)
	ErrRideNotFound        = NotFound("Ride not found", nil)
	ErrTripNotFound        = NotFound("Trip not found", nil)
	ErrPaymentNotFound     = NotFound("Payment not found", nil)

	ErrNoDriversAvailable  = NotFound("No drivers available in the area", nil)
	ErrDriverNotAvailable  = Conflict("Driver is not available", nil)
	ErrRideAlreadyAssigned = Conflict("Ride is already assigned to a driver", nil)
	ErrTripAlreadyCompleted = Conflict("Trip is already completed", nil)

	ErrInvalidStatus       = BadRequest("Invalid status transition", nil)
	ErrInvalidCoordinates  = BadRequest("Invalid coordinates", nil)
	ErrInvalidVehicleType  = BadRequest("Invalid vehicle type", nil)
	ErrInvalidPaymentMethod = BadRequest("Invalid payment method", nil)

	ErrDuplicateRequest    = Conflict("Duplicate request detected", nil)
	ErrRateLimitExceeded   = &AppError{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "Rate limit exceeded. Please try again later",
		Status:  http.StatusTooManyRequests,
	}
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError attempts to convert an error to AppError
func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	// Return generic internal error if not an AppError
	return Internal("An unexpected error occurred", err)
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapAppError wraps an AppError with additional context
func WrapAppError(appErr *AppError, message string) *AppError {
	if appErr == nil {
		return nil
	}
	return &AppError{
		Code:    appErr.Code,
		Message: fmt.Sprintf("%s: %s", message, appErr.Message),
		Status:  appErr.Status,
		Err:     appErr.Err,
	}
}
