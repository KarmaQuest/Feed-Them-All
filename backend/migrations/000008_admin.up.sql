-- Migration 000008 — Admin Dashboard support
--
-- Ajoute deux éléments nécessaires au dashboard admin :
--   1. users.is_banned  — permettre à un admin de bannir un compte sans le supprimer
--   2. level_thresholds — paliers XP configurables depuis le dashboard
--                         (remplace les valeurs hardcodées dans users/service.go)
--   3. Étend le CHECK constraint users.role pour inclure 'admin'

-- 1. Étendre le CHECK constraint role pour inclure 'admin'
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check
    CHECK (role IN ('feeder', 'giver', 'association', 'admin'));

-- 2. Ajout de la colonne is_banned sur users
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_banned BOOLEAN NOT NULL DEFAULT FALSE;

-- 3. Table des paliers de level (configurable depuis le dashboard admin)
CREATE TABLE IF NOT EXISTS level_thresholds (
    level  INT PRIMARY KEY,          -- numéro du level (1, 2, 3...)
    min_xp INT NOT NULL              -- XP minimum requis pour atteindre ce level
);

-- Seed initial : mêmes valeurs que les paliers hardcodés dans users/service.go
INSERT INTO level_thresholds (level, min_xp) VALUES
    (1, 0),
    (2, 100),
    (3, 250),
    (4, 500),
    (5, 900),
    (6, 1400),
    (7, 2100),
    (8, 3000),
    (9, 4500),
    (10, 7000)
ON CONFLICT (level) DO NOTHING;


-- 2. Table des paliers de level (configurable depuis le dashboard admin)
CREATE TABLE IF NOT EXISTS level_thresholds (
    level  INT PRIMARY KEY,          -- numéro du level (1, 2, 3...)
    min_xp INT NOT NULL              -- XP minimum requis pour atteindre ce level
);

-- Seed initial : mêmes valeurs que les paliers hardcodés dans users/service.go
INSERT INTO level_thresholds (level, min_xp) VALUES
    (1, 0),
    (2, 100),
    (3, 250),
    (4, 500),
    (5, 900),
    (6, 1400),
    (7, 2100),
    (8, 3000),
    (9, 4500),
    (10, 7000)
ON CONFLICT (level) DO NOTHING;
