-- Add usage columns to sessions table
ALTER TABLE sessions ADD COLUMN input_tokens INTEGER DEFAULT 0;
ALTER TABLE sessions ADD COLUMN output_tokens INTEGER DEFAULT 0;
ALTER TABLE sessions ADD COLUMN total_cost_usd REAL DEFAULT 0;
