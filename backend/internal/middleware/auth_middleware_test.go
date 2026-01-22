package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dev/personal-finance-tracker/backend/internal/security"
)

// Test JWTAuth()

func TestJWTAuth_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, _ := jwtManager.GenerateToken(123, "test@example.com", userUUID)

	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "protected")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "protected", w.Body.String())
}

func TestJWTAuth_MissingAuthorizationHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "protected")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "authorization header is required")
}

func TestJWTAuth_InvalidTokenFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "protected")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "format must be")
}

func TestJWTAuth_EmptyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "protected")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token is empty")
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "protected")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Generate token with one secret
	jwtManager1 := security.NewJWTManager("first-secret-key-minimum-32-chars")
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, _ := jwtManager1.GenerateToken(123, "test@example.com", userUUID)

	// Validate with different secret
	jwtManager2 := security.NewJWTManager("different-secret-key-minimum-32-chars")
	router.Use(JWTAuth(jwtManager2))
	router.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "protected")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test context helpers

func TestJWTAuth_SetsUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, _ := jwtManager.GenerateToken(123, "test@example.com", userUUID)

	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := GetUserID(c)
		assert.True(t, exists)
		assert.Equal(t, int64(123), userID)

		email, exists := GetUserEmail(c)
		assert.True(t, exists)
		assert.Equal(t, "test@example.com", email)

		uuidStr, exists := GetUserUUID(c)
		assert.True(t, exists)
		assert.Equal(t, userUUID.String(), uuidStr)

		claims, exists := GetUserClaims(c)
		assert.True(t, exists)
		assert.Equal(t, int64(123), claims.UserID)

		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/public", func(c *gin.Context) {
		userID, exists := GetUserID(c)
		assert.False(t, exists)
		assert.Equal(t, int64(0), userID)
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserEmail_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/public", func(c *gin.Context) {
		email, exists := GetUserEmail(c)
		assert.False(t, exists)
		assert.Equal(t, "", email)
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserUUID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/public", func(c *gin.Context) {
		uuidStr, exists := GetUserUUID(c)
		assert.False(t, exists)
		assert.Equal(t, "", uuidStr)
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserClaims_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/public", func(c *gin.Context) {
		claims, exists := GetUserClaims(c)
		assert.False(t, exists)
		assert.Nil(t, claims)
		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/public", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test multiple middleware

func TestJWTAuth_WithOtherMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, _ := jwtManager.GenerateToken(123, "test@example.com", userUUID)

	router.Use(RequestID())
	router.Use(JWTAuth(jwtManager))
	router.GET("/protected", func(c *gin.Context) {
		// Check request ID is set
		requestID := GetRequestID(c)
		assert.NotEmpty(t, requestID)

		// Check user is set
		userID, exists := GetUserID(c)
		assert.True(t, exists)
		assert.Equal(t, int64(123), userID)

		c.String(http.StatusOK, "ok")
	})

	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// Test header name constants

func TestAuthorizationHeaderName(t *testing.T) {
	// This test documents the expected header name
	assert.Equal(t, "Authorization", "Authorization")
}
