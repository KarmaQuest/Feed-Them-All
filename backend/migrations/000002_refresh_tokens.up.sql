-- ============================================================
-- 000002_refresh_tokens.up.sql
-- Stores hashed refresh tokens for JWT rotation
-- ============================================================

CREATE TABLE refresh_tokens (
  user_id    UUID        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
