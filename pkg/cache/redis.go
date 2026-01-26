package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration
type Config struct {
	Host        string
	Port        string
	Password    string
	DB          int
	MaxRetries  int
	PoolSize    int
	MinIdleConn int
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

// NewRedisClient creates a new Redis client
func NewRedisClient(cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConn,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// Close gracefully closes the Redis client
func Close(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// GetClientStats returns Redis client statistics
func GetClientStats(client *redis.Client) map[string]interface{} {
	stats := client.PoolStats()
	return map[string]interface{}{
		"hits":          stats.Hits,
		"misses":        stats.Misses,
		"timeouts":      stats.Timeouts,
		"total_conns":   stats.TotalConns,
		"idle_conns":    stats.IdleConns,
		"stale_conns":   stats.StaleConns,
	}
}

// Helper functions for common operations

// SetWithExpiry sets a key-value pair with expiration
func SetWithExpiry(ctx context.Context, client *redis.Client, key string, value interface{}, expiry time.Duration) error {
	return client.Set(ctx, key, value, expiry).Err()
}

// Get retrieves a value by key
func Get(ctx context.Context, client *redis.Client, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Delete removes a key
func Delete(ctx context.Context, client *redis.Client, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists checks if key exists
func Exists(ctx context.Context, client *redis.Client, key string) (bool, error) {
	count, err := client.Exists(ctx, key).Result()
	return count > 0, err
}

// SetNX sets a key only if it doesn't exist (for distributed locks)
func SetNX(ctx context.Context, client *redis.Client, key string, value interface{}, expiry time.Duration) (bool, error) {
	return client.SetNX(ctx, key, value, expiry).Result()
}

// Incr increments a counter
func Incr(ctx context.Context, client *redis.Client, key string) (int64, error) {
	return client.Incr(ctx, key).Result()
}

// Expire sets expiration on a key
func Expire(ctx context.Context, client *redis.Client, key string, expiry time.Duration) error {
	return client.Expire(ctx, key, expiry).Err()
}

// GetMultiple retrieves multiple keys at once
func GetMultiple(ctx context.Context, client *redis.Client, keys []string) ([]interface{}, error) {
	return client.MGet(ctx, keys...).Result()
}
