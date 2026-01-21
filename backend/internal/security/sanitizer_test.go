package security

import (
	"testing"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
)

func TestSanitizer_CleanInput(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "normal string",
			input:    "Bank of America",
			maxLen:   50,
			expected: "Bank of America",
		},
		{
			name:     "truncates long string",
			input:    "A very long string that exceeds maximum length",
			maxLen:   20,
			expected: "A very long string t",
		},
		{
			name:     "removes null bytes",
			input:    "Valid\x00Text",
			maxLen:   50,
			expected: "ValidText",
		},
		{
			name:     "removes control characters",
			input:    "Text\x00More",
			maxLen:   50,
			expected: "TextMore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.CleanInput(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("CleanInput() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizer_ValidatePeriod(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"daily", "daily", "daily"},
		{"weekly", "weekly", "weekly"},
		{"monthly", "monthly", "monthly"},
		{"DAILY", "DAILY", "daily"},     // Invalid → default
		{"invalid", "invalid", "daily"},   // Invalid → default
		{"'; DROP TABLE", "'; DROP TABLE", "daily"}, // SQLi attempt → default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.ValidatePeriod(tt.input)
			if result != tt.expected {
				t.Errorf("ValidatePeriod(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizer_ValidateTransactionType(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"in", "in", true},
		{"out", "out", true},
		{"IN", "IN", false},           // Uppercase invalid
		{"admin", "admin", false},       // SQLi attempt
		{"' OR '1'='1", "' OR '1'='1", false}, // SQLi attempt
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		result := sanitizer.ValidateTransactionType(tt.input)
		if result != tt.expected {
			t.Errorf("ValidateTransactionType(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	})
	}
}

func TestSanitizer_ValidateCategory(t *testing.T) {
	sanitizer := NewSanitizer()

	// All valid categories should pass
	validCategories := []string{
		"Food", "Transportation", "Housing", "Utilities",
		"Entertainment", "Healthcare", "Shopping", "Education",
		"Salary", "Investment", "Transfer", "Other",
		"", // Empty is allowed
	}

	for _, category := range validCategories {
		if !sanitizer.ValidateCategory(category) {
			t.Errorf("ValidateCategory(%q) returned false, expected true", category)
		}
	}

	// Invalid categories should fail
	invalidCategories := []string{
		"Food' OR '1'='1",
		"<script>alert('xss')</script>",
		"../../etc/passwd",
	}

	for _, category := range invalidCategories {
		if sanitizer.ValidateCategory(category) {
			t.Errorf("ValidateCategory(%q) returned true, expected false", category)
		}
	}
}

func TestSanitizer_ValidateAmount(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		amount   float64
		expected bool
	}{
		{100.50, true},
		{0.01, true},
		{999999999999.99, true}, // Max valid amount
		{0, false},                   // Zero not allowed
		{-100, false},                // Negative not allowed
		{999999999999.999, false},   // Over max
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := sanitizer.ValidateAmount(tt.amount)
			if result != tt.expected {
				t.Errorf("ValidateAmount(%v) = %v, want %v", tt.amount, result, tt.expected)
			}
		})
	}
}

func TestSanitizer_ValidateSQLInput(t *testing.T) {
	sanitizer := NewSanitizer()

	// Safe inputs should pass
	safeInputs := []string{
		"Bank of America",
		"Grocery Store",
		"Transfer to John",
	}

	for _, input := range safeInputs {
		if !sanitizer.ValidateSQLInput(input) {
			t.Errorf("ValidateSQLInput(%q) returned false, expected true", input)
		}
	}

	// Dangerous patterns should fail
	dangerousInputs := []string{
		"admin'; DROP TABLE transactions; --",
		"admin' OR '1'='1",
		"'; EXEC xp_cmdshell('dir') --",
		"${HOME}",
		"`whoami`",
	}

	for _, input := range dangerousInputs {
		if sanitizer.ValidateSQLInput(input) {
			t.Errorf("ValidateSQLInput(%q) returned true, expected false", input)
		}
	}
}

func TestSanitizer_SanitizeSQLIdentifier(t *testing.T) {
	sanitizer := NewSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid identifier",
			input:    "table_name",
			expected: "table_name",
		},
		{
			name:     "valid with underscore",
			input:    "my_table_123",
			expected: "my_table_123",
		},
		{
			name:     "quoted identifier",
			input:    "table-name",
			expected: `"table-name"`,
		},
		{
			name:     "quoted with quotes",
			input:    `my"table`,
			expected: `"my""table"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.SanitizeSQLIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeSQLIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test that sanitizer integrates with domain constants
func TestSanitizer_IntegrationWithDomain(t *testing.T) {
	sanitizer := NewSanitizer()

	// Test that sanitizer respects domain max lengths
	if sanitizer.CleanInput(string(make([]byte, 200)), domain.MaxSourceLength) == string(make([]byte, 200)) {
		t.Error("Source should be truncated to MaxSourceLength")
	}

	if sanitizer.CleanInput(string(make([]byte, 1000)), domain.MaxDescriptionLength) == string(make([]byte, 1000)) {
		t.Error("Description should be truncated to MaxDescriptionLength")
	}
}
