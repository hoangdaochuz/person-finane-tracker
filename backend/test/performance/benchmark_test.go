package performance

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
	"github.com/dev/personal-finance-tracker/backend/internal/repository"
)

// setupBenchmarkDB creates a test database for benchmarking
func setupBenchmarkDB(b *testing.B) *gorm.DB {
	b.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("benchdb"),
		postgres.WithUsername("bench"),
		postgres.WithPassword("bench"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)
	require.NoError(b, err)

	b.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			b.Logf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(b, err)

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	require.NoError(b, err)

	err = db.AutoMigrate(&domain.Transaction{})
	require.NoError(b, err)

	return db
}

// Benchmark CreateTransaction

func BenchmarkCreateTransaction(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := &domain.Transaction{
			Amount:          100.50,
			Type:            domain.TransactionTypeOut,
			Category:        "Food",
			Description:     "Benchmark transaction",
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		if err := repo.Create(tx); err != nil {
			b.Fatalf("failed to create transaction: %v", err)
		}
	}
}

// Benchmark CreateInBatch

func BenchmarkCreateInBatch_10(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	batchSize := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transactions := make([]domain.Transaction, batchSize)
		for j := 0; j < batchSize; j++ {
			transactions[j] = domain.Transaction{
				Amount:          float64(100 + j),
				Type:            domain.TransactionTypeOut,
				Category:        "Food",
				Source:          "Test Bank",
				TransactionDate: time.Now().Truncate(time.Second),
			}
		}
		if err := repo.CreateInBatch(transactions); err != nil {
			b.Fatalf("failed to create batch: %v", err)
		}
	}
}

func BenchmarkCreateInBatch_50(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	batchSize := 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transactions := make([]domain.Transaction, batchSize)
		for j := 0; j < batchSize; j++ {
			transactions[j] = domain.Transaction{
				Amount:          float64(100 + j),
				Type:            domain.TransactionTypeOut,
				Category:        "Food",
				Source:          "Test Bank",
				TransactionDate: time.Now().Truncate(time.Second),
			}
		}
		if err := repo.CreateInBatch(transactions); err != nil {
			b.Fatalf("failed to create batch: %v", err)
		}
	}
}

func BenchmarkCreateInBatch_100(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	batchSize := 100

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		transactions := make([]domain.Transaction, batchSize)
		for j := 0; j < batchSize; j++ {
			transactions[j] = domain.Transaction{
				Amount:          float64(100 + j),
				Type:            domain.TransactionTypeOut,
				Category:        "Food",
				Source:          "Test Bank",
				TransactionDate: time.Now().Truncate(time.Second),
			}
		}
		if err := repo.CreateInBatch(transactions); err != nil {
			b.Fatalf("failed to create batch: %v", err)
		}
	}
}

// Benchmark FindByID

func BenchmarkFindByID(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create a transaction to find
	tx := &domain.Transaction{
		Amount:          100.50,
		Type:            domain.TransactionTypeOut,
		Source:          "Test Bank",
		TransactionDate: time.Now().Truncate(time.Second),
	}
	require.NoError(b, repo.Create(tx))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.FindByID(tx.ID)
		if err != nil {
			b.Fatalf("failed to find by ID: %v", err)
		}
	}
}

// Benchmark List transactions

func BenchmarkList_10Records(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create 10 transactions
	for i := 0; i < 10; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	params := domain.ListTransactionsQueryParams{
		Page:     1,
		PageSize: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(params)
		if err != nil {
			b.Fatalf("failed to list: %v", err)
		}
	}
}

func BenchmarkList_100Records(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create 100 transactions
	for i := 0; i < 100; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	params := domain.ListTransactionsQueryParams{
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(params)
		if err != nil {
			b.Fatalf("failed to list: %v", err)
		}
	}
}

func BenchmarkList_1000Records(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create 1000 transactions
	for i := 0; i < 1000; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	params := domain.ListTransactionsQueryParams{
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(params)
		if err != nil {
			b.Fatalf("failed to list: %v", err)
		}
	}
}

// Benchmark GetSummary

func BenchmarkGetSummary_10Records(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create 10 transactions
	for i := 0; i < 10; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetSummary()
		if err != nil {
			b.Fatalf("failed to get summary: %v", err)
		}
	}
}

func BenchmarkGetSummary_1000Records(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create 1000 transactions
	for i := 0; i < 1000; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetSummary()
		if err != nil {
			b.Fatalf("failed to get summary: %v", err)
		}
	}
}

// Benchmark GetBreakdownBySource

func BenchmarkGetBreakdownBySource_100Records(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	sources := []string{"Bank A", "Bank B", "Bank C", "Wallet"}
	// Create 100 transactions
	for i := 0; i < 100; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          sources[i%len(sources)],
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetBreakdownBySource()
		if err != nil {
			b.Fatalf("failed to get breakdown: %v", err)
		}
	}
}

// Benchmark GetBreakdownByCategory

func BenchmarkGetBreakdownByCategory_100Records(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	categories := []string{"Food", "Transportation", "Utilities", "Shopping", "Entertainment"}
	// Create 100 transactions
	for i := 0; i < 100; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Category:        categories[i%len(categories)],
			Source:          "Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetBreakdownByCategory()
		if err != nil {
			b.Fatalf("failed to get breakdown: %v", err)
		}
	}
}

// Benchmark GetTrends

func BenchmarkGetTrends_Daily(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create transactions across different days
	now := time.Now()
	for i := 0; i < 30; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          "Bank",
			TransactionDate: now.AddDate(0, 0, -i).Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := repo.GetTrends("daily")
		if err != nil {
			b.Fatalf("failed to get trends: %v", err)
		}
	}
}

// Benchmark filtered queries

func BenchmarkList_WithTypeFilter(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	// Create mixed transaction types
	for i := 0; i < 100; i++ {
		txType := domain.TransactionTypeOut
		if i%2 == 0 {
			txType = domain.TransactionTypeIn
		}
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            txType,
			Source:          "Test Bank",
			TransactionDate: time.Now().Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	params := domain.ListTransactionsQueryParams{
		Page:     1,
		PageSize: 20,
		Type:     domain.TransactionTypeOut,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(params)
		if err != nil {
			b.Fatalf("failed to list: %v", err)
		}
	}
}

func BenchmarkList_WithDateFilter(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	now := time.Now()
	// Create transactions across different dates
	for i := 0; i < 100; i++ {
		tx := &domain.Transaction{
			Amount:          float64(100 + i),
			Type:            domain.TransactionTypeOut,
			Source:          "Test Bank",
			TransactionDate: now.AddDate(0, 0, -i).Truncate(time.Second),
		}
		require.NoError(b, repo.Create(tx))
	}

	params := domain.ListTransactionsQueryParams{
		Page:      1,
		PageSize:  20,
		StartDate: now.AddDate(0, 0, -30).Format("2006-01-02"),
		EndDate:   now.Format("2006-01-02"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := repo.List(params)
		if err != nil {
			b.Fatalf("failed to list: %v", err)
		}
	}
}

// Benchmark parallel inserts

func BenchmarkParallelInserts(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tx := &domain.Transaction{
				Amount:          float64(100 + i),
				Type:            domain.TransactionTypeOut,
				Source:          "Test Bank",
				TransactionDate: time.Now().Truncate(time.Second),
			}
			if err := repo.Create(tx); err != nil {
				b.Errorf("failed to create transaction: %v", err)
			}
			i++
		}
	})
}

// Benchmark sequential vs batch inserts

func BenchmarkSequentialInserts_100(b *testing.B) {
	db := setupBenchmarkDB(b)
	repo := repository.NewTransactionRepository(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			tx := &domain.Transaction{
				Amount:          float64(100 + j),
				Type:            domain.TransactionTypeOut,
				Source:          "Test Bank",
				TransactionDate: time.Now().Truncate(time.Second),
			}
			if err := repo.Create(tx); err != nil {
				b.Fatalf("failed to create transaction: %v", err)
			}
		}
		// Clear table for next iteration
		db.Exec("DELETE FROM transactions")
	}
}
