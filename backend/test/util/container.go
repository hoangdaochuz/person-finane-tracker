package util

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresTestContainer manages a PostgreSQL testcontainer
type PostgresTestContainer struct {
	Container *postgres.PostgresContainer
	ctx       context.Context
	database  string
}

// postgresConfig holds the configuration for PostgreSQL container
type postgresConfig struct {
	image      string
	database   string
	username   string
	password   string
	timeout    time.Duration
	occurrence int
}

// DefaultPostgresConfig returns the default configuration for PostgreSQL container
func DefaultPostgresConfig() postgresConfig {
	return postgresConfig{
		image:      "postgres:16-alpine",
		database:   "testdb",
		username:   "test",
		password:   "test",
		timeout:    5 * time.Second,
		occurrence: 2,
	}
}

// E2EPostgresConfig returns the configuration for E2E tests PostgreSQL container
func E2EPostgresConfig() postgresConfig {
	return postgresConfig{
		image:      "postgres:16-alpine",
		database:   "finance_tracker",
		username:   "test_user",
		password:   "test_pass",
		timeout:    10 * time.Second,
		occurrence: 2,
	}
}

// SetupPostgresContainer creates and starts a PostgreSQL testcontainer
func SetupPostgresContainer(t *testing.T, config postgresConfig) *PostgresTestContainer {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		config.image,
		postgres.WithDatabase(config.database),
		postgres.WithUsername(config.username),
		postgres.WithPassword(config.password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(config.occurrence).
				WithStartupTimeout(config.timeout)),
	)
	require.NoError(t, err, "failed to start postgres container")

	// Cleanup on test completion
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	return &PostgresTestContainer{
		Container: pgContainer,
		ctx:       ctx,
		database:  config.database,
	}
}

// SetupPostgresContainerWithDefaults creates a PostgreSQL container with default config
func SetupPostgresContainerWithDefaults(t *testing.T) *PostgresTestContainer {
	t.Helper()
	return SetupPostgresContainer(t, DefaultPostgresConfig())
}

// SetupPostgresContainerForE2E creates a PostgreSQL container for E2E tests
func SetupPostgresContainerForE2E(t *testing.T) *PostgresTestContainer {
	t.Helper()
	return SetupPostgresContainer(t, E2EPostgresConfig())
}

// ConnectionString returns the connection string for the container
func (pc *PostgresTestContainer) ConnectionString(t *testing.T) string {
	t.Helper()
	connStr, err := pc.Container.ConnectionString(pc.ctx, "sslmode=disable")
	require.NoError(t, err, "failed to get connection string")
	return connStr
}

// NewGORMDB creates a new GORM DB connection from the container
func (pc *PostgresTestContainer) NewGORMDB(t *testing.T) *gorm.DB {
	t.Helper()

	connStr := pc.ConnectionString(t)
	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err, "failed to create GORM connection")

	return db
}

// NewGORMDBWithAutoMigrate creates a new GORM DB connection and runs auto-migration
func (pc *PostgresTestContainer) NewGORMDBWithAutoMigrate(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()

	db := pc.NewGORMDB(t)

	if len(models) > 0 {
		err := db.AutoMigrate(models...)
		require.NoError(t, err, "failed to auto-migrate schema")
	}

	return db
}

// Terminate terminates the PostgreSQL container
func (pc *PostgresTestContainer) Terminate(t *testing.T) {
	t.Helper()

	if err := pc.Container.Terminate(pc.ctx); err != nil {
		t.Logf("failed to terminate container: %v", err)
	}
}

// SharedPostgresContainer manages a shared PostgreSQL container across multiple tests
type SharedPostgresContainer struct {
	*PostgresTestContainer
	mu          sync.Mutex
	instanceNum int
}

var (
	sharedContainer     *SharedPostgresContainer
	sharedContainerOnce sync.Once
)

// SetupSharedPostgresContainer creates a single shared PostgreSQL container for all tests in a package
// This should be called in TestMain to set up the container once for all tests
func SetupSharedPostgresContainer(t *testing.T, config postgresConfig) *SharedPostgresContainer {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		config.image,
		postgres.WithDatabase(config.database),
		postgres.WithUsername(config.username),
		postgres.WithPassword(config.password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(config.occurrence).
				WithStartupTimeout(config.timeout)),
	)
	require.NoError(t, err, "failed to start shared postgres container")

	return &SharedPostgresContainer{
		PostgresTestContainer: &PostgresTestContainer{
			Container: pgContainer,
			ctx:       ctx,
		},
	}
}

// NewUniqueDatabase creates a new unique database in the shared container
// This allows each test to have its own isolated database
func (sc *SharedPostgresContainer) NewUniqueDatabase(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()

	sc.mu.Lock()
	sc.instanceNum++
	instanceNum := sc.instanceNum
	sc.mu.Unlock()

	dbName := fmt.Sprintf("%s_%d", sc.database, instanceNum)

	// Connect to the default postgres database to create a new database
	connStr, err := sc.Container.ConnectionString(sc.ctx, "sslmode=disable&dbname=postgres")
	require.NoError(t, err, "failed to get connection string")

	db, err := gorm.Open(gormpostgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err, "failed to create GORM connection")

	// Create the database
	err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)).Error
	require.NoError(t, err, "failed to create database")

	// Close the connection to postgres database
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	// Connect to the new database
	newConnStr, err := sc.Container.ConnectionString(sc.ctx, fmt.Sprintf("sslmode=disable&dbname=%s", dbName))
	require.NoError(t, err, "failed to get connection string for new database")

	newDB, err := gorm.Open(gormpostgres.Open(newConnStr), &gorm.Config{})
	require.NoError(t, err, "failed to create GORM connection to new database")

	// Auto migrate schema if models provided
	if len(models) > 0 {
		err = newDB.AutoMigrate(models...)
		require.NoError(t, err, "failed to auto-migrate schema")
	}

	// Cleanup: drop the database after test
	t.Cleanup(func() {
		// Connect to postgres database to drop the test database
		dropConnStr, _ := sc.Container.ConnectionString(sc.ctx, "sslmode=disable&dbname=postgres")
		dropDB, _ := gorm.Open(gormpostgres.Open(dropConnStr), &gorm.Config{})
		_ = dropDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
		sqlDB, _ := dropDB.DB()
		_ = sqlDB.Close()
	})

	return newDB
}

// TerminateShared terminates the shared container
func (sc *SharedPostgresContainer) TerminateShared(t *testing.T) {
	t.Helper()
	sc.PostgresTestContainer.Terminate(t)
}
