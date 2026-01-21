# SQL Injection Protection

This document explains the SQL injection protection mechanisms in the Finance Tracker backend.

## Defense Layers

### Layer 1: GORM Parameterized Queries (Primary Protection)

GORM automatically uses **parameterized queries** which prevents SQL injection:

```go
// This becomes: WHERE type = $1 with the value passed as a parameter
query.Where("type = ?", userInput)
```

**Why this works**: The database treats parameters as data, not executable code. Even if `userInput` contains `"; DROP TABLE transactions; --`, it will be treated as a literal string value, not SQL commands.

### Layer 2: Input Sanitization (Defense-in-Depth)

The `security.Sanitizer` provides additional protection:

| Function | Purpose |
|----------|---------|
| `CleanInput()` | Truncates to max length, removes control characters |
| `ValidatePeriod()` | Whitelist validation for period parameter |
| `ValidateTransactionType()` | Whitelist validation for transaction type |
| `ValidateCategory()` | Whitelist validation for categories |
| `ValidateSQLInput()` | Detects common SQL injection patterns |

### Layer 3: Domain Layer Validation

The domain layer (`domain/transaction.go`) validates:
- Maximum amounts
- Maximum field lengths
- Date format (RFC3339)
- Category whitelist

## Protected Query Examples

### Example 1: Parameterized WHERE Clause
```go
// Before (VULNERABLE - not used):
// query := "WHERE source = '" + userInput + "'"

// After (SAFE - what we use):
query.Where("source = ?", userInput)
// Becomes: WHERE source = $1 with parameter binding
```

### Example 2: Hardcoded SQL Templates
```go
// SAFE - No user input in the SQL string
switch period {
case "daily":
    query = `SELECT ... FROM transactions GROUP BY date`  // Hardcoded
case "weekly":
    query = `SELECT ... FROM transactions GROUP BY date`  // Hardcoded
}
```

## How to Test SQL Injection Protection

Run the security tests:
```bash
go test ./test/security/... -v -run TestSQL
```

## Security Best Practices Applied

| Practice | Implementation |
|----------|----------------|
| **Parameterized Queries** | GORM's `?` placeholder (lines 71, 77, 83) |
| **Input Whitelisting** | ValidateTransactionType, ValidateCategory (lines 70, 81) |
| **Input Sanitization** | CleanInput with max length (lines 76, 82) |
| **Hardcoded SQL** | GetTrends uses switch with templates (lines 161-215) |
| **Type Safety** | Go structs prevent raw SQL string manipulation |

## Why the Security Tests Pass

The tests in `test/security/auth_test.go` pass because:

1. **Malicious source field** (`Bank'; DROP TABLE transactions; --`):
   - Gets sanitized by `CleanInput()`
   - Still parameterized via `?` placeholder
   - Treated as literal string in database

2. **Invalid category** (`Food' OR '1'='1`):
   - Fails `ValidateCategory()` whitelist check
   - Never reaches the database query
   - Returns 400 Bad Request

3. **Oversized input** (10000 characters):
   - Truncated to `MaxSourceLength` (100 chars)
   - Truncated to `MaxDescriptionLength` (1000 chars)

## Additional Protections

### XSS Protection
- JSON encoding prevents script injection
- Go templates automatically escape HTML

### DoS Protection
- Max page size: 100 records
- Max amount: 999,999,999.99
- Request body size limit: 10 MB

### Authentication
- API key required for webhook endpoints
- All other endpoints are read-only (safe for public access)
