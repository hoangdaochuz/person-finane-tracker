package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
)

// Mock service for testing
type mockTransactionService struct {
	createFunc           func(req *domain.CreateTransactionRequest) (*domain.Transaction, error)
	createBatchFunc      func(req *domain.BatchTransactionRequest) ([]domain.Transaction, error)
	findByIDFunc         func(id int64) (*domain.Transaction, error)
	listFunc             func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error)
	getSummaryFunc       func() (*domain.SummaryResponse, error)
	getTrendsFunc        func(period string) (*domain.TrendsResponse, error)
	getBreakdownSource   func() ([]domain.BreakdownResponse, error)
	getBreakdownCategory func() ([]domain.BreakdownResponse, error)
}

func (m *mockTransactionService) CreateTransaction(req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
	if m.createFunc != nil {
		return m.createFunc(req)
	}
	return &domain.Transaction{ID: 1}, nil
}

func (m *mockTransactionService) CreateBatchTransaction(req *domain.BatchTransactionRequest) ([]domain.Transaction, error) {
	if m.createBatchFunc != nil {
		return m.createBatchFunc(req)
	}
	return []domain.Transaction{{ID: 1}}, nil
}

func (m *mockTransactionService) GetTransactionByID(id int64) (*domain.Transaction, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return &domain.Transaction{ID: id}, nil
}

func (m *mockTransactionService) ListTransactions(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(params)
	}
	return []domain.Transaction{}, 0, nil
}

func (m *mockTransactionService) GetSummary() (*domain.SummaryResponse, error) {
	if m.getSummaryFunc != nil {
		return m.getSummaryFunc()
	}
	return &domain.SummaryResponse{}, nil
}

func (m *mockTransactionService) GetTrends(period string) (*domain.TrendsResponse, error) {
	if m.getTrendsFunc != nil {
		return m.getTrendsFunc(period)
	}
	return &domain.TrendsResponse{}, nil
}

func (m *mockTransactionService) GetBreakdownBySource() ([]domain.BreakdownResponse, error) {
	if m.getBreakdownSource != nil {
		return m.getBreakdownSource()
	}
	return []domain.BreakdownResponse{}, nil
}

func (m *mockTransactionService) GetBreakdownByCategory() ([]domain.BreakdownResponse, error) {
	if m.getBreakdownCategory != nil {
		return m.getBreakdownCategory()
	}
	return []domain.BreakdownResponse{}, nil
}

func setupTestRouter(handler *WebhookHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/webhook/transaction", handler.CreateTransaction)
	router.POST("/webhook/batch", handler.CreateBatchTransaction)
	return router
}

// Test WebhookHandler CreateTransaction

func TestWebhookHandler_CreateTransaction_Success(t *testing.T) {
	now := time.Now()
	expectedTx := &domain.Transaction{
		ID:              1,
		Amount:          100.50,
		Type:            domain.TransactionTypeOut,
		Category:        "Food",
		Description:     "Lunch",
		Source:          "Bank ABC",
		TransactionDate: now,
	}

	mockService := &mockTransactionService{
		createFunc: func(req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
			return expectedTx, nil
		},
	}

	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"amount": 100.50,
		"type": "out",
		"category": "Food",
		"description": "Lunch",
		"source": "Bank ABC",
		"transaction_date": "` + now.Format(time.RFC3339) + `"
	}`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response domain.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
	assert.Equal(t, 100.50, response.Amount)
}

func TestWebhookHandler_CreateTransaction_InvalidJSON(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `invalid json`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestWebhookHandler_CreateTransaction_MissingRequiredField(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"amount": 100.50,
		"type": "out"
	}`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_CreateTransaction_InvalidType(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"amount": 100.50,
		"type": "invalid",
		"source": "Bank ABC",
		"transaction_date": "2026-01-15T12:00:00Z"
	}`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_CreateTransaction_ValidationError(t *testing.T) {
	mockService := &mockTransactionService{
		createFunc: func(req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
			return nil, &domain.ValidationError{
				Field:   "category",
				Message: "invalid category",
			}
		},
	}

	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"amount": 100.50,
		"type": "out",
		"category": "InvalidCategory",
		"source": "Bank ABC",
		"transaction_date": "2026-01-15T12:00:00Z"
	}`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestWebhookHandler_CreateTransaction_ServiceError(t *testing.T) {
	mockService := &mockTransactionService{
		createFunc: func(req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"amount": 100.50,
		"type": "out",
		"source": "Bank ABC",
		"transaction_date": "2026-01-15T12:00:00Z"
	}`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestWebhookHandler_CreateTransaction_NegativeAmount(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"amount": -10,
		"type": "out",
		"source": "Bank ABC",
		"transaction_date": "2026-01-15T12:00:00Z"
	}`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_CreateTransaction_ZeroAmount(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"amount": 0,
		"type": "out",
		"source": "Bank ABC",
		"transaction_date": "2026-01-15T12:00:00Z"
	}`

	req := httptest.NewRequest("POST", "/webhook/transaction", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Test WebhookHandler CreateBatchTransaction

func TestWebhookHandler_CreateBatchTransaction_Success(t *testing.T) {
	now := time.Now()
	expectedTxs := []domain.Transaction{
		{ID: 1, Amount: 100, Type: domain.TransactionTypeOut},
		{ID: 2, Amount: 200, Type: domain.TransactionTypeIn},
	}

	mockService := &mockTransactionService{
		createBatchFunc: func(req *domain.BatchTransactionRequest) ([]domain.Transaction, error) {
			return expectedTxs, nil
		},
	}

	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"transactions": [
			{
				"amount": 100,
				"type": "out",
				"source": "Bank ABC",
				"transaction_date": "` + now.Format(time.RFC3339) + `"
			},
			{
				"amount": 200,
				"type": "in",
				"source": "Bank XYZ",
				"transaction_date": "` + now.Format(time.RFC3339) + `"
			}
		]
	}`

	req := httptest.NewRequest("POST", "/webhook/batch", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["created"])
	assert.Contains(t, response, "transactions")
}

func TestWebhookHandler_CreateBatchTransaction_EmptyBatch(t *testing.T) {
	mockService := &mockTransactionService{
		createBatchFunc: func(req *domain.BatchTransactionRequest) ([]domain.Transaction, error) {
			return nil, &domain.ValidationError{
				Field:   "transactions",
				Message: "at least one transaction is required",
			}
		},
	}

	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"transactions": []
	}`

	req := httptest.NewRequest("POST", "/webhook/batch", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Gin's binding validation catches empty array before service layer
	// So we get 400 Bad Request instead of 500
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_CreateBatchTransaction_InvalidJSON(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `invalid json`

	req := httptest.NewRequest("POST", "/webhook/batch", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_CreateBatchTransaction_ValidationError(t *testing.T) {
	mockService := &mockTransactionService{
		createBatchFunc: func(req *domain.BatchTransactionRequest) ([]domain.Transaction, error) {
			return nil, &domain.ValidationError{
				Field:   "transactions",
				Message: "invalid category",
				Index:   0,
			}
		},
	}

	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"transactions": [
			{
				"amount": 100,
				"type": "out",
				"category": "InvalidCategory",
				"source": "Bank ABC",
				"transaction_date": "2026-01-15T12:00:00Z"
			}
		]
	}`

	req := httptest.NewRequest("POST", "/webhook/batch", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestWebhookHandler_CreateBatchTransaction_ExceedsMaximum(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	// Create 101 transactions (exceeds max of 100)
	transactions := make([]map[string]interface{}, 101)
	for i := 0; i < 101; i++ {
		transactions[i] = map[string]interface{}{
			"amount":           float64(100),
			"type":             "out",
			"source":           "Bank ABC",
			"transaction_date": "2026-01-15T12:00:00Z",
		}
	}

	body, _ := json.Marshal(map[string]interface{}{
		"transactions": transactions,
	})

	req := httptest.NewRequest("POST", "/webhook/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestWebhookHandler_CreateBatchTransaction_ServiceError(t *testing.T) {
	mockService := &mockTransactionService{
		createBatchFunc: func(req *domain.BatchTransactionRequest) ([]domain.Transaction, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewWebhookHandler(mockService)
	router := setupTestRouter(handler)

	body := `{
		"transactions": [
			{
				"amount": 100,
				"type": "out",
				"source": "Bank ABC",
				"transaction_date": "2026-01-15T12:00:00Z"
			}
		]
	}`

	req := httptest.NewRequest("POST", "/webhook/batch", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
