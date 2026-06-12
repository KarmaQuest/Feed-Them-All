<div align="center">

# 🐾 FeedThemAll

**Connecter le surplus alimentaire des restaurants avec les animaux errants — en temps réel, avec une communauté gamifiée.**

[![CI Backend](https://github.com/KarmaQuest/Feed-Them-All/actions/workflows/ci-backend.yml/badge.svg)](https://github.com/KarmaQuest/Feed-Them-All/actions/workflows/ci-backend.yml)
[![CI Frontend](https://github.com/KarmaQuest/Feed-Them-All/actions/workflows/ci-frontend.yml/badge.svg)](https://github.com/KarmaQuest/Feed-Them-All/actions/workflows/ci-frontend.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## Concept

FeedThemAll est une plateforme cross-platform (Web + Mobile) qui met en relation deux types d'utilisateurs autour d'une cause simple : ne pas laisser les animaux errants affamés.

- **Les Givers** (restaurateurs, particuliers) publient leurs invendus alimentaires sur la carte.
- **Les Feeders** (bénévoles) localisent les animaux, récupèrent la nourriture et vont nourrir sur le terrain.
- **Les Associations** partenaires valident les fiches animaux, organisent les sauvetages et lancent des collectes de dons.

Chaque action est récompensée via un système de gamification inspiré de Pokémon : XP, badges, quêtes, guildes de quartier, classements saisonniers.

---

## Fonctionnalités clés

| Domaine | Fonctionnalité |
|---|---|
| 🗺️ Carte temps réel | Pings animaux + nourriture via WebSocket, avatars Feeders actifs |
| 📸 Preuve sociale | Photo/vidéo de nourrissage, confirmation communautaire |
| 🎮 Gamification | XP, niveaux, badges, quêtes hebdo, guildes, saisons, titres géo |
| 🐾 Suivi animal | Fiche persistante, statut adoption, heatmap densité |
| 🤝 Associations | Dashboard partenaire, dons ciblés (zéro commission FTA) |
| 💳 Soutien app | Abonnement Premium volontaire (5$/10$/libre), dons one-shot |
| 📱 Mobile | Mode hors-ligne, notifications push, widget home screen |
| 🌍 Scalabilité | Multi-villes/pays, API publique données anonymisées |

---

## Stack technique

| Couche | Technologie |
|---|---|
| Backend | **Go 1.22** — goroutines, REST + WebSockets |
| Base de données | **PostgreSQL 15 + PostGIS 3.3** — requêtes géospatiales |
| Web Frontend | **React 19 + TypeScript** — Vite, Leaflet/OSM |
| Mobile | **React Native + Expo** — GPS natif, caméra |
| Auth | **JWT** access token (15 min) + refresh token HttpOnly (7j) |
| Paiements | **Stripe** — Billing récurrent, payment_intent, Stripe Connect associations |
| Temps réel | **gorilla/websocket** — broadcast par bounding box |
| Design | **Pixel art** style Pokémon FireRed/LeafGreen — Aseprite + PixelLab.ai |
| CI/CD | **GitHub Actions** — lint + test au push sur `dev`/`main` |

---

## Architecture locale (Phase 1)

```
[Client React localhost:3000]
        │
        ├─ REST  ────────────► [Go Server localhost:8080]
        │                              │
        └─ WebSocket ─────────►        ├─ PostgreSQL + PostGIS (Docker :5432)
                                       └─ uploads/ (fichiers locaux)
```

> Phase 1 : 100% local. Aucune dépendance cloud requise pour développer.

---

## Lancer le projet en local

### Prérequis

- [Go 1.22+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Docker](https://rancherdesktop.io/) (Rancher Desktop ou Docker Desktop)

### 1 — Base de données

```bash
docker compose up -d
```

Vérifie que PostGIS est actif :

```bash
docker exec -it feedthemallfta-db-1 psql -U fta -d feedthemall -c "SELECT PostGIS_version();"
```

### 2 — Backend Go

```bash
cd backend
cp .env.example .env   # puis remplis JWT_SECRET
go run ./cmd/api
```

### 3 — Frontend Web

```bash
cd frontend-web
npm install
npm run dev
```

Frontend disponible sur `http://localhost:5173`.

---

## Structure du monorepo

```
FeedThemAll/
├── backend/              # Go — API REST + WebSockets
│   ├── cmd/api/          # Point d'entrée du serveur
│   ├── internal/         # Packages métier (auth, pings, users, ws...)
│   ├── migrations/       # SQL versionnés via golang-migrate
│   └── uploads/          # Médias uploadés (gitignored)
├── frontend-web/         # React + TypeScript (Vite)
├── frontend-mobile/      # React Native + Expo
├── shared/               # Types TypeScript partagés Web/Mobile
├── references/design/    # Sprites pixel art, palette Pico-8
├── .github/
│   ├── workflows/        # CI GitHub Actions
│   └── skills/           # Instructions pour agents IA (Copilot)
├── docker-compose.yml    # PostgreSQL + PostGIS local
├── Concept.md            # Vision du projet
├── Context.md            # Conventions techniques & architecture
└── Tasks.md              # Suivi des tâches (source de vérité)
```

---

## Modèle économique

L'application est **gratuite et sans fonctionnalité bloquée**. Elle se finance via :

1. **Publicités** — pour les utilisateurs gratuits (AdMob mobile, AdSense web)
2. **Premium volontaire** — abonnement 5$/mois · 10$/mois · montant libre. Avantages : sans pub, badge profil, cosmétiques exclusifs. Géré via **Stripe Billing**.
3. **Dons one-shot** — soutien direct à FeedThemAll via `payment_intent` Stripe.
4. **Dons vers associations** — FeedThemAll fournit la plateforme, chaque association gère son propre compte Stripe. **FTA ne prend aucune commission.**

---

## Gamification — aperçu

| Élément | Description |
|---|---|
| XP | Chaque action terrain rapporte des points |
| Niveaux | Progression du rang (Apprenti Feeder → Légende) |
| Badges | Déverrouillés par jalons (premier nourrissage, 100 pings...) |
| Quêtes | Objectifs hebdomadaires avec bonus XP |
| Guildes | Équipes de quartier, score collectif |
| Saisons | Classement remis à zéro tous les 3 mois + récompenses top 10 |
| Titres géo | "Gardien du 11ème", "Légende de Belleville" |
| Avatar | Personnage pixel art customisable, visible sur la carte |

---

## Design — Pixel Art

Style visuel inspiré de **Pokémon FireRed / LeafGreen**.

- Palette : **Pico-8** (16 couleurs)
- Sprites : 16×24 px, spritesheet 3 cols × 4 rows (3 frames × 4 directions)
- Affiché en ×2 CSS (`image-rendering: pixelated`)
- Outils : **Aseprite** + **PixelLab.ai**

---

## Contribuer

1. Fork le repo & crée une branche depuis `dev` : `git checkout -b feat/ma-feature`
2. Respecte les [Conventional Commits](https://www.conventionalcommits.org/) : `feat:`, `fix:`, `chore:`...
3. Les PR cibles `dev` (jamais directement `main`)
4. La CI doit passer (lint + tests) avant merge

---

## Statut du projet

> Phase 0 (Setup & Infrastructure) : ✅ **Terminée**
> Phase 1 (Backend Auth & DB) : 🔄 En cours

Voir [Tasks.md](Tasks.md) pour le détail complet des tâches.

---

<div align="center">
Made with ❤️ for stray animals — <a href="https://github.com/KarmaQuest/Feed-Them-All">github.com/KarmaQuest/Feed-Them-All</a>
</div>
