package domain

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test RegisterRequest.Validate()

func TestRegisterRequestValidate_ValidInput(t *testing.T) {
	req := &RegisterRequest{
		Email:    "user@example.com",
		Password: "ValidPass123",
		Name:     "John Doe",
	}

	err := req.Validate()

	assert.NoError(t, err)
}

func TestRegisterRequestValidate_InvalidEmail(t *testing.T) {
	req := &RegisterRequest{
		Email:    "not-an-email",
		Password: "ValidPass123",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok, "error should be ValidationError")
	assert.Equal(t, "email", validationErr.Field)
}

func TestRegisterRequestValidate_ShortPassword(t *testing.T) {
	req := &RegisterRequest{
		Email:    "user@example.com",
		Password: "short1",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErr.Field)
	assert.Contains(t, validationErr.Message, "at least")
}

func TestRegisterRequestValidate_PasswordWithoutNumber(t *testing.T) {
	req := &RegisterRequest{
		Email:    "user@example.com",
		Password: "onlyletters",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErr.Field)
	assert.Contains(t, validationErr.Message, "number")
}

func TestRegisterRequestValidate_PasswordWithoutLetter(t *testing.T) {
	req := &RegisterRequest{
		Email:    "user@example.com",
		Password: "1234567890", // Only numbers, fails complexity
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErr.Field)
	assert.Contains(t, validationErr.Message, "at least 3")
}

func TestRegisterRequestValidate_LongPassword(t *testing.T) {
	req := &RegisterRequest{
		Email:    "user@example.com",
		Password: strings.Repeat("a", 130) + "1",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErr.Field)
	assert.Contains(t, validationErr.Message, "at most")
}

func TestRegisterRequestValidate_MinimumValidPassword(t *testing.T) {
	req := &RegisterRequest{
		Email:    "user@example.com",
		Password: "Abc1234567", // Exactly 10 chars, has letter, number, and uppercase
	}

	err := req.Validate()

	assert.NoError(t, err)
}

// Test LoginRequest.Validate()

func TestLoginRequestValidate_ValidInput(t *testing.T) {
	req := &LoginRequest{
		Email:    "user@example.com",
		Password: "anypassword",
	}

	err := req.Validate()

	assert.NoError(t, err)
}

func TestLoginRequestValidate_InvalidEmail(t *testing.T) {
	req := &LoginRequest{
		Email:    "not-an-email",
		Password: "password123",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "email", validationErr.Field)
}

func TestLoginRequestValidate_EmptyPassword(t *testing.T) {
	req := &LoginRequest{
		Email:    "user@example.com",
		Password: "",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "password", validationErr.Field)
}

// Test User.ToResponse()

func TestUserToResponse_ExcludesSensitiveFields(t *testing.T) {
	user := &User{
		ID:           1,
		UUID:         uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		Email:        "user@example.com",
		PasswordHash: "hashed-password-should-not-appear",
		Name:         "John Doe",
		APIKey:       "api-key-123",
		IsActive:     true,
	}

	response := user.ToResponse()

	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, user.UUID, response.UUID)
	assert.Equal(t, "user@example.com", response.Email)
	assert.Equal(t, "John Doe", response.Name)
	assert.Equal(t, "api-key-123", response.APIKey)
	assert.True(t, response.IsActive)
}

// Test User.TableName()

func TestUser_TableName(t *testing.T) {
	user := User{}
	assert.Equal(t, "users", user.TableName())
}

// Test validatePasswordComplexity()

func TestValidatePasswordComplexity_ValidPasswords(t *testing.T) {
	validPasswords := []string{
		"Password123!", // Has letters, numbers, uppercase, special - meets 4 criteria, 12 chars
		"Abc1234567",   // Has letters, numbers, uppercase - meets 3 criteria, 10 chars
		"Pass@123456",  // Has letters, numbers, uppercase, special - meets 4 criteria, 11 chars
		"Test123!@#",   // Has letters, numbers, uppercase, special - meets 4 criteria, 11 chars
		"abc123!@#$",   // Has letters, numbers, special - meets 3 criteria, 10 chars
		"ABC123!@#$",   // Has uppercase, numbers, special - meets 3 criteria, 10 chars
	}

	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			assert.True(t, validatePasswordComplexity(password), "password should be valid: %s", password)
		})
	}
}

func TestValidatePasswordComplexity_InvalidPasswords(t *testing.T) {
	invalidPasswords := []struct {
		name     string
		password string
	}{
		{"only letters", "abcdefgh"},
		{"only numbers", "12345678"},
		{"only special chars", "!@#$%^&*"},
		{"only uppercase", "ABCDEFGH"},
		{"only lowercase", "abcdefgh"},
		{"empty", ""},
	}

	for _, tc := range invalidPasswords {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, validatePasswordComplexity(tc.password), "password should be invalid: %s", tc.password)
		})
	}
}

// Test EmailRegex

func TestEmailRegex_ValidEmails(t *testing.T) {
	validEmails := []string{
		"user@example.com",
		"test.user@domain.co.uk",
		"admin+tag@example.org",
		"user123@test-domain.com",
		"first.last@sub.domain.example.com",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			assert.True(t, EmailRegex.MatchString(email), "email should be valid: %s", email)
		})
	}
}

func TestEmailRegex_InvalidEmails(t *testing.T) {
	invalidEmails := []struct {
		name  string
		email string
	}{
		{"no @", "notanemail"},
		{"no domain", "user@"},
		{"no user", "@example.com"},
		{"double @", "user@@example.com"},
		{"spaces", "user @example.com"},
		{"special chars", "user@exa!mple.com"},
	}

	for _, tc := range invalidEmails {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, EmailRegex.MatchString(tc.email), "email should be invalid: %s", tc.email)
		})
	}
}
