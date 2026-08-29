-- =============================================================
-- Migration 000002_refresh_tokens.down.sql — Suppression table refresh tokens
-- =============================================================
-- Annule la migration 000002 en supprimant la table refresh_tokens.
-- =============================================================

DROP TABLE IF EXISTS refresh_tokens;
