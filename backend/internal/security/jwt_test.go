package security

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const testJWTSecret = "test-secret-key-for-jwt-signing-minimum-32-chars"

// Test JWTManager.GenerateToken()

func TestJWTManager_GenerateToken_ValidInput(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(1, "user@example.com", userUUID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	// JWT tokens have 3 parts separated by dots
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3, "JWT token should have 3 parts")
}

func TestJWTManager_GenerateToken_ContainsCorrectClaims(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	userID := int64(123)
	email := "test@example.com"
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	token, err := manager.GenerateToken(userID, email, userUUID)
	assert.NoError(t, err)

	// Verify the token by parsing it
	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, userUUID.String(), claims.UUID)
	assert.Equal(t, "personal-finance-tracker", claims.Issuer)
	assert.Equal(t, email, claims.Subject)
}

func TestJWTManager_GenerateToken_ExpiryTime(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now()

	token, err := manager.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)

	// Check expiry is approximately 7 days from now
	expectedExpiry := now.Add(7 * 24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 1 second tolerance for test execution time
	diff := actualExpiry.Sub(expectedExpiry)
	assert.Less(t, diff.Abs(), 1*time.Second, "expiry should be approximately 7 days from now")
}

func TestJWTManager_GenerateToken_EmptySecret(t *testing.T) {
	manager := NewJWTManager("")

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(1, "user@example.com", userUUID)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "secret key")
}

// Test JWTManager.ValidateToken()

func TestJWTManager_ValidateToken_ValidToken(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	claims, err := manager.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, int64(1), claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
}

func TestJWTManager_ValidateToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	invalidTokens := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"not a jwt", "not-a-jwt-token"},
		{"incomplete", "header.payload"},
		{"wrong format", "abc.def.ghi"},
	}

	for _, tc := range invalidTokens {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := manager.ValidateToken(tc.token)
			assert.Error(t, err, "invalid token should return error")
			assert.Nil(t, claims)
		})
	}
}

func TestJWTManager_ValidateToken_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager(testJWTSecret)
	manager2 := NewJWTManager("different-secret-key-for-testing-32-chars")

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager1.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	// Try to validate with different secret
	claims, err := manager2.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_ValidateToken_ExpiredToken(t *testing.T) {
	// Create a manager with very short expiry for testing
	manager := NewJWTManager(testJWTSecret)
	manager.SetTokenExpiry(1 * time.Millisecond)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	claims, err := manager.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

func TestJWTManager_ValidateToken_WrongIssuer(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)
	manager.SetIssuer("wrong-issuer")

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	// Reset issuer for validation
	manager.SetIssuer("personal-finance-tracker")

	claims, err := manager.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "issuer")
}

// Test JWTManager.RefreshToken()

func TestJWTManager_RefreshToken_ValidToken(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	oldToken, err := manager.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	newToken, err := manager.RefreshToken(oldToken)

	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, oldToken, newToken, "refreshed token should be different")
}

func TestJWTManager_RefreshToken_InvalidToken(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	newToken, err := manager.RefreshToken("invalid-token")

	assert.Error(t, err)
	assert.Empty(t, newToken)
}

// Test ExtractToken()

func TestExtractToken_ValidBearerToken(t *testing.T) {
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"

	token, err := ExtractToken("Bearer " + testToken)

	assert.NoError(t, err)
	assert.Equal(t, testToken, token)
}

func TestExtractToken_MissingBearer(t *testing.T) {
	testToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test"

	_, err := ExtractToken(testToken)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Bearer")
}

func TestExtractToken_EmptyToken(t *testing.T) {
	_, err := ExtractToken("Bearer ")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}

func TestExtractToken_EmptyHeader(t *testing.T) {
	_, err := ExtractToken("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}

// Test JWTManager.SetTokenExpiry()

func TestJWTManager_SetTokenExpiry(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	// Set custom expiry
	manager.SetTokenExpiry(24 * time.Hour)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)

	// Check expiry is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time

	// Allow 1 second tolerance
	diff := actualExpiry.Sub(expectedExpiry)
	assert.Less(t, diff.Abs(), 1*time.Second)
}

// Test JWTManager.SetIssuer()

func TestJWTManager_SetIssuer(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)
	customIssuer := "custom-issuer-app"
	manager.SetIssuer(customIssuer)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, err := manager.GenerateToken(1, "user@example.com", userUUID)
	assert.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, customIssuer, claims.Issuer)
}

// Test Claims structure

func TestClaims_AllFieldsSet(t *testing.T) {
	manager := NewJWTManager(testJWTSecret)

	userID := int64(42)
	email := "claims@test.com"
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	token, err := manager.GenerateToken(userID, email, userUUID)
	assert.NoError(t, err)

	claims, err := manager.ValidateToken(token)
	assert.NoError(t, err)

	// Check all required claims
	assert.NotEmpty(t, claims.ID, "JWT ID should be set")
	assert.NotEmpty(t, claims.IssuedAt, "IssuedAt should be set")
	assert.NotEmpty(t, claims.ExpiresAt, "ExpiresAt should be set")
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, userUUID.String(), claims.UUID)
	assert.Equal(t, email, claims.Subject)
}
