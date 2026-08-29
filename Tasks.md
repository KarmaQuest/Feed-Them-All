# FeedThemAll — Task List

> Légende : `[ ]` À faire · `[x]` Terminé · `[~]` En cours · `[!]` Bloqué

---

## Tests validés en session (smoke tests manuels)

> Historique des validations manuelles effectuées lors des sessions de développement.
> Chaque ligne = un test exécuté et confirmé OK par le développeur.

| Date | Endpoint / Fonctionnalité | Résultat | Notes |
|---|---|---|---|
| 2026-06-13 | `POST /auth/register` | ✅ 201 | Création compte OK |
| 2026-06-13 | `POST /auth/login` | ✅ 200 | JWT retourné |
| 2026-06-13 | `POST /auth/refresh` | ✅ 200 | Refresh token cookie HttpOnly |
| 2026-06-13 | `POST /auth/logout` | ✅ 204 | Cookie supprimé |
| 2026-06-13 | `POST /pings` | ✅ 201 | Ping créé avec coords |
| 2026-06-13 | `GET /pings?lat=&lon=&radius=` | ✅ 200 | ST_DWithin retourne les pings proches |
| 2026-06-13 | `PATCH /pings/:id/confirm` | ✅ 204 | updated_at mis à jour |
| 2026-06-13 | `PATCH /pings/:id/fed` | ✅ 204 | fed_at enregistré |
| 2026-06-13 | `DELETE /pings/:id` | ✅ 204 | Soft delete is_active=false |
| 2026-06-13 | `POST /pings/:id/media` | ✅ 201 | Upload JPEG/PNG, chemin retourné |
| 2026-06-13 | `GET /pings/:id/media` | ✅ 200 | Liste des médias |
| 2026-06-13 | `GET /uploads/*` | ✅ 200 | Fichier servi statiquement |
| 2026-06-13 | `POST /pings/:id/report` | ✅ 201 / 409 doublon | Signalement créé, contrainte unique OK |
| 2026-06-13 | `GET /pings/:id/reports` | ✅ 200 | Liste avec scores up/down |
| 2026-06-13 | `POST /pings/:id/reports/:id/vote` | ✅ 204 | Vote up↔down changeable (upsert) |
| 2026-06-13 | Tests unitaires `auth` (20/20) | ✅ PASS | `go test ./internal/auth/...` |
| 2026-06-13 | Tests unitaires `pings` (37/37) | ✅ PASS | `go test ./internal/pings/...` |
| 2026-06-13 | `GET /shop/items` | ✅ 200 | 9 items (3 default, 4 quête, 2 payants) |
| 2026-06-13 | `GET /users/me/inventory` | ✅ 200 | Inventaire vide sur nouveau compte |
| 2026-06-13 | `POST /shop/items/:id/purchase` (payant) | ✅ 200 | `client_secret` Stripe retourné (`pi_3Thk...`) |
| 2026-06-13 | Stripe webhook `listen` | ✅ Connecté | `localhost:8080/shop/webhook` visible dans le dashboard Stripe |
| 2026-06-13 | `GET /admin/users` (no JWT) | ✅ 401 | `RequireAdmin` bloque sans token |
| 2026-06-13 | `GET /admin/users` (non-admin JWT) | ✅ 403 | `RequireAdmin` bloque le rôle `feeder` |
| 2026-06-13 | `GET /admin/users` (admin JWT) | ✅ 200 | 5 utilisateurs retournés |
| 2026-06-13 | `GET /admin/xp-actions` | ✅ 200 | 5 actions retournées |
| 2026-06-13 | `GET /admin/level-thresholds` | ✅ 200 | 10 paliers retournés |
| 2026-06-13 | `GET /admin/badges` | ✅ 200 | 10 badges retournés |
| 2026-06-13 | `GET /admin/shop-items` | ✅ 200 | 9 items retournés |
| 2026-06-13 | `GET /admin/pings` | ✅ 200 | 2 pings retournés avec report_count |
| 2026-06-13 | `PATCH /admin/users/:id` (is_banned) | ✅ 204 | Bannissement appliqué |
| 2026-06-13 | `PUT /admin/xp-actions/feed` | ✅ 204 | xp_value modifié |
| 2026-06-13 | `POST /admin/badges` | ✅ 201 | Badge créé, UUID retourné |
| 2026-06-13 | `DELETE /admin/badges/:id` | ✅ 204 | Badge supprimé |
| 2026-06-13 | `POST /pings` avec `animal_type=cat` + `animal_count=3` | ✅ 201 | Champs retournés OK |
| 2026-06-13 | `POST /pings/:id/feedings` (note + animal_count_seen) | ✅ 201 | FeedingEvent créé, fed_at mis à jour, XP feed accordé |
| 2026-06-13 | `GET /pings/:id/feedings` | ✅ 200 | Historique retourné (1 événement) |
| 2026-06-13 | Tests unitaires `pings` après refactor (57/57) | ✅ PASS | `go test ./internal/pings/...` |
| 2026-06-28 | `GET /sprites/default/characters/male/south.png` 64×64 | ✅ 200 | Sprite 64×64 servi |
| 2026-06-28 | `GET /admin/sprites` (arborescence) | ✅ JSON | Tous les dossiers/fichiers listés |
| 2026-06-28 | `POST /admin/sprites/upload` PNG valide | ✅ 201 | Upload réussi |
| 2026-06-28 | `POST /admin/sprites/upload` fichier non-PNG | ✅ 400 | Rejeté |
| 2026-06-28 | `POST /admin/sprites/upload` path traversal | ✅ 400 | Rejeté |
| 2026-06-28 | `DELETE /admin/sprites?path=...` | ✅ 204 | Supprimé |
| 2026-06-28 | `POST /admin/shop-items/{id}/sprite` → `south.png` | ✅ 201 | Fichier nommé correctement |
| 2026-06-28 | `POST /admin/sprites/upload` avec `filename=south.png` | ✅ 201 | Override de nom fonctionnel |

---

## Phase 0 — Setup & Infrastructure Locale

- [x] **P0-01** Initialiser le monorepo Git et pousser la structure sur GitHub (`Feed-Them-All`)
- [x] **P0-02** Créer `docker-compose.yml` avec PostgreSQL 15 + extension PostGIS
- [x] **P0-03** Vérifier que PostgreSQL démarre localement et que PostGIS est actif
- [x] **P0-04** Initialiser le module Go (`go mod init github.com/KarmaQuest/feed-them-all`)
- [x] **P0-05** Initialiser le projet React Web (`Vite + TypeScript`) dans `frontend-web/`
- [x] **P0-06** Configurer ESLint + Prettier dans `frontend-web/`
- [x] **P0-07** Créer le fichier `.env.example` à la racine du backend
- [x] **P0-08** Créer le fichier `.gitignore` (exclure `.env`, `uploads/`, `node_modules/`, binaires Go)
- [x] **P0-09** Mettre en place GitHub Actions : CI lint + test au push sur `dev`

---

## Phase 1 — Backend : Base & Auth

- [x] **P1-01** Créer la migration initiale : tables `users`, `pings`, `animal_profiles`, `xp_actions`, `badges`, `user_badges`, `subscriptions`, `ping_animal_links`
- [x] **P1-02** Ajouter l'index spatial PostGIS sur `pings.location` (inclus dans `000001_init.up.sql`)
- [x] **P1-03** Mettre en place `golang-migrate` pour versionner les migrations
- [x] **P1-04** Implémenter l'inscription (`POST /auth/register`) — hash bcrypt du mot de passe
- [x] **P1-05** Implémenter la connexion (`POST /auth/login`) — retourne access token JWT (15 min)
- [x] **P1-06** Implémenter le refresh token (`POST /auth/refresh`) — cookie HttpOnly 7 jours
- [x] **P1-07** Middleware d'authentification JWT pour les routes protégées
- [x] **P1-08** Rate limiting sur `/auth/register` et `/auth/login`
- [x] **P1-09** Tests unitaires : package `auth`

---

## Phase 2 — Backend : Pings & Géolocalisation

- [x] **P2-01** `GET /pings?lat=&lon=&radius=` — récupérer les pings dans un rayon (ST_DWithin)
- [x] **P2-02** `POST /pings` — créer un ping (animal ou nourriture), position GPS requise
- [x] **P2-03** `PATCH /pings/:id/confirm` — confirmer la présence d'un animal (avec photo optionnelle)
- [x] **P2-04** `PATCH /pings/:id/fed` — marquer un animal comme nourri
- [x] **P2-05** `DELETE /pings/:id` — désactiver un ping (soft delete `is_active = false`)
- [x] **P2-06** Upload de photo de preuve (`POST /pings/:id/media`) — validation MIME, 10 Mo max, stockage `uploads/`
- [x] **P2-07** Servir les fichiers `uploads/` via un handler Go statique
- [x] **P2-08** Signalement et votes sur les pings
  - `POST /pings/:id/report` — signaler un problème sur un ping
    - Accessible à **tout utilisateur authentifié, y compris le créateur du ping**
    - Champs : `reason` (enum : `wrong_location` \| `animal_gone` \| `duplicate` \| `inappropriate`) + `comment` (optionnel)
    - Un utilisateur ne peut signaler qu'une fois le même ping (contrainte DB unique sur `ping_id + reported_by`)
  - `POST /pings/:id/reports/:report_id/vote` — voter pour ou contre un signalement
    - Accessible à **tout utilisateur authentifié, y compris le créateur du ping ou du report**
    - Champs : `value` (enum : `up` \| `down`)
    - Un utilisateur ne peut voter qu'une fois par signalement (contrainte unique `report_id + user_id`)
    - Les votes prouvent la véracité du ping : un taux élevé de `up` sur un report `animal_gone` peut marquer le ping comme douteux sur la carte
  - Migrations :
    - `000004_ping_reports` : table `ping_reports` (id, ping_id FK, reported_by FK, reason, comment, created_at)
    - `000005_ping_report_votes` : table `ping_report_votes` (id, report_id FK cascade, user_id FK cascade, value `'up'|'down'`, created_at, UNIQUE(report_id, user_id))
- [x] **P2-09** Tests : package `pings`

---

## Phase 3 — Backend : WebSocket & Temps Réel

- [x] **P3-01** Mettre en place le serveur WebSocket (`gorilla/websocket`)
- [x] **P3-02** Système d'abonnement par bounding box (le client envoie sa zone visible)
- [x] **P3-03** Broadcast d'un nouveau ping aux clients abonnés à la zone correspondante
- [x] **P3-04** Broadcast de la position des Feeders actifs (avatar en temps réel sur la carte)
- [x] **P3-05** Gestion des déconnexions et nettoyage des abonnements
- [x] **P3-06** Limiter la fréquence de mise à jour de position GPS (max 1 push/seconde par client)

---

## Phase 4 — Backend : Gamification

- [x] **P4-01** Créer la table `xp_actions` avec les valeurs par défaut (signaler, nourrir, uploader, confirmer)
- [x] **P4-02** Fonction `award_xp(user_id, action)` — calcul côté serveur, rate limiting anti-triche
- [x] **P4-03** Appeler `award_xp` automatiquement après chaque action éligible
- [x] **P4-04** Créer la table `badges` et les règles de déverrouillage
- [x] **P4-05** Job asynchrone (goroutine) pour vérifier et déverrouiller les badges
- [x] **P4-06** `GET /users/:id/profile` — retourne XP, niveau, badges, avatar config
- [x] **P4-07** `GET /leaderboard` — top 20 utilisateurs par XP (avec cache en mémoire, TTL 5 min)
- [x] **P4-08** Système d'inventaire avatars
  - Migration : table `avatar_items` (id, slug, name, category `skin|outfit|accessory`, price_cents, unlock_condition)
  - Migration : table `user_avatar_items` (user_id FK, item_id FK, acquired_at, source `quest|purchase`)
  - `GET /users/me/inventory` — liste les items possédés par l'utilisateur connecté
  - `POST /shop/purchase` — achat one-shot Stripe d'un item → `payment_intent` → webhook → crédit inventaire
  - Déverrouillage automatique via quête : appelé depuis `award_xp()` quand un seuil de quête est atteint

---

## Phase 4-bis — Backend : Dashboard Admin

> Objectif : permettre la gestion des variables métier (XP, levels, badges, boutique, users) sans toucher au code.
> Toutes les routes `/admin/*` sont protégées par le rôle `admin` (middleware dédié).
> Aucune route publique ou Feeder n'est impactée.

- [x] **PA-01** Migration : table `level_thresholds` (level INT, min_xp INT) + colonne `users.is_banned BOOLEAN`
  - Initialisation avec les paliers actuels `[0,100,250,500,900,1400,2100,3000,4500,7000]`
  - `users.service.go` charge les paliers depuis la DB au démarrage (avec fallback hardcodé si table vide)
- [x] **PA-02** Package `admin/` — middleware `RequireAdmin` (vérifie `users.role = 'admin'`)
- [x] **PA-03** Dashboard utilisateurs
  - `GET /admin/users?page=&search=` — liste paginée avec filtre nom/email
  - `PATCH /admin/users/:id` — modifier role (`feeder`/`giver`/`association`/`admin`) ou `is_banned`
- [x] **PA-04** Dashboard XP & levels
  - `GET /admin/xp-actions` — liste toutes les actions avec xp_value + daily_limit
  - `PUT /admin/xp-actions/:action` — modifier xp_value ou daily_limit
  - `GET /admin/level-thresholds` — liste les paliers de level
  - `PUT /admin/level-thresholds` — remplacer tous les paliers (tableau JSON)
- [x] **PA-05** Dashboard badges
  - `GET /admin/badges` — liste tous les badges
  - `POST /admin/badges` — créer un nouveau badge
  - `PUT /admin/badges/:id` — modifier label, description ou condition
  - `DELETE /admin/badges/:id` — supprimer un badge
- [x] **PA-06** Dashboard boutique skins
  - `GET /admin/shop-items` — liste tous les items avatar (skin/outfit/accessory)
  - `POST /admin/shop-items` — ajouter un nouvel item (slug, name, category, price_cents, unlock_condition)
  - `PUT /admin/shop-items/:id` — modifier un item existant
  - `DELETE /admin/shop-items/:id` — retirer un item de la boutique
- [x] **PA-07** Dashboard modération pings
  - `GET /admin/pings?active=true&flagged=true` — liste des pings avec nombre de reports
  - `DELETE /admin/pings/:id` — désactivation forcée par un admin (sans vérification owner)
- [x] **PA-08** Frontend admin (`/admin`) — React, protégé par rôle
  - Sidebar avec sections : Utilisateurs · XP & Levels · Badges · Boutique · Modération
  - Chaque section = tableau éditable (inline edit) + boutons action
  - Pas de Leaflet, pas de pixel art — interface sobre et fonctionnelle

---

## Phase 5 — Frontend Web : Carte & Pings

- [x] **P5-01** Intégrer **Leaflet + React-Leaflet** avec fond de carte OpenStreetMap
- [x] **P5-02** Afficher les pings récupérés depuis l'API sous forme de marqueurs custom (SVG inline via `L.divIcon`)
- [x] **P5-03** Mise à jour des marqueurs en temps réel via WebSocket (ajout/suppression de pings)
- [x] **P5-04** Clic sur un marqueur → sidebar slideout droite avec détails du ping (photos, activités, date)
- [x] **P5-05** Bouton "Signaler un animal" dans la sidebar → formulaire de création de ping avec GPS auto-détecté ou clic sur carte
- [x] **P5-06** Bouton "J'ai nourri" → FeedForm inline dans la sidebar + upload photo optionnel + historique mis à jour
- [x] **P5-07** Géolocalisation browser (`navigator.geolocation`) — centrer la carte sur la position utilisateur
- [x] **P5-08** Fallback si géolocalisation refusée (carte centrée sur la ville par défaut, toast d'avertissement)

---

## Phase 6 — Frontend Web : Auth & Profil

- [x] **P6-01** Page Inscription (`/register`) — formulaire + appel API
- [x] **P6-02** Page Connexion utilisateur (`/user-login`) + Connexion admin (`/login`) — formulaires + stockage token
- [x] **P6-03** Gestion du refresh token automatique (intercepteur Axios + `initialize()` au démarrage de l'app)
- [x] **P6-04** Page Profil (`/profile`) — affichage XP, badges, avatar
- [x] **P6-05** Store Zustand : état auth (user, isLogged, initialize, login, logout)

---

## Phase 7 — Frontend Web : Avatar Pixel Art

- [x] **P7-01** Créer le composant `AvatarSprite` — affiche le sprite selon la config `{gender, skin, outfit, accessory}` (fallback icône couleur en attendant les sprites)

---

## Phase 8 — CSS Design System & UI Admin

- [x] **P8-01** Créer `src/styles/tokens.css` — variables CSS globales (couleurs, spacing, radius, fonts, transitions)
- [x] **P8-02** Créer `src/styles/components.css` — classes réutilisables (`.btn`, `.card`, `.modal`, `.badge`, `.table`, `.input`)
- [x] **P8-03** Créer `src/styles/utilities.css` — classes utilitaires (`.u-text-*`, `.u-flex-*`, `.u-gap-*`)
- [x] **P8-04** Réécrire `AdminPage.css` et `LoginPage.css` avec `var()` (tokens)
- [x] **P8-05** Sidebar admin : section "Paliers de Level" séparée de "Actions XP" (`LevelsSection.tsx`)
- [x] **P8-06** Fix boutons toolbar/pagination admin (`.btn-page`)

---

## Bugs & Fixes (2026-06-14)

- [x] **FIX-01** `ProtectedRoute` redirigait immédiatement avant que `initialize()` soit résolu → boucle infinie sur `/admin`. Fix : ajout de `initialized: boolean` dans le store auth — `ProtectedRoute` affiche `null` jusqu'à ce que la session soit restaurée.
- [x] **FIX-02** Autofill navigateur (Chrome/Firefox) non capté par React `useState` → login envoyait des champs vides. Fix : lecture des valeurs depuis `e.currentTarget.elements` dans `AuthForm.handleSubmit`.
- [x] **FIX-03** `MapSidebar.tsx` corrompu (BOM UTF-8 + double-encodage) via `Set-Content` PowerShell 5.1. Fix : fichier réécrit via `create_file` (VS Code agent). **RÈGLE** : ne jamais utiliser `Set-Content` sur des fichiers source.
- [x] **FIX-04** Upload sprite avec `filename=south.png` ignoré — le Path retourné utilisait `filepath.Base(filePath)` au lieu de la variable `name`. Fix : `service.go:364` — retourner `name` au lieu de `filepath.Base(filePath)`.

---

## Sidebar Carte (UX Refactor — 2026-06-14)

- [x] **UX-01** Supprimer la topbar horizontale — remplacée par un bouton FAB flottant (logo, coin haut-droit)
- [x] **UX-02** Créer `MapSidebar.tsx` — panneau slideout droite avec animation CSS (`transform: translateX`)
- [x] **UX-03** Panneau `nav` : stats (🐾 🍖 ● Live), user badge, boutons Signaler/Admin/Déconnexion
- [x] **UX-04** Panneau `signal` : `SignalForm` inline (mode GPS ou clic sur carte)
- [x] **UX-05** Panneau `ping` : détails ping + activités (historique nourrissages) + FeedForm + Confirmer présence
- [x] **UX-06** Fix critique : `NavPanel`, `SignalPanel`, `PingPanel` extraits hors du composant parent → plus de démontage/remontage React
- [x] **UX-07** Fix `ListFeedingEvents` backend : JOIN `users` pour retourner `username` (était UUID seul)
- [x] **UX-08** Fix `FeedingEvent` frontend : champ `fed_at` (était `created_at` → `Invalid Date`)
- [x] **P7-02** Afficher l'avatar du Feeder connecté sur la carte Leaflet (marqueur `L.divIcon` avec sprite statique)
  - [x] Copier les sprites PNG dans `frontend-web/public/assets/sprites/characters/{male,female}/south.png`
  - [x] Refonte `AvatarSprite.tsx` : remplacer emoji/fond couleur par `<img>` vers le sprite PNG
  - [x] Nettoyer `AvatarSprite.css` : supprimer styles couleur/emoji, garder dimensions + borders + shadow
  - [x] `markers.ts` : ajouter `createAvatarIcon(config)` avec `L.divIcon` + `<img>`
  - [x] `MapView.tsx` : remplacer `userIcon` par `createAvatarIcon(user?.avatar_config)`
- [ ] **P7-03** Mise à jour de la position de l'avatar en temps réel (push GPS → WebSocket)
- [ ] **P7-04** Page customisation avatar (`/avatar`) — sélecteur visuel de skin/tenue/accessoire
- [ ] **P7-05** Sauvegarder la config avatar en DB via `PATCH /users/me/avatar`
- [ ] **P7-06** Afficher les avatars des autres Feeders actifs sur la carte (en semi-transparent)
- [ ] **P7-07** Animations de proximité sur la carte
  - Feeder à < 30 m d'un ping animal → jouer animation "nourrir" (frame dédiée du spritesheet)
  - Feeder à < 30 m d'un ping Giver → jouer animation "entrer dans le bâtiment" (sprite disparaît derrière le marqueur)
  - Calcul de proximité côté client (distance Haversine, aucun changement backend)
- [ ] **P7-08** Page boutique avatar (`/shop`)
  - Affiche tous les items disponibles (skins, tenues, accessoires) avec leur prix ou condition de déverrouillage
  - Deux onglets : **Quêtes** (gratuit, déblocage via XP/succès) · **Boutique** (achat one-shot Stripe)
  - Appel `POST /shop/purchase` → crédite l'item dans l'inventaire utilisateur
  - Appel `GET /users/me/inventory` → liste les items possédés

---

## SM — Sprite Management System (Dashboard Upload)

### Passe 1 — Upload statique (PNG unique)

- [x] **SM-01** Ajouter `SPRITES_DIR` à la config Go (variable d'env, default: `./sprites`)
- [x] **SM-02** Route `GET /sprites/*` dans Go — servir les fichiers statiques depuis `backend/sprites/`
- [x] **SM-03** Migrer les sprites existants de `frontend-web/public/assets/sprites/` → `backend/sprites/default/characters/{male,female}/south.png`
- [x] **SM-04** Créer dossier `backend/sprites/default/markers/` pour les futurs PNGs d'icônes
- [x] **SM-05** Routes admin : `POST /admin/sprites/upload`, `GET /admin/sprites`, `DELETE /admin/sprites/{type}/{slug}/{file}`
- [x] **SM-06** Validation upload (MIME type PNG, anti-path-traversal, taille max 5 Mo)
- [x] **SM-07** Frontend : nouvelles fonctions API `listSprites`, `deleteSprite`, `uploadShopItemSprite` dans `api/admin.ts`
- [x] **SM-08** Onglet "Sprites" minimaliste dans l'admin : liste arborescente + preview thumbnail + suppression
- [x] **SM-09** Mettre à jour `createAvatarIcon()` et `AvatarSprite.tsx` : résolution chemin `/api/sprites/...` (shop → default), sprite 64×64, suppression bord/shadow
- [x] **SM-10** Mettre à jour icônes ping (animal/food/fed) : essai PNG `/api/sprites/default/markers/{type}.png` → fallback SVG, taille 48px
- [x] **SM-11** Upload sprite intégré au formulaire Boutique : `POST /admin/shop-items/{id}/sprite` stocke dans `shop/{slug}/south.png`
- [x] **SM-11b** Upload avec override de nom de fichier (`filename=south.png`) — backend + frontend
- [x] **SM-11c** Sélecteur de sprite dans le formulaire boutique (grille thumbnails 64px depuis `shop/`)
- [x] **SM-11d** Auto-slug depuis le nom de l'item avec vérification collision + incrémentation
- [x] **SM-11e** Preview animation spritesheet hover dans l'arborescence Sprites + bouton ▶ dans le sélecteur boutique

### Passe 2 — Animations (spritesheet + CSS stepping)
>
> Progress : Preview frontend ✅ (hover dans SpritesSection, bouton ▶ dans sélecteur boutique).
> Reste à faire : upload backend + résolution + rendu carte.

- [ ] **SM-12** Support upload spritesheet PNG + JSON metadata (backend)
- [ ] **SM-13** Dashboard : champ "Ajouter une animation" par item (idle/walk/open_door)
- [ ] **SM-14** Résolution par animation : `shop/{slug}/{animation}/spritesheet.png`
- [ ] **SM-15** Marqueur carte animé : CSS `background-image` + `background-position` stepping sur `L.divIcon`
- [ ] **SM-16** AvatarSprite animé avec détection auto de la direction (south/west/east/north)

---

## Phase 8 — Design & Assets Pixel Art

- [ ] **P8-01** Fixer la palette de couleurs définitive dans Aseprite (basée Pico-8)
- [ ] **P8-02** Créer le sprite Avatar Feeder générique (4 directions × 3 frames) avec Aseprite + PixelLab.ai
- [ ] **P8-03** Créer le sprite Avatar Giver (toque/tablier) — même format
- [ ] **P8-04** Créer le sprite Chat errant (idle + animation)
- [ ] **P8-05** Créer le sprite Chien errant (idle + animation)
- [ ] **P8-06** Créer les icônes marqueurs : gamelle, patte, étoile (zone nourrie)
- [ ] **P8-07** Créer la barre XP pixel art (composant React CSS)
- [ ] **P8-08** Créer la fenêtre de notification style dialogue RPG
- [ ] **P8-09** Créer les sprites de customisation MVP : 3 tenues, 3 couleurs de cheveux, 2 accessoires (débloquables via quêtes)
- [ ] **P8-10** Créer les sprites boutique (items premium) : 2 skins exclusifs, 2 tenues achetables — visuellement distincts des items gratuits
- [ ] **P8-11** Créer les frames d'animation de proximité : "nourrir" (4 frames) · "entrer bâtiment" (3 frames fade-out)

---

## Phase 9 — Mobile (React Native)

- [ ] **P9-01** Initialiser le projet React Native (`Expo`) dans `frontend-mobile/`
- [ ] **P9-02** Partager les types TypeScript depuis `shared/types/`
- [ ] **P9-03** Partager la couche `api/` avec le frontend web
- [ ] **P9-04** Intégrer la carte (React Native Maps ou Leaflet via WebView)
- [ ] **P9-05** Géolocalisation native (Expo Location)
- [ ] **P9-06** Affichage des pings et avatars (parité avec le web)

---

---

## Audit Qualité — Architecture & Sécurité (2026-06-13)

> Résultats de l'audit pré-Phase 5. À corriger avant de commencer le frontend.
> Légende : `[ ]` À faire · `[x]` Terminé

### A — Architecture

- [x] **A1** — `frontend-web/.vite/deps/` mal ignoré par `.gitignore` (apparaissait comme untracked)
- [x] **A2** — `references/design/sprites/` contient des binaires PNG/ZIP trackés dans git → **exclus du repo** (protection des sprites contre récupération externe)
- [x] **A3** — Dossier `shared/types/` absent → créé avec `index.ts` placeholder (types réels ajoutés en Phase 9)
- [x] **A4** — Pas de package `config/` centralisé → créé `internal/config/config.go`, `main.go` migré

### S — Sécurité

- [x] **S1** — CORS restreint à `localhost:5173` + `https://feedthemall.org` (prod) (OWASP A05)
- [x] **S2** — WebSocket `CheckOrigin` corrigé vers `feedthemall.org` (OWASP A05)
- [x] **S3** — `STRIPE_SECRET_KEY` via `os.Getenv` uniquement — confirmé ok (OWASP A02)
- [x] **S4** — Validation mot de passe renforcée : 8 chars + 1 chiffre obligatoire (OWASP A07)
- [x] **S5** — Pagination capée à `page <= 100` côté serveur (OWASP A04)
- [x] **S6** — Arrondi GPS 3 décimales (~100m) via `math.Round` sur `GET /pings` et `GET /admin/pings` (OWASP A01)
- [x] **S7** — `config.Load()` refuse le démarrage en prod avec secrets insecures (OWASP A02)
- [x] **S8** — Handler statique `uploads/` via `http.StripPrefix` + `http.Dir` — confirmé ok (OWASP A01)
- [x] **S9** — Cookie refresh token avec `SameSite=Strict` — confirmé ok (OWASP A01)

---

## Backlog / Phase 2+

### Suivi des animaux
- [ ] Fiche animal créée par n'importe quel utilisateur (Feeder/Giver) depuis un ping ou manuellement
- [ ] Fiche associative mise en avant quand une association a créé/validé la fiche (badge "Suivi par [Association]", contact, actions en cours) — fiche communautaire reste visible en secondaire
- [ ] Statut d'adoption / prise en charge refuge (modifiable par tous, confirmable par une association)
- [ ] Heatmap densité animaux par zone

### Comptes Association
- [ ] Type utilisateur `association` (3ème rôle)
- [ ] Tableau de bord association (stats, zones, bénévoles actifs)
- [ ] Validation de fiches animaux par les associations
- [ ] Plateforme de dons ciblés portée par les associations partenaires (FeedThemAll = intermédiaire technique, fonds → association directement)

### Gamification avancée
- [ ] Système de quêtes (hebdomadaires, style Pokémon) — récompenses : XP, badges, **skins/tenues exclusifs** débloqués dans l'inventaire
- [ ] Guildes de quartier + classement inter-quartiers
- [ ] Saisons (remise à zéro classement tous les 3 mois + récompenses top 10)
- [ ] Titres géographiques ("Gardien du 11ème", "Légende de Belleville")

### Carte & UX
- [ ] Mode nuit/jour (carte adaptée à l'heure)
- [ ] Zones dangereuses (signalement chantier, circulation)
- [ ] Itinéraire de nourrissage optimisé

### Mobile
- [ ] Mode hors-ligne (cache local + sync)
- [ ] Notifications push (Firebase FCM)
- [ ] Widget home screen

### Infrastructure & Monétisation
- [ ] **Publicités** — stratégie validée (2026-06-13)
  - [ ] **Web** : Google AdSense — bannière fixe en bas de la page profil + bas de la page shop (jamais sur la carte)
  - [ ] **Mobile** : Google AdMob — bannière bas d'écran (hors carte) + rewarded video (regarder une pub = +50 XP)
  - [ ] Intégrer dès le MVP (pas de seuil minimum de trafic — l'app a besoin de revenus dès le départ)
  - [ ] Placement mobile : jamais d'interstitiel pendant une action (nourrir, créer ping) — uniquement entre sessions (retour accueil)
  - [ ] Premium : option "supprimer les pubs" incluse dans l'abonnement Premium
- [ ] **Analytics — Matomo self-hosted** (validé 2026-06-13)
  - [ ] Déployer Matomo via Docker sur le même VPS que le backend
  - [ ] Intégrer le script Matomo dans `frontend-web/index.html` + `frontend-mobile/`
  - [ ] Configurer les événements custom : `ping_created`, `animal_fed`, `badge_unlocked`, `shop_purchase`
  - [ ] RGPD : activer l'anonymisation IP + respecter le Do Not Track
  - [ ] Dashboard objectifs : retention J7, entonnoir création ping → nourrissage, top zones géographiques
- [ ] **Abonnement Premium volontaire** — paliers 5$/mois · 10$/mois · montant libre (Stripe Billing, récurrent)
  - [ ] Intégration Stripe (clés API, webhook secret dans `.env`)
  - [ ] Table `subscriptions` en DB (user_id, stripe_customer_id, stripe_subscription_id, plan, status, currency)
  - [ ] Webhook Stripe → activer/désactiver `is_premium` sur le compte utilisateur
  - [ ] Page paiement Web (`/support`) — 3 boutons + champ montant libre
  - [ ] Gestion multi-devises : USD par défaut, EUR/GBP/CAD phase 2 (détection IP, taux Stripe)
- [ ] **Boutique in-app skins** — achat one-shot de skins/tenues/accessoires exclusifs via Stripe `payment_intent`
  - Items créés en DB (`avatar_items`), prix fixé par l'équipe
  - Webhook Stripe → `user_avatar_items` (source = `purchase`)
  - Item aussitôt disponible dans le sélecteur avatar
  - Principe : les items boutique sont **cosmétiques uniquement** — aucun avantage gameplay
- [ ] **Dons one-shot à FeedThemAll** — `payment_intent` Stripe (pas de récurrence), même page `/support`
- [ ] **Dons ciblés associations** — `payment_intent` Stripe vers compte Stripe propre de l'association (Stripe Connect, FTA zéro commission)
- [ ] API publique données anonymisées
- [ ] Migration stockage médias vers Cloudflare R2
- [ ] Cache Redis pour leaderboard et pings chauds
- [ ] Déploiement VPS (Docker Compose en production)
- [ ] Support multi-villes / multi-pays
