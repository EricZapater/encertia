-- Allow anonymous (unregistered) players in live matches
ALTER TABLE match_players ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE match_players ADD COLUMN IF NOT EXISTS player_token VARCHAR(255);
