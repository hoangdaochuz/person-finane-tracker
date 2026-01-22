package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTManager handles JWT token creation and validation
type JWTManager struct {
	secretKey string
	issuer    string
	// TokenExpiry is how long tokens are valid for
	TokenExpiry time.Duration
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	UUID   string `json:"uuid"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey:   secretKey,
		issuer:      "personal-finance-tracker",
		TokenExpiry: 7 * 24 * time.Hour, // 7 days
	}
}

// GenerateToken generates a new JWT token for a user
func (j *JWTManager) GenerateToken(userID int64, email string, userUUID uuid.UUID) (string, error) {
	if j.secretKey == "" {
		return "", errors.New("JWT secret key is not configured")
	}

	now := time.Now()
	expiresAt := now.Add(j.TokenExpiry)

	claims := &Claims{
		UserID: userID,
		Email:  email,
		UUID:   userUUID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   email,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(), // Unique JWT ID
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	if j.secretKey == "" {
		return nil, errors.New("JWT secret key is not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Verify issuer
	if claims.Issuer != j.issuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiry
// This can be used to implement token refresh logic
func (j *JWTManager) RefreshToken(oldTokenString string) (string, error) {
	claims, err := j.ValidateToken(oldTokenString)
	if err != nil {
		return "", err
	}

	// Parse the UUID from claims
	userUUID, err := uuid.Parse(claims.UUID)
	if err != nil {
		return "", fmt.Errorf("invalid user UUID in token: %w", err)
	}

	// Generate a new token with updated expiry
	return j.GenerateToken(claims.UserID, claims.Email, userUUID)
}

// ExtractToken extracts the bearer token from the Authorization header
func ExtractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Check if it starts with "Bearer "
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) {
		return "", errors.New("authorization header format must be: Bearer {token}")
	}

	if authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header format must be: Bearer {token}")
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

// SetTokenExpiry sets a custom token expiry duration
func (j *JWTManager) SetTokenExpiry(duration time.Duration) {
	j.TokenExpiry = duration
}

// SetIssuer sets a custom issuer
func (j *JWTManager) SetIssuer(issuer string) {
	j.issuer = issuer
}
