-- Add model column to sessions (default to haiku)
ALTER TABLE sessions ADD COLUMN model TEXT DEFAULT 'haiku';
