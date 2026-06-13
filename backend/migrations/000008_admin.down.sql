-- Migration 000008 — rollback Admin Dashboard support
ALTER TABLE users DROP COLUMN IF EXISTS is_banned;
DROP TABLE IF EXISTS level_thresholds;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('feeder', 'giver', 'association'));
