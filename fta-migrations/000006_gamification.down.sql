-- =============================================================
-- Migration 000006_gamification.down.sql
-- =============================================================
DELETE FROM badges WHERE slug IN (
  'first_signal','feeder_5','feeder_25','feeder_100',
  'photographer','confirmer_10',
  'xp_100','xp_500','xp_1000','xp_5000'
);

DROP TABLE IF EXISTS xp_log;
