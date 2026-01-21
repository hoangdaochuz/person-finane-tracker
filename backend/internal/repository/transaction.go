package repository

import (
	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/security"
)

// TransactionRepository handles database operations for transactions
type TransactionRepository interface {
	Create(tx *domain.Transaction) error
	CreateInBatch(transactions []domain.Transaction) error
	FindByID(id int64) (*domain.Transaction, error)
	List(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error)
	GetSummary() (*domain.SummaryResponse, error)
	GetTrends(period string) ([]domain.TrendDataPoint, error)
	GetBreakdownBySource() ([]domain.BreakdownResponse, error)
	GetBreakdownByCategory() ([]domain.BreakdownResponse, error)
}

type transactionRepository struct {
	db         *gorm.DB
	sanitizer *security.Sanitizer
}

const (
	maxPageSize = 100
)

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{
		db:         db,
		sanitizer: security.NewSanitizer(),
	}
}

func (r *transactionRepository) Create(tx *domain.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *transactionRepository) CreateInBatch(transactions []domain.Transaction) error {
	if len(transactions) == 0 {
		return nil
	}
	return r.db.CreateInBatches(transactions, 100).Error
}

func (r *transactionRepository) FindByID(id int64) (*domain.Transaction, error) {
	var tx domain.Transaction
	// Use parameterized query (implicit protection via GORM)
	err := r.db.First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) List(params domain.ListTransactionsQueryParams) ([]domain.Transaction, int64, error) {
	var transactions []domain.Transaction
	var total int64

	query := r.db.Model(&domain.Transaction{})

	// Apply filters with explicit sanitization (defense-in-depth)
	// GORM's ? placeholder provides parameterized query protection
	if params.Type != "" {
		// Validate transaction type before using in query
		typeStr := string(params.Type)
		if r.sanitizer.ValidateTransactionType(typeStr) {
			query = query.Where("type = ?", params.Type)
		}
	}
	if params.Source != "" {
		// Sanitize source input to prevent injection
		safeSource := r.sanitizer.CleanInput(params.Source, domain.MaxSourceLength)
		query = query.Where("source = ?", safeSource)
	}
	if params.Category != "" {
		// Validate category against whitelist
		if r.sanitizer.ValidateCategory(params.Category) {
			safeCategory := r.sanitizer.CleanInput(params.Category, domain.MaxCategoryLength)
			query = query.Where("category = ?", safeCategory)
		}
	}
	if params.StartDate != "" {
		// Dates are always in ISO 8601 format (RFC3339), which is safe
		// The validation in domain layer ensures proper format
		query = query.Where("transaction_date >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("transaction_date <= ?", params.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Enforce maximum page size (prevent DoS via large page sizes)
	pageSize := params.PageSize
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if pageSize <= 0 {
		pageSize = 20 // default page size
	}

	// Apply pagination with offset validation
	if params.Page < 1 {
		params.Page = 1
	}
	offset := (params.Page - 1) * pageSize

	err := query.
		Order("transaction_date DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

func (r *transactionRepository) GetSummary() (*domain.SummaryResponse, error) {
	var result struct {
		TotalIncome      float64
		TotalExpense     float64
		TransactionCount int64
	}

	// Raw SQL is safe here - no user input, hardcoded query
	err := r.db.Model(&domain.Transaction{}).
		Select(`
			COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) as total_income,
			COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as total_expense,
			COUNT(*) as transaction_count
		`).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &domain.SummaryResponse{
		TotalIncome:      result.TotalIncome,
		TotalExpense:     result.TotalExpense,
		CurrentBalance:   result.TotalIncome - result.TotalExpense,
		TransactionCount: result.TransactionCount,
	}, nil
}

func (r *transactionRepository) GetTrends(period string) ([]domain.TrendDataPoint, error) {
	var results []domain.TrendDataPoint

	// Validate and sanitize period parameter (whitelist approach)
	safePeriod := r.sanitizer.ValidatePeriod(period)

	// Use safe, hardcoded query templates based on validated period
	// This prevents SQL injection as the SQL templates are completely hardcoded
	var query string
	switch safePeriod {
	case "daily":
		query = `
			SELECT
				TO_CHAR(transaction_date, 'YYYY-MM-DD') as date,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) as income,
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as expense,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) -
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as net
			FROM transactions
			GROUP BY date
			ORDER BY date DESC
			LIMIT 30
		`
	case "weekly":
		query = `
			SELECT
				TO_CHAR(transaction_date, 'YYYY-"W"IW') as date,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) as income,
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as expense,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) -
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as net
			FROM transactions
			GROUP BY date
			ORDER BY date DESC
			LIMIT 30
		`
	case "monthly":
		query = `
			SELECT
				TO_CHAR(transaction_date, 'YYYY-MM') as date,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) as income,
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as expense,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) -
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as net
			FROM transactions
			GROUP BY date
			ORDER BY date DESC
			LIMIT 30
		`
	default:
		// Default to daily if invalid period is provided
		query = `
			SELECT
				TO_CHAR(transaction_date, 'YYYY-MM-DD') as date,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) as income,
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as expense,
				COALESCE(SUM(CASE WHEN type = 'in' THEN amount ELSE 0 END), 0) -
				COALESCE(SUM(CASE WHEN type = 'out' THEN amount ELSE 0 END), 0) as net
			FROM transactions
			GROUP BY date
			ORDER BY date DESC
			LIMIT 30
		`
	}

	// Raw SQL is safe here - query is completely hardcoded, no user input interpolation
	err := r.db.Raw(query).Scan(&results).Error
	return results, err
}

func (r *transactionRepository) GetBreakdownBySource() ([]domain.BreakdownResponse, error) {
	var results []domain.BreakdownResponse

	// First get total amount
	var totalExpense float64
	r.db.Model(&domain.Transaction{}).
		Where("type = ?", domain.TransactionTypeOut).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	// Get breakdown by source - hardcoded SQL with no user input
	query := `
		SELECT
			source as label,
			COALESCE(SUM(amount), 0) as amount,
			COUNT(*) as count
		FROM transactions
		WHERE type = 'out'
		GROUP BY source
		ORDER BY amount DESC
	`

	err := r.db.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Calculate percentages
	for i := range results {
		if totalExpense > 0 {
			results[i].Percentage = (results[i].Amount / totalExpense) * 100
		}
	}

	return results, nil
}

func (r *transactionRepository) GetBreakdownByCategory() ([]domain.BreakdownResponse, error) {
	var results []domain.BreakdownResponse

	// First get total amount
	var totalExpense float64
	r.db.Model(&domain.Transaction{}).
		Where("type = ?", domain.TransactionTypeOut).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalExpense)

	// Get breakdown by category - hardcoded SQL with no user input
	query := `
		SELECT
			COALESCE(category, 'Uncategorized') as label,
			COALESCE(SUM(amount), 0) as amount,
			COUNT(*) as count
		FROM transactions
		WHERE type = 'out'
		GROUP BY category
		ORDER BY amount DESC
	`

	err := r.db.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	// Calculate percentages
	for i := range results {
		if totalExpense > 0 {
			results[i].Percentage = (results[i].Amount / totalExpense) * 100
		}
	}

	return results, nil
}
