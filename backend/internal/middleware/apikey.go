package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// nolint:gosec // G101 - This is a header name, not actual credentials
	APIKeyHeader = "X-API-Key"
)

// APIKeyAuth validates the API key in the request header
func APIKeyAuth(validAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(APIKeyHeader)

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key is required",
			})
			c.Abort()
			return
		}

		if apiKey != validAPIKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ErrorHandler is a global error handler middleware
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
}

// CORS middleware for handling Cross-Origin requests with origin whitelist
func CORS(config CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		if origin != "" {
			for _, allowed := range config.AllowedOrigins {
				if allowed == "*" || allowed == origin {
					allowedOrigin = allowed
					break
				}
				// Support subdomain matching (e.g., *.example.com)
				if strings.HasPrefix(allowed, "*.") {
					domain := strings.TrimPrefix(allowed, "*.")
					if strings.HasSuffix(origin, "."+domain) || origin == domain {
						allowedOrigin = origin
						break
					}
				}
			}
		}

		// If origin is not allowed, return empty for Access-Control-Allow-Origin
		// which will cause the browser to block the request
		if allowedOrigin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
