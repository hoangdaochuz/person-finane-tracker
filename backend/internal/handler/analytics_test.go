package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
)

func setupAnalyticsRouter(handler *AnalyticsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/analytics/summary", handler.GetSummary)
	router.GET("/analytics/trends", handler.GetTrends)
	router.GET("/analytics/breakdown/source", handler.GetBreakdownBySource)
	router.GET("/analytics/breakdown/category", handler.GetBreakdownByCategory)
	router.GET("/transactions", handler.ListTransactions)
	router.GET("/transactions/:id", handler.GetTransactionByID)
	return router
}

// Test AnalyticsHandler GetSummary

func TestAnalyticsHandler_GetSummary_Success(t *testing.T) {
	expectedSummary := &domain.SummaryResponse{
		TotalIncome:      1000.00,
		TotalExpense:     500.00,
		CurrentBalance:   500.00,
		TransactionCount: 10,
	}

	mockService := &mockTransactionService{
		getSummaryFunc: func() (*domain.SummaryResponse, error) {
			return expectedSummary, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.SummaryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1000.00, response.TotalIncome)
	assert.Equal(t, 500.00, response.TotalExpense)
	assert.Equal(t, 500.00, response.CurrentBalance)
	assert.Equal(t, int64(10), response.TransactionCount)
}

func TestAnalyticsHandler_GetSummary_DatabaseError(t *testing.T) {
	mockService := &mockTransactionService{
		getSummaryFunc: func() (*domain.SummaryResponse, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

// Test AnalyticsHandler GetTrends

func TestAnalyticsHandler_GetTrends_DailySuccess(t *testing.T) {
	expectedTrends := &domain.TrendsResponse{
		Period: "daily",
		Data: []domain.TrendDataPoint{
			{Date: "2026-01-15", Income: 100, Expense: 50, Net: 50},
		},
	}

	mockService := &mockTransactionService{
		getTrendsFunc: func(period string) (*domain.TrendsResponse, error) {
			return expectedTrends, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/trends?period=daily", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.TrendsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "daily", response.Period)
	assert.Len(t, response.Data, 1)
}

func TestAnalyticsHandler_GetTrends_WeeklySuccess(t *testing.T) {
	expectedTrends := &domain.TrendsResponse{
		Period: "weekly",
		Data: []domain.TrendDataPoint{
			{Date: "2026-W02", Income: 500, Expense: 200, Net: 300},
		},
	}

	mockService := &mockTransactionService{
		getTrendsFunc: func(period string) (*domain.TrendsResponse, error) {
			return expectedTrends, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/trends?period=weekly", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetTrends_MonthlySuccess(t *testing.T) {
	expectedTrends := &domain.TrendsResponse{
		Period: "monthly",
		Data: []domain.TrendDataPoint{
			{Date: "2026-01", Income: 2000, Expense: 1000, Net: 1000},
		},
	}

	mockService := &mockTransactionService{
		getTrendsFunc: func(period string) (*domain.TrendsResponse, error) {
			return expectedTrends, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/trends?period=monthly", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetTrends_DefaultsToDaily(t *testing.T) {
	expectedTrends := &domain.TrendsResponse{
		Period: "daily",
		Data:   []domain.TrendDataPoint{},
	}

	mockService := &mockTransactionService{
		getTrendsFunc: func(period string) (*domain.TrendsResponse, error) {
			assert.Equal(t, "daily", period, "should default to daily when no period provided")
			return expectedTrends, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/trends", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_GetTrends_InvalidPeriod(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/trends?period=invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	assert.Contains(t, response["error"], "period must be one of")
}

func TestAnalyticsHandler_GetTrends_ServiceError(t *testing.T) {
	mockService := &mockTransactionService{
		getTrendsFunc: func(period string) (*domain.TrendsResponse, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/trends?period=daily", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Test AnalyticsHandler GetBreakdownBySource

func TestAnalyticsHandler_GetBreakdownBySource_Success(t *testing.T) {
	expectedBreakdown := []domain.BreakdownResponse{
		{Label: "Bank ABC", Amount: 500, Percentage: 50.0, Count: 5},
		{Label: "Bank XYZ", Amount: 300, Percentage: 30.0, Count: 3},
	}

	mockService := &mockTransactionService{
		getBreakdownSource: func() ([]domain.BreakdownResponse, error) {
			return expectedBreakdown, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/breakdown/source", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.BreakdownResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Bank ABC", response[0].Label)
}

func TestAnalyticsHandler_GetBreakdownBySource_DatabaseError(t *testing.T) {
	mockService := &mockTransactionService{
		getBreakdownSource: func() ([]domain.BreakdownResponse, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/breakdown/source", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Test AnalyticsHandler GetBreakdownByCategory

func TestAnalyticsHandler_GetBreakdownByCategory_Success(t *testing.T) {
	expectedBreakdown := []domain.BreakdownResponse{
		{Label: "Food", Amount: 300, Percentage: 30.0, Count: 10},
		{Label: "Transportation", Amount: 200, Percentage: 20.0, Count: 5},
	}

	mockService := &mockTransactionService{
		getBreakdownCategory: func() ([]domain.BreakdownResponse, error) {
			return expectedBreakdown, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/breakdown/category", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.BreakdownResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, "Food", response[0].Label)
}

func TestAnalyticsHandler_GetBreakdownByCategory_DatabaseError(t *testing.T) {
	mockService := &mockTransactionService{
		getBreakdownCategory: func() ([]domain.BreakdownResponse, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/analytics/breakdown/category", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Test AnalyticsHandler ListTransactions

func TestAnalyticsHandler_ListTransactions_DefaultPagination(t *testing.T) {
	expectedTxs := []domain.Transaction{
		{ID: 1, Amount: 100, Type: domain.TransactionTypeOut},
		{ID: 2, Amount: 200, Type: domain.TransactionTypeIn},
	}

	mockService := &mockTransactionService{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			assert.Equal(t, 1, params.Page, "should default to page 1")
			assert.Equal(t, 20, params.PageSize, "should default to page size 20")
			return expectedTxs, int64(2), nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "data")
	assert.Contains(t, response, "pagination")
}

func TestAnalyticsHandler_ListTransactions_WithPagination(t *testing.T) {
	expectedTxs := []domain.Transaction{{ID: 1}}

	mockService := &mockTransactionService{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			assert.Equal(t, 2, params.Page)
			assert.Equal(t, 10, params.PageSize)
			return expectedTxs, int64(25), nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions?page=2&page_size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	pagination := response["pagination"].(map[string]interface{})
	assert.Equal(t, float64(2), pagination["page"])
	assert.Equal(t, float64(10), pagination["page_size"])
	assert.Equal(t, float64(25), pagination["total"])
	assert.Equal(t, float64(3), pagination["total_pages"])
}

func TestAnalyticsHandler_ListTransactions_WithFilters(t *testing.T) {
	expectedTxs := []domain.Transaction{{ID: 1}}

	mockService := &mockTransactionService{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			assert.Equal(t, domain.TransactionTypeOut, params.Type)
			assert.Equal(t, "Bank ABC", params.Source)
			assert.Equal(t, "Food", params.Category)
			return expectedTxs, int64(1), nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions?type=out&source=Bank+ABC&category=Food", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_ListTransactions_WithDateFilters(t *testing.T) {
	expectedTxs := []domain.Transaction{{ID: 1}}

	mockService := &mockTransactionService{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			assert.Equal(t, "2026-01-01", params.StartDate)
			assert.Equal(t, "2026-01-31", params.EndDate)
			return expectedTxs, int64(1), nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAnalyticsHandler_ListTransactions_ServiceError(t *testing.T) {
	mockService := &mockTransactionService{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Test AnalyticsHandler GetTransactionByID

func TestAnalyticsHandler_GetTransactionByID_Success(t *testing.T) {
	expectedTx := &domain.Transaction{
		ID:     1,
		Amount: 100.50,
		Type:   domain.TransactionTypeOut,
		Source: "Bank ABC",
	}

	mockService := &mockTransactionService{
		findByIDFunc: func(id int64) (*domain.Transaction, error) {
			assert.Equal(t, int64(1), id)
			return expectedTx, nil
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response domain.Transaction
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), response.ID)
}

func TestAnalyticsHandler_GetTransactionByID_InvalidID(t *testing.T) {
	mockService := &mockTransactionService{}
	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestAnalyticsHandler_GetTransactionByID_NotFound(t *testing.T) {
	mockService := &mockTransactionService{
		findByIDFunc: func(id int64) (*domain.Transaction, error) {
			return nil, errors.New("transaction not found")
		},
	}

	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetTransactionByID_NegativeID(t *testing.T) {
	mockService := &mockTransactionService{
		findByIDFunc: func(id int64) (*domain.Transaction, error) {
			return nil, errors.New("invalid transaction ID")
		},
	}
	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions/-1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be processed as -1 is a valid number format but service returns error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAnalyticsHandler_GetTransactionByID_ZeroID(t *testing.T) {
	mockService := &mockTransactionService{
		findByIDFunc: func(id int64) (*domain.Transaction, error) {
			return nil, errors.New("invalid transaction ID")
		},
	}
	handler := NewAnalyticsHandler(mockService)
	router := setupAnalyticsRouter(handler)

	req := httptest.NewRequest("GET", "/transactions/0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should be processed as 0 is a valid number format but service returns error
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
