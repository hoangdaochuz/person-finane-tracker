package util

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
)

// CreateTestTransaction creates a valid test transaction
func CreateTestTransaction() *domain.Transaction {
	now := time.Now()
	return &domain.Transaction{
		ID:              1,
		Amount:          100.50,
		Type:            domain.TransactionTypeOut,
		Category:        "Food",
		Description:     "Test transaction",
		Source:          "Test Bank",
		SourceAccount:   "1234",
		Recipient:       "",
		TransactionDate: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// CreateTestTransactionRequest creates a valid test transaction request
func CreateTestTransactionRequest() *domain.CreateTransactionRequest {
	return &domain.CreateTransactionRequest{
		Amount:          100.50,
		Type:            domain.TransactionTypeOut,
		Category:        "Food",
		Description:     "Test transaction",
		Source:          "Test Bank",
		SourceAccount:   "1234",
		TransactionDate: time.Now().Format(time.RFC3339),
	}
}

// AssertJSONBody asserts that the response body contains valid JSON matching expected
func AssertJSONBody(t *testing.T, body string, expected interface{}) {
	t.Helper()
	var actual interface{}
	if err := json.Unmarshal([]byte(body), &actual); err != nil {
		t.Fatalf("failed to parse response body as JSON: %v", err)
	}

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("failed to marshal expected: %v", err)
	}
	var expectedParsed interface{}
	if err := json.Unmarshal(expectedJSON, &expectedParsed); err != nil {
		t.Fatalf("failed to parse expected as JSON: %v", err)
	}

	if !jsonEqual(actual, expectedParsed) {
		t.Errorf("response body does not match expected.\nGot: %s\nWant: %s", body, string(expectedJSON))
	}
}

// jsonEqual compares two JSON values for equality
func jsonEqual(a, b interface{}) bool {
	aBytes, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bBytes, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return bytes.Equal(aBytes, bBytes)
}

// AssertHTTPError asserts HTTP error response
func AssertHTTPError(t *testing.T, w *httptest.ResponseRecorder, code int) {
	t.Helper()
	if w.Code != code {
		t.Errorf("expected status code %d, got %d", code, w.Code)
	}
}

// SetupTestGin sets up Gin for testing
func SetupTestGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ParseResponse parses JSON response into target
func ParseResponse(t *testing.T, body string, target interface{}) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), target); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !contains(s, substr) {
		t.Errorf("expected string to contain %q, got %q", substr, s)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsInner(s, substr))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
