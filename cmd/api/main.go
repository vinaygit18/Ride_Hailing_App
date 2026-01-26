package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocomet/ride-hailing/internal/api/handlers"
	"github.com/gocomet/ride-hailing/internal/api/routes"
	"github.com/gocomet/ride-hailing/internal/config"
	"github.com/gocomet/ride-hailing/pkg/cache"
	"github.com/gocomet/ride-hailing/pkg/database"
	"github.com/gocomet/ride-hailing/pkg/logger"
	"github.com/gocomet/ride-hailing/pkg/monitoring"
	"github.com/gocomet/ride-hailing/pkg/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
		Output: cfg.Log.Output,
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Starting GoComet Ride-Hailing Application",
		logger.String("env", cfg.Server.Env),
		logger.String("port", cfg.Server.Port),
	)

	// Initialize New Relic
	nrApp, err := monitoring.New(monitoring.Config{
		LicenseKey: cfg.NewRelic.LicenseKey,
		AppName:    cfg.NewRelic.AppName,
		Enabled:    cfg.NewRelic.Enabled,
		LogLevel:   cfg.NewRelic.LogLevel,
	})
	if err != nil {
		appLogger.Warn("Failed to initialize New Relic", logger.Err(err))
	} else if nrApp.IsEnabled() {
		appLogger.Info("New Relic APM initialized successfully",
			logger.String("app_name", cfg.NewRelic.AppName),
			logger.Bool("enabled", true))
	} else {
		appLogger.Info("New Relic APM disabled")
	}
	defer nrApp.Shutdown(10 * time.Second)

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cache.Config{
		Host:        cfg.Redis.Host,
		Port:        cfg.Redis.Port,
		Password:    cfg.Redis.Password,
		DB:          cfg.Redis.DB,
		MaxRetries:  cfg.Redis.MaxRetries,
		PoolSize:    cfg.Redis.PoolSize,
		DialTimeout: cfg.Redis.DialTimeout,
		ReadTimeout: cfg.Redis.ReadTimeout,
	})
	if err != nil {
		appLogger.Fatal("Failed to connect to Redis", logger.Err(err))
	}
	defer cache.Close(redisClient)

	appLogger.Info("Connected to Redis successfully")

	// Initialize PostgreSQL
	postgresDB, err := database.NewPostgresDB(database.Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
		DBName:   "gocomet",
		SSLMode:  "disable",
		MaxConns: 25,
		MaxIdle:  5,
	})
	if err != nil {
		appLogger.Fatal("Failed to connect to PostgreSQL", logger.Err(err))
	}
	defer postgresDB.Close()

	appLogger.Info("Connected to PostgreSQL successfully")

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(appLogger)
	go wsHub.Run()

	// Initialize handlers with dependencies
	h := handlers.NewHandlers(postgresDB, redisClient, appLogger, wsHub)

	// Initialize Gin router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup all routes
	var nrApplication *monitoring.NewRelicApp
	if nrApp.IsEnabled() {
		nrApplication = nrApp
	}
	routes.SetupRoutes(router, h, nrApplication.Application)

	appLogger.Info("Routes configured successfully")

	// Serve static files
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("./web/templates/*")

	// Serve web pages
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/driver", func(c *gin.Context) {
		c.HTML(http.StatusOK, "driver.html", nil)
	})
	router.GET("/rider", func(c *gin.Context) {
		c.HTML(http.StatusOK, "rider.html", nil)
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info("Server starting", logger.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", logger.Err(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", logger.Err(err))
	}

	appLogger.Info("Server stopped gracefully")
}
