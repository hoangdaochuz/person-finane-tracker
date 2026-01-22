package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/internal/security"
	"github.com/dev/personal-finance-tracker/backend/internal/service"
)

// mockAuthService is a mock implementation of AuthService for testing
type mockAuthService struct {
	registerResp *domain.AuthResponse
	registerErr  error
	loginResp    *domain.AuthResponse
	loginErr     error
	validateResp *security.Claims
	validateErr  error
	jwtManager   *security.JWTManager
}

func (m *mockAuthService) Register(req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	if m.registerErr != nil {
		return nil, m.registerErr
	}
	return m.registerResp, nil
}

func (m *mockAuthService) Login(req *domain.LoginRequest) (*domain.AuthResponse, error) {
	if m.loginErr != nil {
		return nil, m.loginErr
	}
	return m.loginResp, nil
}

func (m *mockAuthService) ValidateToken(token string) (*security.Claims, error) {
	if m.validateErr != nil {
		return nil, m.validateErr
	}
	return m.validateResp, nil
}

func (m *mockAuthService) GetJWTManager() *security.JWTManager {
	if m.jwtManager == nil {
		return security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	}
	return m.jwtManager
}

// setupAuthTestRouter creates a test router with the auth handler
func setupAuthTestRouter(authService service.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	authHandler := NewAuthHandler(authService)
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	return router
}

// helper to create a test auth response
func createTestAuthResponse(t *testing.T) *domain.AuthResponse {
	t.Helper()
	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	jwtManager := security.NewJWTManager("test-jwt-secret-minimum-32-chars")
	token, err := jwtManager.GenerateToken(1, "test@example.com", userUUID)
	assert.NoError(t, err)

	return &domain.AuthResponse{
		Token: token,
		User: domain.UserResponse{
			ID:       1,
			UUID:     userUUID,
			Email:    "test@example.com",
			Name:     "Test User",
			APIKey:   "test-api-key",
			IsActive: true,
		},
	}
}

// Test Register()

func TestAuthHandler_Register_Success(t *testing.T) {
	mockAuth := &mockAuthService{
		registerResp: createTestAuthResponse(t),
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
		"name":     "New User",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response domain.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "test@example.com", response.User.Email)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	mockAuth := &mockAuthService{
		registerErr: &domain.ValidationError{
			Field:   "email",
			Message: "invalid email format",
		},
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "invalid-email",
		"password": "short",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	mockAuth := &mockAuthService{
		registerErr: repository.ErrUserAlreadyExists,
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "existing@example.com",
		"password": "Password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "user with this email already exists", response["error"])
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthService{}
	router := setupAuthTestRouter(mockAuth)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Login()

func TestAuthHandler_Login_Success(t *testing.T) {
	mockAuth := &mockAuthService{
		loginResp: createTestAuthResponse(t),
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, "test@example.com", response.User.Email)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	mockAuth := &mockAuthService{
		loginErr: service.ErrInvalidCredentials,
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "WrongPassword",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "invalid email or password", response["error"])
}

func TestAuthHandler_Login_InactiveUser(t *testing.T) {
	mockAuth := &mockAuthService{
		loginErr: service.ErrUserInactive,
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "inactive@example.com",
		"password": "Password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	// Should use generic message to prevent user enumeration
	assert.Equal(t, "invalid email or password", response["error"])
}

func TestAuthHandler_Login_ValidationError(t *testing.T) {
	mockAuth := &mockAuthService{
		loginErr: &domain.ValidationError{
			Field:   "email",
			Message: "invalid email format",
		},
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "valid@example.com",
		"password": "ValidPassword123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Contains(t, response, "field")
}

func TestAuthHandler_Login_BindingError(t *testing.T) {
	mockAuth := &mockAuthService{}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "invalid-email",
		"password": "",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthService{}
	router := setupAuthTestRouter(mockAuth)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test error scenarios

func TestAuthHandler_Register_ServiceError(t *testing.T) {
	mockAuth := &mockAuthService{
		registerErr: errors.New("database error"),
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "failed to create user", response["error"])
}

func TestAuthHandler_Login_ServiceError(t *testing.T) {
	mockAuth := &mockAuthService{
		loginErr: errors.New("database connection failed"),
	}
	router := setupAuthTestRouter(mockAuth)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "Password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Equal(t, "authentication failed", response["error"])
}
