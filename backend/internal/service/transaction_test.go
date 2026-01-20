package service

import (
	"errors"
	"testing"
	"time"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
)

// mockRepository is a mock implementation of TransactionRepository for testing
type mockRepository struct {
	createFunc           func(tx *domain.Transaction) error
	createInBatchFunc    func(transactions []domain.Transaction) error
	findByIDFunc         func(id int64) (*domain.Transaction, error)
	listFunc             func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error)
	getSummaryFunc       func() (*domain.SummaryResponse, error)
	getTrendsFunc        func(period string) ([]domain.TrendDataPoint, error)
	getBreakdownSource   func() ([]domain.BreakdownResponse, error)
	getBreakdownCategory func() ([]domain.BreakdownResponse, error)
}

func (m *mockRepository) Create(tx *domain.Transaction) error {
	if m.createFunc != nil {
		return m.createFunc(tx)
	}
	return nil
}

func (m *mockRepository) CreateInBatch(transactions []domain.Transaction) error {
	if m.createInBatchFunc != nil {
		return m.createInBatchFunc(transactions)
	}
	return nil
}

func (m *mockRepository) FindByID(id int64) (*domain.Transaction, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return &domain.Transaction{ID: id}, nil
}

func (m *mockRepository) List(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(params)
	}
	return []domain.Transaction{}, 0, nil
}

func (m *mockRepository) GetSummary() (*domain.SummaryResponse, error) {
	if m.getSummaryFunc != nil {
		return m.getSummaryFunc()
	}
	return &domain.SummaryResponse{}, nil
}

func (m *mockRepository) GetTrends(period string) ([]domain.TrendDataPoint, error) {
	if m.getTrendsFunc != nil {
		return m.getTrendsFunc(period)
	}
	return []domain.TrendDataPoint{}, nil
}

func (m *mockRepository) GetBreakdownBySource() ([]domain.BreakdownResponse, error) {
	if m.getBreakdownSource != nil {
		return m.getBreakdownSource()
	}
	return []domain.BreakdownResponse{}, nil
}

func (m *mockRepository) GetBreakdownByCategory() ([]domain.BreakdownResponse, error) {
	if m.getBreakdownCategory != nil {
		return m.getBreakdownCategory()
	}
	return []domain.BreakdownResponse{}, nil
}

// Test CreateTransaction

func TestCreateTransaction_Success(t *testing.T) {
	mockRepo := &mockRepository{
		createFunc: func(tx *domain.Transaction) error {
			tx.ID = 1
			return nil
		},
	}
	service := NewTransactionService(mockRepo)

	req := &domain.CreateTransactionRequest{
		Amount:          100.50,
		Type:            domain.TransactionTypeOut,
		Category:        "Food",
		Description:     "Lunch",
		Source:          "Bank ABC",
		SourceAccount:   "1234",
		TransactionDate: time.Now().Format(time.RFC3339),
	}

	tx, err := service.CreateTransaction(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}
	if tx.Amount != 100.50 {
		t.Errorf("expected amount 100.50, got %f", tx.Amount)
	}
	if tx.Type != domain.TransactionTypeOut {
		t.Errorf("expected type 'out', got %s", tx.Type)
	}
}

func TestCreateTransaction_ExcessiveAmount(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	req := &domain.CreateTransactionRequest{
		Amount:          999999999999, // Exceeds MaxAmount
		Type:            domain.TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: time.Now().Format(time.RFC3339),
	}

	tx, err := service.CreateTransaction(req)

	if err == nil {
		t.Error("expected error for excessive amount, got nil")
	}
	if tx != nil {
		t.Error("expected nil transaction, got non-nil")
	}
}

func TestCreateTransaction_InvalidCategory(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	req := &domain.CreateTransactionRequest{
		Amount:          100,
		Type:            domain.TransactionTypeOut,
		Category:        "InvalidCategory",
		Source:          "Bank ABC",
		TransactionDate: time.Now().Format(time.RFC3339),
	}

	tx, err := service.CreateTransaction(req)

	if err == nil {
		t.Error("expected error for invalid category, got nil")
	}
	if tx != nil {
		t.Error("expected nil transaction, got non-nil")
	}
}

func TestCreateTransaction_FutureDate(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	futureDate := time.Now().Add(24 * time.Hour)
	req := &domain.CreateTransactionRequest{
		Amount:          100,
		Type:            domain.TransactionTypeOut,
		Category:        "Food",
		Source:          "Bank ABC",
		TransactionDate: futureDate.Format(time.RFC3339),
	}

	tx, err := service.CreateTransaction(req)

	if err == nil {
		t.Error("expected error for future date, got nil")
	}
	if tx != nil {
		t.Error("expected nil transaction, got non-nil")
	}
}

func TestCreateTransaction_InvalidDate(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	req := &domain.CreateTransactionRequest{
		Amount:          100,
		Type:            domain.TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: "invalid-date",
	}

	tx, err := service.CreateTransaction(req)

	if err == nil {
		t.Error("expected error for invalid date, got nil")
	}
	if tx != nil {
		t.Error("expected nil transaction, got non-nil")
	}
}

func TestCreateTransaction_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		createFunc: func(tx *domain.Transaction) error {
			return errors.New("database error")
		},
	}
	service := NewTransactionService(mockRepo)

	req := &domain.CreateTransactionRequest{
		Amount:          100,
		Type:            domain.TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: time.Now().Format(time.RFC3339),
	}

	tx, err := service.CreateTransaction(req)

	if err == nil {
		t.Error("expected error from repository, got nil")
	}
	if tx != nil {
		t.Error("expected nil transaction, got non-nil")
	}
}

// Test CreateBatchTransaction

func TestCreateBatchTransaction_Success(t *testing.T) {
	mockRepo := &mockRepository{
		createInBatchFunc: func(transactions []domain.Transaction) error {
			for i := range transactions {
				transactions[i].ID = int64(i + 1)
			}
			return nil
		},
	}
	service := NewTransactionService(mockRepo)

	req := &domain.BatchTransactionRequest{
		Transactions: []domain.CreateTransactionRequest{
			{
				Amount:          100,
				Type:            domain.TransactionTypeOut,
				Source:          "Bank ABC",
				TransactionDate: time.Now().Format(time.RFC3339),
			},
			{
				Amount:          200,
				Type:            domain.TransactionTypeIn,
				Source:          "Bank XYZ",
				TransactionDate: time.Now().Format(time.RFC3339),
			},
		},
	}

	transactions, err := service.CreateBatchTransaction(req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(transactions) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(transactions))
	}
}

func TestCreateBatchTransaction_Empty(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	req := &domain.BatchTransactionRequest{
		Transactions: []domain.CreateTransactionRequest{},
	}

	transactions, err := service.CreateBatchTransaction(req)

	if err == nil {
		t.Error("expected error for empty batch, got nil")
	}
	if transactions != nil {
		t.Error("expected nil transactions, got non-nil")
	}
}

func TestCreateBatchTransaction_ExcessiveAmount(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	req := &domain.BatchTransactionRequest{
		Transactions: []domain.CreateTransactionRequest{
			{
				Amount:          999999999999, // Exceeds MaxAmount
				Type:            domain.TransactionTypeOut,
				Source:          "Bank ABC",
				TransactionDate: time.Now().Format(time.RFC3339),
			},
		},
	}

	transactions, err := service.CreateBatchTransaction(req)

	if err == nil {
		t.Error("expected error for excessive amount, got nil")
	}
	if transactions != nil {
		t.Error("expected nil transactions, got non-nil")
	}
}

// Test GetTransactionByID

func TestGetTransactionByID_Success(t *testing.T) {
	expectedTx := &domain.Transaction{
		ID:     1,
		Amount: 100,
		Type:   domain.TransactionTypeOut,
	}
	mockRepo := &mockRepository{
		findByIDFunc: func(id int64) (*domain.Transaction, error) {
			return expectedTx, nil
		},
	}
	service := NewTransactionService(mockRepo)

	tx, err := service.GetTransactionByID(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tx.ID != 1 {
		t.Errorf("expected ID 1, got %d", tx.ID)
	}
}

func TestGetTransactionByID_InvalidID(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	tx, err := service.GetTransactionByID(0)

	if err == nil {
		t.Error("expected error for ID 0, got nil")
	}
	if tx != nil {
		t.Error("expected nil transaction, got non-nil")
	}
}

func TestGetTransactionByID_NegativeID(t *testing.T) {
	mockRepo := &mockRepository{}
	service := NewTransactionService(mockRepo)

	tx, err := service.GetTransactionByID(-1)

	if err == nil {
		t.Error("expected error for negative ID, got nil")
	}
	if tx != nil {
		t.Error("expected nil transaction, got non-nil")
	}
}

// Test ListTransactions

func TestListTransactions_DefaultPagination(t *testing.T) {
	mockRepo := &mockRepository{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			return []domain.Transaction{{ID: 1}}, 1, nil
		},
	}
	service := NewTransactionService(mockRepo)

	params := domain.ListTransactionsQueryParams{}
	transactions, total, err := service.ListTransactions(params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if transactions == nil {
		t.Error("expected transactions, got nil")
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestListTransactions_PageSetToZero(t *testing.T) {
	mockRepo := &mockRepository{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			if params.Page != 1 {
				t.Errorf("expected page to be defaulted to 1, got %d", params.Page)
			}
			return []domain.Transaction{}, 0, nil
		},
	}
	service := NewTransactionService(mockRepo)

	params := domain.ListTransactionsQueryParams{Page: 0}
	_, _, err := service.ListTransactions(params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListTransactions_PageSizeTooLarge(t *testing.T) {
	mockRepo := &mockRepository{
		listFunc: func(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
			if params.PageSize != 20 {
				t.Errorf("expected page size to be defaulted to 20, got %d", params.PageSize)
			}
			return []domain.Transaction{}, 0, nil
		},
	}
	service := NewTransactionService(mockRepo)

	params := domain.ListTransactionsQueryParams{PageSize: 200}
	_, _, err := service.ListTransactions(params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

// Test GetSummary

func TestGetSummary_Success(t *testing.T) {
	expectedSummary := &domain.SummaryResponse{
		TotalIncome:      1000,
		TotalExpense:     500,
		CurrentBalance:   500,
		TransactionCount: 10,
	}
	mockRepo := &mockRepository{
		getSummaryFunc: func() (*domain.SummaryResponse, error) {
			return expectedSummary, nil
		},
	}
	service := NewTransactionService(mockRepo)

	summary, err := service.GetSummary()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if summary.TotalIncome != 1000 {
		t.Errorf("expected total income 1000, got %f", summary.TotalIncome)
	}
}

// Test GetTrends

func TestGetTrends_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getTrendsFunc: func(period string) ([]domain.TrendDataPoint, error) {
			return []domain.TrendDataPoint{
				{Date: "2026-01-15", Income: 100, Expense: 50, Net: 50},
			}, nil
		},
	}
	service := NewTransactionService(mockRepo)

	trends, err := service.GetTrends("daily")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if trends.Period != "daily" {
		t.Errorf("expected period 'daily', got %s", trends.Period)
	}
	if len(trends.Data) != 1 {
		t.Errorf("expected 1 data point, got %d", len(trends.Data))
	}
}

// Test GetBreakdownBySource

func TestGetBreakdownBySource_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getBreakdownSource: func() ([]domain.BreakdownResponse, error) {
			return []domain.BreakdownResponse{
				{Label: "Bank ABC", Amount: 500, Percentage: 50, Count: 5},
			}, nil
		},
	}
	service := NewTransactionService(mockRepo)

	breakdown, err := service.GetBreakdownBySource()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(breakdown) != 1 {
		t.Errorf("expected 1 breakdown item, got %d", len(breakdown))
	}
}

// Test GetBreakdownByCategory

func TestGetBreakdownByCategory_Success(t *testing.T) {
	mockRepo := &mockRepository{
		getBreakdownCategory: func() ([]domain.BreakdownResponse, error) {
			return []domain.BreakdownResponse{
				{Label: "Food", Amount: 300, Percentage: 30, Count: 10},
			}, nil
		},
	}
	service := NewTransactionService(mockRepo)

	breakdown, err := service.GetBreakdownByCategory()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(breakdown) != 1 {
		t.Errorf("expected 1 breakdown item, got %d", len(breakdown))
	}
}
