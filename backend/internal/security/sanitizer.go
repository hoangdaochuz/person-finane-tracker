package security

import (
	"regexp"
	"strings"
	"unicode"

	"gorm.io/gorm"
)

// Sanitizer provides input sanitization functions
type Sanitizer struct{}

// NewSanitizer creates a new sanitizer instance
func NewSanitizer() *Sanitizer {
	return &Sanitizer{}
}

// SanitizeString removes potentially dangerous characters from strings
// This is a defense-in-depth measure alongside parameterized queries
func (s *Sanitizer) SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newline, tab, carriage return
	var result strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) {
			// Allow \n, \t, \r
			if r != '\n' && r != '\t' && r != '\r' {
				continue
			}
		}
		result.WriteRune(r)
	}

	return result.String()
}

// SanitizeSQLIdentifier safely escapes SQL identifiers (table names, column names)
// Note: This should rarely be needed with proper parameterized queries
func (s *Sanitizer) SanitizeSQLIdentifier(identifier string) string {
	// PostgreSQL identifier rules:
	// - Must start with letter or underscore
	// - Can contain letters, numbers, underscores
	// - Wrap in quotes if it contains special characters

	// Use regex to validate
	validIdentifier := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if validIdentifier.MatchString(identifier) {
		return identifier
	}

	// If invalid, return a safe default or quote it
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

// ValidateSQLInput checks for common SQL injection patterns
// Returns true if the input appears safe, false otherwise
func (s *Sanitizer) ValidateSQLInput(input string) bool {
	// Common SQL injection patterns to detect
	sqlInjectionPatterns := []string{
		`(?i)\b(SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|CREATE|EXEC|UNION|SCRIPT)\b`,
		`(?i)\b(OR|AND)\s+\d+\s*=\s*\d+`,              // "OR 1=1" type attacks
		`(?i)\-\-.*--`,                                  // SQL comments
		`(?i)[;'"\\]`,                                  // Quotes and backslashes
		`(?i)\bEXEC\b\s*\(|\bXP_CMDSHELL\b`,        // Command execution
		`(?i)\bWAITFOR\b\s+\bDELAY\b`,                // Delay attacks
		`(?i)\bCAST\b`,                                // CAST operations
		`\$\{[^}]*\}`,                                 // Shell variable expansion ${VAR}
		"`[^`]*`",                                      // Shell command substitution `cmd`
	}

	for _, pattern := range sqlInjectionPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return false
		}
	}

	return true
}

// TruncateString safely truncates a string to max length
func (s *Sanitizer) TruncateString(input string, maxLength int) string {
	runes := []rune(input)
	if len(runes) > maxLength {
		return string(runes[:maxLength])
	}
	return input
}

// SafeWhereClause builds a safe WHERE clause with proper escaping
// This provides an additional layer of security on top of GORM's parameterization
func (s *Sanitizer) SafeWhereClause(db *gorm.DB, field string, operator string, value interface{}) *gorm.DB {
	// Only allow safe operators
	safeOperators := map[string]bool{
		"=":  true,
		"!=": true,
		">":  true,
		"<":  true,
		">=": true,
		"<=": true,
		"LIKE": true,
		"IN":  true,
		"NOT IN": true,
	}

	op := strings.ToUpper(operator)
	if !safeOperators[op] {
		// Default to equals if operator is unsafe
		op = "="
	}

	return db.Where(field+" "+op+" ?", value)
}

// ValidatePeriod validates the period parameter for trends queries
func (s *Sanitizer) ValidatePeriod(period string) string {
	// Whitelist approach for period validation
	validPeriods := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	if validPeriods[period] {
		return period
	}

	// Default to daily if invalid
	return "daily"
}

// ValidateTransactionType validates transaction type
func (s *Sanitizer) ValidateTransactionType(txType string) bool {
	validTypes := map[string]bool{
		"in":  true,
		"out": true,
	}

	return validTypes[txType]
}

// ValidateCategory validates category against whitelist
func (s *Sanitizer) ValidateCategory(category string) bool {
	if category == "" {
		return true // Empty category is allowed
	}

	validCategories := map[string]bool{
		"Food":           true,
		"Transportation": true,
		"Housing":        true,
		"Utilities":      true,
		"Entertainment":  true,
		"Healthcare":     true,
		"Shopping":       true,
		"Education":      true,
		"Salary":         true,
		"Investment":     true,
		"Transfer":       true,
		"Other":          true,
	}

	return validCategories[category]
}

// CleanInput performs comprehensive cleaning of user input
func (s *Sanitizer) CleanInput(input string, maxLength int) string {
	// Step 1: Truncate to max length
	input = s.TruncateString(input, maxLength)

	// Step 2: Remove null bytes and dangerous control characters
	input = s.SanitizeString(input)

	// Step 3: Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// ValidateAmount validates that amount is positive and within reasonable bounds
func (s *Sanitizer) ValidateAmount(amount float64) bool {
	// Amount must be positive
	if amount <= 0 {
		return false
	}

	// Reasonable upper bound (999 billion)
	const maxAmount = 999999999999.99
	if amount > maxAmount {
		return false
	}

	return true
}

// EscapeLikeForPattern escapes special characters used in LIKE patterns
// This prevents % and _ wildcards from being injected
func (s *Sanitizer) EscapeLikePattern(pattern string) string {
	// Escape % and _ which are wildcards in SQL LIKE
	pattern = strings.ReplaceAll(pattern, "%", "\\%")
	pattern = strings.ReplaceAll(pattern, "_", "\\_")
	return pattern
}
