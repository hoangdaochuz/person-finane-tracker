package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/handler"
	"github.com/dev/personal-finance-tracker/backend/internal/middleware"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/internal/service"
	"github.com/dev/personal-finance-tracker/backend/test/util"
)

func setupTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	pc := util.SetupPostgresContainerWithDefaults(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})

	// Setup dependencies
	repo := repository.NewTransactionRepository(db)
	svc := service.NewTransactionService(repo)

	// Setup handlers and routes
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{AllowedOrigins: []string{"*"}}))

	// Use test API key
	testAPIKey := "test-api-key"
	router.Use(middleware.APIKeyAuth(testAPIKey))
	router.Use(middleware.ErrorHandler())

	webhookHandler := handler.NewWebhookHandler(svc)
	analyticsHandler := handler.NewAnalyticsHandler(svc)

	router.POST("/webhook/transaction", webhookHandler.CreateTransaction)
	router.POST("/webhook/batch", webhookHandler.CreateBatchTransaction)
	router.GET("/analytics/summary", analyticsHandler.GetSummary)
	router.GET("/analytics/trends", analyticsHandler.GetTrends)
	router.GET("/analytics/breakdown/source", analyticsHandler.GetBreakdownBySource)
	router.GET("/analytics/breakdown/category", analyticsHandler.GetBreakdownByCategory)
	router.GET("/transactions", analyticsHandler.ListTransactions)
	router.GET("/transactions/:id", analyticsHandler.GetTransactionByID)

	server := httptest.NewServer(router)
	return server, testAPIKey
}

func makeRequest(t *testing.T, server *httptest.Server, method, path string, body interface{}, apiKey string) *http.Response {
	t.Helper()

	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	url := server.URL + path
	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// Test API Integration: Create transaction then retrieve it

func TestAPI_Integration_CreateAndRetrieve(t *testing.T) {
	server, apiKey := setupTestServer(t)
	defer server.Close()

	t.Run("Create transaction", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount":           100.50,
			"type":             "out",
			"category":         "Food",
			"description":      "Test lunch",
			"source":           "Test Bank",
			"transaction_date": time.Now().Format(time.RFC3339),
		}

		resp := makeRequest(t, server, "POST", "/webhook/transaction", reqBody, apiKey)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var txResp domain.Transaction
		err := json.NewDecoder(resp.Body).Decode(&txResp)
		require.NoError(t, err)
		assert.Greater(t, txResp.ID, int64(0))

		t.Run("Retrieve transaction", func(t *testing.T) {
			resp = makeRequest(t, server, "GET", fmt.Sprintf("/transactions/%d", txResp.ID), nil, apiKey)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var getResp domain.Transaction
			err = json.NewDecoder(resp.Body).Decode(&getResp)
			require.NoError(t, err)
			assert.Equal(t, txResp.ID, getResp.ID)
			assert.Equal(t, 100.50, getResp.Amount)
		})
	})
}

// Test API Integration: Batch create then list

func TestAPI_Integration_BatchCreateAndList(t *testing.T) {
	server, apiKey := setupTestServer(t)
	defer server.Close()

	reqBody := map[string]interface{}{
		"transactions": []map[string]interface{}{
			{
				"amount":           100,
				"type":             "out",
				"category":         "Food",
				"source":           "Bank A",
				"transaction_date": time.Now().Format(time.RFC3339),
			},
			{
				"amount":           200,
				"type":             "out",
				"category":         "Transportation",
				"source":           "Bank B",
				"transaction_date": time.Now().Format(time.RFC3339),
			},
			{
				"amount":           300,
				"type":             "in",
				"category":         "Salary",
				"source":           "Bank A",
				"transaction_date": time.Now().Format(time.RFC3339),
			},
		},
	}

	resp := makeRequest(t, server, "POST", "/webhook/batch", reqBody, apiKey)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var batchResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&batchResp)
	require.NoError(t, err)
	assert.Equal(t, float64(3), batchResp["created"])

	t.Run("List all transactions", func(t *testing.T) {
		resp = makeRequest(t, server, "GET", "/transactions?page=1&page_size=10", nil, apiKey)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		data := listResp["data"].([]interface{})
		assert.Len(t, data, 3)

		pagination := listResp["pagination"].(map[string]interface{})
		assert.Equal(t, float64(3), pagination["total"])
	})

	t.Run("Filter by type", func(t *testing.T) {
		resp = makeRequest(t, server, "GET", "/transactions?type=out", nil, apiKey)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var listResp map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&listResp)
		require.NoError(t, err)

		data := listResp["data"].([]interface{})
		assert.Len(t, data, 2)

		pagination := listResp["pagination"].(map[string]interface{})
		assert.Equal(t, float64(2), pagination["total"])
	})
}

// Test API Integration: Create then query analytics

func TestAPI_Integration_CreateThenQueryAnalytics(t *testing.T) {
	server, apiKey := setupTestServer(t)
	defer server.Close()

	// Create transactions
	reqBody := map[string]interface{}{
		"transactions": []map[string]interface{}{
			{"amount": 1000, "type": "in", "category": "Salary", "source": "Bank A", "transaction_date": time.Now().Format(time.RFC3339)},
			{"amount": 500, "type": "in", "category": "Investment", "source": "Bank A", "transaction_date": time.Now().Format(time.RFC3339)},
			{"amount": 200, "type": "out", "category": "Food", "source": "Bank A", "transaction_date": time.Now().Format(time.RFC3339)},
			{"amount": 300, "type": "out", "category": "Transportation", "source": "Bank B", "transaction_date": time.Now().Format(time.RFC3339)},
		},
	}

	resp := makeRequest(t, server, "POST", "/webhook/batch", reqBody, apiKey)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	t.Run("Get summary", func(t *testing.T) {
		resp = makeRequest(t, server, "GET", "/analytics/summary", nil, apiKey)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summary domain.SummaryResponse
		err := json.NewDecoder(resp.Body).Decode(&summary)
		require.NoError(t, err)
		assert.Equal(t, 1500.00, summary.TotalIncome)
		assert.Equal(t, 500.00, summary.TotalExpense)
		assert.Equal(t, 1000.00, summary.CurrentBalance)
		assert.Equal(t, int64(4), summary.TransactionCount)
	})

	t.Run("Get breakdown by source", func(t *testing.T) {
		resp = makeRequest(t, server, "GET", "/analytics/breakdown/source", nil, apiKey)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var breakdown []domain.BreakdownResponse
		err := json.NewDecoder(resp.Body).Decode(&breakdown)
		require.NoError(t, err)
		assert.Len(t, breakdown, 2)
	})

	t.Run("Get breakdown by category", func(t *testing.T) {
		resp = makeRequest(t, server, "GET", "/analytics/breakdown/category", nil, apiKey)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var breakdown []domain.BreakdownResponse
		err := json.NewDecoder(resp.Body).Decode(&breakdown)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(breakdown), 2)
	})
}

// Test API error responses

func TestAPI_Integration_ErrorResponses(t *testing.T) {
	server, apiKey := setupTestServer(t)
	defer server.Close()

	t.Run("Invalid JSON", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), "POST", server.URL+"/webhook/transaction", bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Missing API key", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/analytics/summary", nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Wrong API key", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/analytics/summary", nil, "wrong-key")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Missing required fields", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount": 100,
			// Missing type and source
		}

		resp := makeRequest(t, server, "POST", "/webhook/transaction", reqBody, apiKey)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Invalid transaction type", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount":           100,
			"type":             "invalid",
			"source":           "Bank",
			"transaction_date": time.Now().Format(time.RFC3339),
		}

		resp := makeRequest(t, server, "POST", "/webhook/transaction", reqBody, apiKey)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Future date rejected", func(t *testing.T) {
		futureDate := time.Now().Add(24 * time.Hour)
		reqBody := map[string]interface{}{
			"amount":           100,
			"type":             "out",
			"source":           "Bank",
			"transaction_date": futureDate.Format(time.RFC3339),
		}

		resp := makeRequest(t, server, "POST", "/webhook/transaction", reqBody, apiKey)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Invalid category", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount":           100,
			"type":             "out",
			"category":         "InvalidCategory",
			"source":           "Bank",
			"transaction_date": time.Now().Format(time.RFC3339),
		}

		resp := makeRequest(t, server, "POST", "/webhook/transaction", reqBody, apiKey)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Not found transaction", func(t *testing.T) {
		resp := makeRequest(t, server, "GET", "/transactions/99999", nil, apiKey)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestAPI_Integration_InvalidTrendsPeriod(t *testing.T) {
	server, apiKey := setupTestServer(t)
	defer server.Close()

	resp := makeRequest(t, server, "GET", "/analytics/trends?period=invalid", nil, apiKey)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp["error"], "period must be one of")
}

func TestAPI_Integration_CORSHeaders(t *testing.T) {
	server, apiKey := setupTestServer(t)
	defer server.Close()

	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL+"/analytics/summary", nil)
	require.NoError(t, err)
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "*", resp.Header.Get("Access-Control-Allow-Origin"))
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Methods"))
}
