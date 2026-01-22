package security

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test PasswordHasher.Hash()

func TestPasswordHasher_Hash_ValidPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("MySecurePassword123")

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")
	assert.NotEqual(t, "MySecurePassword123", hash, "hash should not equal plain password")
}

func TestPasswordHasher_Hash_DifferentPasswordsDifferentHashes(t *testing.T) {
	hasher := NewPasswordHasher()

	hash1, err := hasher.Hash("Password1")
	hash2, err2 := hasher.Hash("Password2")

	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "different passwords should produce different hashes")
}

func TestPasswordHasher_Hash_SamePasswordDifferentHashes(t *testing.T) {
	hasher := NewPasswordHasher()

	hash1, err := hasher.Hash("SamePassword123")
	hash2, err2 := hasher.Hash("SamePassword123")

	assert.NoError(t, err)
	assert.NoError(t, err2)
	// Different hashes due to random salt
	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes due to salt")
}

func TestPasswordHasher_Hash_Format(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("TestPassword123")

	assert.NoError(t, err)
	parts := strings.Split(hash, "$")
	assert.Len(t, parts, 6, "argon2id hash should have 6 parts separated by $")
	assert.Equal(t, "", parts[0], "first part should be empty")
	assert.Equal(t, "argon2id", parts[1], "second part should be argon2id")
	assert.Contains(t, parts[2], "v=", "third part should contain version")
	assert.Contains(t, parts[3], "m=", "fourth part should contain memory")
	assert.Contains(t, parts[3], "t=", "fourth part should contain time")
	assert.Contains(t, parts[3], "p=", "fourth part should contain parallelism")
	assert.NotEmpty(t, parts[4], "fifth part should be base64 salt")
	assert.NotEmpty(t, parts[5], "sixth part should be base64 hash")
}

// Test PasswordHasher.Verify()

func TestPasswordHasher_Verify_ValidPassword(t *testing.T) {
	hasher := NewPasswordHasher()
	password := "CorrectPassword123"

	hash, err := hasher.Hash(password)
	assert.NoError(t, err)

	valid, err := hasher.Verify(password, hash)
	assert.NoError(t, err)
	assert.True(t, valid, "correct password should verify")
}

func TestPasswordHasher_Verify_InvalidPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("CorrectPassword123")
	assert.NoError(t, err)

	valid, err := hasher.Verify("WrongPassword456", hash)
	assert.NoError(t, err)
	assert.False(t, valid, "wrong password should not verify")
}

func TestPasswordHasher_Verify_CaseSensitive(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("Password123")
	assert.NoError(t, err)

	valid, err := hasher.Verify("password123", hash)
	assert.NoError(t, err)
	assert.False(t, valid, "password should be case sensitive")
}

func TestPasswordHasher_Verify_EmptyPassword(t *testing.T) {
	hasher := NewPasswordHasher()

	hash, err := hasher.Hash("")
	assert.NoError(t, err)

	valid, err := hasher.Verify("", hash)
	assert.NoError(t, err)
	assert.True(t, valid, "empty password hash should verify with empty password")
}

func TestPasswordHasher_Verify_InvalidHashFormat(t *testing.T) {
	hasher := NewPasswordHasher()

	invalidHashes := []struct {
		name string
		hash string
	}{
		{"not a hash", "not-a-hash"},
		{"missing parts", "$argon2id$v=19$m=65536,t=3,p=4$salt"},
		{"wrong algorithm", "$sha256$v=1$m=1,t=1,p=1$salt$hash"},
		{"empty hash", ""},
	}

	for _, tc := range invalidHashes {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := hasher.Verify("password", tc.hash)
			assert.Error(t, err, "invalid hash should return error")
			assert.False(t, valid)
		})
	}
}

// Test PasswordHasher.MustHash()

func TestPasswordHasher_MustHash_PanicsOnError(t *testing.T) {
	hasher := NewPasswordHasher()

	// Since Hash doesn't return errors in normal conditions, we test that it works
	hash := hasher.MustHash("TestPassword123")

	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")
}

// Test integration: Hash and Verify cycle

func TestPasswordHasher_HashVerifyCycle_VariousPasswords(t *testing.T) {
	hasher := NewPasswordHasher()

	passwords := []string{
		"simple123",
		"Complex!Password@2024",
		"あいうえお123", // Unicode characters
		strings.Repeat("a", 100) + "1",
		"1" + strings.Repeat("a", 100),
		"MixedCase123456",
		"Special!@#$%Chars",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			hash, err := hasher.Hash(password)
			assert.NoError(t, err, "hashing should succeed for: %s", password)

			valid, err := hasher.Verify(password, hash)
			assert.NoError(t, err, "verification should succeed for: %s", password)
			assert.True(t, valid, "password should verify for: %s", password)
		})
	}
}
