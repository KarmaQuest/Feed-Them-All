ALTER TABLE ping_media
  DROP COLUMN IF EXISTS feeding_event_id;

DROP TABLE IF EXISTS ping_feeding_events;
