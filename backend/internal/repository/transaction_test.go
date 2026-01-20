package repository

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open gorm connection: %v", err)
	}

	return gormDB, mock, sqlDB
}

// Test Create

func TestTransactionRepository_Create_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	now := time.Now().Truncate(time.Second)
	tx := &domain.Transaction{
		Amount:          100.50,
		Type:            domain.TransactionTypeOut,
		Category:        "Food",
		Description:     "Lunch",
		Source:          "Bank ABC",
		SourceAccount:   "1234",
		TransactionDate: now,
	}

	// GORM uses Query with returning for INSERT in PostgreSQL
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(tx)

	assert.NoError(t, err)
}

func TestTransactionRepository_Create_DatabaseError(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	tx := &domain.Transaction{
		Amount:          100.50,
		Type:            domain.TransactionTypeOut,
		Source:          "Bank ABC",
		TransactionDate: time.Now(),
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO").
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	err := repo.Create(tx)

	assert.Error(t, err)
}

// Test CreateInBatch

func TestTransactionRepository_CreateInBatch_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	now := time.Now().Truncate(time.Second)
	transactions := []domain.Transaction{
		{
			Amount:          100.50,
			Type:            domain.TransactionTypeOut,
			Source:          "Bank ABC",
			TransactionDate: now,
		},
		{
			Amount:          200.00,
			Type:            domain.TransactionTypeIn,
			Source:          "Bank XYZ",
			TransactionDate: now,
		},
	}

	// GORM uses Query with RETURNING for batch inserts
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
	mock.ExpectCommit()

	err := repo.CreateInBatch(transactions)

	assert.NoError(t, err)
}

func TestTransactionRepository_CreateInBatch_Empty(t *testing.T) {
	db, _, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	err := repo.CreateInBatch([]domain.Transaction{})

	assert.NoError(t, err)
}

// Test FindByID

func TestTransactionRepository_FindByID_Found(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"id", "amount", "type", "category", "description", "source", "source_account", "recipient", "transaction_date", "created_at", "updated_at"}).
		AddRow(1, 100.50, "out", "Food", "Lunch", "Bank ABC", "1234", "", now, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE "transactions"."id" = $1 ORDER BY "transactions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	tx, err := repo.FindByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, int64(1), tx.ID)
	assert.Equal(t, 100.50, tx.Amount)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionRepository_FindByID_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE "transactions"."id" = $1 ORDER BY "transactions"."id" LIMIT $2`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	tx, err := repo.FindByID(999)

	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test List

func TestTransactionRepository_List_DefaultPagination(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	now := time.Now().Truncate(time.Second)

	// Mock count query
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(countRows)

	// Mock select query
	rows := sqlmock.NewRows([]string{"id", "amount", "type", "category", "description", "source", "source_account", "recipient", "transaction_date", "created_at", "updated_at"}).
		AddRow(1, 100.50, "out", "Food", "Lunch", "Bank ABC", "1234", "", now, now, now)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	params := domain.ListTransactionsQueryParams{
		Page:     1,
		PageSize: 20,
	}

	transactions, total, err := repo.List(params)

	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, int64(1), total)
}

func TestTransactionRepository_List_WithFilters(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	now := time.Now().Truncate(time.Second)

	// Mock count query with filters
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(countRows)

	// Mock select query with filters
	rows := sqlmock.NewRows([]string{"id", "amount", "type", "category", "description", "source", "source_account", "recipient", "transaction_date", "created_at", "updated_at"}).
		AddRow(1, 100.50, "out", "Food", "Lunch", "Bank ABC", "1234", "", now, now, now)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	params := domain.ListTransactionsQueryParams{
		Page:     1,
		PageSize: 20,
		Type:     domain.TransactionTypeOut,
		Source:   "Bank ABC",
		Category: "Food",
	}

	transactions, total, err := repo.List(params)

	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, int64(1), total)
}

func TestTransactionRepository_List_WithDateFilters(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	now := time.Now().Truncate(time.Second)

	// Mock count query with date filters
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(countRows)

	// Mock select query with date filters
	rows := sqlmock.NewRows([]string{"id", "amount", "type", "category", "description", "source", "source_account", "recipient", "transaction_date", "created_at", "updated_at"}).
		AddRow(1, 100.50, "out", "Food", "Lunch", "Bank ABC", "1234", "", now, now, now)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	params := domain.ListTransactionsQueryParams{
		Page:      1,
		PageSize:  20,
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
	}

	transactions, total, err := repo.List(params)

	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, int64(1), total)
}

// Test GetSummary

func TestTransactionRepository_GetSummary_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	rows := sqlmock.NewRows([]string{"total_income", "total_expense", "transaction_count"}).
		AddRow(1000.00, 500.00, 10)

	// Use a simpler regex pattern that matches the SELECT query
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	summary, err := repo.GetSummary()

	assert.NoError(t, err)
	assert.Equal(t, 1000.00, summary.TotalIncome)
	assert.Equal(t, 500.00, summary.TotalExpense)
	assert.Equal(t, 500.00, summary.CurrentBalance)
	assert.Equal(t, int64(10), summary.TransactionCount)
}

// Test GetTrends

func TestTransactionRepository_GetTrends_Daily(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	rows := sqlmock.NewRows([]string{"date", "income", "expense", "net"}).
		AddRow("2026-01-15", 100.00, 50.00, 50.00)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	trends, err := repo.GetTrends("daily")

	assert.NoError(t, err)
	assert.Len(t, trends, 1)
	assert.Equal(t, "2026-01-15", trends[0].Date)
	assert.Equal(t, 100.00, trends[0].Income)
}

func TestTransactionRepository_GetTrends_Monthly(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	rows := sqlmock.NewRows([]string{"date", "income", "expense", "net"}).
		AddRow("2026-01", 1000.00, 500.00, 500.00)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	trends, err := repo.GetTrends("monthly")

	assert.NoError(t, err)
	assert.Len(t, trends, 1)
}

// Test GetBreakdownBySource

func TestTransactionRepository_GetBreakdownBySource_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	// Mock total expense query
	totalRows := sqlmock.NewRows([]string{"COALESCE(SUM(amount), 0)"}).AddRow(1000.00)
	mock.ExpectQuery("SELECT").WillReturnRows(totalRows)

	// Mock breakdown query
	rows := sqlmock.NewRows([]string{"label", "amount", "count"}).
		AddRow("Bank ABC", 500.00, 5).
		AddRow("Bank XYZ", 300.00, 3)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	breakdown, err := repo.GetBreakdownBySource()

	assert.NoError(t, err)
	assert.Len(t, breakdown, 2)
	assert.Equal(t, "Bank ABC", breakdown[0].Label)
	assert.Equal(t, 500.00, breakdown[0].Amount)
	assert.Equal(t, 50.0, breakdown[0].Percentage)
}

// Test GetBreakdownByCategory

func TestTransactionRepository_GetBreakdownByCategory_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	// Mock total expense query
	totalRows := sqlmock.NewRows([]string{"COALESCE(SUM(amount), 0)"}).AddRow(1000.00)
	mock.ExpectQuery("SELECT").WillReturnRows(totalRows)

	// Mock breakdown query
	rows := sqlmock.NewRows([]string{"label", "amount", "count"}).
		AddRow("Food", 300.00, 10).
		AddRow("Transportation", 200.00, 5)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	breakdown, err := repo.GetBreakdownByCategory()

	assert.NoError(t, err)
	assert.Len(t, breakdown, 2)
	assert.Equal(t, "Food", breakdown[0].Label)
	assert.Equal(t, 300.00, breakdown[0].Amount)
	assert.Equal(t, 30.0, breakdown[0].Percentage)
}

func TestTransactionRepository_GetBreakdownByCategory_Uncategorized(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewTransactionRepository(db)

	// Mock total expense query
	totalRows := sqlmock.NewRows([]string{"COALESCE(SUM(amount), 0)"}).AddRow(500.00)
	mock.ExpectQuery("SELECT").WillReturnRows(totalRows)

	// Mock breakdown query with NULL category
	rows := sqlmock.NewRows([]string{"label", "amount", "count"}).
		AddRow("Uncategorized", 100.00, 2)

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	breakdown, err := repo.GetBreakdownByCategory()

	assert.NoError(t, err)
	assert.Len(t, breakdown, 1)
	assert.Equal(t, "Uncategorized", breakdown[0].Label)
}
