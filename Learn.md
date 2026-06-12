# FeedThemAll — Guide de compréhension pour non-développeurs

> Ce fichier explique **à quoi sert chaque fichier du projet**, comment ils fonctionnent ensemble,
> et pourquoi certains choix techniques ont été faits. Aucune connaissance en programmation requise.

---

## Vue d'ensemble : comment le projet est organisé

```
FeedThemAll/
│
├── backend/              ← Le "cerveau" : serveur Go qui gère les données et la logique
│   ├── cmd/api/          ← Le point de départ — c'est ici que le serveur démarre
│   ├── internal/         ← Le code métier, organisé par thème
│   │   └── auth/         ← Tout ce qui concerne les comptes et la connexion
│   ├── migrations/       ← Les instructions pour créer/modifier la base de données
│   ├── tests/            ← Pages HTML pour tester l'API depuis un navigateur
│   └── uploads/          ← Photos uploadées par les utilisateurs (en local)
│
├── frontend-web/         ← L'interface web (React, ce que voit l'utilisateur)
├── frontend-mobile/      ← L'application mobile (React Native)
├── shared/               ← Code partagé entre web et mobile
└── docker-compose.yml    ← Lance la base de données PostgreSQL en un clic
```

---

## La base de données — PostgreSQL + PostGIS

### C'est quoi une base de données ?
C'est un endroit où on stocke toutes les informations de l'application de façon organisée.
Imagine un classeur avec des onglets : chaque onglet est une **table**, chaque ligne est une **entrée**.

### C'est quoi PostGIS ?
Une extension de PostgreSQL qui comprend les coordonnées GPS.
Sans PostGIS, stocker "cet animal est à 48.8566°N, 2.3522°E" ne permettrait pas de répondre à
"quels animaux sont à moins de 500 mètres de moi ?". PostGIS rend cette requête possible et rapide.

### C'est quoi Docker ?
Docker permet de lancer la base de données PostgreSQL sur ton ordinateur sans l'installer manuellement.
La commande `docker compose up -d` crée un "container" (une boîte isolée) avec PostgreSQL prêt à l'emploi.

---

## Les migrations SQL — `backend/migrations/`

### C'est quoi une migration ?
Une migration est un fichier SQL numéroté qui **crée ou modifie la structure de la base de données**.
L'outil `golang-migrate` les exécute dans l'ordre, et garde en mémoire lesquelles ont déjà été jouées.

### Pourquoi pas juste modifier la base directement ?
Parce qu'en équipe (ou entre ta machine et le serveur de production), tout le monde doit avoir
**exactement la même structure** de base de données. Les migrations garantissent ça.

| Fichier | Ce qu'il fait |
|---|---|
| `000001_init.up.sql` | Crée toutes les tables initiales (users, pings, animal_profiles, badges...) |
| `000001_init.down.sql` | Supprime ces tables (pour "annuler" — dangereux en prod) |
| `000002_refresh_tokens.up.sql` | Ajoute la table qui stocke les tokens de reconnexion |
| `000002_refresh_tokens.down.sql` | Supprime cette table |

### Les tables et leur rôle

| Table | Rôle |
|---|---|
| `users` | Un compte utilisateur : email, mot de passe hashé, rôle (feeder/giver/association), XP, avatar |
| `pings` | Un signalement sur la carte : "j'ai vu un chat ici" ou "j'ai de la nourriture à donner" |
| `animal_profiles` | La fiche d'un animal : son surnom, son espèce, son historique, son statut (errant/adopté) |
| `ping_animal_links` | Relie un ping à la fiche d'un animal connu (un ping peut concerner "Moustache le chat roux") |
| `xp_actions` | Le barème XP : "signaler un animal = 10 points", "nourrir = 25 points"... |
| `badges` | La définition des badges : "Premier nourrissage", "100 pings"... |
| `user_badges` | Qui a gagné quel badge et quand |
| `subscriptions` | Les abonnements Premium et dons via Stripe |
| `refresh_tokens` | Les tokens de reconnexion automatique (hashés, pour la sécurité) |

---

## Le Backend Go — `backend/`

### C'est quoi Go (Golang) ?
Go est un langage de programmation créé par Google. Il est très rapide, gère bien les connexions
simultanées (des milliers d'utilisateurs en même temps) et consomme peu de mémoire.
C'est le choix idéal pour un serveur temps réel comme FeedThemAll.

### Comment le code est-il organisé ?

Le backend suit un principe : **chaque couche a un rôle unique**.

```
Navigateur / App Mobile
        │
        ▼
   [ Handler ]     ← Reçoit la requête HTTP, vérifie les données basiques, répond en JSON
        │
        ▼
   [ Service ]     ← Applique les règles métier (validation, logique, calculs)
        │
        ▼
   [ Repository ]  ← Parle à la base de données (SQL uniquement)
        │
        ▼
   [ PostgreSQL ]  ← Stocke les données
```

---

## Le package `auth` — `backend/internal/auth/`

C'est le module qui gère tout ce qui concerne les comptes utilisateurs et la sécurité de connexion.

### `model.go` — Les structures de données
Définit les "formulaires" utilisés dans tout le module :
- `User` : à quoi ressemble un utilisateur (id, email, username, rôle, XP...)
- `RegisterRequest` : ce que le client envoie pour s'inscrire
- `LoginRequest` : ce que le client envoie pour se connecter
- `TokenResponse` : ce que le serveur renvoie après connexion réussie

### `store.go` — Le contrat avec la base de données
Définit une **interface** : une liste de promesses que la couche base de données doit tenir.
Concrètement, elle liste toutes les opérations SQL nécessaires (créer un user, récupérer par email...).

> **Analogie** : c'est comme une fiche de poste. Le store.go dit "j'ai besoin de quelqu'un capable de faire X, Y, Z". La vraie implémentation (repository.go) est le candidat qui accepte ce poste.

### `repository.go` — Les vraies requêtes SQL
Implémente le contrat de `store.go` avec PostgreSQL réel.
Chaque méthode = une requête SQL précise.
Ce fichier ne décide **jamais** de la logique — il exécute ce qu'on lui demande.

### `service.go` — La logique métier
C'est le cerveau. Il :
- Valide le mot de passe (min 8 caractères)
- Hash le mot de passe avec **bcrypt** avant de le stocker
  *(bcrypt est une fonction à sens unique : impossible de retrouver le mot de passe depuis le hash)*
- Génère les **tokens JWT** (voir section suivante)
- Vérifie les credentials lors d'un login
- Gère la rotation des refresh tokens

### `handler.go` — Les routes HTTP
Expose les 4 routes aux clients (navigateur / app mobile) :
- `POST /auth/register` → créer un compte
- `POST /auth/login` → se connecter
- `POST /auth/refresh` → renouveler sa session sans se reconnecter
- `POST /auth/logout` → se déconnecter

### `middleware.go` — Le gardien des routes protégées
S'insère automatiquement avant chaque route qui nécessite une connexion.
Lit le token JWT dans le header, le valide, et laisse passer ou bloque.

---

## La sécurité des mots de passe — bcrypt

Quand tu t'inscris avec le mot de passe "monMotDePasse1", voici ce qui se passe :
1. `bcrypt` transforme "monMotDePasse1" en quelque chose comme `$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy`
2. **Seul ce hash est stocké en base** — jamais le mot de passe original
3. Lors du login, bcrypt re-transforme ce que tu tapes et compare les deux hashs
4. Résultat : même si la base de données est piratée, les mots de passe sont inutilisables

---

## Les tokens JWT — comment fonctionne la session

Un **JWT (JSON Web Token)** est un "badge" numérique signé par le serveur.

### Access Token (courte durée — 15 minutes)
- Envoyé à chaque requête dans le header : `Authorization: Bearer <token>`
- Prouve que tu es bien qui tu prétends être
- Expire après 15 min pour limiter les risques si quelqu'un l'intercepte

### Refresh Token (longue durée — 7 jours)
- Stocké dans un **cookie HttpOnly** : le navigateur l'envoie automatiquement, JavaScript ne peut pas le lire (protection contre les attaques XSS)
- Sert uniquement à obtenir un nouvel access token quand celui-ci expire
- Quand il est utilisé, le serveur en génère un nouveau et invalide l'ancien (**rotation**)

### Schéma simplifié
```
[Client]                          [Serveur]
   │  POST /auth/login                │
   │ ──────────────────────────────► │
   │                                  │ Vérifie password
   │ ◄────────────────────────────── │ Retourne access_token (15min) + cookie refresh_token (7j)
   │                                  │
   │  GET /pings (+ access_token)     │
   │ ──────────────────────────────► │
   │ ◄────────────────────────────── │ Répond avec les pings
   │                                  │
   │  [15 minutes plus tard...]       │
   │                                  │
   │  POST /auth/refresh (cookie auto)│
   │ ──────────────────────────────► │
   │ ◄────────────────────────────── │ Retourne nouvel access_token + nouveau cookie
```

---

## `cmd/api/main.go` — Le point de départ

C'est le fichier qui démarre tout le serveur. Dans l'ordre :
1. Lit les variables d'environnement (`.env`) : mot de passe DB, clés JWT, port...
2. Ouvre une connexion vers PostgreSQL (via un "pool" de connexions réutilisables)
3. Crée les objets repository → service → handler pour le module auth
4. Configure le routeur chi avec les middlewares (logs, timeouts, CORS...)
5. Démarre le serveur sur le port 8080

---

## Les variables d'environnement — `backend/.env`

Ce fichier (jamais committé sur Git) contient les "secrets" de l'application :

```
DATABASE_URL=postgres://fta:fta@localhost:5432/feedthemall  ← Adresse de la base de données
JWT_SECRET=...                                               ← Clé pour signer les access tokens
JWT_REFRESH_SECRET=...                                       ← Clé pour signer les refresh tokens
PORT=8080                                                    ← Port d'écoute du serveur
ENV=development                                              ← Mode de fonctionnement
```

> Ces valeurs ne sont **jamais** écrites dans le code source pour éviter qu'elles se retrouvent sur GitHub.

---

## `docker-compose.yml` — La base de données en local

Ce fichier décrit comment lancer PostgreSQL sur ta machine avec une seule commande.
Il configure automatiquement : le nom de la base (`feedthemall`), l'utilisateur (`fta`),
le mot de passe (`fta`), le port (`5432`), et la persistance des données (volume `pgdata`).

```bash
docker compose up -d    # Démarrer PostgreSQL
docker compose down     # Arrêter (données conservées)
docker compose down -v  # Arrêter + SUPPRIMER toutes les données
```

---

## `backend/tests/test-auth.html` — Interface de test

Une page HTML simple, servie par le serveur Go sur `http://localhost:8080/tests/test-auth.html`.
Elle permet de tester tous les endpoints d'authentification depuis un navigateur, sans outil externe.
Disponible uniquement quand `ENV=development`.

---

## Ce qui vient ensuite

| Phase | Ce qui sera créé |
|---|---|
| P2 — Pings | Routes pour créer/lire les signalements sur la carte, upload de photos |
| P3 — WebSocket | Mise à jour de la carte en temps réel |
| P4 — Gamification | Système XP, badges, leaderboard |
| P5/6 — Frontend Web | Interface React avec la carte Leaflet |
| P7 — Avatars | Sprites pixel art des personnages |
| P8 — Design | Assets Aseprite / PixelLab.ai |
| P9 — Mobile | Application React Native |
