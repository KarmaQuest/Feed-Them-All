-- Rollback migration 000014
ALTER TABLE users DROP COLUMN IF EXISTS is_private;
