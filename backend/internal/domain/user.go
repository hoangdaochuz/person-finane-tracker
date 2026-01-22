package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxNameLength is the maximum length for user name
	MaxNameLength = 100
	// MaxEmailLength is the maximum length for email
	MaxEmailLength = 255
	// MinPasswordLength is the minimum length for password
	MinPasswordLength = 10
	// MaxPasswordLength is the maximum length for password
	MaxPasswordLength = 128
	// APIKeyLength is the length of generated API keys
	APIKeyLength = 32
)

var (
	// EmailRegex is a simple email validation regex
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// validatePasswordComplexity checks if password meets complexity requirements
// Password must satisfy at least 3 of the following 4 criteria:
// 1. Contains letters (a-z or A-Z)
// 2. Contains numbers (0-9)
// 3. Contains special characters
// 4. Contains uppercase letters (A-Z)
// Go's regexp (RE2) doesn't support lookahead, so we check this separately
func validatePasswordComplexity(password string) bool {
	if len(password) < MinPasswordLength {
		return false
	}

	hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	var satisfied int
	if hasLetter {
		satisfied++
	}
	if hasNumber {
		satisfied++
	}
	if hasSpecial {
		satisfied++
	}
	if regexp.MustCompile(`[A-Z]`).MatchString(password) {
		satisfied++
	}

	return satisfied >= 3
}

// User represents a user account
type User struct {
	ID           int64      `json:"id" gorm:"primaryKey"`
	UUID         uuid.UUID  `json:"uuid" gorm:"type:uuid;not null;unique"`
	Email        string     `json:"email" gorm:"type:varchar(255);not null;unique"`
	PasswordHash string     `json:"-" gorm:"type:varchar(255);not null"` // never expose in JSON
	Name         string     `json:"name" gorm:"type:varchar(100)"`
	APIKey       string     `json:"api_key" gorm:"type:varchar(255);not null;unique"`
	IsActive     bool       `json:"is_active" gorm:"not null;default:true"`
	LastLoginAt  *time.Time `json:"last_login_at" gorm:"type:timestamp"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=10,max=128"`
	Name     string `json:"name" binding:"omitempty,max=100"`
}

// Validate performs additional validation beyond struct tags
func (r *RegisterRequest) Validate() error {
	// Email format validation
	if !EmailRegex.MatchString(r.Email) {
		return &ValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	// Password length validation
	if len(r.Password) < MinPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters", MinPasswordLength),
		}
	}

	if len(r.Password) > MaxPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at most %d characters", MaxPasswordLength),
		}
	}

	// Password complexity validation (must satisfy at least 3 of 4 criteria)
	if !validatePasswordComplexity(r.Password) {
		return &ValidationError{
			Field:   "password",
			Message: "password must satisfy at least 3 of the following: contain letters, numbers, special characters, or uppercase letters",
		}
	}

	return nil
}

// LoginRequest is the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required"`
}

// Validate performs validation on login request
func (r *LoginRequest) Validate() error {
	// Email format validation
	if !EmailRegex.MatchString(r.Email) {
		return &ValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	// Password should not be empty
	if r.Password == "" {
		return &ValidationError{
			Field:   "password",
			Message: "password is required",
		}
	}

	return nil
}

// AuthResponse is the response body for successful authentication
type AuthResponse struct {
	Token string      `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse is a safe user representation (without sensitive data)
type UserResponse struct {
	ID       int64     `json:"id"`
	UUID     uuid.UUID `json:"uuid"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	APIKey   string    `json:"api_key"`
	IsActive bool      `json:"is_active"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		UUID:     u.UUID,
		Email:    u.Email,
		Name:     u.Name,
		APIKey:   u.APIKey,
		IsActive: u.IsActive,
	}
}

// SanitizeEmail returns a cleaned and validated email
func SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	return strings.ToLower(email)
}

// GenerateAPIKey generates a random API key
// This is a placeholder - actual implementation will be in security package
func GenerateAPIKey() string {
	return uuid.New().String()[:APIKeyLength]
}
