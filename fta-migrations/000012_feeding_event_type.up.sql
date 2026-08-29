-- Migration 000010 : ajoute event_type sur ping_feeding_events.
--
-- event_type distingue l'événement de création du ping ('signal') des nourrissages normaux ('feeding').
-- DEFAULT 'feeding' pour ne pas casser les lignes existantes.
-- Le backend insère automatiquement un événement 'signal' lors de la création d'un ping.

ALTER TABLE ping_feeding_events
  ADD COLUMN event_type VARCHAR(10) NOT NULL DEFAULT 'feeding'
             CHECK (event_type IN ('signal', 'feeding'));
