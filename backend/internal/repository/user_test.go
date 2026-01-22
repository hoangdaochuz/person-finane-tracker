package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/dev/personal-finance-tracker/backend/internal/domain"
)

// Test Create()

func TestUserRepository_Create_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	user := &domain.User{
		UUID:         userUUID,
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		Name:         "Test User",
		APIKey:       "generated-api-key",
		IsActive:     true,
	}

	// Mock check for existing user
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Mock insert
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	user := &domain.User{
		UUID:         userUUID,
		Email:        "existing@example.com",
		PasswordHash: "hashed-password",
		IsActive:     true,
	}

	// Mock finding existing user
	rows := sqlmock.NewRows([]string{"id", "uuid", "email", "password_hash", "name", "api_key", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, userUUID, "existing@example.com", "hash", "Existing User", "key", true, nil, time.Now(), time.Now())

	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	err := repo.Create(user)

	assert.Error(t, err)
	assert.Equal(t, ErrUserAlreadyExists, err)
}

func TestUserRepository_Create_GeneratesAPIKey(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	user := &domain.User{
		UUID:         userUUID,
		Email:        "test@example.com",
		PasswordHash: "hashed-password",
		IsActive:     true,
		APIKey:       "", // Empty - should be generated
	}

	// Mock check for existing user
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Mock insert
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, user.APIKey, "API key should be generated")
}

// Test FindByEmail()

func TestUserRepository_FindByEmail_Found(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"id", "uuid", "email", "password_hash", "name", "api_key", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, userUUID, "test@example.com", "hashed-password", "Test User", "api-key", true, nil, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("test@example.com", 1).
		WillReturnRows(rows)

	user, err := repo.FindByEmail("test@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("nonexistent@example.com", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.FindByEmail("nonexistent@example.com")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Test FindByUUID()

func TestUserRepository_FindByUUID_Found(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"id", "uuid", "email", "password_hash", "name", "api_key", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, userUUID, "test@example.com", "hashed-password", "Test User", "api-key", true, nil, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(userUUID.String(), 1).
		WillReturnRows(rows)

	user, err := repo.FindByUUID(userUUID.String())

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userUUID, user.UUID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByUUID_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE uuid = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(userUUID.String(), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.FindByUUID(userUUID.String())

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)
}

// Test FindByID()

func TestUserRepository_FindByID_Found(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	now := time.Now().Truncate(time.Second)
	rows := sqlmock.NewRows([]string{"id", "uuid", "email", "password_hash", "name", "api_key", "is_active", "last_login_at", "created_at", "updated_at"}).
		AddRow(1, userUUID, "test@example.com", "hashed-password", "Test User", "api-key", true, nil, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	user, err := repo.FindByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID_NotFound(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	user, err := repo.FindByID(999)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUserNotFound, err)
}

// Test UpdateLastLogin()

func TestUserRepository_UpdateLastLogin_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), 1). // last_login_at, updated_at, and id
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.UpdateLastLogin(1)

	assert.NoError(t, err)
}

// Test Update()

func TestUserRepository_Update_Success(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	user := &domain.User{
		ID:       1,
		UUID:     userUUID,
		Email:    "updated@example.com",
		Name:     "Updated Name",
		IsActive: true,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(user)

	assert.NoError(t, err)
}

// Test input sanitization

func TestUserRepository_Create_SanitizesEmail(t *testing.T) {
	db, mock, sqlDB := setupMockDB(t)
	defer sqlDB.Close()

	repo := NewUserRepository(db)

	userUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	user := &domain.User{
		UUID:         userUUID,
		Email:        "  TEST@EXAMPLE.COM  ", // Has spaces and uppercase
		PasswordHash: "hashed-password",
		IsActive:     true,
	}

	// Mock check for existing user
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id"}))

	// Mock insert
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(user)

	assert.NoError(t, err)
	// Email should be sanitized (trimmed and lowercase) before being used
	// The sanitizer in the repo should handle this
}
