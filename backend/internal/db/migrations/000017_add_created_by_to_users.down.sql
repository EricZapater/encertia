DROP INDEX IF EXISTS idx_users_created_by;
ALTER TABLE users DROP COLUMN IF EXISTS created_by;
