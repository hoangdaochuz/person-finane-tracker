package security

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/handler"
	"github.com/dev/personal-finance-tracker/backend/internal/middleware"
)

// Mock service for security testing
type mockSecurityService struct {
	createFunc func(req *domain.CreateTransactionRequest) (*domain.Transaction, error)
}

func (m *mockSecurityService) CreateTransaction(req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
	// Call validation to replicate real service behavior
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if m.createFunc != nil {
		return m.createFunc(req)
	}
	return &domain.Transaction{ID: 1}, nil
}

func (m *mockSecurityService) CreateBatchTransaction(req *domain.BatchTransactionRequest) ([]domain.Transaction, error) {
	return []domain.Transaction{{ID: 1}}, nil
}

func (m *mockSecurityService) GetTransactionByID(id int64) (*domain.Transaction, error) {
	return &domain.Transaction{ID: id}, nil
}

func (m *mockSecurityService) ListTransactions(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
	return []domain.Transaction{}, 0, nil
}

func (m *mockSecurityService) GetSummary() (*domain.SummaryResponse, error) {
	return &domain.SummaryResponse{}, nil
}

func (m *mockSecurityService) GetTrends(period string) (*domain.TrendsResponse, error) {
	return &domain.TrendsResponse{}, nil
}

func (m *mockSecurityService) GetBreakdownBySource() ([]domain.BreakdownResponse, error) {
	return []domain.BreakdownResponse{}, nil
}

func (m *mockSecurityService) GetBreakdownByCategory() ([]domain.BreakdownResponse, error) {
	return []domain.BreakdownResponse{}, nil
}

func setupSecurityRouter(apiKey string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{AllowedOrigins: []string{"*"}}))
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.APIKeyAuth(apiKey))

	mockService := &mockSecurityService{}
	webhookHandler := handler.NewWebhookHandler(mockService)
	analyticsHandler := handler.NewAnalyticsHandler(mockService)

	router.POST("/webhook/transaction", webhookHandler.CreateTransaction)
	router.GET("/analytics/summary", analyticsHandler.GetSummary)
	router.GET("/transactions/:id", analyticsHandler.GetTransactionByID)

	return router
}

// Test Authentication Tests

func TestSecurity_RequestWithoutAPIKey(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestSecurity_RequestWithWrongAPIKey(t *testing.T) {
	router := setupSecurityRouter("correct-api-key")

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	req.Header.Set("X-API-Key", "wrong-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestSecurity_RequestWithEmptyAPIKey(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	req.Header.Set("X-API-Key", "")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSecurity_RequestWithCorrectAPIKey(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
}

func TestSecurity_APIKeyWithWhitespace(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	// Test with leading/trailing whitespace
	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	req.Header.Set("X-API-Key", " test-api-key ")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail because keys don't match with whitespace
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSecurity_APIKeyCaseSensitive(t *testing.T) {
	router := setupSecurityRouter("Test-API-Key")

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail because keys are case-sensitive
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test Input Validation Security

func TestSQL_InjectionInSourceField(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	// Attempt SQL injection in source field
	payload := map[string]interface{}{
		"amount":           100,
		"type":             "out",
		"source":           "Bank'; DROP TABLE transactions; --",
		"transaction_date": "2026-01-15T12:00:00Z",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Request should either be rejected by validation or handled safely
	// The important thing is it shouldn't cause a 500 error or crash
	assert.NotEqual(t, http.StatusInternalServerError, w.Code)
}

func TestSQL_InjectionInCategoryField(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	payload := map[string]interface{}{
		"amount":           100,
		"type":             "out",
		"category":         "Food' OR '1'='1",
		"source":           "Bank",
		"transaction_date": "2026-01-15T12:00:00Z",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Invalid category should be rejected with 400 (validation error)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestXSS_InDescriptionField(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
		"<svg onload=alert('xss')>",
	}

	for _, payload := range xssPayloads {
		t.Run(payload, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"amount":           100,
				"type":             "out",
				"source":           "Bank",
				"description":      payload,
				"transaction_date": "2026-01-15T12:00:00Z",
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-api-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// The request should be handled without errors
			// Response should not execute the script
			assert.NotEqual(t, http.StatusInternalServerError, w.Code)
		})
	}
}

// Test Input Size Limits

func TestSecurity_OversizedDescription(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	// Create description that exceeds max length
	oversizedDescription := string(make([]byte, 10000))

	payload := map[string]interface{}{
		"amount":           100,
		"type":             "out",
		"source":           "Bank",
		"description":      oversizedDescription,
		"transaction_date": "2026-01-15T12:00:00Z",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be rejected by validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecurity_OversizedSource(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	oversizedSource := string(make([]byte, 200))

	payload := map[string]interface{}{
		"amount":           100,
		"type":             "out",
		"source":           oversizedSource,
		"transaction_date": "2026-01-15T12:00:00Z",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be rejected by validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Special Characters

func TestSecurity_SpecialCharactersInFields(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	specialStrings := []string{
		"../../etc/passwd",
		"C:\\Windows\\System32\\config\\sam",
		"${HOME}",
		"`whoami`",
		";ls -la;",
		"| cat /etc/passwd",
		"$(rm -rf /)",
		"`touch /tmp/pwned`",
	}

	for _, specialStr := range specialStrings {
		t.Run(specialStr, func(t *testing.T) {
			payload := map[string]interface{}{
				"amount":           100,
				"type":             "out",
				"source":           "Bank",
				"description":      specialStr,
				"transaction_date": "2026-01-15T12:00:00Z",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-api-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should be handled safely without crashes
			assert.NotEqual(t, http.StatusInternalServerError, w.Code)
		})
	}
}

// Test Negative Amount

func TestSecurity_NegativeAmount(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	payload := map[string]interface{}{
		"amount":           -100,
		"type":             "out",
		"source":           "Bank",
		"transaction_date": "2026-01-15T12:00:00Z",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be rejected by validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSecurity_ZeroAmount(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	payload := map[string]interface{}{
		"amount":           0,
		"type":             "out",
		"source":           "Bank",
		"transaction_date": "2026-01-15T12:00:00Z",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be rejected by validation (amount must be > 0)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test Malformed JSON

func TestSecurity_MalformedJSON(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	malformedJSONs := []string{
		"{{{",
		"}",
		"{{}}",
		"{'amount': 100}", // Single quotes instead of double
		"{amount: 100}",   // Missing quotes
		"null",
		"undefined",
	}

	for _, malformed := range malformedJSONs {
		t.Run(malformed, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(malformed)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-api-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should return bad request
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// Test Future Date Prevention

func TestSecurity_FutureDateRejected(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	futureDates := []string{
		"2099-01-01T00:00:00Z",
		"2027-12-31T23:59:59Z",
		"3000-01-01T00:00:00Z",
	}

	for _, futureDate := range futureDates {
		t.Run(futureDate, func(t *testing.T) {
			payload := map[string]interface{}{
				"amount":           100,
				"type":             "out",
				"source":           "Bank",
				"transaction_date": futureDate,
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-api-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should be rejected with 400 (validation error)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// Test Invalid Transaction Type

func TestSecurity_InvalidTransactionType(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	invalidTypes := []string{
		"admin",
		"root",
		"su",
		"exec",
		"eval",
		"system",
		"IN", // Uppercase should fail (only lowercase)
		"OUT",
		"",
	}

	for _, txType := range invalidTypes {
		t.Run(txType, func(t *testing.T) {
			payload := map[string]interface{}{
				"amount":           100,
				"type":             txType,
				"source":           "Bank",
				"transaction_date": "2026-01-15T12:00:00Z",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-API-Key", "test-api-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Should be rejected
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// Test Path Traversal in ID Parameter

func TestSecurity_PathTraversalInID(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	pathTraversalAttempts := []string{
		"../../etc/passwd",
		"....//....//....//etc/passwd",
		"%2e%2e%2fetc/passwd",
		"..%5c..%5c..%5cetc/passwd",
	}

	for _, attempt := range pathTraversalAttempts {
		t.Run(attempt, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/transactions/"+attempt, nil)
			req.Header.Set("X-API-Key", "test-api-key")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// Gin returns 404 for paths that don't match the route pattern
			// This is correct behavior - the path doesn't match /transactions/:id
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}

// Test CORS

func TestSecurity_CORSHeadersSet(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	req := httptest.NewRequest("OPTIONS", "/analytics/summary", nil)
	req.Header.Set("Origin", "http://malicious-site.com")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// CORS headers should be present
	origin := w.Header().Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, origin)
}

func TestSecurity_POSTRequestWithCORS(t *testing.T) {
	router := setupSecurityRouter("test-api-key")

	payload := map[string]interface{}{
		"amount":           100,
		"type":             "out",
		"source":           "Bank",
		"transaction_date": "2026-01-15T12:00:00Z",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader(body))
	req.Header.Set("Origin", "http://malicious-site.com")
	req.Header.Set("X-API-Key", "test-api-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Even with different origin, if API key is correct, request should succeed
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}
