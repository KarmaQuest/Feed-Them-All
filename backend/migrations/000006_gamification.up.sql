-- =============================================================
-- Migration 000006_gamification.up.sql
-- =============================================================
-- Complète le système de gamification :
--   xp_log      → journal de chaque gain XP (anti-triche : comptage daily_limit)
--   Seeding badges → badges de départ prêts à l'emploi
--
-- Les tables xp_actions et badges sont déjà créées dans 000001_init.
-- Ici on ajoute le log et on insère les définitions de badges.
-- =============================================================

-- ============================================================
-- xp_log
-- Enregistre chaque attribution d'XP.
-- Utilisé pour deux choses :
--   1. Limiter la fréquence anti-triche : COUNT par (user_id, action, date)
--   2. Vérifier les conditions de badges de type "action_count"
-- ============================================================
CREATE TABLE xp_log (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action     VARCHAR(50) NOT NULL REFERENCES xp_actions(action),
  xp_earned  INTEGER     NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index sur (user_id, created_at) pour le comptage daily_limit
CREATE INDEX idx_xp_log_user_date   ON xp_log(user_id, created_at);
-- Index sur (user_id, action) pour le comptage total par action (condition badges)
CREATE INDEX idx_xp_log_user_action ON xp_log(user_id, action);

-- ============================================================
-- Seed : badges par défaut
-- condition JSONB supporte deux types :
--   {"type":"xp_threshold","value":N}           → users.xp >= N
--   {"type":"action_count","action":"X","value":N} → xp_log COUNT(action=X) >= N
-- ============================================================
INSERT INTO badges (slug, label, description, condition) VALUES
  ('first_signal',   'Premier pas',      'Signaler un animal pour la première fois',
   '{"type":"action_count","action":"signal_animal","value":1}'),

  ('feeder_5',       'Nourrisseur',      'Nourrir 5 animaux',
   '{"type":"action_count","action":"feed","value":5}'),

  ('feeder_25',      'Pro du gamelle',   'Nourrir 25 animaux',
   '{"type":"action_count","action":"feed","value":25}'),

  ('feeder_100',     'Légendaire',       'Nourrir 100 animaux',
   '{"type":"action_count","action":"feed","value":100}'),

  ('photographer',   'Photographe',      'Uploader 10 photos de preuve',
   '{"type":"action_count","action":"upload_photo","value":10}'),

  ('confirmer_10',   'Temoin fiable',    'Confirmer 10 presences animaux',
   '{"type":"action_count","action":"confirm_presence","value":10}'),

  ('xp_100',         'En route !',       'Atteindre 100 XP',
   '{"type":"xp_threshold","value":100}'),

  ('xp_500',         'Investi',          'Atteindre 500 XP',
   '{"type":"xp_threshold","value":500}'),

  ('xp_1000',        'Champion',         'Atteindre 1000 XP',
   '{"type":"xp_threshold","value":1000}'),

  ('xp_5000',        'Héros de la rue',  'Atteindre 5000 XP',
   '{"type":"xp_threshold","value":5000}');
