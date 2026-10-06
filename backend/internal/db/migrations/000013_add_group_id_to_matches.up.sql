-- Add group_id to matches table
ALTER TABLE matches ADD COLUMN group_id UUID REFERENCES groups(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_matches_group_id ON matches (group_id);
