-- =============================================================
-- Migration 000007_avatar_shop.up.sql
-- =============================================================
-- Ajoute le système de boutique / inventaire avatar :
--
--   avatar_items      → catalogue des items (skins, tenues, accessoires)
--                       price_cents = 0  → item gratuit, débloqué par quête
--                       price_cents > 0  → item payant, acheté via Stripe
--                       unlock_condition → JSONB, même format que badges.condition
--                         null  = gratuit sans condition (item de base)
--                         {"type":"xp_threshold","value":500}
--                         {"type":"action_count","action":"feed","value":10}
--
--   user_avatar_items → inventaire de chaque utilisateur
--                       source = 'quest'    → débloqué automatiquement par la gamification
--                       source = 'purchase' → acheté via Stripe (webhook payment_intent.succeeded)
--                       source = 'default'  → item de base attribué à l'inscription
--
--   shop_orders       → trace de chaque payment_intent Stripe (pour le webhook idempotent)
-- =============================================================

-- ============================================================
-- avatar_items — catalogue
-- ============================================================
CREATE TABLE avatar_items (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  slug             VARCHAR(60) NOT NULL UNIQUE,
  name             TEXT        NOT NULL,
  category         VARCHAR(15) NOT NULL CHECK (category IN ('skin', 'outfit', 'accessory')),
  price_cents      INTEGER     NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
  currency         VARCHAR(3)  NOT NULL DEFAULT 'usd',
  unlock_condition JSONB,                  -- null = always free (base item or quest with no specific condition)
  is_active        BOOLEAN     NOT NULL DEFAULT TRUE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_avatar_items_category ON avatar_items(category) WHERE is_active = TRUE;

-- ============================================================
-- user_avatar_items — inventaire
-- ============================================================
CREATE TABLE user_avatar_items (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id     UUID        NOT NULL REFERENCES avatar_items(id) ON DELETE CASCADE,
  acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  source      VARCHAR(10) NOT NULL CHECK (source IN ('default', 'quest', 'purchase')),
  UNIQUE (user_id, item_id)
);

CREATE INDEX idx_user_avatar_items_user ON user_avatar_items(user_id);

-- ============================================================
-- shop_orders — traçabilité Stripe (idempotence webhook)
-- ============================================================
CREATE TABLE shop_orders (
  id                 UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id            UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  item_id            UUID        NOT NULL REFERENCES avatar_items(id),
  stripe_payment_intent_id TEXT NOT NULL UNIQUE,
  amount_cents       INTEGER     NOT NULL,
  currency           VARCHAR(3)  NOT NULL DEFAULT 'usd',
  status             VARCHAR(20) NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending', 'succeeded', 'failed')),
  created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shop_orders_user ON shop_orders(user_id);

-- ============================================================
-- Seed : items de base (default — attribués à tous)
-- ============================================================
INSERT INTO avatar_items (slug, name, category, price_cents, unlock_condition) VALUES
  ('skin_default',      'Default Skin',       'skin',      0, NULL),
  ('outfit_default',    'Basic Outfit',       'outfit',    0, NULL),
  ('accessory_none',    'No Accessory',       'accessory', 0, NULL);

-- ============================================================
-- Seed : items débloqués par quête (gratuits, avec condition)
-- ============================================================
INSERT INTO avatar_items (slug, name, category, price_cents, unlock_condition) VALUES
  ('outfit_feeder',     'Feeder Jacket',      'outfit',    0,
   '{"type":"action_count","action":"feed","value":5}'),

  ('outfit_champion',   'Champion Hoodie',    'outfit',    0,
   '{"type":"xp_threshold","value":500}'),

  ('accessory_bag',     'Food Bag',           'accessory', 0,
   '{"type":"action_count","action":"feed","value":10}'),

  ('skin_golden',       'Golden Skin',        'skin',      0,
   '{"type":"xp_threshold","value":1000}');

-- ============================================================
-- Seed : items payants (boutique)
-- ============================================================
INSERT INTO avatar_items (slug, name, category, price_cents, unlock_condition) VALUES
  ('skin_night',        'Night Rider Skin',   'skin',      499, NULL),
  ('outfit_ninja',      'Ninja Outfit',       'outfit',    699, NULL);
