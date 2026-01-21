package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/handler"
	"github.com/dev/personal-finance-tracker/backend/internal/middleware"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/internal/service"
	"github.com/dev/personal-finance-tracker/backend/test/util"
)

// skipIfNoDocker skips E2E tests if Docker is not available for testcontainers
func skipIfNoDocker(t *testing.T) {
	t.Helper()

	// Check for explicit environment variable to skip E2E tests
	if os.Getenv("SKIP_E2E") == "1" {
		t.Skip("Skipping E2E test: SKIP_E2E=1")
	}
}

// E2ETestServer wraps the test server with all dependencies
type E2ETestServer struct {
	Container *util.PostgresTestContainer
	Server    *httptest.Server
	DB        *gorm.DB
	APIKey    string
}

// SetupE2EServer creates a complete test server for E2E testing
func SetupE2EServer(t *testing.T) *E2ETestServer {
	t.Helper()

	pc := util.SetupPostgresContainerForE2E(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})

	// Setup full application stack
	repo := repository.NewTransactionRepository(db)
	svc := service.NewTransactionService(repo)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{AllowedOrigins: []string{"*"}}))
	router.Use(middleware.ErrorHandler())

	// Use test API key
	testAPIKey := "e2e-test-api-key"
	router.Use(middleware.APIKeyAuth(testAPIKey))

	webhookHandler := handler.NewWebhookHandler(svc)
	analyticsHandler := handler.NewAnalyticsHandler(svc)

	// Register all routes
	router.POST("/webhook/transaction", webhookHandler.CreateTransaction)
	router.POST("/webhook/batch", webhookHandler.CreateBatchTransaction)
	router.GET("/analytics/summary", analyticsHandler.GetSummary)
	router.GET("/analytics/trends", analyticsHandler.GetTrends)
	router.GET("/analytics/breakdown/source", analyticsHandler.GetBreakdownBySource)
	router.GET("/analytics/breakdown/category", analyticsHandler.GetBreakdownByCategory)
	router.GET("/transactions", analyticsHandler.ListTransactions)
	router.GET("/transactions/:id", analyticsHandler.GetTransactionByID)

	server := httptest.NewServer(router)

	return &E2ETestServer{
		Server:    server,
		DB:        db,
		APIKey:    testAPIKey,
		Container: pc,
	}
}

// Close cleans up the E2E test server
func (s *E2ETestServer) Close(t *testing.T) {
	t.Helper()
	s.Container.Terminate(t)
}

// MakeRequest is a helper to make HTTP requests to the test server
func (s *E2ETestServer) MakeRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	t.Helper()

	var bodyReader *bytes.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}

	url := s.Server.URL + path
	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// Test E2E: Complete user flow - create transaction, verify in summary

func TestE2E_CompleteUserFlow(t *testing.T) {
	skipIfNoDocker(t)
	server := SetupE2EServer(t)
	defer server.Close(t)

	t.Run("Step 1: Create income transaction", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount":           5000,
			"type":             "in",
			"category":         "Salary",
			"description":      "Monthly salary",
			"source":           "Main Bank",
			"transaction_date": time.Now().Format(time.RFC3339),
		}

		resp := server.MakeRequest(t, "POST", "/webhook/transaction", reqBody)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var tx domain.Transaction
		err := json.NewDecoder(resp.Body).Decode(&tx)
		require.NoError(t, err)
		assert.Equal(t, 5000.0, tx.Amount)
		assert.Equal(t, domain.TransactionTypeIn, tx.Type)
	})

	t.Run("Step 2: Create expense transactions", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"transactions": []map[string]interface{}{
				{
					"amount":           150,
					"type":             "out",
					"category":         "Food",
					"description":      "Groceries",
					"source":           "Main Bank",
					"transaction_date": time.Now().Format(time.RFC3339),
				},
				{
					"amount":           50,
					"type":             "out",
					"category":         "Transportation",
					"description":      "Gas",
					"source":           "Main Bank",
					"transaction_date": time.Now().Format(time.RFC3339),
				},
				{
					"amount":           200,
					"type":             "out",
					"category":         "Utilities",
					"description":      "Electric bill",
					"source":           "Main Bank",
					"transaction_date": time.Now().Format(time.RFC3339),
				},
			},
		}

		resp := server.MakeRequest(t, "POST", "/webhook/batch", reqBody)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, float64(3), result["created"])
	})

	t.Run("Step 3: Verify summary reflects all transactions", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/analytics/summary", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var summary domain.SummaryResponse
		err := json.NewDecoder(resp.Body).Decode(&summary)
		require.NoError(t, err)

		assert.Equal(t, 5000.0, summary.TotalIncome)
		assert.Equal(t, 400.0, summary.TotalExpense)
		assert.Equal(t, 4600.0, summary.CurrentBalance)
		assert.Equal(t, int64(4), summary.TransactionCount)
	})

	t.Run("Step 4: Verify breakdown by category", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/analytics/breakdown/category", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var breakdown []domain.BreakdownResponse
		err := json.NewDecoder(resp.Body).Decode(&breakdown)
		require.NoError(t, err)

		// Should have Food, Transportation, and Utilities categories
		categoryMap := make(map[string]float64)
		for _, item := range breakdown {
			categoryMap[item.Label] = item.Amount
		}

		assert.Equal(t, 150.0, categoryMap["Food"])
		assert.Equal(t, 50.0, categoryMap["Transportation"])
		assert.Equal(t, 200.0, categoryMap["Utilities"])
	})

	t.Run("Step 5: List transactions with pagination", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/transactions?page=1&page_size=10", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		data := result["data"].([]interface{})
		pagination := result["pagination"].(map[string]interface{})

		assert.Equal(t, float64(4), pagination["total"])
		assert.Len(t, data, 4)
	})
}

// Test E2E: Authentication workflow

func TestE2E_AuthenticationFlow(t *testing.T) {
	skipIfNoDocker(t)
	server := SetupE2EServer(t)
	defer server.Close(t)

	t.Run("Request without API key is rejected", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), "GET", server.Server.URL+"/analytics/summary", nil)
		require.NoError(t, err)
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Request with wrong API key is rejected", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), "GET", server.Server.URL+"/analytics/summary", nil)
		require.NoError(t, err)
		req.Header.Set("X-API-Key", "wrong-api-key")
		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Request with correct API key is accepted", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/analytics/summary", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// Test E2E: Error handling workflow

func TestE2E_ErrorHandlingFlow(t *testing.T) {
	skipIfNoDocker(t)
	server := SetupE2EServer(t)
	defer server.Close(t)

	t.Run("Invalid data returns error", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount": -100,
			"type":   "out",
			"source": "Bank",
		}

		resp := server.MakeRequest(t, "POST", "/webhook/transaction", reqBody)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Future date rejected", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount":           100,
			"type":             "out",
			"source":           "Bank",
			"transaction_date": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		}

		resp := server.MakeRequest(t, "POST", "/webhook/transaction", reqBody)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp["error"], "future")
	})

	t.Run("Invalid category rejected", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount":           100,
			"type":             "out",
			"category":         "InvalidCategory",
			"source":           "Bank",
			"transaction_date": time.Now().Format(time.RFC3339),
		}

		resp := server.MakeRequest(t, "POST", "/webhook/transaction", reqBody)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp["error"], "invalid category")
	})
}

// Test E2E: Pagination with large dataset

func TestE2E_PaginationWorkflow(t *testing.T) {
	skipIfNoDocker(t)
	server := SetupE2EServer(t)
	defer server.Close(t)

	// Create 50 transactions
	transactions := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		transactions[i] = map[string]interface{}{
			"amount":           float64(100 + i),
			"type":             "out",
			"category":         "Food",
			"source":           "Bank",
			"transaction_date": time.Now().Format(time.RFC3339),
		}
	}

	reqBody := map[string]interface{}{"transactions": transactions}
	resp := server.MakeRequest(t, "POST", "/webhook/batch", reqBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	t.Run("First page", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/transactions?page=1&page_size=20", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		pagination := result["pagination"].(map[string]interface{})

		assert.Len(t, data, 20)
		assert.Equal(t, float64(50), pagination["total"])
		assert.Equal(t, float64(3), pagination["total_pages"])
	})

	t.Run("Second page", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/transactions?page=2&page_size=20", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		assert.Len(t, data, 20)
	})

	t.Run("Third page (partial)", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/transactions?page=3&page_size=20", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		data := result["data"].([]interface{})
		assert.Len(t, data, 10)
	})
}

// Test E2E: Analytics after data insertion

func TestE2E_AnalyticsFlow(t *testing.T) {
	skipIfNoDocker(t)
	server := SetupE2EServer(t)
	defer server.Close(t)

	// Create diverse transactions
	reqBody := map[string]interface{}{
		"transactions": []map[string]interface{}{
			{"amount": 1000, "type": "in", "category": "Salary", "source": "Bank A", "transaction_date": time.Now().Format(time.RFC3339)},
			{"amount": 500, "type": "in", "category": "Investment", "source": "Bank B", "transaction_date": time.Now().Format(time.RFC3339)},
			{"amount": 200, "type": "out", "category": "Food", "source": "Bank A", "transaction_date": time.Now().Format(time.RFC3339)},
			{"amount": 150, "type": "out", "category": "Food", "source": "Bank B", "transaction_date": time.Now().Format(time.RFC3339)},
			{"amount": 100, "type": "out", "category": "Transportation", "source": "Bank A", "transaction_date": time.Now().Format(time.RFC3339)},
		},
	}

	resp := server.MakeRequest(t, "POST", "/webhook/batch", reqBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	t.Run("Verify summary", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/analytics/summary", nil)
		var summary domain.SummaryResponse
		json.NewDecoder(resp.Body).Decode(&summary)

		assert.Equal(t, 1500.0, summary.TotalIncome)
		assert.Equal(t, 450.0, summary.TotalExpense)
		assert.Equal(t, 1050.0, summary.CurrentBalance)
	})

	t.Run("Verify trends", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/analytics/trends?period=daily", nil)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var trends domain.TrendsResponse
		json.NewDecoder(resp.Body).Decode(&trends)

		assert.Equal(t, "daily", trends.Period)
		assert.NotEmpty(t, trends.Data)
	})

	t.Run("Verify source breakdown", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/analytics/breakdown/source", nil)
		var breakdown []domain.BreakdownResponse
		json.NewDecoder(resp.Body).Decode(&breakdown)

		assert.Len(t, breakdown, 2)

		sourceMap := make(map[string]float64)
		for _, item := range breakdown {
			sourceMap[item.Label] = item.Amount
		}
		assert.Equal(t, 300.0, sourceMap["Bank A"])
		assert.Equal(t, 150.0, sourceMap["Bank B"])
	})

	t.Run("Verify category breakdown", func(t *testing.T) {
		resp := server.MakeRequest(t, "GET", "/analytics/breakdown/category", nil)
		var breakdown []domain.BreakdownResponse
		json.NewDecoder(resp.Body).Decode(&breakdown)

		assert.GreaterOrEqual(t, len(breakdown), 2)
	})
}

// Test E2E: Create and retrieve individual transaction

func TestE2E_CreateAndRetrieveTransaction(t *testing.T) {
	skipIfNoDocker(t)
	server := SetupE2EServer(t)
	defer server.Close(t)

	t.Run("Create transaction", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"amount":           250.75,
			"type":             "out",
			"category":         "Shopping",
			"description":      "New shoes",
			"source":           "Credit Card",
			"source_account":   "1234",
			"transaction_date": time.Now().Format(time.RFC3339),
		}

		resp := server.MakeRequest(t, "POST", "/webhook/transaction", reqBody)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdTx domain.Transaction
		json.NewDecoder(resp.Body).Decode(&createdTx)
		assert.Greater(t, createdTx.ID, int64(0))

		t.Run("Retrieve by ID", func(t *testing.T) {
			url := fmt.Sprintf("/transactions/%d", createdTx.ID)
			resp := server.MakeRequest(t, "GET", url, nil)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var retrievedTx domain.Transaction
			json.NewDecoder(resp.Body).Decode(&retrievedTx)

			assert.Equal(t, createdTx.ID, retrievedTx.ID)
			assert.Equal(t, 250.75, retrievedTx.Amount)
			assert.Equal(t, "Shopping", retrievedTx.Category)
			assert.Equal(t, "New shoes", retrievedTx.Description)
			assert.Equal(t, "Credit Card", retrievedTx.Source)
		})
	})
}
