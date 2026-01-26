package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	NewRelic    NewRelicConfig
	JWT         JWTConfig
	Pricing     PricingConfig
	Matching    MatchingConfig
	RateLimit   RateLimitConfig
	WebSocket   WebSocketConfig
	Cache       CacheConfig
	Log         LogConfig
	CORS        CORSConfig
	Features    FeatureFlags
}

type ServerConfig struct {
	Port string
	Env  string
	Host string
}

type DatabaseConfig struct {
	Host            string
	Port            string
	Name            string
	User            string
	Password        string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
	MaxLifetime     time.Duration
}

type RedisConfig struct {
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

type NewRelicConfig struct {
	LicenseKey string
	AppName    string
	Enabled    bool
	LogLevel   string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type PricingConfig struct {
	BaseFare struct {
		Economy int
		Premium int
		Luxury  int
	}
	PerKMRate struct {
		Economy int
		Premium int
		Luxury  int
	}
	PerMinuteRate struct {
		Economy int
		Premium int
		Luxury  int
	}
	MaxSurgeMultiplier float64
	MinSurgeMultiplier float64
}

type MatchingConfig struct {
	MaxRadiusKM      float64
	MaxTimeout       time.Duration
	MaxCandidates    int
}

type RateLimitConfig struct {
	LocationUpdatesPerSecond int
	RideRequestsPerMinute    int
	GeneralPerMinute         int
}

type WebSocketConfig struct {
	ReadBufferSize       int
	WriteBufferSize      int
	HeartbeatInterval    time.Duration
}

type CacheConfig struct {
	TTLActiveRides     time.Duration
	TTLDriverLocations time.Duration
	TTLIdempotency     time.Duration
}

type LogConfig struct {
	Level  string
	Format string
	Output string
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

type FeatureFlags struct {
	EnableSurgePricing    bool
	EnableAutoMatching    bool
	EnableRealTimeUpdates bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("SERVER_ENV", "development"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			Name:            getEnv("DB_NAME", "gocomet"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 100),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
			MaxLifetime:     time.Duration(getEnvAsInt("DB_MAX_LIFETIME_MINUTES", 30)) * time.Minute,
		},
		Redis: RedisConfig{
			Host:        getEnv("REDIS_HOST", "localhost"),
			Port:        getEnv("REDIS_PORT", "6379"),
			Password:    getEnv("REDIS_PASSWORD", ""),
			DB:          getEnvAsInt("REDIS_DB", 0),
			MaxRetries:  getEnvAsInt("REDIS_MAX_RETRIES", 3),
			PoolSize:    getEnvAsInt("REDIS_POOL_SIZE", 100),
			MinIdleConn: 10,
			DialTimeout: 5 * time.Second,
			ReadTimeout: 3 * time.Second,
		},
		NewRelic: NewRelicConfig{
			LicenseKey: getEnv("NEW_RELIC_LICENSE_KEY", ""),
			AppName:    getEnv("NEW_RELIC_APP_NAME", "GoComet-RideHailing"),
			Enabled:    getEnvAsBool("NEW_RELIC_ENABLED", true),
			LogLevel:   getEnv("NEW_RELIC_LOG_LEVEL", "info"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your_jwt_secret_key_here"),
			Expiry: parseDuration(getEnv("JWT_EXPIRY", "24h"), 24*time.Hour),
		},
		Matching: MatchingConfig{
			MaxRadiusKM:   getEnvAsFloat64("MAX_MATCHING_RADIUS_KM", 5.0),
			MaxTimeout:    time.Duration(getEnvAsInt("MAX_MATCHING_TIMEOUT_SECONDS", 30)) * time.Second,
			MaxCandidates: getEnvAsInt("MAX_DRIVER_CANDIDATES", 10),
		},
		RateLimit: RateLimitConfig{
			LocationUpdatesPerSecond: getEnvAsInt("RATE_LIMIT_LOCATION_UPDATES_PER_SECOND", 2),
			RideRequestsPerMinute:    getEnvAsInt("RATE_LIMIT_RIDE_REQUESTS_PER_MINUTE", 5),
			GeneralPerMinute:         getEnvAsInt("RATE_LIMIT_GENERAL_PER_MINUTE", 100),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:    getEnvAsInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize:   getEnvAsInt("WS_WRITE_BUFFER_SIZE", 1024),
			HeartbeatInterval: time.Duration(getEnvAsInt("WS_HEARTBEAT_INTERVAL_SECONDS", 30)) * time.Second,
		},
		Cache: CacheConfig{
			TTLActiveRides:     time.Duration(getEnvAsInt("CACHE_TTL_ACTIVE_RIDES", 300)) * time.Second,
			TTLDriverLocations: time.Duration(getEnvAsInt("CACHE_TTL_DRIVER_LOCATIONS", 300)) * time.Second,
			TTLIdempotency:     time.Duration(getEnvAsInt("CACHE_TTL_IDEMPOTENCY", 86400)) * time.Second,
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
			Output: getEnv("LOG_OUTPUT", "stdout"),
		},
		Features: FeatureFlags{
			EnableSurgePricing:    getEnvAsBool("ENABLE_SURGE_PRICING", true),
			EnableAutoMatching:    getEnvAsBool("ENABLE_AUTO_MATCHING", true),
			EnableRealTimeUpdates: getEnvAsBool("ENABLE_REAL_TIME_UPDATES", true),
		},
	}

	// Set pricing configuration
	cfg.Pricing.BaseFare.Economy = getEnvAsInt("BASE_FARE_ECONOMY", 50)
	cfg.Pricing.BaseFare.Premium = getEnvAsInt("BASE_FARE_PREMIUM", 100)
	cfg.Pricing.BaseFare.Luxury = getEnvAsInt("BASE_FARE_LUXURY", 200)

	cfg.Pricing.PerKMRate.Economy = getEnvAsInt("PER_KM_RATE_ECONOMY", 10)
	cfg.Pricing.PerKMRate.Premium = getEnvAsInt("PER_KM_RATE_PREMIUM", 15)
	cfg.Pricing.PerKMRate.Luxury = getEnvAsInt("PER_KM_RATE_LUXURY", 25)

	cfg.Pricing.PerMinuteRate.Economy = getEnvAsInt("PER_MINUTE_RATE_ECONOMY", 2)
	cfg.Pricing.PerMinuteRate.Premium = getEnvAsInt("PER_MINUTE_RATE_PREMIUM", 3)
	cfg.Pricing.PerMinuteRate.Luxury = getEnvAsInt("PER_MINUTE_RATE_LUXURY", 5)

	cfg.Pricing.MaxSurgeMultiplier = getEnvAsFloat64("MAX_SURGE_MULTIPLIER", 3.0)
	cfg.Pricing.MinSurgeMultiplier = getEnvAsFloat64("MIN_SURGE_MULTIPLIER", 1.0)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("SERVER_PORT is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("REDIS_HOST is required")
	}
	if c.JWT.Secret == "your_jwt_secret_key_here" && c.Server.Env == "production" {
		return fmt.Errorf("JWT_SECRET must be set in production")
	}
	return nil
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	return defaultValue
}
