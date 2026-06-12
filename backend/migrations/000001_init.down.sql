-- ============================================================
-- 000001_init.down.sql
-- Rollback initial schema
-- ============================================================

DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS user_badges;
DROP TABLE IF EXISTS badges;
DROP TABLE IF EXISTS xp_actions;
DROP TABLE IF EXISTS ping_animal_links;
DROP TABLE IF EXISTS animal_profiles;
DROP TABLE IF EXISTS pings;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";
DROP EXTENSION IF EXISTS postgis;
