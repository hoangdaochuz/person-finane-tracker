package repository

import (
	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
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
	db *gorm.DB
}

const (
	maxPageSize = 100
)

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
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

	// Apply filters
	if params.Type != "" {
		query = query.Where("type = ?", params.Type)
	}
	if params.Source != "" {
		query = query.Where("source = ?", params.Source)
	}
	if params.Category != "" {
		query = query.Where("category = ?", params.Category)
	}
	if params.StartDate != "" {
		query = query.Where("transaction_date >= ?", params.StartDate)
	}
	if params.EndDate != "" {
		query = query.Where("transaction_date <= ?", params.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Enforce maximum page size
	pageSize := params.PageSize
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	if pageSize <= 0 {
		pageSize = 20 // default page size
	}

	// Apply pagination
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

	// Use safe, hardcoded query templates to prevent SQL injection
	var query string
	switch period {
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

	// Get breakdown by source
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

	// Get breakdown by category
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
