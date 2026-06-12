-- Migration 000004 — ping_reports
-- Crée la table des signalements de pings.
-- Un utilisateur ne peut signaler qu'une fois le même ping (contrainte unique ping_id + reported_by).
-- Accessible à tout utilisateur authentifié, y compris le créateur du ping.

CREATE TABLE ping_reports (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    ping_id     UUID        NOT NULL REFERENCES pings(id) ON DELETE CASCADE,
    reported_by UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason      VARCHAR(32) NOT NULL CHECK (reason IN ('wrong_location', 'animal_gone', 'duplicate', 'inappropriate')),
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_ping_reports_user UNIQUE (ping_id, reported_by)
);

CREATE INDEX idx_ping_reports_ping_id ON ping_reports (ping_id);
