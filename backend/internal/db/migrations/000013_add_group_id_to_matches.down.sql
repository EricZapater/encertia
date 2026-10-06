-- Remove group_id from matches table
DROP INDEX IF EXISTS idx_matches_group_id;
ALTER TABLE matches DROP COLUMN IF EXISTS group_id;
