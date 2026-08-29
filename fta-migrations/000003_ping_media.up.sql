-- =============================================================
-- Migration 000003_ping_media.up.sql — Table des médias de pings
-- =============================================================
-- Cette table stocke les chemins des photos uploadées par les utilisateurs
-- comme preuves de nourrissage ou de signalement.
--
-- Pourquoi une table séparée et pas une colonne dans pings ?
--   Un ping peut avoir plusieurs photos (avant/après nourrissage, plusieurs angles).
--   Une relation N:1 (plusieurs médias → un ping) est plus flexible.
--
-- Structure :
--   id         → UUID unique du média
--   ping_id    → référence vers le ping concerné (suppression en cascade)
--   file_path  → chemin relatif dans le dossier uploads/ (ex: "abc-uuid/photo.jpg")
--   created_at → date d'upload
--
-- Le fichier physique est stocké dans le dossier uploads/ du serveur.
-- En production, ce dossier sera remplacé par un stockage S3/R2.
-- =============================================================

CREATE TABLE ping_media (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  ping_id    UUID        NOT NULL REFERENCES pings(id) ON DELETE CASCADE,
  file_path  TEXT        NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index pour récupérer rapidement tous les médias d'un ping
CREATE INDEX idx_ping_media_ping_id ON ping_media(ping_id);
