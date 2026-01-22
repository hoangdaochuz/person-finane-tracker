package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/service"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Call service to register user
	response, err := h.authService.Register(&req)
	if err != nil {
		// Check for specific error types
		var validationErr *domain.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Message,
				"field": validationErr.Field,
			})
			return
		}

		// Check for duplicate email error
		if service.IsDuplicateEmailError(err) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "user with this email already exists",
			})
			return
		}

		// All other errors are internal server errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Call service to authenticate user
	response, err := h.authService.Login(&req)
	if err != nil {
		// Check for validation errors
		var validationErr *domain.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": validationErr.Message,
				"field": validationErr.Field,
			})
			return
		}

		// For auth failures (invalid credentials, inactive user), return 401
		// Use generic message to prevent user enumeration
		if errors.Is(err, service.ErrInvalidCredentials) ||
			errors.Is(err, service.ErrUserInactive) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email or password",
			})
			return
		}

		// All other errors are internal server errors
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "authentication failed",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
