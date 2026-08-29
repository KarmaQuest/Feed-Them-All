-- Migration 000015 : ajoute le champ animal_breed aux pings.
--
-- animal_breed : race spécifique de l'animal (ex: 'carlin', 'labrador', 'siamois').
--   NULL autorisé (ping legacy ou si race non précisée).
--   VARCHAR(50) pour supporter les noms de races longs.

ALTER TABLE pings
  ADD COLUMN animal_breed VARCHAR(50) DEFAULT NULL;
