package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
	"github.com/dev/personal-finance-tracker/backend/test/util"
)

func setupTestDB(t *testing.T) *util.PostgresTestContainer {
	t.Helper()

	pc := util.SetupPostgresContainerWithDefaults(t)
	return pc
}

// Test complete transaction CRUD flow

func TestIntegration_TransactionCRUD(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	t.Run("Create transaction", func(t *testing.T) {
		tx := &domain.Transaction{
			Amount:          100.50,
			Type:            domain.TransactionTypeOut,
			Category:        "Food",
			Description:     "Lunch",
			Source:          "Bank ABC",
			SourceAccount:   "1234",
			TransactionDate: time.Now().Truncate(time.Second),
		}

		err := repo.Create(tx)
		assert.NoError(t, err)
		assert.Greater(t, tx.ID, int64(0))
	})

	t.Run("Find by ID", func(t *testing.T) {
		tx := &domain.Transaction{
			Amount:          200.00,
			Type:            domain.TransactionTypeIn,
			Category:        "Salary",
			Description:     "Monthly salary",
			Source:          "Bank XYZ",
			TransactionDate: time.Now().Truncate(time.Second),
		}

		err := repo.Create(tx)
		require.NoError(t, err)

		found, err := repo.FindByID(tx.ID)
		assert.NoError(t, err)
		assert.Equal(t, tx.ID, found.ID)
		assert.Equal(t, 200.00, found.Amount)
		assert.Equal(t, domain.TransactionTypeIn, found.Type)
	})

	t.Run("List transactions", func(t *testing.T) {
		// Create multiple transactions
		for i := 0; i < 5; i++ {
			tx := &domain.Transaction{
				Amount:          float64(100 + i),
				Type:            domain.TransactionTypeOut,
				Category:        "Food",
				Source:          "Test Bank",
				TransactionDate: time.Now().Truncate(time.Second),
			}
			err := repo.Create(tx)
			require.NoError(t, err)
		}

		params := domain.ListTransactionsQueryParams{
			Page:     1,
			PageSize: 10,
		}

		transactions, total, err := repo.List(params)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(transactions), 5)
		assert.GreaterOrEqual(t, total, int64(5))
	})

	t.Run("Filter by type", func(t *testing.T) {
		// Create specific transaction
		tx := &domain.Transaction{
			Amount:          300.00,
			Type:            domain.TransactionTypeOut,
			Category:        "Shopping",
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		err := repo.Create(tx)
		require.NoError(t, err)

		params := domain.ListTransactionsQueryParams{
			Page:     1,
			PageSize: 10,
			Type:     domain.TransactionTypeOut,
		}

		transactions, total, err := repo.List(params)
		assert.NoError(t, err)
		assert.Greater(t, total, int64(0))
		// Verify all returned are type 'out'
		for _, tx := range transactions {
			assert.Equal(t, domain.TransactionTypeOut, tx.Type)
		}
	})
}

func TestIntegration_BatchInsert(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	transactions := make([]domain.Transaction, 10)
	for i := 0; i < 10; i++ {
		transactions[i] = domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Category:        "Food",
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
	}

	err := repo.CreateInBatch(transactions)
	assert.NoError(t, err)

	// Verify all were inserted
	params := domain.ListTransactionsQueryParams{
		Page:     1,
		PageSize: 100,
	}

	all, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
	assert.Len(t, all, 10)
}

func TestIntegration_GetSummary(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	// Create test transactions
	transactions := []domain.Transaction{
		{Amount: 1000, Type: domain.TransactionTypeIn, Source: "Bank", TransactionDate: time.Now().Truncate(time.Second)},
		{Amount: 500, Type: domain.TransactionTypeIn, Source: "Bank", TransactionDate: time.Now().Truncate(time.Second)},
		{Amount: 200, Type: domain.TransactionTypeOut, Source: "Bank", TransactionDate: time.Now().Truncate(time.Second)},
		{Amount: 300, Type: domain.TransactionTypeOut, Source: "Bank", TransactionDate: time.Now().Truncate(time.Second)},
	}

	for _, tx := range transactions {
		err := repo.Create(&tx)
		require.NoError(t, err)
	}

	summary, err := repo.GetSummary()
	assert.NoError(t, err)
	assert.Equal(t, 1500.00, summary.TotalIncome)
	assert.Equal(t, 500.00, summary.TotalExpense)
	assert.Equal(t, 1000.00, summary.CurrentBalance)
	assert.Equal(t, int64(4), summary.TransactionCount)
}

func TestIntegration_GetBreakdownBySource(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	// Create transactions from different sources
	transactions := []domain.Transaction{
		{Amount: 100, Type: domain.TransactionTypeOut, Source: "Bank A", TransactionDate: time.Now()},
		{Amount: 200, Type: domain.TransactionTypeOut, Source: "Bank A", TransactionDate: time.Now()},
		{Amount: 150, Type: domain.TransactionTypeOut, Source: "Bank B", TransactionDate: time.Now()},
	}

	for _, tx := range transactions {
		err := repo.Create(&tx)
		require.NoError(t, err)
	}

	breakdown, err := repo.GetBreakdownBySource()
	assert.NoError(t, err)
	assert.Len(t, breakdown, 2)

	// Verify breakdown
	sourceAmounts := make(map[string]float64)
	for _, item := range breakdown {
		sourceAmounts[item.Label] = item.Amount
	}
	assert.Equal(t, 300.00, sourceAmounts["Bank A"])
	assert.Equal(t, 150.00, sourceAmounts["Bank B"])
}

func TestIntegration_GetBreakdownByCategory(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	// Create transactions with different categories
	transactions := []domain.Transaction{
		{Amount: 100, Type: domain.TransactionTypeOut, Category: "Food", Source: "Bank", TransactionDate: time.Now()},
		{Amount: 200, Type: domain.TransactionTypeOut, Category: "Food", Source: "Bank", TransactionDate: time.Now()},
		{Amount: 150, Type: domain.TransactionTypeOut, Category: "Transportation", Source: "Bank", TransactionDate: time.Now()},
		{Amount: 50, Type: domain.TransactionTypeOut, Category: "", Source: "Bank", TransactionDate: time.Now()},
	}

	for _, tx := range transactions {
		err := repo.Create(&tx)
		require.NoError(t, err)
	}

	breakdown, err := repo.GetBreakdownByCategory()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(breakdown), 2)

	// Verify uncategorized is handled
	categoryAmounts := make(map[string]float64)
	for _, item := range breakdown {
		categoryAmounts[item.Label] = item.Amount
	}
	assert.Equal(t, 300.00, categoryAmounts["Food"])
	assert.Equal(t, 150.00, categoryAmounts["Transportation"])
}

func TestIntegration_GetTrends(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	// Create transactions across different days
	now := time.Now().Truncate(time.Second)
	transactions := []domain.Transaction{
		{Amount: 100, Type: domain.TransactionTypeOut, Source: "Bank", TransactionDate: now.Add(-24 * time.Hour)},
		{Amount: 200, Type: domain.TransactionTypeIn, Source: "Bank", TransactionDate: now.Add(-24 * time.Hour)},
		{Amount: 150, Type: domain.TransactionTypeOut, Source: "Bank", TransactionDate: now},
		{Amount: 300, Type: domain.TransactionTypeIn, Source: "Bank", TransactionDate: now},
	}

	for _, tx := range transactions {
		err := repo.Create(&tx)
		require.NoError(t, err)
	}

	trends, err := repo.GetTrends("daily")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(trends), 1)
}

func TestIntegration_TransactionNotFound(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	tx, err := repo.FindByID(99999)
	assert.Error(t, err)
	assert.Nil(t, tx)
}

func TestIntegration_EmptyDatabase(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	t.Run("Empty list", func(t *testing.T) {
		params := domain.ListTransactionsQueryParams{
			Page:     1,
			PageSize: 10,
		}

		transactions, total, err := repo.List(params)
		assert.NoError(t, err)
		assert.Len(t, transactions, 0)
		assert.Equal(t, int64(0), total)
	})

	t.Run("Empty summary", func(t *testing.T) {
		summary, err := repo.GetSummary()
		assert.NoError(t, err)
		assert.Equal(t, 0.00, summary.TotalIncome)
		assert.Equal(t, 0.00, summary.TotalExpense)
		assert.Equal(t, 0.00, summary.CurrentBalance)
		assert.Equal(t, int64(0), summary.TransactionCount)
	})

	t.Run("Empty breakdown", func(t *testing.T) {
		breakdown, err := repo.GetBreakdownBySource()
		assert.NoError(t, err)
		assert.Len(t, breakdown, 0)
	})
}

func TestIntegration_DateFilters(t *testing.T) {
	pc := setupTestDB(t)
	db := pc.NewGORMDBWithAutoMigrate(t, &domain.Transaction{})
	repo := repository.NewTransactionRepository(db)

	now := time.Now().Truncate(time.Second)

	// Create transactions with different dates
	oldTransaction := &domain.Transaction{
		Amount:          100,
		Type:            domain.TransactionTypeOut,
		Source:          "Bank",
		TransactionDate: now.Add(-30 * 24 * time.Hour), // 30 days ago
	}
	recentTransaction := &domain.Transaction{
		Amount:          200,
		Type:            domain.TransactionTypeOut,
		Source:          "Bank",
		TransactionDate: now.Add(-24 * time.Hour), // 1 day ago
	}

	err := repo.Create(oldTransaction)
	require.NoError(t, err)
	err = repo.Create(recentTransaction)
	require.NoError(t, err)

	// Filter by date range
	startDate := now.Add(-7 * 24 * time.Hour).Format("2006-01-02")
	endDate := now.Format("2006-01-02")

	params := domain.ListTransactionsQueryParams{
		Page:      1,
		PageSize:  10,
		StartDate: startDate,
		EndDate:   endDate,
	}

	transactions, total, err := repo.List(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, transactions, 1)
	assert.Equal(t, 200.00, transactions[0].Amount)
}
