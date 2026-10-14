-- Add created_by column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id) ON DELETE SET NULL;

-- Backfill existing users: assign created_by to the first admin user
UPDATE users
SET created_by = (
    SELECT id FROM users WHERE role = 'admin' ORDER BY created_at ASC LIMIT 1
)
WHERE created_by IS NULL;

-- Create index on created_by for fast scoping queries
CREATE INDEX IF NOT EXISTS idx_users_created_by ON users(created_by);
