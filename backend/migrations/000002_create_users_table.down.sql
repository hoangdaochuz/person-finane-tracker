-- Rollback migration for users table
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF NOT EXISTS idx_users_api_key;
DROP TABLE IF EXISTS users;
