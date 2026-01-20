# Comprehensive Software Testing Plan

## Overview

This document outlines a comprehensive testing strategy for the Personal Finance Tracker backend, covering all major testing types in software engineering.

**Project**: Personal Finance Tracker Backend
**Language**: Golang
**Framework**: Gin
**Last Updated**: 2026-01-16

---

## Testing Pyramid

```
                    E2E Tests
                   /         \
                  /           \
                /               \
             Integration Tests
              /                   \
             /                     \
           /                         \
      Unit Tests                   Component Tests
```

- **Unit Tests**: Many, fast, isolated (70% of tests)
- **Integration Tests**: Moderate number, medium speed (20% of tests)
- **E2E Tests**: Few, slow, realistic scenarios (10% of tests)

---

## 1. Unit Tests

### Purpose
Test individual functions, methods, or components in isolation.

### Current Status
- ✅ Service layer tests (21 tests)
- ✅ Middleware tests (13 tests)
- ❌ Repository layer tests
- ❌ Handler tests
- ❌ Domain/Entity tests

### Implementation Plan

#### 1.1 Repository Layer Tests

**Target**: Test database operations with mock or test database

**Files to Create**:
- `internal/repository/transaction_test.go`

**Test Cases**:
```go
- Create() - successful insert
- Create() - database error handling
- CreateInBatch() - batch insert
- FindByID() - found
- FindByID() - not found
- List() - with filters
- List() - pagination
- GetSummary() - aggregation
- GetTrends() - date grouping
- GetBreakdownBySource() - category breakdown
- GetBreakdownByCategory() - source breakdown
```

**Tools**:
- `github.com/DATA-DOG/go-sqlmock` for SQL mocking
- OR: `testcontainers-go` for real PostgreSQL in tests

**Example Structure**:
```go
func TestTransactionRepository_Create_Success(t *testing.T) {
    // Setup test database with sqlmock
    db, mock, err := sqlmock.New()
    // Test insert logic
    // Verify expectations
}
```

#### 1.2 Handler Tests

**Target**: Test HTTP endpoints without real network

**Files to Create**:
- `internal/handler/webhook_test.go`
- `internal/handler/analytics_test.go`

**Test Cases**:
```go
webhook_test.go:
- CreateTransaction - success
- CreateTransaction - invalid JSON
- CreateTransaction - validation error
- CreateTransaction - unauthorized (missing API key)
- CreateTransaction - wrong API key
- CreateBatchTransaction - success
- CreateBatchTransaction - empty batch
- CreateBatchTransaction - validation error

analytics_test.go:
- GetSummary - success
- GetSummary - database error
- GetTrends - success
- GetTrends - invalid period
- ListTransactions - pagination
- GetTransactionByID - found
- GetTransactionByID - not found
```

**Tools**:
- `net/http/httptest` for HTTP testing
- Mock repository for service layer

**Example Structure**:
```go
func TestWebhookHandler_CreateTransaction_Success(t *testing.T) {
    // Setup mock service
    // Create httptest recorder
    // Create request with valid body
    // Call handler
    // Assert status code and response
}
```

#### 1.3 Domain/Entity Tests

**Target**: Test domain logic, validation, transformations

**Files to Create**:
- `internal/domain/transaction_test.go`

**Test Cases**:
```go
- ToTransaction() - valid input
- ToTransaction() - invalid date
- Validate() - valid category
- Validate() - invalid category
- Validate() - future date rejected
- Validate() - amount exceeds maximum
- ValidationError.Error() - message format
- BatchTransactionRequest.Validate() - batch validation
```

---

## 2. Integration Tests

### Purpose
Test how multiple components work together.

### Implementation Plan

#### 2.1 Database Integration Tests

**Target**: Test real database operations

**Files to Create**:
- `test/integration/database_test.go`

**Test Cases**:
```go
- Complete transaction flow (create → read → update → delete)
- Transaction with relations (if added later)
- Migration application and rollback
- Connection pool behavior
- Transaction isolation levels
```

**Tools**:
- `testcontainers-go` for real PostgreSQL in Docker
- Database migrations before each test

**Example Structure**:
```go
func TestIntegration_TransactionFlow(t *testing.T) {
    // Start PostgreSQL container
    // Run migrations
    // Create repository
    // Perform full CRUD operations
    // Verify database state
    // Cleanup
}
```

#### 2.2 API Integration Tests

**Target**: Test HTTP handlers with real database

**Files to Create**:
- `test/integration/api_test.go`

**Test Cases**:
```go
- POST /webhook/transaction → GET /transactions/:id (verify persistence)
- Batch create → List transactions (verify all stored)
- Analytics endpoints after data insertion
- Error responses with database failures
```

**Example Structure**:
```go
func TestAPI_Integration_CreateAndRetrieve(t *testing.T) {
    // Setup test server with real DB
    // POST transaction
    // GET transaction
    // Verify data matches
}
```

---

## 3. Component Tests

### Purpose
Test complete components or modules with external dependencies mocked.

### Implementation Plan

#### 3.1 Service Component Tests

**Target**: Test complete service behavior with repository mocked

**Current Status**: ✅ Already covered by unit tests

#### 3.2 Handler Component Tests

**Target**: Test HTTP layer with service mocked

**Current Status**: ❌ Need to implement

---

## 4. End-to-End (E2E) Tests

### Purpose
Test the entire system from HTTP request to database response.

### Implementation Plan

#### 4.1 API E2E Tests

**Target**: Test complete request/response cycles

**Files to Create**:
- `test/e2e/api_test.go`

**Test Cases**:
```go
- Complete webhook flow with authentication
- Create transaction → Query via analytics
- Batch creation → Verify summary
- Error scenarios (invalid data, auth failures)
- Pagination with large datasets
```

**Tools**:
- `testcontainers-go` for PostgreSQL
- Real HTTP server (`httptest.Server` or real Gin instance)
- Environment-specific test configuration

**Example Structure**:
```go
func TestE2E_WebhookToAnalytics(t *testing.T) {
    // Start PostgreSQL container
    // Start HTTP server
    // POST transaction to webhook
    // GET summary from analytics
    // Verify data flows correctly
}
```

---

## 5. Contract Tests

### Purpose
Verify API contracts (OpenAPI spec) match implementation.

### Implementation Plan

#### 5.1 API Contract Tests

**Target**: Ensure responses match OpenAPI specification

**Files to Create**:
- `test/contract/api_test.go`

**Test Cases**:
```go
- Verify response schema matches spec
- Verify required fields present
- Verify data types correct
- Verify status codes match spec
```

**Tools**:
- OpenAPI validation against actual responses
- `github.com/stretchr/testify` for assertions

---

## 6. Performance Tests

### Purpose
Measure system performance under load.

### Implementation Plan

#### 6.1 Load Tests

**Target**: Test system under concurrent load

**Files to Create**:
- `test/performance/load_test.go`

**Test Scenarios**:
```go
- 100 concurrent webhook requests
- 1000 transaction insertions
- Analytics query performance
- Batch insertion performance (1000 transactions)
```

**Tools**:
- `github.com/rakyll/go-testbenchmark` or `vegeta`
- Go's built-in `testing.B` for benchmarks

**Targets**:
- Webhook: < 100ms p95 latency
- Analytics: < 200ms p95 latency
- Batch (100 items): < 500ms

**Example**:
```go
func BenchmarkCreateTransaction(b *testing.B) {
    // Setup DB connection
    // Run b.N times
    // Report allocations and timing
}
```

#### 6.2 Stress Tests

**Target**: Find breaking points

**Files to Create**:
- `test/performance/stress_test.go`

**Test Scenarios**:
```go
- Gradual load increase until failure
- Memory leak detection
- Connection pool exhaustion
- Database connection limits
```

---

## 7. Security Tests

### Purpose
Identify security vulnerabilities.

### Implementation Plan

#### 7.1 Authentication Tests

**Target**: Verify API key enforcement

**Files to Create**:
- `test/security/auth_test.go`

**Test Cases**:
```go
- Request without API key → 401
- Request with wrong API key → 401
- Request with expired API key (if implemented)
- SQL injection attempts
- XSS in description field
```

#### 7.2 Input Validation Tests

**Target**: Ensure malicious inputs rejected

**Test Cases**:
```go
- Negative amounts
- Future dates
- Invalid categories
- Oversized strings (buffer overflow prevention)
- Special characters in fields
```

---

## 8. Property-Based Tests

### Purpose
Find edge cases through random input generation.

### Implementation Plan

#### 8.1 Property Tests

**Target**: Validate invariants across many random inputs

**Files to Create**:
- `test/property/transaction_test.go`

**Properties to Test**:
```go
- For any transaction: amount > 0
- For batch: size matches input size
- For pagination: page_size ≤ returned count
- For dates: transaction_date ≤ created_at
```

**Tools**:
- `github.com/stretchr/testify/assert`
- Custom random generators

**Example**:
```go
func TestProperty_AmountAlwaysPositive(t *testing.T) {
    for i := 0; i < 1000; i++ {
        tx := generateRandomTransaction()
        assert.True(t, tx.Amount > 0)
    }
}
```

---

## 9. Mutation Tests

### Purpose
Ensure test suite quality by verifying tests fail when code breaks.

### Implementation Plan

#### 9.1 Mutation Testing

**Target**: Verify tests catch intentional bugs

**Tools**:
- `github.com/agudeloa/go-mutation` or `grep -r` manual checks

**Process**:
```bash
# Install gremlins
go install github.com/agudeloa/go-mutation/cmd/gremlins@latest

# Run mutation tests
gremlins ./...

# If tests still pass after mutation, tests are incomplete
```

---

## 10. Chaos Engineering

### Purpose
Test system resilience under failure conditions.

### Implementation Plan

#### 10.1 Failure Injection Tests

**Target**: Verify graceful degradation

**Files to Create**:
- `test/chaos/failure_test.go`

**Test Scenarios**:
```go
- Database connection drops mid-request
- Database slow queries (timeouts)
- Out of memory conditions
- Network delays
- Disk space full
```

**Tools**:
- Chaos Monkey for service termination
- Custom failure injection

---

## Test Organization

### Directory Structure

```
backend/
├── internal/
│   ├── domain/
│   │   └── transaction_test.go          # Unit: Domain
│   ├── service/
│   │   └── transaction_test.go          # Unit: Service
│   ├── repository/
│   │   └── transaction_test.go          # Unit: Repository (mock)
│   ├── handler/
│   │   ├── webhook_test.go              # Unit: Handler
│   │   └── analytics_test.go            # Unit: Handler
│   └── middleware/
│       └── apikey_test.go                # Unit: Middleware
├── test/
│   ├── integration/
│   │   ├── database_test.go             # Integration: DB
│   │   └── api_test.go                  # Integration: API
│   ├── e2e/
│   │   └── workflow_test.go             # E2E: Full system
│   ├── contract/
│   │   └── api_spec_test.go             # Contract: OpenAPI validation
│   ├── performance/
│   │   ├── load_test.go                 # Performance: Load
│   │   ├── benchmark_test.go            # Performance: Benchmark
│   │   └── stress_test.go               # Performance: Stress
│   ├── security/
│   │   ├── auth_test.go                 # Security: Auth
│   │   └── injection_test.go            # Security: Input validation
│   ├── property/
│   │   └── transaction_test.go          # Property-based
│   └── chaos/
│       └── failure_test.go              # Chaos: Failure injection
└── testdata/
    ├── fixtures/                         # Test data fixtures
    ├── requests/                        # Sample request bodies
    └── responses/                       # Expected responses
```

---

## Test Configuration

### Test Database

**Option 1: SQLite in-memory**
```go
// Fast, but different from production DB
db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
```

**Option 2: PostgreSQL with testcontainers**
```go
// Real PostgreSQL, slower but accurate
postgres, err := testcontainers.GenericContainer(ctx, ...)
```

**Option 3: SQL Mocking**
```go
// No database needed, fast but brittle
db, mock, err := sqlmock.New()
```

### Test Coverage Goals

| Layer | Target Coverage |
|-------|---------------|
| Domain | 95%+ |
| Service | 90%+ |
| Repository | 85%+ |
| Handler | 80%+ |
| Middleware | 85%+ |
| **Overall** | **80%+** |

---

## CI/CD Integration

### GitHub Actions Updates

**`.github/workflows/test.yaml`**:
```yaml
name: Test

on: [push, pull_request]

jobs:
  unit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - name: Unit tests
        run: go test ./internal/... -coverprofile=coverage.out
      - name: Coverage
        run: go tool cover -func=coverage.out

  integration:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
    steps:
      - name: Integration tests
        run: go test ./test/integration/...

  e2e:
    runs-on: ubuntu-latest
    steps:
      - name: E2E tests
        run: go test ./test/e2e/...

  performance:
    runs-on: ubuntu-latest
    steps:
      - name: Benchmarks
        run: go test -bench=. -benchmem
```

---

## Implementation Phases

### Phase 1: Foundation (1-2 days)
- Repository layer tests (mock or testcontainers)
- Domain/Entity tests
- Test infrastructure setup

### Phase 2: Component Tests (2-3 days)
- Handler tests (webhook, analytics)
- Complete service layer coverage
- Middleware edge cases

### Phase 3: Integration (2-3 days)
- Database integration tests
- API integration tests
- Migration testing

### Phase 4: Quality & Performance (2-3 days)
- E2E workflow tests
- Performance benchmarks
- Security tests
- Load tests

### Phase 5: Advanced (Optional)
- Property-based tests
- Mutation testing
- Chaos engineering

**Total Estimated Time**: 9-14 days

---

## Test Utilities

### Helper Functions to Create

**`test/util/testutil.go`**:
```go
package util

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *gorm.DB

// CleanupTestDB drops test tables
func CleanupTestDB(t *testing.T, db *gorm.DB)

// CreateTestTransaction creates a valid test transaction
func CreateTestTransaction() *domain.Transaction

// AssertHTTPError asserts HTTP error response
func AssertHTTPError(t *testing.T, w *httptest.ResponseRecorder, code int)

// AssertJSONBody asserts JSON response body
func AssertJSONBody(t *testing.T, body string, expected interface{})
```

---

## Test Data Management

### Fixtures

**`testdata/fixtures/transactions.json`**:
```json
[
  {
    "amount": 100.50,
    "type": "out",
    "category": "Food",
    "description": "Lunch",
    "source": "Bank ABC",
    "transaction_date": "2026-01-15T12:00:00Z"
  }
]
```

### Golden Files

**`testdata/responses/summary.json`**:
```json
{
  "total_income": 1000.00,
  "total_expense": 500.00,
  "current_balance": 500.00,
  "transaction_count": 10
}
```

---

## Best Practices

### Unit Test Guidelines
1. **Fast**: Each test should run in < 10ms
2. **Isolated**: No external dependencies
3. **Deterministic**: Same input = same output
4. **Readable**: Test name describes what is being tested
5. **One assertion per test**: Prefer multiple small tests over one large test

### Integration Test Guidelines
1. **Realistic**: Use real components (database, HTTP)
2. **Clean**: Reset state between tests
3. **Idempotent**: Can run multiple times safely

### E2E Test Guidelines
1. **User-focused**: Test user workflows
2. **Minimal**: Only test critical paths
3. **Stable**: Avoid flaky tests

---

## Common Test Patterns

### Table-Driven Tests

```go
func TestValidateCategory(t *testing.T) {
    tests := []struct {
        name    string
        category string
        wantErr bool
    }{
        {"Valid category", "Food", false},
        {"Valid empty", "", false},
        {"Invalid", "Invalid", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### Setup/Teardown Pattern

```go
func TestSuite(t *testing.T) {
    // Setup (run once)
    db := SetupTestDB(t)
    defer CleanupTestDB(t, db)

    t.Run("Test1", func(t *testing.T) {
        // Test1
    })

    t.Run("Test2", func(t *testing.T) {
        // Test2
    })
}
```

---

## Monitoring Test Quality

### Metrics to Track

1. **Coverage**: Percentage of code exercised by tests
2. **Pass Rate**: Percentage of tests passing
3. **Duration**: How long tests take to run
4. **Flakiness**: How often tests fail intermittently

### Coverage Command

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage by function
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

---

## Next Steps

1. Start with **Repository Layer Tests** (highest value)
2. Add **Handler Tests** (exposes integration issues)
3. Set up **Test Infrastructure** (testcontainers, fixtures)
4. Implement **Integration Tests** (validates data flow)
5. Add **Performance Benchmarks** (ensures scalability)
6. Implement **E2E Tests** (validates user journeys)
