-- ============================================================
-- 000001_init.up.sql
-- Initial schema for FeedThemAll
-- ============================================================

-- Extensions
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- users
-- ============================================================
CREATE TABLE users (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT        NOT NULL UNIQUE,
  password_hash TEXT        NOT NULL,
  username      TEXT        NOT NULL UNIQUE,
  role          VARCHAR(15) NOT NULL DEFAULT 'feeder'
                            CHECK (role IN ('feeder', 'giver', 'association')),
  is_premium    BOOLEAN     NOT NULL DEFAULT FALSE,
  xp            INTEGER     NOT NULL DEFAULT 0,
  avatar_config JSONB       NOT NULL DEFAULT '{}',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- pings
-- ============================================================
CREATE TABLE pings (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  type        VARCHAR(10) NOT NULL CHECK (type IN ('animal', 'food')),
  location    GEOGRAPHY(POINT, 4326) NOT NULL,  -- lon, lat WGS84
  created_by  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  is_active   BOOLEAN     NOT NULL DEFAULT TRUE,
  fed_at      TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Spatial index (P1-02)
CREATE INDEX idx_pings_location ON pings USING GIST(location);
-- Partial index: active pings only (most queries filter on this)
CREATE INDEX idx_pings_active ON pings(is_active) WHERE is_active = TRUE;

-- ============================================================
-- animal_profiles
-- A persistent profile for a stray animal, created by any user.
-- Association-validated profiles are surfaced first in the UI.
-- ============================================================
CREATE TABLE animal_profiles (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  nickname         TEXT,                              -- community-given name
  species          VARCHAR(10) NOT NULL DEFAULT 'unknown'
                               CHECK (species IN ('cat', 'dog', 'unknown')),
  description      TEXT,
  location_hint    TEXT,                              -- e.g. "near the bakery on Rue de Rivoli"
  status           VARCHAR(20) NOT NULL DEFAULT 'stray'
                               CHECK (status IN ('stray', 'adopted', 'sheltered', 'deceased')),
  -- Association ownership: NULL = community profile
  association_id   UUID        REFERENCES users(id) ON DELETE SET NULL,
  validated_by_asso BOOLEAN    NOT NULL DEFAULT FALSE,
  created_by       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index to quickly fetch association-validated profiles
CREATE INDEX idx_animal_profiles_asso ON animal_profiles(association_id)
  WHERE association_id IS NOT NULL;
CREATE INDEX idx_animal_profiles_status ON animal_profiles(status);

-- ============================================================
-- ping <-> animal_profile link (many-to-many: a ping can reference a profile)
-- ============================================================
CREATE TABLE ping_animal_links (
  ping_id    UUID NOT NULL REFERENCES pings(id) ON DELETE CASCADE,
  profile_id UUID NOT NULL REFERENCES animal_profiles(id) ON DELETE CASCADE,
  linked_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (ping_id, profile_id)
);

-- ============================================================
-- xp_actions
-- ============================================================
CREATE TABLE xp_actions (
  action      VARCHAR(50) PRIMARY KEY,
  xp_value    INTEGER     NOT NULL,
  daily_limit INTEGER     NOT NULL DEFAULT 10
);

INSERT INTO xp_actions (action, xp_value, daily_limit) VALUES
  ('signal_animal',    10, 20),
  ('feed',             25, 10),
  ('upload_photo',     15, 15),
  ('confirm_presence',  5, 30),
  ('create_profile',   20, 5);

-- ============================================================
-- badges
-- ============================================================
CREATE TABLE badges (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  slug        VARCHAR(50) NOT NULL UNIQUE,
  label       TEXT        NOT NULL,
  description TEXT,
  condition   JSONB       NOT NULL  -- e.g. {"type":"xp_threshold","value":100}
);

CREATE TABLE user_badges (
  user_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  badge_id  UUID        NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
  earned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, badge_id)
);

-- ============================================================
-- subscriptions (Premium + one-shot donations)
-- ============================================================
CREATE TABLE subscriptions (
  id                     UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id                UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stripe_customer_id     TEXT        NOT NULL,
  stripe_subscription_id TEXT,                    -- NULL for one-shot donations
  plan                   VARCHAR(20) NOT NULL CHECK (plan IN ('premium_5', 'premium_10', 'custom', 'donation')),
  amount_cents           INTEGER     NOT NULL,     -- in USD cents
  currency               VARCHAR(3)  NOT NULL DEFAULT 'usd',
  status                 VARCHAR(20) NOT NULL DEFAULT 'active'
                                     CHECK (status IN ('active', 'canceled', 'past_due', 'unpaid')),
  created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
