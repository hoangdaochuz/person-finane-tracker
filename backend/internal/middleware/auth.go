package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dev/personal-finance-tracker/backend/internal/security"
)

const (
	// UserContextKey is the context key for user information
	UserContextKey = "user"
	// UserIDContextKey is the context key for user ID
	UserIDContextKey = "user_id"
	// UserEmailContextKey is the context key for user email
	UserEmailContextKey = "user_email"
	// UserUUIDContextKey is the context key for user UUID
	UserUUIDContextKey = "user_uuid"
)

// JWTAuth validates JWT tokens and adds user info to context
func JWTAuth(jwtManager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			c.Abort()
			return
		}

		// Extract token from Bearer format
		token, err := security.ExtractToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Add user info to context for downstream handlers
		c.Set(UserIDContextKey, claims.UserID)
		c.Set(UserEmailContextKey, claims.Email)
		c.Set(UserUUIDContextKey, claims.UUID)
		c.Set(UserContextKey, claims)

		c.Next()
	}
}

// GetUserID retrieves the user ID from context
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(UserIDContextKey)
	if !exists {
		return 0, false
	}
	id, ok := userID.(int64)
	return id, ok
}

// GetUserEmail retrieves the user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(UserEmailContextKey)
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}

// GetUserUUID retrieves the user UUID from context
func GetUserUUID(c *gin.Context) (string, bool) {
	uuid, exists := c.Get(UserUUIDContextKey)
	if !exists {
		return "", false
	}
	u, ok := uuid.(string)
	return u, ok
}

// GetUserClaims retrieves the full user claims from context
func GetUserClaims(c *gin.Context) (*security.Claims, bool) {
	claims, exists := c.Get(UserContextKey)
	if !exists {
		return nil, false
	}
	userClaims, ok := claims.(*security.Claims)
	return userClaims, ok
}
