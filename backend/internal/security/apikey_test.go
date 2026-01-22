package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test APIKeyGenerator.Generate()

func TestAPIKeyGenerator_Generate_ValidKey(t *testing.T) {
	gen := NewAPIKeyGenerator()

	key, err := gen.Generate()

	assert.NoError(t, err)
	assert.NotEmpty(t, key)
	// Base64 encoding of 32 bytes produces approximately 43 characters
	assert.GreaterOrEqual(t, len(key), 40, "key should be at least 40 characters")
	assert.LessOrEqual(t, len(key), 50, "key should be at most 50 characters")
}

func TestAPIKeyGenerator_Generate_UniqueKeys(t *testing.T) {
	gen := NewAPIKeyGenerator()

	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := gen.Generate()
		assert.NoError(t, err)
		assert.False(t, keys[key], "generated key should be unique")
		keys[key] = true
	}
	assert.Len(t, keys, 100, "should generate 100 unique keys")
}

func TestAPIKeyGenerator_Generate_Base64URLSafe(t *testing.T) {
	gen := NewAPIKeyGenerator()

	key, err := gen.Generate()
	assert.NoError(t, err)

	// Base64 URL encoding uses A-Z, a-z, 0-9, -, _, and = (for padding)
	// It should NOT contain + or / which are in standard base64
	for _, c := range key {
		isValid := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '='
		assert.True(t, isValid,
			"key should only contain URL-safe base64 characters, got: %c", c)
	}
}

// Test APIKeyGenerator.Validate()

func TestAPIKeyGenerator_ValidKey_ValidFormat(t *testing.T) {
	gen := NewAPIKeyGenerator()

	// Generate a key and validate it
	key, err := gen.Generate()
	assert.NoError(t, err)

	valid := gen.Validate(key)
	assert.True(t, valid, "generated key should be valid")
}

func TestAPIKeyGenerator_Validate_EmptyKey(t *testing.T) {
	gen := NewAPIKeyGenerator()

	valid := gen.Validate("")
	assert.False(t, valid, "empty key should be invalid")
}

func TestAPIKeyGenerator_Validate_TooShort(t *testing.T) {
	gen := NewAPIKeyGenerator()

	shortKeys := []string{
		"abc",
		"short-key-123",
		"too-short-key-",
	}

	for _, key := range shortKeys {
		t.Run(key, func(t *testing.T) {
			valid := gen.Validate(key)
			assert.False(t, valid, "short key should be invalid: %s", key)
		})
	}
}

func TestAPIKeyGenerator_Validate_ValidBase64(t *testing.T) {
	gen := NewAPIKeyGenerator()

	// Use an actually generated key for this test
	key, err := gen.Generate()
	assert.NoError(t, err)

	valid := gen.Validate(key)
	assert.True(t, valid, "generated key should pass validation")
}

func TestAPIKeyGenerator_Validate_InvalidBase64(t *testing.T) {
	gen := NewAPIKeyGenerator()

	invalidBase64 := "abc123!@#$%^&*()"
	valid := gen.Validate(invalidBase64)
	assert.False(t, valid, "invalid base64 string should fail validation")
}

// Test APIKeyGenerator.SanitizeKey()

func TestAPIKeyGenerator_SanitizeKey_TrimsWhitespace(t *testing.T) {
	gen := NewAPIKeyGenerator()

	input := "  abc-123_xyz  "
	sanitized := gen.SanitizeKey(input)

	assert.Equal(t, "abc-123_xyz", sanitized)
}

func TestAPIKeyGenerator_SanitizeKey_RemovesTabs(t *testing.T) {
	gen := NewAPIKeyGenerator()

	input := "abc-123\t\t_xyz"
	sanitized := gen.SanitizeKey(input)

	assert.Equal(t, "abc-123_xyz", sanitized)
}

func TestAPIKeyGenerator_SanitizeKey_RemovesNewlines(t *testing.T) {
	gen := NewAPIKeyGenerator()

	input := "abc\n-123\n_xyz\n"
	sanitized := gen.SanitizeKey(input)

	assert.Equal(t, "abc-123_xyz", sanitized)
}

func TestAPIKeyGenerator_SanitizeKey_MultipleSpaces(t *testing.T) {
	gen := NewAPIKeyGenerator()

	input := "abc   123   xyz"
	sanitized := gen.SanitizeKey(input)

	assert.Equal(t, "abc123xyz", sanitized)
}

// Test APIKeyGenerator.GenerateWithPrefix()

func TestAPIKeyGenerator_GenerateWithPrefix_WithPrefix(t *testing.T) {
	gen := NewAPIKeyGenerator()

	key, err := gen.GenerateWithPrefix("test")

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "test_"), "key should start with prefix")
	// Remove prefix and validate the rest
	keyWithoutPrefix := strings.TrimPrefix(key, "test_")
	assert.True(t, gen.Validate(keyWithoutPrefix), "key part should be valid")
}

func TestAPIKeyGenerator_GenerateWithPrefix_EmptyPrefix(t *testing.T) {
	gen := NewAPIKeyGenerator()

	key, err := gen.GenerateWithPrefix("")

	assert.NoError(t, err)
	// No prefix means no underscore separator added
	// But the generated key might contain _ as valid base64 URL char
	assert.False(t, strings.HasPrefix(key, "_"), "key should not start with underscore")
}

func TestAPIKeyGenerator_GenerateWithPrefix_TrailingUnderscore(t *testing.T) {
	gen := NewAPIKeyGenerator()

	key, err := gen.GenerateWithPrefix("test_")

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "test_"), "key should start with prefix")
	// Should not have double underscore
	assert.False(t, strings.HasPrefix(key, "test__"), "should not add extra underscore")
}

func TestAPIKeyGenerator_GenerateWithPrefix_Unique(t *testing.T) {
	gen := NewAPIKeyGenerator()

	keys := make(map[string]bool)
	for i := 0; i < 50; i++ {
		key, err := gen.GenerateWithPrefix("app")
		assert.NoError(t, err)
		assert.False(t, keys[key], "generated key should be unique")
		keys[key] = true
	}
	assert.Len(t, keys, 50, "should generate 50 unique keys with prefix")
}

// Test APIKeyGenerator.MustGenerate()

func TestAPIKeyGenerator_MustGenerate_Works(t *testing.T) {
	gen := NewAPIKeyGenerator()

	// Should not panic
	key := gen.MustGenerate()

	assert.NotEmpty(t, key)
	assert.True(t, gen.Validate(key))
}

// Test key properties

func TestAPIKeyGenerator_NoSpecialCharsBesidesBase64(t *testing.T) {
	gen := NewAPIKeyGenerator()

	for i := 0; i < 50; i++ {
		key, err := gen.Generate()
		assert.NoError(t, err)

		for _, c := range key {
			isLetter := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
			isDigit := c >= '0' && c <= '9'
			isHyphen := c == '-'
			isUnderscore := c == '_'
			isPadding := c == '='

			assert.True(t, isLetter || isDigit || isHyphen || isUnderscore || isPadding,
				"key should only contain URL-safe base64 characters, got: %c in %s", c, key)
		}
	}
}
