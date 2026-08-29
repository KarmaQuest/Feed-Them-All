-- =============================================================
-- Migration 000002_refresh_tokens.up.sql — Table des refresh tokens
-- =============================================================
-- Cette table stocke les refresh tokens des utilisateurs connectés.
--
-- Pourquoi stocker le token en base ?
--   Les JWT sont normalement "stateless" (le serveur ne stocke rien).
--   Mais pour pouvoir invalider un refresh token (au logout, ou si compromis),
--   on doit garder une trace. On ne stocke PAS le token brut (risque si fuite)
--   mais son hash SHA-256 : irréversible, inutilisable sans le token original.
--
-- Structure :
--   user_id    → clé primaire + clé étrangère vers users.id
--              (un seul refresh token actif par utilisateur — le login écrase l'ancien)
--   token_hash → SHA-256 du refresh token brut (hex, 64 caractères)
--   created_at → date de création (pour audit éventuel)
-- =============================================================

CREATE TABLE refresh_tokens (
  user_id    UUID        PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  token_hash TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
