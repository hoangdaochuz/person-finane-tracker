package middleware

import (
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/dev/personal-finance-tracker/backend/internal/logger"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
)

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already in header
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			// Generate new UUID
			requestID = uuid.New().String()
		}

		// Set request ID in header for response
		c.Header(RequestIDHeader, requestID)

		// Store request ID in context
		c.Set(RequestIDKey, requestID)

		// Add request ID to logger context
		log := logger.Get().With().Str("request_id", requestID).Logger()
		c.Set("logger", log)

		// Add request ID to gin context for easy access
		c.Set("request_id", requestID)

		c.Next()
	}
}

// GetLogger retrieves the request-scoped logger from context
func GetLogger(c *gin.Context) zerolog.Logger {
	if log, exists := c.Get("logger"); exists {
		if logger, ok := log.(zerolog.Logger); ok {
			return logger
		}
	}
	// Fallback to global logger
	return logger.Get()
}

// GetRequestID retrieves the request ID from context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Get request-scoped logger
		log := GetLogger(c)

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		status := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get request ID
		requestID := GetRequestID(c)

		// Build log entry
		event := log.Info().
			Str("client_ip", clientIP).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", query).
			Int("status", status).
			Dur("latency", latency).
			Str("user_agent", c.Request.UserAgent()).
			Str("request_id", requestID)

		// Add error if present
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			event = log.Error().Err(err.Err)
		}

		// Log based on status code
		switch {
		case status >= 500:
			event.Msg("Server error")
		case status >= 400:
			event.Msg("Client error")
		case status >= 300:
			event.Msg("Redirect")
		default:
			event.Msg("Request completed")
		}
	}
}

// MaxBodySize limits the maximum size of request body
type limitedReadCloser struct {
	io.Reader
	io.Closer
}

func (lrc *limitedReadCloser) Close() error {
	return lrc.Closer.Close()
}

// MaxBodySize limits the maximum size of request body
func MaxBodySize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = &limitedReadCloser{
				Reader: io.LimitReader(c.Request.Body, maxSize),
				Closer: c.Request.Body,
			}
		}
		c.Next()
	}
}
