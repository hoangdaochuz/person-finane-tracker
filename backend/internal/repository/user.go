package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/security"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when attempting to create a user with duplicate email
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// UserRepository handles database operations for users
type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByUUID(uuid string) (*domain.User, error)
	FindByID(id int64) (*domain.User, error)
	UpdateLastLogin(userID int64) error
	Update(user *domain.User) error
}

type userRepository struct {
	db         *gorm.DB
	sanitizer  *security.Sanitizer
	apiKeyGen  *security.APIKeyGenerator
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db:         db,
		sanitizer:  security.NewSanitizer(),
		apiKeyGen:  security.NewAPIKeyGenerator(),
	}
}

func (r *userRepository) Create(user *domain.User) error {
	// Sanitize email
	user.Email = r.sanitizer.CleanInput(user.Email, domain.MaxEmailLength)

	// Sanitize name if provided
	if user.Name != "" {
		user.Name = r.sanitizer.CleanInput(user.Name, domain.MaxNameLength)
	}

	// Generate API key if not provided
	if user.APIKey == "" {
		apiKey, err := r.apiKeyGen.Generate()
		if err != nil {
			return err
		}
		user.APIKey = apiKey
	}

	// Check if user with this email already exists
	var existingUser domain.User
	err := r.db.Where("email = ?", user.Email).First(&existingUser).Error
	if err == nil {
		// User found, return duplicate error
		return ErrUserAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other database error
		return err
	}

	// Create the user
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	// Sanitize email
	safeEmail := r.sanitizer.CleanInput(email, domain.MaxEmailLength)

	var user domain.User
	err := r.db.Where("email = ?", safeEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByUUID(uuidStr string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("uuid = ?", uuidStr).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByID(id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateLastLogin(userID int64) error {
	now := time.Now()
	return r.db.Model(&domain.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now).Error
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}
