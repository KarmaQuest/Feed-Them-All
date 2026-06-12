# Références Design — FeedThemAll

## Direction Artistique Cible
Style **Pixel Art top-down**, inspiré des jeux Pokémon FireRed / LeafGreen (GBA).
Rendu : carte 2D vue de dessus, personnages en sprites animés, interface rétro-cute.

---

## Références Visuelles

### 1. Vue Carte Top-Down (Pokémon FireRed/LeafGreen)
> Référence principale pour le rendu de la carte et le placement du sprite joueur dessus.

![Pokémon FireRed Map](https://www.wikihow.com/images/thumb/1/1e/Get-All-the-HMs-in-Pok%C3%A9mon-FireRed-and-LeafGreen-Step-1-Version-3.jpg/aid1394150-v4-728px-Get-All-the-HMs-in-Pok%C3%A9mon-FireRed-and-LeafGreen-Step-1-Version-3.jpg)

**Ce qu'on retient :**
- Sprite joueur petit (~16×24px) posé sur une carte Leaflet/OSM
- Pas de rotation du sprite — 4 directions (haut/bas/gauche/droite)
- Le fond est la vraie carte OSM (pas un tileset custom)

---

### 2. Spritesheet Personnage (Référence Format)
> Référence pour le format de spritesheet du personnage customisable.

📁 Placer les spritesheets dans : `frontend-web/public/assets/sprites/characters/`

**Format cible (inspiré de la référence jointe) :**
- Grille de sprites : 3 colonnes × 3-4 lignes
- Frames par direction : 3 (idle + 2 pas)
- Directions : bas (row 0), gauche (row 1), droite (row 2), haut (row 3)
- Taille d'un sprite : **16×24 px** ou **32×48 px** (×2 pour lisibilité)
- Format export Aseprite : PNG spritesheet + JSON de mapping

**Workflow de création :**
1. Générer la base dans [PixelLab.ai](https://www.pixellab.ai/) avec la palette définie
2. Affiner et animer dans **Aseprite**
3. Exporter : `File > Export Sprite Sheet` → PNG + JSON
4. Placer dans `frontend-web/public/assets/sprites/characters/{nom_skin}.png`

---

## Palette de Couleurs (à fixer)

Utiliser la palette **[Pico-8](https://lospec.com/palette-list/pico-8)** (16 couleurs) comme base :

| Rôle | Hex |
|---|---|
| Fond sombre | `#000000` |
| Contour sprite | `#1D2B53` |
| Peau claire | `#FFCCAA` |
| Herbe / Terrain | `#00B543` |
| Eau | `#29ADFF` |
| Accent chaud | `#FF004D` |
| Accent froid | `#7E2553` |

> Importer cette palette dans Aseprite : `Palette > Load from file` et dans PixelLab.ai : option "Custom palette".

---

## Sprites à Créer (MVP)

### Personnages
- [ ] Avatar générique (1 skin, 4 directions, 3 frames) — Feeder
- [ ] Avatar Giver (tablier/toque de cuisinier)

### Animaux
- [ ] Chat errant (idle + 1 animation)
- [ ] Chien errant (idle + 1 animation)

### Marqueurs Carte
- [ ] Icône gamelle (ping nourriture disponible)
- [ ] Icône patte (ping animal signalé)
- [ ] Icône étoile (zone récemment nourrie)

### UI
- [ ] Barre XP pixel art
- [ ] Fenêtre de dialogue style RPG (pour les notifications)
