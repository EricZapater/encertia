-- Migration: Remove language column from users table
ALTER TABLE users DROP COLUMN IF EXISTS language;
