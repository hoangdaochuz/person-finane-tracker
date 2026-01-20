package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/dev/personal-finance-tracker/backend/internal/config"
	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/handler"
	"github.com/dev/personal-finance-tracker/backend/internal/logger"
	"github.com/dev/personal-finance-tracker/backend/internal/middleware"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/internal/service"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logger
	logger.Init(cfg)
	log := logger.Get()

	log.Info().Str("config", *configPath).Msg("Configuration loaded")
	log.Info().Str("mode", cfg.Server.Mode).Msg("Starting application")

	// Set Gin mode based on config
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.Server.Mode == "test" {
		gin.SetMode(gin.TestMode)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get database connection")
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info().Msg("Database connected successfully")

	// Auto-migrate the schema (for development and test only)
	// In production, use golang-migrate instead
	if cfg.Server.Mode == "debug" || cfg.Server.Mode == "test" {
		if err := db.AutoMigrate(&domain.Transaction{}); err != nil {
			log.Fatal().Err(err).Msg("Failed to migrate database")
		}
		log.Info().Msg("Database migration completed")
	} else {
		log.Warn().Msg("AutoMigrate disabled in production mode. Use golang-migrate for schema migrations.")
	}

	// Initialize repositories
	txRepo := repository.NewTransactionRepository(db)

	// Initialize services
	txService := service.NewTransactionService(txRepo)

	// Initialize handlers
	webhookHandler := handler.NewWebhookHandler(txService)
	analyticsHandler := handler.NewAnalyticsHandler(txService)

	// Setup router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: cfg.Server.AllowedOrigins,
	}))
	router.Use(middleware.RequestID())
	router.Use(middleware.MaxBodySize(10 << 20)) // 10 MB max body size
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.ErrorHandler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		log := middleware.GetLogger(c)
		log.Debug().Msg("Health check requested")

		// Check database connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sqlDB, err := db.DB()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get database connection")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		// Ping the database
		if err := sqlDB.PingContext(ctx); err != nil {
			log.Error().Err(err).Msg("Database ping failed")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database unreachable",
			})
			return
		}

		// Execute a simple query to verify database is responding
		var result int
		if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
			log.Error().Err(err).Msg("Database query failed")
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database query failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Webhook endpoints (require API key)
		webhook := v1.Group("/webhook")
		webhook.Use(middleware.APIKeyAuth(cfg.APIKey))
		{
			webhook.POST("/transaction", webhookHandler.CreateTransaction)
			webhook.POST("/transactions/batch", webhookHandler.CreateBatchTransaction)
		}

		// Analytics endpoints (no auth required for single-user app)
		analytics := v1.Group("/analytics")
		{
			analytics.GET("/summary", analyticsHandler.GetSummary)
			analytics.GET("/trends", analyticsHandler.GetTrends)
			analytics.GET("/by-source", analyticsHandler.GetBreakdownBySource)
			analytics.GET("/by-category", analyticsHandler.GetBreakdownByCategory)
		}

		// Transaction endpoints
		transactions := v1.Group("/transactions")
		{
			transactions.GET("", analyticsHandler.ListTransactions)
			transactions.GET("/:id", analyticsHandler.GetTransactionByID)
		}
	}

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       time.Duration(cfg.Server.Timeout) * time.Second,
		ReadHeaderTimeout: 5 * time.Second, // Prevent Slowloris attacks
		WriteTimeout:      time.Duration(cfg.Server.Timeout) * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB max header size
	}

	// Start server in a goroutine
	go func() {
		log.Info().Str("address", addr).Str("mode", cfg.Server.Mode).Msg("Server starting")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close database connection
	closeErr := sqlDB.Close()
	if closeErr != nil {
		log.Error().Err(closeErr).Msg("Failed to close database connection")
	}

	log.Info().Msg("Server exited")
}
