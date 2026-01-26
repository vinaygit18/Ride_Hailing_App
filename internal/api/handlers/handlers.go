package handlers

import (
	"database/sql"

	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Handlers holds all handler dependencies
type Handlers struct {
	DB     *sql.DB
	Redis  *redis.Client
	Logger *logger.Logger
	Hub    interface{} // WebSocket hub (interface to avoid circular dependency)
}

// NewHandlers creates a new Handlers instance
func NewHandlers(db *sql.DB, redisClient *redis.Client, logger *logger.Logger, hub interface{}) *Handlers {
	return &Handlers{
		DB:     db,
		Redis:  redisClient,
		Logger: logger,
		Hub:    hub,
	}
}
