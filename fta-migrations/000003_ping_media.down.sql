-- =============================================================
-- Migration 000003_ping_media.down.sql — Suppression table ping_media
-- =============================================================
-- Annule la migration 000003 en supprimant la table ping_media.
-- Note : les fichiers physiques dans uploads/ ne sont PAS supprimés par cette migration.
-- =============================================================

DROP TABLE IF EXISTS ping_media;
