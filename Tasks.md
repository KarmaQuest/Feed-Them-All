# FeedThemAll — Task List

> Légende : `[ ]` À faire · `[x]` Terminé · `[~]` En cours · `[!]` Bloqué

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

- [ ] **P1-01** Créer la migration initiale : tables `users`, `pings`, `xp_actions`, `user_badges`
- [ ] **P1-02** Ajouter l'index spatial PostGIS sur `pings.location`
- [ ] **P1-03** Mettre en place `golang-migrate` pour versionner les migrations
- [ ] **P1-04** Implémenter l'inscription (`POST /auth/register`) — hash bcrypt du mot de passe
- [ ] **P1-05** Implémenter la connexion (`POST /auth/login`) — retourne access token JWT (15 min)
- [ ] **P1-06** Implémenter le refresh token (`POST /auth/refresh`) — cookie HttpOnly 7 jours
- [ ] **P1-07** Middleware d'authentification JWT pour les routes protégées
- [ ] **P1-08** Rate limiting sur `/auth/register` et `/auth/login`
- [ ] **P1-09** Tests unitaires : package `auth`

---

## Phase 2 — Backend : Pings & Géolocalisation

- [ ] **P2-01** `GET /pings?lat=&lon=&radius=` — récupérer les pings dans un rayon (ST_DWithin)
- [ ] **P2-02** `POST /pings` — créer un ping (animal ou nourriture), position GPS requise
- [ ] **P2-03** `PATCH /pings/:id/confirm` — confirmer la présence d'un animal (avec photo optionnelle)
- [ ] **P2-04** `PATCH /pings/:id/fed` — marquer un animal comme nourri
- [ ] **P2-05** `DELETE /pings/:id` — désactiver un ping (soft delete `is_active = false`)
- [ ] **P2-06** Upload de photo de preuve (`POST /pings/:id/media`) — validation MIME, 10 Mo max, stockage `uploads/`
- [ ] **P2-07** Servir les fichiers `uploads/` via un handler Go statique
- [ ] **P2-08** Tests d'intégration : package `pings` avec DB PostgreSQL de test (Docker)

---

## Phase 3 — Backend : WebSocket & Temps Réel

- [ ] **P3-01** Mettre en place le serveur WebSocket (`gorilla/websocket`)
- [ ] **P3-02** Système d'abonnement par bounding box (le client envoie sa zone visible)
- [ ] **P3-03** Broadcast d'un nouveau ping aux clients abonnés à la zone correspondante
- [ ] **P3-04** Broadcast de la position des Feeders actifs (avatar en temps réel sur la carte)
- [ ] **P3-05** Gestion des déconnexions et nettoyage des abonnements
- [ ] **P3-06** Limiter la fréquence de mise à jour de position GPS (max 1 push/seconde par client)

---

## Phase 4 — Backend : Gamification

- [ ] **P4-01** Créer la table `xp_actions` avec les valeurs par défaut (signaler, nourrir, uploader, confirmer)
- [ ] **P4-02** Fonction `award_xp(user_id, action)` — calcul côté serveur, rate limiting anti-triche
- [ ] **P4-03** Appeler `award_xp` automatiquement après chaque action éligible
- [ ] **P4-04** Créer la table `badges` et les règles de déverrouillage
- [ ] **P4-05** Job asynchrone (goroutine) pour vérifier et déverrouiller les badges
- [ ] **P4-06** `GET /users/:id/profile` — retourne XP, niveau, badges, avatar config
- [ ] **P4-07** `GET /leaderboard` — top 20 utilisateurs par XP (avec cache en mémoire, TTL 5 min)

---

## Phase 5 — Frontend Web : Carte & Pings

- [ ] **P5-01** Intégrer **Leaflet + React-Leaflet** avec fond de carte OpenStreetMap
- [ ] **P5-02** Afficher les pings récupérés depuis l'API sous forme de marqueurs custom pixel art
- [ ] **P5-03** Mise à jour des marqueurs en temps réel via WebSocket (ajout/suppression de pings)
- [ ] **P5-04** Clic sur un marqueur → popup avec historique du ping (photos, confirmations, date)
- [ ] **P5-05** Bouton "Signaler un animal" → formulaire de création de ping avec GPS auto-détecté
- [ ] **P5-06** Bouton "J'ai nourri" → marquer le ping comme nourri + upload photo optionnel
- [ ] **P5-07** Géolocalisation browser (`navigator.geolocation`) — centrer la carte sur la position utilisateur
- [ ] **P5-08** Fallback si géolocalisation refusée (carte centrée sur la ville par défaut)

---

## Phase 6 — Frontend Web : Auth & Profil

- [ ] **P6-01** Page Inscription (`/register`) — formulaire + appel API
- [ ] **P6-02** Page Connexion (`/login`) — formulaire + stockage du token
- [ ] **P6-03** Gestion du refresh token automatique (intercepteur Axios/fetch)
- [ ] **P6-04** Page Profil (`/profile`) — affichage XP, badges, avatar
- [ ] **P6-05** Store Zustand : état auth (user, token, isLogged)

---

## Phase 7 — Frontend Web : Avatar Pixel Art

- [ ] **P7-01** Créer le composant `AvatarSprite` — affiche le bon sprite selon la config `{skin, outfit, accessory}`
- [ ] **P7-02** Afficher l'avatar du Feeder connecté sur la carte Leaflet (marqueur `L.divIcon` avec sprite animé)
- [ ] **P7-03** Mise à jour de la position de l'avatar en temps réel (push GPS → WebSocket)
- [ ] **P7-04** Page customisation avatar (`/avatar`) — sélecteur visuel de skin/tenue/accessoire
- [ ] **P7-05** Sauvegarder la config avatar en DB via `PATCH /users/me/avatar`
- [ ] **P7-06** Afficher les avatars des autres Feeders actifs sur la carte (en semi-transparent)

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
- [ ] **P8-09** Créer les sprites de customisation : 3 tenues, 3 couleurs de cheveux, 2 accessoires (MVP)

---

## Phase 9 — Mobile (React Native)

- [ ] **P9-01** Initialiser le projet React Native (`Expo`) dans `frontend-mobile/`
- [ ] **P9-02** Partager les types TypeScript depuis `shared/types/`
- [ ] **P9-03** Partager la couche `api/` avec le frontend web
- [ ] **P9-04** Intégrer la carte (React Native Maps ou Leaflet via WebView)
- [ ] **P9-05** Géolocalisation native (Expo Location)
- [ ] **P9-06** Affichage des pings et avatars (parité avec le web)

---

## Backlog / Phase 2+

### Suivi des animaux
- [ ] Fiche animal persistante (historique, photos, surnom communautaire)
- [ ] Statut d'adoption / prise en charge refuge
- [ ] Heatmap densité animaux par zone

### Comptes Association
- [ ] Type utilisateur `association` (3ème rôle)
- [ ] Tableau de bord association (stats, zones, bénévoles actifs)
- [ ] Validation de fiches animaux par les associations
- [ ] Plateforme de dons ciblés portée par les associations partenaires (FeedThemAll = intermédiaire technique, fonds → association directement)

### Gamification avancée
- [ ] Système de quêtes (hebdomadaires, style Pokémon)
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
- [ ] Publicités (AdMob mobile, Google AdSense web)
- [ ] **Abonnement Premium volontaire** — paliers 5$/mois · 10$/mois · montant libre (Stripe Billing, récurrent)
  - [ ] Intégration Stripe (clés API, webhook secret dans `.env`)
  - [ ] Table `subscriptions` en DB (user_id, stripe_customer_id, stripe_subscription_id, plan, status, currency)
  - [ ] Webhook Stripe → activer/désactiver `is_premium` sur le compte utilisateur
  - [ ] Page paiement Web (`/support`) — 3 boutons + champ montant libre
  - [ ] Gestion multi-devises : USD par défaut, EUR/GBP/CAD phase 2 (détection IP, taux Stripe)
- [ ] **Dons one-shot à FeedThemAll** — `payment_intent` Stripe (pas de récurrence), même page `/support`
- [ ] **Dons ciblés associations** — `payment_intent` Stripe vers compte Stripe propre de l'association (Stripe Connect, FTA zéro commission)
- [ ] API publique données anonymisées
- [ ] Migration stockage médias vers Cloudflare R2
- [ ] Cache Redis pour leaderboard et pings chauds
- [ ] Déploiement VPS (Docker Compose en production)
- [ ] Support multi-villes / multi-pays
