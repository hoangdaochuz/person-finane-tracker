package service

import (
	"errors"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/internal/security"
)

// TransactionService handles business logic for transactions
type TransactionService interface {
	CreateTransaction(req *domain.CreateTransactionRequest) (*domain.Transaction, error)
	CreateBatchTransaction(req *domain.BatchTransactionRequest) ([]domain.Transaction, error)
	GetTransactionByID(id int64) (*domain.Transaction, error)
	ListTransactions(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error)
	GetSummary() (*domain.SummaryResponse, error)
	GetTrends(period string) (*domain.TrendsResponse, error)
	GetBreakdownBySource() ([]domain.BreakdownResponse, error)
	GetBreakdownByCategory() ([]domain.BreakdownResponse, error)
}

type transactionService struct {
	repo       repository.TransactionRepository
	sanitizer *security.Sanitizer
}

// NewTransactionService creates a new transaction service
func NewTransactionService(repo repository.TransactionRepository) TransactionService {
	return &transactionService{
		repo:       repo,
		sanitizer: security.NewSanitizer(),
	}
}

func (s *transactionService) CreateTransaction(req *domain.CreateTransactionRequest) (*domain.Transaction, error) {
	// Perform validation (includes SQL injection prevention)
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Additional explicit sanitization for defense-in-depth
	req.Source = s.sanitizer.CleanInput(req.Source, domain.MaxSourceLength)
	if req.Category != "" {
		req.Category = s.sanitizer.CleanInput(req.Category, domain.MaxCategoryLength)
	}
	req.Description = s.sanitizer.CleanInput(req.Description, domain.MaxDescriptionLength)
	if req.SourceAccount != "" {
		req.SourceAccount = s.sanitizer.CleanInput(req.SourceAccount, domain.MaxAccountLength)
	}
	if req.Recipient != "" {
		req.Recipient = s.sanitizer.CleanInput(req.Recipient, domain.MaxRecipientLength)
	}

	// Convert request to domain
	transaction, err := req.ToTransaction()
	if err != nil {
		return nil, err
	}

	// Create transaction (repository uses parameterized queries)
	if err := s.repo.Create(transaction); err != nil {
		return nil, err
	}

	return transaction, nil
}

func (s *transactionService) CreateBatchTransaction(req *domain.BatchTransactionRequest) ([]domain.Transaction, error) {
	// Perform validation
	if err := req.Validate(); err != nil {
		return nil, err
	}

	transactions := make([]domain.Transaction, 0, len(req.Transactions))

	for _, t := range req.Transactions {
		// Sanitize each transaction's fields
		t.Source = s.sanitizer.CleanInput(t.Source, domain.MaxSourceLength)
		if t.Category != "" {
			t.Category = s.sanitizer.CleanInput(t.Category, domain.MaxCategoryLength)
		}
		t.Description = s.sanitizer.CleanInput(t.Description, domain.MaxDescriptionLength)
		if t.SourceAccount != "" {
			t.SourceAccount = s.sanitizer.CleanInput(t.SourceAccount, domain.MaxAccountLength)
		}
		if t.Recipient != "" {
			t.Recipient = s.sanitizer.CleanInput(t.Recipient, domain.MaxRecipientLength)
		}

		// Convert request to domain
		transaction, err := t.ToTransaction()
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, *transaction)
	}

	// Create transactions in batch
	if err := s.repo.CreateInBatch(transactions); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (s *transactionService) GetTransactionByID(id int64) (*domain.Transaction, error) {
	// Validate ID to prevent path traversal or injection
	if id <= 0 {
		return nil, errors.New("invalid transaction ID")
	}
	return s.repo.FindByID(id)
}

func (s *transactionService) ListTransactions(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
	// Validate pagination params to prevent DoS
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 || params.PageSize > 100 {
		params.PageSize = 20
	}

	// Sanitize filter parameters
	if params.Source != "" {
		params.Source = s.sanitizer.CleanInput(params.Source, domain.MaxSourceLength)
	}
	if params.Category != "" {
		// Validate against whitelist before passing to repository
		if !s.sanitizer.ValidateCategory(params.Category) {
			// Return empty result if invalid category
			return []domain.Transaction{}, 0, nil
		}
		params.Category = s.sanitizer.CleanInput(params.Category, domain.MaxCategoryLength)
	}

	return s.repo.List(params)
}

func (s *transactionService) GetSummary() (*domain.SummaryResponse, error) {
	return s.repo.GetSummary()
}

func (s *transactionService) GetTrends(period string) (*domain.TrendsResponse, error) {
	data, err := s.repo.GetTrends(period)
	if err != nil {
		return nil, err
	}

	return &domain.TrendsResponse{
		Period: period,
		Data:   data,
	}, nil
}

func (s *transactionService) GetBreakdownBySource() ([]domain.BreakdownResponse, error) {
	return s.repo.GetBreakdownBySource()
}

func (s *transactionService) GetBreakdownByCategory() ([]domain.BreakdownResponse, error) {
	return s.repo.GetBreakdownByCategory()
}
