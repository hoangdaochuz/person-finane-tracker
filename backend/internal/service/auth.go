package service

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/internal/security"
)

var (
	// ErrInvalidCredentials is returned when email/password don't match
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserInactive is returned when trying to authenticate an inactive user
	ErrUserInactive = errors.New("user account is inactive")
)

// AuthService handles business logic for authentication
type AuthService interface {
	Register(req *domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(req *domain.LoginRequest) (*domain.AuthResponse, error)
	ValidateToken(token string) (*security.Claims, error)
	GetJWTManager() *security.JWTManager
}

type authService struct {
	userRepo       repository.UserRepository
	passwordHasher *security.PasswordHasher
	jwtManager     *security.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:       userRepo,
		passwordHasher: security.NewPasswordHasher(),
		jwtManager:     security.NewJWTManager(jwtSecret),
	}
}

// Register creates a new user account and returns auth response
func (s *authService) Register(req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Perform validation
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err == nil && existingUser != nil {
		return nil, repository.ErrUserAlreadyExists
	} else if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	// Hash the password
	passwordHash, err := s.passwordHasher.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user domain model
	user := &domain.User{
		UUID:         uuid.New(),
		Email:        req.Email,
		PasswordHash: passwordHash,
		Name:         req.Name,
		IsActive:     true,
	}

	// Save to database
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.UUID)
	if err != nil {
		return nil, err
	}

	// Create auth response with safe user data (using ToResponse method)
	return &domain.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// Login authenticates a user and returns auth response
func (s *authService) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	// Perform validation
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Return generic error to prevent user enumeration
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	valid, err := s.passwordHasher.Verify(req.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(user.ID); err != nil {
		// Log the error but don't fail login
		// This is a non-critical operation
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.UUID)
	if err != nil {
		return nil, err
	}

	// Create auth response with safe user data (using ToResponse method)
	return &domain.AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *authService) ValidateToken(token string) (*security.Claims, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Verify user still exists and is active
	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserInactive
	}

	return claims, nil
}

// GetJWTManager returns the JWT manager for use in middleware
func (s *authService) GetJWTManager() *security.JWTManager {
	return s.jwtManager
}

// IsDuplicateEmailError checks if an error is a duplicate email error
func IsDuplicateEmailError(err error) bool {
	return errors.Is(err, repository.ErrUserAlreadyExists) ||
		errors.Is(err, gorm.ErrDuplicatedKey)
}
