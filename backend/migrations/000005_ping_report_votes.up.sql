-- Migration 000005 — ping_report_votes
-- Crée la table des votes sur les signalements (like / dislike).
-- Un utilisateur ne peut voter qu'une fois par signalement (contrainte unique report_id + user_id).
-- Accessible à tout utilisateur authentifié, y compris le créateur du ping ou du report.

CREATE TABLE ping_report_votes (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_id  UUID        NOT NULL REFERENCES ping_reports(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value      VARCHAR(4)  NOT NULL CHECK (value IN ('up', 'down')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_ping_report_votes_user UNIQUE (report_id, user_id)
);

CREATE INDEX idx_ping_report_votes_report_id ON ping_report_votes (report_id);
