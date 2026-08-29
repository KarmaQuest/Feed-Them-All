-- Migration 000011 : crée la table ping_comments pour les commentaires sur les pings.
--
-- Chaque ligne représente un commentaire posté par un utilisateur sur un ping.
-- content est limité à 500 caractères en base pour éviter les abus.
-- La modération admin peut éditer ou supprimer n'importe quel commentaire.

CREATE TABLE ping_comments (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  ping_id    UUID        NOT NULL REFERENCES pings(id) ON DELETE CASCADE,
  author_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  content    TEXT        NOT NULL CHECK (char_length(content) BETWEEN 1 AND 500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ping_comments_ping_id ON ping_comments(ping_id);
CREATE INDEX idx_ping_comments_created_at ON ping_comments(created_at DESC);
