package driver

import "errors"

var (
	ErrDriverNotFound      = errors.New("driver not found")
	ErrInvalidDriverName   = errors.New("invalid driver name")
	ErrInvalidDriverEmail  = errors.New("invalid driver email")
	ErrInvalidDriverPhone  = errors.New("invalid driver phone")
	ErrInvalidDriverStatus = errors.New("invalid driver status")
	ErrInvalidVehicleType  = errors.New("invalid vehicle type")
	ErrDriverNotAvailable  = errors.New("driver is not available")
)
