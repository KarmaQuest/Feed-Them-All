-- Rollback migration 000010
ALTER TABLE ping_feeding_events DROP COLUMN IF EXISTS event_type;
