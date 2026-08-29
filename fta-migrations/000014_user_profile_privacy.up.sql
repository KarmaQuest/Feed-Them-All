-- Migration 000014 : ajoute is_private sur les utilisateurs.
--
-- is_private : si true, le profil public ne retourne que username + level.
-- DEFAULT false : tous les profils existants restent publics.

ALTER TABLE users
  ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT FALSE;
