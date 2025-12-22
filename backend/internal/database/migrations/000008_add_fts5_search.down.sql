-- Remove FTS5 search
DROP TRIGGER IF EXISTS messages_fts_au;
DROP TRIGGER IF EXISTS messages_fts_ad;
DROP TRIGGER IF EXISTS messages_fts_ai;
DROP TABLE IF EXISTS messages_fts;
