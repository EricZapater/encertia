-- Migration: Add language column to users table with default 'ca' and validation check constraint
ALTER TABLE users ADD COLUMN IF NOT EXISTS language VARCHAR(10) NOT NULL DEFAULT 'ca' CHECK (language IN ('ca', 'es', 'en'));

UPDATE users SET language = 'ca' WHERE language IS NULL OR language = '';
