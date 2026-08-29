-- Rollback migration 000011
ALTER TABLE users DROP COLUMN IF EXISTS roles;
