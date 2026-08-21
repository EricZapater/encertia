DROP INDEX IF EXISTS idx_users_name;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_role;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('teacher', 'student'));

ALTER TABLE users DROP COLUMN IF EXISTS is_active;
