package domain

import (
	"time"
)

// TransactionType represents the direction of money flow
type TransactionType string

const (
	TransactionTypeIn  TransactionType = "in"
	TransactionTypeOut TransactionType = "out"
)

// Valid categories for transactions
var ValidCategories = map[string]bool{
	"Food":           true,
	"Transportation": true,
	"Housing":        true,
	"Utilities":      true,
	"Entertainment":  true,
	"Healthcare":     true,
	"Shopping":       true,
	"Education":      true,
	"Salary":         true,
	"Investment":     true,
	"Transfer":       true,
	"Other":          true,
	"":               true, // Empty category is allowed
}

const (
	// MaxAmount is the maximum allowed amount for a transaction
	MaxAmount = 999999999.99
	// MaxDescriptionLength is the maximum length for description
	MaxDescriptionLength = 1000
	// MaxSourceLength is the maximum length for source
	MaxSourceLength = 100
	// MaxAccountLength is the maximum length for account identifiers
	MaxAccountLength = 100
	// MaxRecipientLength is the maximum length for recipient
	MaxRecipientLength = 100
	// MaxCategoryLength is the maximum length for category
	MaxCategoryLength = 50
)

// Transaction represents a financial transaction from a bank or e-wallet
type Transaction struct {
	Type            TransactionType `json:"type" gorm:"type:varchar(20);not null;index"`
	Category        string          `json:"category" gorm:"type:varchar(50)"`
	Description     string          `json:"description" gorm:"type:text"`
	Source          string          `json:"source" gorm:"type:varchar(100);not null;index"` // Bank/wallet name
	SourceAccount   string          `json:"source_account" gorm:"type:varchar(100)"`        // Account identifier
	Recipient       string          `json:"recipient" gorm:"type:varchar(100)"`             // For transfers
	TransactionDate time.Time       `json:"transaction_date" gorm:"not null;index"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	ID              int64           `json:"id" gorm:"primaryKey"`
	Amount          float64         `json:"amount" gorm:"type:decimal(15,2);not null"`
}

// TableName specifies the table name for GORM
func (Transaction) TableName() string {
	return "transactions"
}

// CreateTransactionRequest is the request body for creating a transaction
type CreateTransactionRequest struct {
	Type            TransactionType `json:"type" binding:"required,oneof=in out"`
	Category        string          `json:"category" binding:"omitempty,max=50"`
	Description     string          `json:"description" binding:"omitempty,max=1000"`
	Source          string          `json:"source" binding:"required,min=1,max=100"`
	SourceAccount   string          `json:"source_account" binding:"omitempty,max=100"`
	Recipient       string          `json:"recipient" binding:"omitempty,max=100"`
	TransactionDate string          `json:"transaction_date" binding:"required"`
	Amount          float64         `json:"amount" binding:"required,gt=0"`
}

// Validate performs additional validation beyond struct tags
func (r *CreateTransactionRequest) Validate() error {
	// Check if amount exceeds maximum
	if r.Amount > MaxAmount {
		return &ValidationError{
			Field:   "amount",
			Message: "amount exceeds maximum allowed value",
		}
	}

	// Validate category if provided
	if r.Category != "" {
		if !ValidCategories[r.Category] {
			return &ValidationError{
				Field:   "category",
				Message: "invalid category. Valid categories are: Food, Transportation, Housing, Utilities, Entertainment, Healthcare, Shopping, Education, Salary, Investment, Transfer, Other",
			}
		}
	}

	// Parse and validate date
	txDate, err := time.Parse(time.RFC3339, r.TransactionDate)
	if err != nil {
		return &ValidationError{
			Field:   "transaction_date",
			Message: "invalid date format. Must be RFC3339 format (e.g., 2026-01-15T12:00:00Z)",
		}
	}

	// Check for future dates (transactions can't be in the future)
	if txDate.After(time.Now().Add(5 * time.Minute)) {
		return &ValidationError{
			Field:   "transaction_date",
			Message: "transaction date cannot be in the future",
		}
	}

	// Check for very old dates (reasonable limit: 10 years ago)
	if txDate.Before(time.Now().AddDate(-10, 0, 0)) {
		return &ValidationError{
			Field:   "transaction_date",
			Message: "transaction date is too far in the past (more than 10 years)",
		}
	}

	return nil
}

// ToTransaction converts CreateTransactionRequest to Transaction domain
func (r *CreateTransactionRequest) ToTransaction() (*Transaction, error) {
	txDate, err := time.Parse(time.RFC3339, r.TransactionDate)
	if err != nil {
		return nil, err
	}

	// Sanitize and limit field lengths
	category := r.Category
	if len(category) > MaxCategoryLength {
		category = category[:MaxCategoryLength]
	}

	description := r.Description
	if len(description) > MaxDescriptionLength {
		description = description[:MaxDescriptionLength]
	}

	source := r.Source
	if len(source) > MaxSourceLength {
		source = source[:MaxSourceLength]
	}

	sourceAccount := r.SourceAccount
	if len(sourceAccount) > MaxAccountLength {
		sourceAccount = sourceAccount[:MaxAccountLength]
	}

	recipient := r.Recipient
	if len(recipient) > MaxRecipientLength {
		recipient = recipient[:MaxRecipientLength]
	}

	return &Transaction{
		Amount:          r.Amount,
		Type:            r.Type,
		Category:        category,
		Description:     description,
		Source:          source,
		SourceAccount:   sourceAccount,
		Recipient:       recipient,
		TransactionDate: txDate,
	}, nil
}

// BatchTransactionRequest is the request body for batch transaction creation
type BatchTransactionRequest struct {
	Transactions []CreateTransactionRequest `json:"transactions" binding:"required,min=1,max=100"`
}

// Validate performs validation on the entire batch
func (r *BatchTransactionRequest) Validate() error {
	if len(r.Transactions) == 0 {
		return &ValidationError{
			Field:   "transactions",
			Message: "at least one transaction is required",
		}
	}

	if len(r.Transactions) > 100 {
		return &ValidationError{
			Field:   "transactions",
			Message: "maximum 100 transactions allowed per batch",
		}
	}

	// Validate each transaction
	for i, tx := range r.Transactions {
		if err := tx.Validate(); err != nil {
			return &ValidationError{
				Field:   "transactions",
				Message: err.Error(),
				Index:   i,
			}
		}
	}

	return nil
}

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string
	Message string
	Index   int // For batch operations
}

func (e *ValidationError) Error() string {
	if e.Index >= 0 {
		return e.Message
	}
	return e.Field + ": " + e.Message
}

// SummaryResponse is the response for analytics summary
type SummaryResponse struct {
	TotalIncome      float64 `json:"total_income"`
	TotalExpense     float64 `json:"total_expense"`
	CurrentBalance   float64 `json:"current_balance"`
	TransactionCount int64   `json:"transaction_count"`
}

// TrendsResponse is the response for analytics trends
type TrendsResponse struct {
	Period string           `json:"period"` // daily, weekly, monthly
	Data   []TrendDataPoint `json:"data"`
}

// TrendDataPoint represents a single data point in trends
type TrendDataPoint struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
}

// BreakdownResponse is the response for source or category breakdown
type BreakdownResponse struct {
	Label      string  `json:"label"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
	Count      int64   `json:"count"`
}

// ListTransactionsQueryParams represents query parameters for listing transactions
type ListTransactionsQueryParams struct {
	Type      TransactionType `form:"type"`
	Source    string          `form:"source"`
	Category  string          `form:"category"`
	StartDate string          `form:"start_date"`
	EndDate   string          `form:"end_date"`
	Page      int             `form:"page,default=1"`
	PageSize  int             `form:"page_size,default=20"`
}
