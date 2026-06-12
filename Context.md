# FeedThemAll — Contexte Technique & Pratiques

## Stack Technique

| Couche | Technologie | Raison |
|---|---|---|
| Backend | **Go (Golang)** | Concurrence native (goroutines) pour le temps réel, faible empreinte mémoire, performances élevées |
| API | **REST + WebSockets** | REST pour les opérations CRUD, WebSockets pour les mises à jour de carte en temps réel |
| Base de données | **PostgreSQL + PostGIS** | PostGIS pour les requêtes géospatiales (pings proches, zone de rayon) |
| Web Frontend | **React** | Composants réutilisables, écosystème mature |
| Mobile | **React Native** | Partage de logique avec le Web, accès natif GPS/caméra |
| Carte | **Leaflet + OpenStreetMap** | 100% gratuit, open source, aucune clé API requise |
| Auth | **JWT (access + refresh tokens)** | Stateless, compatible Web et Mobile |
| Temps réel | **WebSockets (Go gorilla/websocket)** | Push des nouveaux pings carte sans polling |
| Stockage médias | **Système de fichiers local** | Dossier `uploads/` servi par Go en phase locale ; migration S3 prévue en phase 2 |
| CI/CD | **GitHub Actions** | Intégration native avec le repo, gratuit pour projets publics |

---

## Architecture Générale

```
[Phase 1 — 100% Local]

Client (React)
        │
        ├─ REST API  ──────────────────► Go HTTP Server (localhost:8080)
        │                                    │
        └─ WebSocket ──────────────────►     ├─ PostgreSQL + PostGIS (Docker local)
                                             └─ Dossier uploads/ local
```

> **Phase 1 : tout tourne en local sur ta machine.** PostgreSQL via Docker Desktop, serveur Go sur `localhost:8080`, frontend React sur `localhost:3000`. Aucun serveur distant requis.
> **Phase 2 (plus tard)** : déploiement sur un VPS ou service cloud une fois le MVP validé.

- **Monorepo recommandé** : `backend/`, `frontend-web/`, `frontend-mobile/`, `shared/` (types TypeScript partagés)
- Séparer les responsabilités : chaque service Go dans son propre package (`pings`, `users`, `gamification`, `auth`)

---

## Conventions de Code

### Go (Backend)
- Respecter `gofmt` et `golint` — aucun code non formaté accepté en PR
- Erreurs explicites : ne jamais ignorer une erreur (`_ = err` interdit hors cas justifié)
- Logging structuré avec **`slog`** (stdlib Go 1.21+) ou **`zerolog`**
- Les handlers HTTP sont fins — toute logique métier dans un package `service/`
- Variables d'environnement via un fichier `.env` + package `os.Getenv` ; jamais de secrets hardcodés

### TypeScript / React
- **ESLint + Prettier** configurés dans le repo, aucune exception
- Props typées (interfaces TypeScript) — pas de `any`
- Composants fonctionnels uniquement (hooks), pas de class components
- État global via **Zustand** (léger) plutôt que Redux pour ce projet
- Appels API centralisés dans un dossier `api/` (jamais de `fetch` direct dans un composant)

---

## Base de Données

### Schéma clé (PostGIS)
```sql
-- Les pings utilisent un type géographique natif
CREATE TABLE pings (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type        VARCHAR(10) NOT NULL CHECK (type IN ('animal','food')),
  location    GEOGRAPHY(POINT, 4326) NOT NULL,
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  updated_at  TIMESTAMPTZ DEFAULT NOW(),
  created_by  UUID REFERENCES users(id),
  is_active   BOOLEAN DEFAULT TRUE
);

-- Index spatial obligatoire pour les requêtes de proximité
CREATE INDEX idx_pings_location ON pings USING GIST(location);
```

- Utiliser des **UUID v4** comme clés primaires (pas d'entiers auto-incrémentés exposés en API)
- **Migrations versionnées** avec `golang-migrate` — aucune modification de schéma à la main en prod
- Requête de proximité type : `ST_DWithin(location, ST_MakePoint(lon, lat)::geography, rayon_metres)`

---

## Sécurité

- **Authentification** : JWT avec access token court (15 min) + refresh token (7 jours) en cookie HttpOnly
- **Rate limiting** sur les endpoints sensibles (création de ping, inscription) — utiliser `golang.org/x/time/rate`
- **Validation des inputs** côté serveur, toujours (même si le frontend valide déjà)
- **Uploads médias** : valider le type MIME côté serveur (ne pas se fier à l'extension), limiter la taille (ex. 10 Mo max)
- Pas de données personnelles dans les logs
- Les coordonnées GPS affichées publiquement sont **arrondies à ~100 m** pour préserver la vie privée des Feeders

---

## Temps Réel — Stratégie WebSocket

- Chaque client connecté s'abonne à une **zone géographique** (bounding box)
- Quand un nouveau ping est créé dans cette zone, le serveur broadcast via le canal correspondant
- Limiter les messages : ne pusher que les **deltas** (nouveau ping, ping désactivé), pas l'état complet de la carte
- Fallback polling (30 s) si la connexion WebSocket échoue

---

## Gamification — Principes

- Les points XP sont calculés **côté serveur uniquement** — jamais côté client
- Chaque action a un coût fixe défini dans une table `xp_actions` (modifiable sans redéploiement)
- Les badges sont déverrouillés de manière **asynchrone** (job en arrière-plan) pour ne pas bloquer la réponse API
- Anti-triche : limiter le nombre d'actions récompensées par heure/jour par utilisateur (rate limiting XP)

---

## Branching & Workflow Git

```
main          ← Production (protégée, merge via PR uniquement)
dev           ← Intégration continue
feature/*     ← Nouvelles fonctionnalités (ex. feature/ping-creation)
fix/*         ← Corrections de bugs
```

- **Conventional Commits** : `feat:`, `fix:`, `chore:`, `docs:`, `refactor:` — obligatoire pour générer un changelog
- Chaque PR doit passer les tests CI avant merge
- Pas de push direct sur `main` ni `dev`

---

## Tests

- **Backend Go** : tests unitaires avec `testing` stdlib + `testify` pour les assertions ; tests d'intégration avec une base PostgreSQL de test (Docker)
- **Frontend** : Vitest + React Testing Library pour les composants ; Playwright pour les tests E2E critiques (flux inscription, création ping)
- Couverture cible : **>70%** sur la logique métier backend

---

## Performance & Scalabilité

- Les requêtes géospatiales doivent retourner en **< 100 ms** — surveiller avec `EXPLAIN ANALYZE`
- Pagination des résultats (cursor-based plutôt qu'offset pour la carte)
- Cache Redis (optionnel phase 2) pour les pings chauds et les leaderboards XP
- Images servies depuis le dossier `uploads/` local via Go en phase 1 ; CDN en phase 2

---

## Variables d'Environnement (`.env.example` à maintenir à jour)

```
DATABASE_URL=postgres://user:pass@localhost:5432/feedthemall
JWT_SECRET=
JWT_REFRESH_SECRET=
UPLOAD_DIR=./uploads
PORT=8080
ENV=development
```

> Aucune clé API tierce nécessaire en phase locale. OpenStreetMap/Leaflet fonctionne sans clé.

---

## Design — Direction Artistique Pixel Art / Pokémon

### Concept
L'univers visuel s'inspire des jeux Pokémon (pixel art, ambiance rétro-cute) pour créer un attachement émotionnel fort avec les animaux aidés.

### Avatar Joueur Customisable
- Chaque utilisateur (Feeder ou Giver) possède un **avatar pixel art** personnalisable
- Customisation : couleur de peau, vêtements, accessoires, sac à dos (ex. sac de croquettes), couleur de cheveux
- L'avatar est **visible sur la carte** en temps réel à la position GPS du joueur (comme un sprite dans un jeu top-down)
- Techniquement : sprite PNG/GIF exporté depuis un éditeur de pixel art, affiché comme un marqueur Leaflet personnalisé (`L.divIcon` avec une image)

### Marqueurs sur la Carte
| Élément | Style visuel |
|---|---|
| Ping animal (chien/chat errant) | Sprite pixel art de l'animal correspondant |
| Ping nourriture (Giver) | Icône assiette/gamelle pixel art |
| Avatar Feeder en ligne | Sprite du personnage customisé |
| Zone nourrie récemment | Effet brillance / étoiles pixel art |

### Outils
- **[Aseprite](https://www.aseprite.org/)** ✅ (déjà utilisé) — création et animation des sprites, export spritesheet PNG
- **[PixelLab.ai](https://www.pixellab.ai/)** ✅ (déjà utilisé) — génération de sprites pixel art par IA, accélère la création des assets (animaux, décors, icônes)
- **Workflow suggéré** : générer une base avec PixelLab.ai → affiner et animer dans Aseprite → exporter en spritesheet PNG
- **Police** : [Press Start 2P](https://fonts.google.com/specimen/Press+Start+2P) (Google Fonts, gratuit) pour les titres/UI
- **Palette de couleurs** : fixer une palette commune dans Aseprite (16-32 couleurs) et l'importer dans PixelLab.ai pour garantir la cohérence visuelle — ex. palette [Pico-8](https://lospec.com/palette-list/pico-8)

### Faisabilité Technique
- ✅ **Oui, totalement faisable** avec Leaflet : les marqueurs acceptent des icônes HTML/CSS/image custom
- ✅ L'avatar en temps réel sur la carte = WebSocket qui pousse la position GPS du Feeder actif
- ✅ La customisation = un objet JSON `{skin, hair, outfit, accessory}` stocké en DB, rendu côté client
- ⚠️ Créer les sprites prend du temps — prévoir un set minimal pour le MVP (1 avatar générique, 2-3 animaux, 2 icônes)

---

## Workflow de Validation des Tâches

### Principe : Zéro tâche non validée ne débloque la suivante

Chaque tâche/sous-tâche dans `Tasks.md` suit ce cycle obligatoire avant d'être marquée `[x]` :

```
[ ] À faire
 │
 ▼
[~] En cours (travail actif)
 │
 ▼
 ⬛ Critères de validation atteints ?
 │         │
Non       Oui
 │         │
 ▼         ▼
[!]       [x] Terminé + validé
Bloqué
```

### Critères de Validation par Type de Tâche

| Type | Critère de validation |
|---|---|
| Migration DB | `golang-migrate up` passe sans erreur, tables/index visibles dans psql |
| Endpoint API | Réponse correcte testée via `curl` ou Bruno/Insomnia + test unitaire vert |
| WebSocket | Connexion établie, message reçu côté client dans la console navigateur |
| Composant React | Rendu visible dans le navigateur, pas d'erreur console, props typées |
| Sprite pixel art | Asset exporté en PNG dans le bon dossier, visible sur la carte ou dans l'UI |
| Test | `go test ./...` ou `vitest run` passe à 100% sur le package concerné |
| CI/CD | Pipeline GitHub Actions vert (badge ✅ sur le repo) |

### Règles

1. **Une seule tâche `[~]` à la fois** — finir avant de commencer la suivante
2. **Toute nouvelle tâche ou sous-tâche doit être listée dans `Tasks.md` et validée par l'utilisateur avant d'être démarrée**
3. **Pas de merge sur `dev` sans que la tâche correspondante soit `[x]`**
4. Si une tâche révèle un nouveau besoin (bug, dépendance oubliée), créer une nouvelle entrée dans `Tasks.md` et demander validation
5. Les tâches de design (P8-*) sont validées par revue visuelle explicite de l'utilisateur

### Fichiers de Suivi
- `Tasks.md` — source de vérité des tâches (statuts, IDs)
- `Context.md` — workflow, conventions, décisions (ce fichier)
- `docs/` — décisions d'architecture détaillées (ADR) si besoin

---

## Agents & Skills (GitHub Copilot)

Les agents sont des fichiers `SKILL.md` dans `.github/skills/`, chargés automatiquement par GitHub Copilot selon le contexte.

### Agents Créés

Les agents sont des fichiers `SKILL.md` dans `.github/skills/`, chargés automatiquement par GitHub Copilot selon le contexte.

| Agent | Fichier | Spécialité |
|---|---|---|
| **go-backend** | `.github/skills/go-backend/SKILL.md` | Go, handlers, middleware, WebSockets, JWT, upload |
| **react-web** | `.github/skills/react-web/SKILL.md` | React, Leaflet, Zustand, avatar sprite system |
| **react-native** | `.github/skills/react-native/SKILL.md` | Expo, GPS natif, camera, navigation, shared types |
| **postgresql-postgis** | `.github/skills/postgresql-postgis/SKILL.md` | Schema, migrations, requêtes spatiales PostGIS |
| **ci-github** | `.github/skills/ci-github/SKILL.md` | GitHub Actions, branch strategy, Conventional Commits |
| **pixel-art-design** | `.github/skills/pixel-art-design/SKILL.md` | Aseprite + PixelLab.ai workflow, spritesheet format, CSS animation |

> Les skills sont écrits en local — pas de dépendance à des URLs externes.

---

## Features Prévues (validées)

### 🐾 Suivi des animaux
- **Fiche animal persistante** — un animal signalé plusieurs fois crée une fiche avec historique, photos, surnom donné par la communauté
- **Statut d'adoption** — marquer un animal comme "adopté" ou "pris en charge par un refuge"
- **Estimation de population par zone** — heatmap des zones denses en animaux errants

### 🤝 Partenariats & Impact
- **Comptes Association** — 3ème type d'utilisateur (association de protection animale) : valide les fiches, organise captures/stérilisations, voit les zones chaudes
- **Dons ciblés vers associations** — portés par les **associations partenaires** : "Nourrir ce chat pendant 1 mois = 5$", avec tracking de l'impact réel (repas confirmés). FeedThemAll fournit la plateforme, les fonds vont directement à l'association. **FTA ne prend aucune commission.** Chaque association gère son propre compte Stripe.
- **Dons directs à FeedThemAll** — bouton "Soutenir l'app" accessible depuis le profil et la page À propos. Montants suggérés : **5$ · 10$ · montant libre**. Entièrement volontaire, aucune fonctionnalité bloquée. Devise de base : **USD**. Multi-devises prévu (auto-détection pays via IP, taux Stripe).
- **Tableau de bord association** — stats d'animaux nourris, zones actives, bénévoles actifs cette semaine

### 🎮 Gamification avancée
- **Quêtes** — "Nourris 3 animaux différents cette semaine" → XP bonus, style quêtes Pokémon
- **Guildes de quartier** — groupe de Feeders d'un même quartier, score collectif, classement inter-quartiers
- **Saisons** — remise à zéro du classement tous les 3 mois + récompenses top 10
- **Titres géographiques** — "Gardien du 11ème", "Légende de Belleville" selon zone et niveau

### 📍 Carte & Temps réel
- **Mode nuit/jour** — carte qui change visuellement selon l'heure
- **Zones dangereuses** — signaler circulation dense, chantier pour avertir les Feeders
- **Itinéraire de nourrissage** — optimisation de l'ordre des pings à nourrir dans une zone

### 📱 Mobile / UX
- **Mode hors-ligne** — cache local des pings, sync au retour du réseau
- **Notifications push** — "Animal non nourri depuis 24h près de toi", "Quelqu'un a confirmé ton ping"
- **Widget home screen** — compteur d'animaux nourris cette semaine

### 🌍 Scalabilité
- **Multi-villes / Multi-pays** — structure backend multi-région dès le départ
- **API publique** — données anonymisées accessibles aux associations et chercheurs

### 💳 Monétisation
- **Premium volontaire** — abonnement non obligatoire pour soutenir l'app. Paliers : **5$/mois · 10$/mois · montant libre**. Avantages : badge profil Premium, absence de publicités, cosmétiques exclusifs (tenue avatar). Aucune fonctionnalité de jeu bloquée.
- **Paiement** — Stripe (Web + Mobile). Gestion des abonnements récurrents via Stripe Billing. Webhooks Stripe pour activer/désactiver le statut Premium en DB.
- **Multi-devises** — devise de base USD. Stripe gère la conversion automatique. Affichage dans la devise locale de l'utilisateur (détection via IP ou préférence profil). Phase 2 : EUR, GBP, CAD.
- **Dons one-shot** — même interface que le Premium mais sans récurrence (via `payment_intent` Stripe).

### Priorités MVP+1
1. Fiche animal persistante (différenciant émotionnel fort)
2. Quêtes (boost rétention simple à implémenter)
3. Comptes Association (partenariats + monétisation B2B)

---

## Décisions Architecturales Notables

| Décision | Choix | Raison |
|---|---|---|
| Monorepo | Oui | Partage de types, CI unifiée, cohérence versioning |
| Base de données | PostgreSQL seule (pas de MongoDB) | PostGIS indispensable, pas besoin de schéma flexible |
| Mobile | React Native (pas Flutter) | Partage de code JS avec le Web |
| Carte | Leaflet + OpenStreetMap | Zéro coût, aucune clé API, open source |
| Auth | JWT stateless | Simplifie le scaling horizontal |
