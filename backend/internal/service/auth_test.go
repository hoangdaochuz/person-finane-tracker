package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/internal/security"
)

// mockUserRepository is a mock implementation of UserRepository for testing
type mockUserRepository struct {
	createUser         *domain.User
	createErr          error
	findByEmailUser    *domain.User
	findByEmailErr     error
	findByUUIDUser     *domain.User
	findByUUIDErr      error
	findByIDUser       *domain.User
	findByIDErr        error
	updateLastLoginErr error
	updateUser         *domain.User
	updateErr          error
}

func (m *mockUserRepository) Create(user *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	// Simulate database setting ID
	user.ID = 1
	// Simulate API key generation if empty
	if user.APIKey == "" {
		user.APIKey = "generated-api-key-12345"
	}
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*domain.User, error) {
	if m.findByEmailErr != nil {
		return nil, m.findByEmailErr
	}
	return m.findByEmailUser, nil
}

func (m *mockUserRepository) FindByUUID(uuidStr string) (*domain.User, error) {
	if m.findByUUIDErr != nil {
		return nil, m.findByUUIDErr
	}
	return m.findByUUIDUser, nil
}

func (m *mockUserRepository) FindByID(id int64) (*domain.User, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.findByIDUser, nil
}

func (m *mockUserRepository) UpdateLastLogin(userID int64) error {
	return m.updateLastLoginErr
}

func (m *mockUserRepository) Update(user *domain.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	return nil
}

// helper to create a test user with hashed password
func createTestUser(t *testing.T, email, password string) *domain.User {
	t.Helper()
	hasher := security.NewPasswordHasher()
	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	return &domain.User{
		ID:           1,
		UUID:         userUUID,
		Email:        email,
		PasswordHash: hash,
		Name:         "Test User",
		APIKey:       "test-api-key",
		IsActive:     true,
	}
}

// Test Register()

func TestAuthService_Register_Success(t *testing.T) {
	mockRepo := &mockUserRepository{
		findByEmailErr: repository.ErrUserNotFound, // User doesn't exist
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.RegisterRequest{
		Email:    "newuser@example.com",
		Password: "Password123",
		Name:     "New User",
	}

	response, err := authService.Register(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "newuser@example.com", response.User.Email)
	assert.Equal(t, "New User", response.User.Name)
	assert.NotEmpty(t, response.User.APIKey, "API key should be generated")
}

func TestAuthService_Register_ValidationError(t *testing.T) {
	mockRepo := &mockUserRepository{}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.RegisterRequest{
		Email:    "invalid-email",
		Password: "short",
	}

	response, err := authService.Register(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	var validationErr *domain.ValidationError
	assert.True(t, errors.As(err, &validationErr))
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	existingUser := createTestUser(t, "existing@example.com", "Password123")

	mockRepo := &mockUserRepository{
		findByEmailUser: existingUser,
		findByEmailErr:  nil, // User found
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.RegisterRequest{
		Email:    "existing@example.com",
		Password: "Password123",
	}

	response, err := authService.Register(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, repository.ErrUserAlreadyExists, err)
}

// Test Login()

func TestAuthService_Login_Success(t *testing.T) {
	testUser := createTestUser(t, "test@example.com", "CorrectPassword123")

	mockRepo := &mockUserRepository{
		findByEmailUser: testUser,
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.LoginRequest{
		Email:    "test@example.com",
		Password: "CorrectPassword123",
	}

	response, err := authService.Login(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "test@example.com", response.User.Email)
}

func TestAuthService_Login_ValidationError(t *testing.T) {
	mockRepo := &mockUserRepository{}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.LoginRequest{
		Email:    "invalid-email",
		Password: "",
	}

	response, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	var validationErr *domain.ValidationError
	assert.True(t, errors.As(err, &validationErr))
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		findByEmailErr: repository.ErrUserNotFound,
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "Password123",
	}

	response, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	testUser := createTestUser(t, "test@example.com", "CorrectPassword123")

	mockRepo := &mockUserRepository{
		findByEmailUser: testUser,
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword456",
	}

	response, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrInvalidCredentials, err)
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	testUser := createTestUser(t, "test@example.com", "Password123")
	testUser.IsActive = false

	mockRepo := &mockUserRepository{
		findByEmailUser: testUser,
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	req := &domain.LoginRequest{
		Email:    "test@example.com",
		Password: "Password123",
	}

	response, err := authService.Login(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrUserInactive, err)
}

// Test ValidateToken()

func TestAuthService_ValidateToken_ValidToken(t *testing.T) {
	testUser := createTestUser(t, "test@example.com", "Password123")

	mockRepo := &mockUserRepository{
		findByIDUser: testUser,
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	// First, generate a valid token
	token, err := authService.GetJWTManager().GenerateToken(testUser.ID, testUser.Email, testUser.UUID)
	assert.NoError(t, err)

	// Now validate it
	claims, err := authService.ValidateToken(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, testUser.ID, claims.UserID)
	assert.Equal(t, testUser.Email, claims.Email)
}

func TestAuthService_ValidateToken_UserNotFound(t *testing.T) {
	mockRepo := &mockUserRepository{
		findByIDErr: repository.ErrUserNotFound,
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	// Create a token for a non-existent user
	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	token, _ := jwtManager.GenerateToken(999, "nonexistent@example.com", userUUID)

	claims, err := authService.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "not found")
}

func TestAuthService_ValidateToken_InactiveUser(t *testing.T) {
	testUser := createTestUser(t, "test@example.com", "Password123")
	testUser.IsActive = false

	mockRepo := &mockUserRepository{
		findByIDUser: testUser,
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	// Create a token for inactive user
	token, _ := authService.GetJWTManager().GenerateToken(testUser.ID, testUser.Email, testUser.UUID)

	claims, err := authService.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, ErrUserInactive, err)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	mockRepo := &mockUserRepository{}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	claims, err := authService.ValidateToken("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

// Test IsDuplicateEmailError()

func TestIsDuplicateEmailError_DuplicateKey(t *testing.T) {
	err := IsDuplicateEmailError(repository.ErrUserAlreadyExists)
	assert.True(t, err)
}

func TestIsDuplicateEmailError_GormErrDuplicatedKey(t *testing.T) {
	err := IsDuplicateEmailError(gorm.ErrDuplicatedKey)
	assert.True(t, err)
}

func TestIsDuplicateEmailError_OtherError(t *testing.T) {
	err := IsDuplicateEmailError(errors.New("some other error"))
	assert.False(t, err)
}

// Test GetJWTManager()

func TestAuthService_GetJWTManager_ReturnsManager(t *testing.T) {
	mockRepo := &mockUserRepository{}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	manager := authService.GetJWTManager()

	assert.NotNil(t, manager)
	assert.IsType(t, &security.JWTManager{}, manager)
}

// Integration-style tests

func TestAuthService_RegisterLoginFlow(t *testing.T) {
	// Test that a registered user can log in
	mockRepo := &mockUserRepository{
		findByEmailErr: repository.ErrUserNotFound, // User doesn't exist initially
	}
	authService := NewAuthService(mockRepo, "test-jwt-secret-minimum-32-chars")

	// Register
	registerReq := &domain.RegisterRequest{
		Email:    "flowtest@example.com",
		Password: "Password123",
		Name:     "Flow Test",
	}

	registerResp, err := authService.Register(registerReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, registerResp.Token)

	// Now simulate finding the user for login
	testUser := createTestUser(t, registerReq.Email, registerReq.Password)
	mockRepo.findByEmailUser = testUser
	mockRepo.findByEmailErr = nil

	// Login with same credentials
	loginReq := &domain.LoginRequest{
		Email:    "flowtest@example.com",
		Password: "Password123",
	}

	loginResp, err := authService.Login(loginReq)

	assert.NoError(t, err)
	assert.NotNil(t, loginResp)
	assert.NotEmpty(t, loginResp.Token)
	// Tokens should be different due to different issued times
	assert.NotEqual(t, registerResp.Token, loginResp.Token)
}
