-- Migration 000011 : système de rôles multiples pour les utilisateurs.
--
-- Un utilisateur peut être Feeder, Giver, ou les deux simultanément.
-- Association est exclusif. Admin est exclusif.
--
-- roles TEXT[] : tableau des rôles actifs de l'utilisateur.
-- Le champ role (VARCHAR) est conservé comme rôle primaire pour les vérifications admin.
-- Valeurs possibles : 'feeder', 'giver', 'association', 'admin'.

ALTER TABLE users
  ADD COLUMN roles TEXT[] NOT NULL DEFAULT '{}';

-- Migrer les données existantes
UPDATE users SET roles = ARRAY[role];
