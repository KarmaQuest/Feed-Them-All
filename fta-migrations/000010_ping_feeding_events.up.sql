-- Migration 000010 : crée la table ping_feeding_events pour l'historique des nourrissages.
--
-- Chaque ligne représente un nourrissage effectué par un utilisateur sur un ping.
-- Plusieurs événements peuvent être liés au même ping (historique complet).
--
-- animal_count_seen : optionnel — permet de noter combien d'animaux étaient présents
--   au moment précis de ce nourrissage (peut différer du count initial du ping).

CREATE TABLE ping_feeding_events (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  ping_id           UUID        NOT NULL REFERENCES pings(id) ON DELETE CASCADE,
  fed_by            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  fed_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  note              TEXT,
  animal_count_seen INTEGER     CHECK (animal_count_seen IS NULL OR (animal_count_seen >= 1 AND animal_count_seen <= 100))
);

CREATE INDEX idx_ping_feeding_events_ping_id ON ping_feeding_events(ping_id);
CREATE INDEX idx_ping_feeding_events_fed_at  ON ping_feeding_events(fed_at DESC);

-- Ajoute un lien optionnel entre un media et l'événement de nourrissage lors duquel
-- la photo a été prise. NULL = media lié au ping globalement (sans événement spécifique).
ALTER TABLE ping_media
  ADD COLUMN feeding_event_id UUID DEFAULT NULL
    REFERENCES ping_feeding_events(id) ON DELETE SET NULL;
