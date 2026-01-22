package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// PasswordHasher handles password hashing using Argon2id (OWASP recommended)
type PasswordHasher struct {
	// Argon2id parameters (OWASP recommended values for 2024)
	time    uint32 // Iterations
	memory  uint32 // Memory in KiB
	threads uint8  // Parallelism
	keyLen  uint32 // Key length (derived key length)
	saltLen uint32 // Salt length
}

// NewPasswordHasher creates a new password hasher with OWASP recommended parameters
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		time:    3,      // Iterations: OWASP recommends at least 3
		memory:  64 * 1024, // 64 MiB (65536 KiB) - OWASP recommends 64 MiB
		threads: 4,      // Parallelism: number of available CPU cores
		keyLen:  32,     // 256-bit key
		saltLen: 16,     // 128-bit salt
	}
}

// Hash creates an Argon2id hash of the password
// Format: $argon2id$v=19$m=65536,t=3,p=4$<base64salt>$<base64hash>
func (p *PasswordHasher) Hash(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, p.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive the key using Argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.time,
		p.memory,
		p.threads,
		p.keyLen,
	)

	// Encode to the standard PHC format (Password Hashing Competition)
	// Format: $argon2id$v=19$m=65536,t=3,p=4$<base64salt>$<base64hash>
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.time,
		p.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encodedHash, nil
}

// Verify verifies a password against a hash
func (p *PasswordHasher) Verify(password, hash string) (bool, error) {
	// Parse the hash
	params, salt, hashBytes, err := p.parseHash(hash)
	if err != nil {
		return false, err
	}

	// Derive the key from the password using the same parameters
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.time,
		params.memory,
		params.threads,
		params.keyLen,
	)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(hashBytes, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

// hashParams holds the parsed parameters from a hash string
type hashParams struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
}

// parseHash parses an Argon2id hash string and extracts parameters, salt, and hash
func (p *PasswordHasher) parseHash(encodedHash string) (hashParams, []byte, []byte, error) {
	params := hashParams{
		keyLen: p.keyLen, // Default keyLen from our hasher
	}

	// Split the hash into its components
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return params, nil, nil, fmt.Errorf("invalid hash format")
	}

	// Verify algorithm (parts[1] because hash starts with $)
	if parts[1] != "argon2id" {
		return params, nil, nil, fmt.Errorf("algorithm not supported: %s", parts[1])
	}

	// Parse version (we don't use it but verify format)
	// v=19

	// Parse parameters
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.memory, &params.time, &params.threads)
	if err != nil {
		return params, nil, nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params, nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	// Decode hash
	hashBytes, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params, nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return params, salt, hashBytes, nil
}

// MustHash is a helper that hashes a password and panics on error
// Use this only for hardcoded values (like test data)
func (p *PasswordHasher) MustHash(password string) string {
	hash, err := p.Hash(password)
	if err != nil {
		panic(fmt.Sprintf("failed to hash password: %v", err))
	}
	return hash
}
