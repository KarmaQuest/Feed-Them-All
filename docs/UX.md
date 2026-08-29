# UX Briefs — FeedThemAll

> Décisions UX validées. Chaque brief = une fonctionnalité précise.

---

## Signalement Animal — Popup centrée

**Validé le :** 2026-07-11

### Comportement

1. Clic "Signaler un animal" dans la sidebar → **ferme la sidebar** → ouvre une popup centrée (85% de la page)
2. Popup contient tout le formulaire actuel (type animal/nourriture, espèce, nombre, position GPS/carte, submit)
3. Bouton "Choisir sur la carte" → **popup se ferme** → mode pick carte → clic sur la carte → **popup se rouvre** avec les coordonnées
4. Clic en dehors de la popup **ou** bouton Annuler → ferme la popup
5. Validation → popup fermée → ping créé → sidebar rouverte avec le détail du ping
6. Hauteur : scrollable (la popup suit le contenu)

### Layout

- Largeur : 85% de la fenêtre (max 900px)
- Hauteur : auto avec scroll vertical si nécessaire
- Overlay semi-transparent derrière

### Sélecteur de sprite animal

- **Upload :** `default/animals/dogs/` et `default/animals/cats/` (depuis l'onglet Sprites admin)
- **Affichage :** grille `auto-fill` avec thumbnails 64×64
- **Interaction :** clic sur un sprite → sélectionné (highlight border)
- **Search bar :** en haut de la grille, placeholder "Rechercher une race..."
- **État vide :** "Aucun sprite animal disponible" si aucun sprite uploadé
- **Valeur envoyée :** le nom du dossier/breed (ex: `carlin`, `labrador`) → stocké dans `animal_breed`
