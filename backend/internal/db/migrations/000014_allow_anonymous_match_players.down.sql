-- Revert anonymous players changes
ALTER TABLE match_players DROP COLUMN IF EXISTS player_token;
ALTER TABLE match_players ALTER COLUMN user_id SET NOT NULL;
