package domain

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Test ToTransaction

func TestToTransaction_ValidInput(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          100.50,
		Type:            TransactionTypeOut,
		Category:        "Food",
		Description:     "Lunch",
		Source:          "Bank ABC",
		SourceAccount:   "1234",
		Recipient:       "John Doe",
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	tx, err := req.ToTransaction()

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, 100.50, tx.Amount)
	assert.Equal(t, TransactionTypeOut, tx.Type)
	assert.Equal(t, "Food", tx.Category)
	assert.Equal(t, "Lunch", tx.Description)
	assert.Equal(t, "Bank ABC", tx.Source)
	assert.Equal(t, "1234", tx.SourceAccount)
	assert.Equal(t, "John Doe", tx.Recipient)
}

func TestToTransaction_InvalidDate(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: "invalid-date",
	}

	tx, err := req.ToTransaction()

	assert.Error(t, err)
	assert.Nil(t, tx)
}

func TestToTransaction_TruncatesLongFields(t *testing.T) {
	longString := strings.Repeat("a", 200)

	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Category:        longString, // Will be truncated to MaxCategoryLength (50)
		Description:     longString, // Will be truncated to MaxDescriptionLength (1000)
		Source:          longString, // Will be truncated to MaxSourceLength (100)
		SourceAccount:   longString, // Will be truncated to MaxAccountLength (100)
		Recipient:       longString, // Will be truncated to MaxRecipientLength (100)
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	tx, err := req.ToTransaction()

	assert.NoError(t, err)
	// The truncation limits strings to max length
	assert.LessOrEqual(t, len(tx.Category), MaxCategoryLength)
	assert.LessOrEqual(t, len(tx.Description), MaxDescriptionLength)
	assert.LessOrEqual(t, len(tx.Source), MaxSourceLength)
	assert.LessOrEqual(t, len(tx.SourceAccount), MaxAccountLength)
	assert.LessOrEqual(t, len(tx.Recipient), MaxRecipientLength)
}

func TestToTransaction_EmptyFields(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          100.50,
		Type:            TransactionTypeIn,
		Source:          "Bank ABC",
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	tx, err := req.ToTransaction()

	assert.NoError(t, err)
	assert.Equal(t, "", tx.Category)
	assert.Equal(t, "", tx.Description)
	assert.Equal(t, "", tx.SourceAccount)
	assert.Equal(t, "", tx.Recipient)
}

// Test Validate

func TestValidate_ValidTransaction(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Category:        "Food",
		Source:          "Bank ABC",
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	err := req.Validate()

	assert.NoError(t, err)
}

func TestValidate_ValidEmptyCategory(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Category:        "",
		Source:          "Bank ABC",
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	err := req.Validate()

	assert.NoError(t, err)
}

func TestValidate_InvalidCategory(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Category:        "InvalidCategory",
		Source:          "Bank ABC",
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok, "error should be ValidationError")
	assert.Equal(t, "category", validationErr.Field)
}

func TestValidate_AllValidCategories(t *testing.T) {
	validCategories := []string{
		"Food", "Transportation", "Housing", "Utilities",
		"Entertainment", "Healthcare", "Shopping", "Education",
		"Salary", "Investment", "Transfer", "Other", "",
	}

	for _, category := range validCategories {
		t.Run(category, func(t *testing.T) {
			req := &CreateTransactionRequest{
				Amount:          100,
				Type:            TransactionTypeOut,
				Category:        category,
				Source:          "Bank ABC",
				TransactionDate: "2026-01-15T12:00:00Z",
			}

			err := req.Validate()
			assert.NoError(t, err, "category %s should be valid", category)
		})
	}
}

func TestValidate_ExcessiveAmount(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          MaxAmount + 1,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "amount", validationErr.Field)
}

func TestValidate_MaximumAllowedAmount(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          MaxAmount,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: "2026-01-15T12:00:00Z",
	}

	err := req.Validate()

	assert.NoError(t, err, "MaxAmount should be valid")
}

func TestValidate_InvalidDateFormat(t *testing.T) {
	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: "not-a-date",
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "transaction_date", validationErr.Field)
}

func TestValidate_FutureDate(t *testing.T) {
	futureDate := time.Now().Add(10 * time.Minute)

	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: futureDate.Format(time.RFC3339),
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "transaction_date", validationErr.Field)
	assert.Contains(t, validationErr.Message, "future")
}

func TestValidate_CurrentDateAllowed(t *testing.T) {
	now := time.Now()

	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: now.Format(time.RFC3339),
	}

	err := req.Validate()

	assert.NoError(t, err, "current date should be valid")
}

func TestValidate_TooOldDate(t *testing.T) {
	oldDate := time.Now().AddDate(-11, 0, 0)

	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: oldDate.Format(time.RFC3339),
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "transaction_date", validationErr.Field)
	assert.Contains(t, validationErr.Message, "too far in the past")
}

func TestValidate_TenYearsAgoAllowed(t *testing.T) {
	// Use a date just slightly less than 10 years ago to be within the valid range
	// The validation uses Before() so 10 years exactly would fail
	nineYearsAnd11MonthsAgo := time.Now().AddDate(-9, -11, 0)

	req := &CreateTransactionRequest{
		Amount:          100,
		Type:            TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: nineYearsAnd11MonthsAgo.Format(time.RFC3339),
	}

	err := req.Validate()

	assert.NoError(t, err, "date just under 10 years ago should be valid")
}

// Test ValidationError.Error()

func TestValidationError_WithIndex(t *testing.T) {
	err := &ValidationError{
		Field:   "category",
		Message: "invalid category",
		Index:   5,
	}

	msg := err.Error()

	assert.Equal(t, "invalid category", msg)
}

func TestValidationError_WithoutIndex(t *testing.T) {
	err := &ValidationError{
		Field:   "category",
		Message: "invalid category",
		Index:   -1,
	}

	msg := err.Error()

	assert.Equal(t, "category: invalid category", msg)
}

// Test BatchTransactionRequest.Validate()

func TestBatchValidate_ValidBatch(t *testing.T) {
	req := &BatchTransactionRequest{
		Transactions: []CreateTransactionRequest{
			{
				Amount:          100,
				Type:            TransactionTypeOut,
				Category:        "Food",
				Source:          "Bank ABC",
				TransactionDate: "2026-01-15T12:00:00Z",
			},
			{
				Amount:          200,
				Type:            TransactionTypeIn,
				Category:        "Salary",
				Source:          "Bank XYZ",
				TransactionDate: "2026-01-15T12:00:00Z",
			},
		},
	}

	err := req.Validate()

	assert.NoError(t, err)
}

func TestBatchValidate_EmptyBatch(t *testing.T) {
	req := &BatchTransactionRequest{
		Transactions: []CreateTransactionRequest{},
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, "transactions", validationErr.Field)
}

func TestBatchValidate_ExceedsMaximum(t *testing.T) {
	transactions := make([]CreateTransactionRequest, 101)
	for i := 0; i < 101; i++ {
		transactions[i] = CreateTransactionRequest{
			Amount:          100,
			Type:            TransactionTypeOut,
			Source:          "Bank ABC",
			TransactionDate: "2026-01-15T12:00:00Z",
		}
	}

	req := &BatchTransactionRequest{
		Transactions: transactions,
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Contains(t, validationErr.Message, "maximum 100")
}

func TestBatchValidate_ExactMaximum(t *testing.T) {
	transactions := make([]CreateTransactionRequest, 100)
	for i := 0; i < 100; i++ {
		transactions[i] = CreateTransactionRequest{
			Amount:          100,
			Type:            TransactionTypeOut,
			Source:          "Bank ABC",
			TransactionDate: "2026-01-15T12:00:00Z",
		}
	}

	req := &BatchTransactionRequest{
		Transactions: transactions,
	}

	err := req.Validate()

	assert.NoError(t, err, "100 transactions should be valid")
}

func TestBatchValidate_InvalidTransactionInBatch(t *testing.T) {
	req := &BatchTransactionRequest{
		Transactions: []CreateTransactionRequest{
			{
				Amount:          100,
				Type:            TransactionTypeOut,
				Category:        "Food",
				Source:          "Bank ABC",
				TransactionDate: "2026-01-15T12:00:00Z",
			},
			{
				Amount:          200,
				Type:            TransactionTypeIn,
				Category:        "InvalidCategory",
				Source:          "Bank XYZ",
				TransactionDate: "2026-01-15T12:00:00Z",
			},
		},
	}

	err := req.Validate()

	assert.Error(t, err)
	validationErr, ok := err.(*ValidationError)
	assert.True(t, ok)
	assert.Equal(t, 1, validationErr.Index, "should report index of second transaction")
}

// Test TransactionType constants

func TestTransactionType_Constants(t *testing.T) {
	assert.Equal(t, TransactionType("in"), TransactionTypeIn)
	assert.Equal(t, TransactionType("out"), TransactionTypeOut)
}

// Test TableName

func TestTransaction_TableName(t *testing.T) {
	tx := Transaction{}
	assert.Equal(t, "transactions", tx.TableName())
}

// Test ValidCategories map

func TestValidCategories_AllExpectedCategoriesPresent(t *testing.T) {
	expectedCategories := []string{
		"Food", "Transportation", "Housing", "Utilities",
		"Entertainment", "Healthcare", "Shopping", "Education",
		"Salary", "Investment", "Transfer", "Other", "",
	}

	for _, category := range expectedCategories {
		assert.True(t, ValidCategories[category], "category %s should be valid", category)
	}
}

func TestValidCategories_InvalidCategoryNotPresent(t *testing.T) {
	assert.False(t, ValidCategories["InvalidCategory"])
	assert.False(t, ValidCategories["XYZ"])
}
