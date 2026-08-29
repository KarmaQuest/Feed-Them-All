-- Migration 000009 : ajoute le type d'animal et le nombre d'animaux aux pings.
--
-- animal_type : espèce observée ('cat', 'dog', 'other').
--   NULL autorisé → obligatoire uniquement pour type='animal' (contrôlé en application).
-- animal_count : nombre d'animaux vus au moment du signalement.
--   DEFAULT 1, minimum 1, maximum 100.

ALTER TABLE pings
  ADD COLUMN animal_type  VARCHAR(10) DEFAULT NULL
                          CHECK (animal_type IN ('cat', 'dog', 'other')),
  ADD COLUMN animal_count INTEGER     NOT NULL DEFAULT 1
                          CHECK (animal_count >= 1 AND animal_count <= 100);
