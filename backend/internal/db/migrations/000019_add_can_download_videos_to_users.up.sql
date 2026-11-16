-- Add can_download_videos column to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS can_download_videos BOOLEAN NOT NULL DEFAULT FALSE;

-- Ensure all admin users have can_download_videos set to TRUE by default
UPDATE users SET can_download_videos = TRUE WHERE role = 'admin';
