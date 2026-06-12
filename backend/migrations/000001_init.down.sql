-- =============================================================
-- Migration 000001_init.down.sql — Suppression du schéma initial
-- =============================================================
-- Ce fichier ANNULE la migration 000001_init.up.sql.
-- Il supprime toutes les tables dans l'ordre inverse de leur création
-- pour respecter les dépendances (clés étrangères).
--
-- ATTENTION : exécuter ce fichier supprime TOUTES LES DONNÉES.
-- Utilisé uniquement en développement pour réinitialiser la base.
-- Commande : migrate -path ./migrations -database "..." down 1
-- =============================================================

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
