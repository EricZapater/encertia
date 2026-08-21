-- Add is_active column if not exists
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- Update role check constraint to include 'admin'
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'teacher', 'student'));

-- Indexes for performance on users table
CREATE INDEX IF NOT EXISTS idx_users_role ON users (role) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users (is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_name ON users (first_name, last_name) WHERE deleted_at IS NULL;
