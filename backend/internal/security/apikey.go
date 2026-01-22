package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// APIKeyGenerator generates cryptographically secure API keys
type APIKeyGenerator struct {
	// KeyLength is the number of random bytes to generate
	// The final base64-encoded string will be approximately 1.33x longer
	KeyLength int
}

// NewAPIKeyGenerator creates a new API key generator
func NewAPIKeyGenerator() *APIKeyGenerator {
	return &APIKeyGenerator{
		KeyLength: 32, // 32 random bytes = 256 bits of entropy
	}
}

// Generate generates a new random API key
// The key is base64-encoded for safe use in HTTP headers and JSON
func (g *APIKeyGenerator) Generate() (string, error) {
	// Generate random bytes
	bytes := make([]byte, g.KeyLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64 for easy transport
	// Using URL-safe encoding to avoid issues with + and / characters
	key := base64.URLEncoding.EncodeToString(bytes)

	return key, nil
}

// MustGenerate generates a new API key and panics on error
// Use this only for test scenarios
func (g *APIKeyGenerator) MustGenerate() string {
	key, err := g.Generate()
	if err != nil {
		panic(fmt.Sprintf("failed to generate API key: %v", err))
	}
	return key
}

// Validate validates an API key format
func (g *APIKeyGenerator) Validate(key string) bool {
	if key == "" {
		return false
	}

	// Check length - base64 encoding of 32 bytes is approximately 43-44 characters
	// Allow some flexibility for future changes
	if len(key) < 32 || len(key) > 128 {
		return false
	}

	// Verify it's valid base64
	_, err := base64.URLEncoding.DecodeString(key)
	return err == nil
}

// SanitizeKey removes any whitespace or problematic characters from an API key
func (g *APIKeyGenerator) SanitizeKey(key string) string {
	// Remove whitespace
	key = strings.TrimSpace(key)
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, "\t", "")
	key = strings.ReplaceAll(key, "\n", "")
	return key
}

// GenerateWithPrefix generates an API key with a prefix for easy identification
// Useful for distinguishing between different types of keys
func (g *APIKeyGenerator) GenerateWithPrefix(prefix string) (string, error) {
	key, err := g.Generate()
	if err != nil {
		return "", err
	}

	// Add prefix if provided
	if prefix != "" {
		// Ensure prefix doesn't have trailing underscore
		prefix = strings.TrimSuffix(prefix, "_")
		key = prefix + "_" + key
	}

	return key, nil
}
