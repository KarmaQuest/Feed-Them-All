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
│   │   ├── auth/         ← Tout ce qui concerne les comptes et la connexion
│   │   └── pings/        ← Tout ce qui concerne les signalements sur la carte
│   ├── migrations/       ← Les instructions pour créer/modifier la base de données (5 fichiers)
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
| `000003_ping_media.up.sql` | Ajoute la table `ping_media` pour les photos uploadées |
| `000004_ping_reports.up.sql` | Ajoute la table `ping_reports` pour les signalements |
| `000005_ping_report_votes.up.sql` | Ajoute la table `ping_report_votes` pour les votes sur signalements |

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

| Phase | Ce qui sera créé | Statut |
|---|---|---|
| P3 — WebSocket | Mise à jour de la carte en temps réel | ✅ Terminé |
| P4 — Gamification | Système XP, badges, leaderboard | ✅ Terminé |
| P4-08 — Shop | Boutique avatars, Stripe, inventaire | ✅ Terminé |
| PA — Admin Dashboard | Interface d'administration backend + frontend | ⬜ À faire |
| P5/6 — Frontend Web | Interface React avec la carte Leaflet | ⬜ À faire |
| P7 — Avatars | Sprites pixel art des personnages | ⬜ À faire |
| P8 — Design | Assets Aseprite / PixelLab.ai | ⬜ À faire |
| P9 — Mobile | Application React Native | ⬜ À faire |

---

## Polling vs WebSocket — deux façons de recevoir des données en temps réel

### Le polling — "est-ce qu'il y a du nouveau ?"

Sans WebSocket, la seule façon pour le client de savoir si quelque chose a changé est de **demander régulièrement** au serveur.

```
[Client]          [Serveur]
   │  GET /pings       │
   │ ────────────────► │ → "rien de nouveau"
   │                   │
   │  (1 seconde)      │
   │  GET /pings       │
   │ ────────────────► │ → "rien de nouveau"
   │                   │
   │  (1 seconde)      │
   │  GET /pings       │
   │ ────────────────► │ → "voilà 2 nouveaux pings !"
```

C'est comme appuyer sur F5 toutes les secondes pour actualiser une page.

**Problèmes du polling :**
- Gaspille de la bande passante et des ressources serveur (1000 utilisateurs = 1000 requêtes/seconde)
- Le délai entre la création d'un ping et son apparition chez les autres = jusqu'à 1 seconde
- Si on augmente la fréquence pour réduire le délai, ça empire le gaspillage

### WebSocket — "je t'enverrai un message quand il y a du nouveau"

Un **WebSocket** est une connexion permanente et bidirectionnelle entre le client et le serveur.
Une fois ouverte, les deux peuvent s'envoyer des messages à tout moment — sans que le client ait besoin de demander.

```
[Client]                    [Serveur]
   │  Connexion WebSocket       │
   │ ─────────────────────────► │  ← connexion établie une seule fois
   │                            │
   │                            │  (quelqu'un crée un ping)
   │ ◄───────────────────────── │  → "nouveau ping ici !"  (immédiat)
   │                            │
   │                            │  (un feeder se déplace)
   │ ◄───────────────────────── │  → "le feeder X est maintenant là"
   │                            │
   │  (connexion reste ouverte) │
```

**Avantages du WebSocket :**
- La carte se met à jour **instantanément** quand quelqu'un crée un ping
- Le serveur n'envoie des données **que quand il y a quelque chose à dire**
- Une seule connexion par utilisateur au lieu d'une requête par seconde

### Pourquoi le WebSocket avant le frontend ?

Si on construisait d'abord le frontend (la carte) avec du polling, puis qu'on ajoutait le WebSocket après,
il faudrait réécrire toute la logique de mise à jour de la carte.
En faisant le WebSocket d'abord, le frontend peut se connecter directement à une vraie connexion temps réel.

---

## Hébergement — quel serveur pour FeedThemAll ?

### WebSocket et serverless — incompatibilité

Un WebSocket a besoin d'une **connexion qui reste ouverte** entre le client et le serveur.
Certains hébergements dits "serverless" (AWS Lambda, Vercel, Netlify...) **éteignent automatiquement le serveur** après 30 secondes d'inactivité pour économiser des ressources. Résultat : la connexion WebSocket est coupée.

> Analogie : c'est comme louer une salle de réunion qui se verrouille automatiquement toutes les 30 secondes. Pour une réunion longue, il faut une salle classique qui reste ouverte.

Pour FeedThemAll, il faut donc un **serveur qui tourne en permanence**.

### Comparaison des options

| Option | Prix | Facilité | WebSocket | PostGIS | Photos uploadées |
|---|---|---|---|---|---|
| **Hetzner VPS CX23** | ~4.49€/mois | Moyenne (Linux) | ✅ | ✅ via Docker | ✅ disque local |
| **Fly.io** | ~5-10€/mois | Facile | ✅ | ✅ | ⚠️ disparaissent au redéploiement |
| **Railway** | ~5€/mois | Très facile | ✅ | ⚠️ non garanti | ⚠️ disparaissent |
| **Render** | ~7€/mois | Facile | ✅ | ✅ | ⚠️ disparaissent |

Voir aussi : https://www.hetzner.com/cloud/cost-optimized

**"Photos disparaissent"** : sur les hébergements cloud modernes, chaque mise à jour du serveur repart d'une image vierge. Les photos uploadées par les utilisateurs sont effacées. Solution : les stocker sur un service externe comme **Cloudflare R2** ou **AWS S3** qui garde les fichiers séparément du serveur.

### Recommandation selon le stade du projet

**Phase actuelle (dev local) → tout tourne sur ta machine.** Aucun serveur requis.

**MVP / premiers utilisateurs → Hetzner VPS CX23 (~4.49€/mois)**
- Un seul serveur Linux, Docker Compose : Go + PostgreSQL+PostGIS + photos sur le même disque
- Même configuration qu'en local, juste sur une machine distante
- Pas de services séparés à gérer, pas de surprise de facturation

**Si l'app grandit → architecture séparée**
- Backend Go sur Fly.io
- PostgreSQL+PostGIS managé via Supabase
- Photos sur Cloudflare R2 (très peu cher, compatible S3)

---

## Le package `pings` — `backend/internal/pings/`

C'est le module qui gère les signalements sur la carte : créer un ping, le retrouver par GPS,
uploader une photo, le signaler comme problématique, voter sur les signalements.

### Les nouvelles tables en base

| Table | Rôle |
|---|---|
| `ping_media` | Photos de preuve liées à un ping (chemin du fichier sur disque) |
| `ping_reports` | Signalements : "cet animal a disparu", "mauvaise position"... Un seul report par utilisateur par ping |
| `ping_report_votes` | Votes up/down sur un signalement — permet de mesurer sa crédibilité. Un vote **remplace** le précédent si on change d'avis |

### Le système de signalement (report)

N'importe quel utilisateur connecté peut signaler un ping — **y compris son créateur**.
Raisons possibles : `wrong_location`, `animal_gone`, `duplicate`, `inappropriate`.

Chaque signalement accumule des votes :
- **Vote up** → "je confirme ce signalement est valide"
- **Vote down** → "ce signalement est faux"
- **Score** = nombre de up − nombre de down

Le frontend peut utiliser ce score pour afficher un badge "douteux" sur un ping très signalé.

### Soft delete — qu'est-ce que c'est ?

Quand on "supprime" un ping, on ne l'efface pas vraiment de la base.
On passe juste `is_active = false`. Le ping reste en base pour l'historique.

> Avantage : si un animal réapparaît au même endroit, on peut retrouver l'historique.
> Si on supprimait vraiment, l'information serait perdue pour toujours.

### Upload de photos

Quand un utilisateur envoie une photo :
1. Le serveur lit les **512 premiers octets** du fichier pour détecter son vrai format (JPEG ou PNG)
2. Il **ne fait pas confiance** au nom du fichier ni au header envoyé par le client — uniquement aux octets réels
3. Le fichier est sauvegardé dans `uploads/<pingID>/<uuid>.jpg` sur le disque du serveur
4. Le chemin relatif est enregistré en base dans `ping_media`

---

## Les tests automatisés — `backend/internal/pings/*_test.go`

### Pourquoi écrire des tests ?

Un test automatisé est un programme qui **vérifie que le code se comporte correctement**.
Au lieu de devoir retester manuellement après chaque modification, les tests s'exécutent
en quelques secondes et signalent immédiatement si quelque chose est cassé.

> Analogie : c'est comme une checklist de vol qu'un pilote parcourt avant chaque décollage.
> Sauf qu'ici la checklist se complète toute seule et prend 0.6 secondes.

### Les 3 fichiers de tests du package `pings`

#### `fake_store_test.go` — La fausse base de données

Pour tester sans lancer PostgreSQL, on crée un **fakeStore** : une implémentation de l'interface `Store`
qui stocke tout en mémoire (dans des maps Go) au lieu d'une vraie base de données.

```
Test démarre
    │
    ▼
fakeStore créé (vide, en mémoire)
    │
    ▼
Service branché sur fakeStore (pas sur PostgreSQL)
    │
    ▼
Test s'exécute — les données vivent dans la RAM
    │
    ▼
Test termine — tout est effacé automatiquement
```

Avantages :
- Aucun Docker nécessaire
- Chaque test repart d'une base vide → pas de pollution entre les tests
- 100x plus rapide qu'une vraie base de données

#### `service_test.go` — Tests de la logique métier

Ces tests vérifient que les **règles** sont bien appliquées :

| Test | Ce qu'il vérifie |
|---|---|
| `TestService_Create_InvalidType` | Créer un ping avec `type = "zombie"` → retourne `ErrInvalidType` |
| `TestService_Deactivate_NotOwner` | Un autre utilisateur tente de supprimer ton ping → retourne `ErrNotOwner` |
| `TestService_Report_CreatorCanReport` | Le créateur d'un ping peut aussi le signaler lui-même |
| `TestService_Report_AlreadyReported` | Signaler deux fois le même ping → retourne `ErrAlreadyReported` |
| `TestService_VoteReport_ChangeVote` | Voter up puis changer pour down → le vote est mis à jour sans erreur |
| `TestService_ListReports_WithScore` | 1 vote up + 1 vote down → score = 0 |

#### `handler_test.go` — Tests des réponses HTTP

Ces tests vérifient que le serveur renvoie les **bons codes HTTP** dans les bons cas :

| Code HTTP | Signification | Exemple |
|---|---|---|
| `201 Created` | Ressource créée avec succès | Ping créé, report créé |
| `204 No Content` | Action réussie, rien à retourner | Confirmer présence, voter |
| `400 Bad Request` | Données invalides | Type de ping inconnu, reason invalide |
| `401 Unauthorized` | Non connecté | Tenter de créer un ping sans token JWT |
| `403 Forbidden` | Connecté mais pas autorisé | Supprimer le ping de quelqu'un d'autre |
| `404 Not Found` | Ressource inexistante | Voter sur un report qui n'existe pas |
| `409 Conflict` | Doublon détecté | Signaler un ping qu'on a déjà signalé |

Pour simuler un utilisateur connecté dans les tests (sans vrais tokens JWT), on utilise
`auth.NewContextWithUserID(ctx, "user-id")` — une fonction helper qui injecte directement
l'ID utilisateur dans le contexte de la requête, comme si le middleware JWT l'avait fait.

### Comment lancer les tests

```bash
cd backend
go test ./internal/...        # Lance tous les tests (auth + pings)
go test ./internal/pings/...  # Lance uniquement les tests pings
go test ./internal/pings/... -v  # Affiche chaque test individuellement
```

Résultat actuel : **57 tests, 0 échec** (auth : 20, pings : 37).

### C'est quoi une "sentinel error" (erreur sentinelle) ?

C'est une erreur nommée et fixe, déclarée une fois au niveau du package :

```go
var ErrNotOwner = errors.New("you are not the owner of this ping")
```

Avantage : le Handler peut comparer avec `errors.Is(err, ErrNotOwner)` pour savoir exactement
quoi retourner au client. Si on retournait juste un texte libre, impossible de le détecter précisément.

> Analogie : c'est comme des codes d'erreur sur un appareil électroménager.
> "E3" veut toujours dire la même chose — on sait quoi faire.
> Un message en texte libre ("quelque chose ne va pas") ne dit rien.

### C'est quoi un upsert ?

Un **upsert** = INSERT ou UPDATE selon si la ligne existe déjà.

Pour les votes : au lieu de bloquer avec une erreur si l'utilisateur a déjà voté,
on **remplace son vote** par le nouveau. SQL : `INSERT ... ON CONFLICT DO UPDATE SET value = EXCLUDED.value`.

C'est le comportement de YouTube (tu peux changer ton pouce en bas) ou Reddit.

---

## Le package `websocket` — `backend/internal/websocket/`

### Le Hub — chef d'orchestre des connexions

Le **Hub** est un objet unique qui tourne dans une goroutine dédiée et gère **tous** les clients connectés.
Il est le seul à modifier la liste des clients — cela évite les conflits si deux utilisateurs
se connectent exactement en même temps (**race condition**).

```
[Client A connecté]  →  register channel  →  [HUB]  →  [liste des clients]
[Client B déconnecté] →  unregister channel →  [HUB]  →  [liste mise à jour]
[Nouveau ping créé]   →  broadcast channel  →  [HUB]  →  envoie à tous les clients concernés
```

> Analogie : le Hub est comme un standardiste téléphonique. Tout passe par lui.
> Personne n'appelle directement quelqu'un d'autre — on demande au standardiste de transmettre.

### BoundingBox — recevoir seulement ce qui est sur ton écran

Quand une application affiche une carte, elle n'a pas besoin de recevoir les pings du monde entier —
seulement ceux visibles dans la **zone de la carte affichée sur l'écran**.

Une `BoundingBox` est un rectangle GPS : `{ nord, sud, est, ouest }`.

```
┌─────────────────────────────────┐  ← nord (ex: 48.90)
│                                 │
│   Tu vois la carte ici          │
│   Le serveur n'envoie que ce    │
│   qui est dans ce rectangle     │
│                                 │
└─────────────────────────────────┘  ← sud (ex: 48.80)
  ouest (ex: 2.28)        est (ex: 2.42)
```

Quand tu fais défiler la carte vers le nord, ton client envoie une nouvelle BoundingBox au serveur.

### Goroutine de lecture / écriture par client

Chaque client WebSocket a **deux goroutines** qui tournent en parallèle :

| Goroutine | Rôle |
|---|---|
| `ReadPump` | Lit les messages entrants du client (position GPS, changement de zone) |
| `WritePump` | Envoie les messages sortants vers le client (nouveaux pings, positions) |

Séparer lecture et écriture en deux goroutines est la méthode recommandée par gorilla/websocket :
les deux peuvent agir en même temps sans se bloquer.

### Keepalive — ping/pong

Si une connexion WebSocket est inerte trop longtemps (WiFi coupé, onglet en arrière-plan),
le serveur ne peut pas toujours détecter que l'autre côté est parti.

Le mécanisme **ping/pong** résout ça :
- Toutes les 54 secondes, le serveur envoie un message "ping" au client
- Si le client ne répond pas avec "pong" dans les 60 secondes → connexion fermée et client retiré

### Rate limiting GPS — 1 mise à jour par seconde

Si chaque feeder envoie sa position GPS 60 fois par seconde (mise à jour accélération du téléphone),
1000 feeders = 60 000 messages/seconde. C'est trop.

On utilise un **token bucket** (seau de jetons) pour limiter à 1 message GPS par seconde par client :
- Le seau se remplit de 1 jeton par seconde
- Chaque message GPS consomme 1 jeton
- Si le seau est vide → le message est ignoré silencieusement

> Analogie : tu as un robinet qui te donne 1 bille par seconde. Chaque fois que tu veux envoyer
> ta position, tu utilises 1 bille. Si tu n'en as plus, tu attends la prochaine bille.

### Précision GPS arrondie

Les coordonnées GPS envoyées par les clients sont arrondies à **4 décimales** (~11 mètres de précision).
Raison : éviter un déluge de mises à jour quand le téléphone oscille de 0.00001° sans bouger.

---

## Le package `gamification` — `backend/internal/gamification/`

### C'est quoi le système XP ?

L'XP (expérience) est un compteur qui augmente quand tu fais des actions utiles dans l'app.
Chaque type d'action a un nombre de points et une limite journalière.

| Action | Identifiant | XP | Limite/jour |
|---|---|---|---|
| Signaler un animal | `signal_animal` | 10 | 5× |
| Confirmer une présence | `confirm_presence` | 15 | 10× |
| Nourrir un animal | `feed` | 25 | 5× |
| Uploader une photo | `upload_photo` | 20 | 3× |

Ces valeurs sont configurées dans la **table `xp_actions`** en base de données — pas dans le code.
Un admin peut les modifier sans redéployer le serveur.

### LogAndAwardXP — opération atomique avec CTE

Quand un utilisateur gagne des XP, deux choses doivent se passer ensemble :
1. Enregistrer l'action dans `xp_log` (historique)
2. Incrémenter `users.xp`

On utilise une **CTE (Common Table Expression)** pour faire les deux en une seule requête SQL :

```sql
WITH inserted AS (
    INSERT INTO xp_log (user_id, action, xp_earned) VALUES ($1, $2, $3)
    RETURNING xp_earned
)
UPDATE users SET xp = xp + (SELECT xp_earned FROM inserted) WHERE id = $1
```

> Pourquoi une CTE ? Si on faisait deux requêtes séparées et que le serveur crashait entre les deux,
> l'utilisateur aurait son log mais pas ses XP (ou l'inverse). La CTE garantit le "tout ou rien".

### Les badges — conditions configurables en base

Un badge est défini par :
- Son identifiant (`first_signal`, `feeder_5`...)
- Son nom affiché ("Premier signal", "Nourrisseur débutant"...)
- Une **condition** en JSON : `{"type": "action_count", "action": "feed", "threshold": 5}`

Types de conditions :
- `xp_threshold` → "avoir accumulé X points XP total"
- `action_count` → "avoir effectué X fois une action précise"

Avantage : ajouter un badge = insérer une ligne en base, sans modifier le code.

### Vérification asynchrone des badges

Vérifier si un utilisateur vient de débloquer un badge après chaque action serait lent
si fait de façon synchrone (l'utilisateur attend la réponse).

Solution : la vérification se lance dans une **goroutine séparée** (fil d'exécution parallèle) :

```
POST /pings/:id/fed
    │
    ▼
Service.MarkFed() → base de données mise à jour → réponse 204 envoyée immédiatement
    │
    └──► (goroutine) AwardXP() → CheckBadges() → CheckQuestItems()
              [s'exécute en arrière-plan, sans bloquer le client]
```

L'utilisateur reçoit sa réponse instantanément. Les badges apparaissent quelques millisecondes plus tard.

### Le leaderboard — cache en mémoire

Le classement global des 20 meilleurs feeders est **calculé en mémoire et mis en cache** pour 5 minutes.

Pourquoi ? Si 1000 utilisateurs regardent le leaderboard en même temps, on ne veut pas faire
1000 requêtes SQL identiques. On fait la requête une fois, on garde le résultat 5 minutes.

```
Requête 1 : calcul SQL → résultat stocké en mémoire avec timestamp
Requêtes 2 à 10000 : résultat servi depuis la mémoire (instantané)
[5 minutes plus tard] Requête 10001 : nouveau calcul SQL → nouveau cache
```

Le cache utilise `sync.RWMutex` : plusieurs goroutines peuvent lire simultanément,
mais une seule peut écrire (pendant le rafraîchissement). Cela évite les corruptions.

### Calcul du niveau

Le niveau est calculé à la volée à partir des XP — il n'est **pas stocké en base**.
Si les seuils changent (décision de game design), tous les utilisateurs voient leur niveau mis à jour
automatiquement sans migration de données.

| Niveau | XP requis |
|---|---|
| 1 | 0 |
| 2 | 100 |
| 3 | 250 |
| 4 | 500 |
| 5 | 900 |
| 6 | 1 400 |
| 7 | 2 100 |
| 8 | 3 000 |
| 9 | 4 500 |
| 10 | 7 000 |

---

## Le package `shop` — `backend/internal/shop/`

### Les 3 types d'items

| Type | Comment le débloquer | Exemple |
|---|---|---|
| **Default** | Disponible dès le départ, sans condition | Skin de base, tenue par défaut |
| **Quest** | Remplir une condition (XP ou actions) | "Tenue Feeder" après 5 nourrissages |
| **Paid** | Payer via Stripe | "Ninja Outfit" 6,99$ |

### Le flux d'achat Stripe

Quand un utilisateur achète un item payant, voici ce qui se passe :

```
[Utilisateur clique "Acheter"]
        │
        ▼
POST /shop/items/:id/purchase
        │
        ▼
Backend crée un PaymentIntent chez Stripe
        │
        ▼
Retourne un client_secret au frontend
        │
        ▼
Frontend affiche le formulaire de carte Stripe (stripe.js)
        │
        ▼
Utilisateur entre sa carte → Stripe traite le paiement
        │
        ▼
Stripe envoie un webhook POST /shop/webhook → "payment_intent.succeeded"
        │
        ▼
Backend vérifie la signature du webhook (STRIPE_WEBHOOK_SECRET)
        │
        ▼
CompleteOrder : UPDATE shop_orders + INSERT user_avatar_items (atomique via CTE)
        │
        ▼
L'item apparaît dans l'inventaire de l'utilisateur
```

### C'est quoi un webhook ?

Un webhook est une notification que Stripe envoie **au serveur** quand un événement se produit.
C'est l'inverse d'un appel API normal : ici c'est Stripe qui appelle notre serveur, pas l'inverse.

> Analogie : au lieu d'appeler le restaurant toutes les 5 minutes pour demander "ma pizza est prête ?",
> le restaurant t'appelle quand elle est prête. C'est un webhook.

**Pourquoi ne pas faire confiance au frontend ?**
Si le frontend disait au serveur "le paiement a réussi, donne-moi l'item", n'importe qui pourrait
envoyer ce message sans avoir vraiment payé. Le webhook vient **directement de Stripe**, avec une
signature cryptographique que le serveur vérifie — impossible à falsifier.

### Vérification de signature webhook

Stripe joint à chaque webhook une valeur `Stripe-Signature` dans les headers.
Le backend la vérifie avec `STRIPE_WEBHOOK_SECRET` (une clé partagée secrète).

Si la signature ne correspond pas → requête rejetée immédiatement avec 400.
Cela protège contre quelqu'un qui essaierait de simuler un faux paiement.

### Idempotence — éviter de doubler les items

Si Stripe envoie le même webhook deux fois (réseau instable, retry automatique),
le serveur ne doit pas créditer l'item deux fois.

Protection : la table `shop_orders` a une contrainte `UNIQUE` sur `stripe_payment_intent_id`.
Si `CompleteOrder` est appelé deux fois avec le même ID Stripe → la deuxième tentative retourne
`ErrOrderExists` et l'action est ignorée silencieusement.

> Idempotence = une opération qui donne le même résultat peu importe combien de fois on la répète.

---

## Injection de dépendances via interfaces — éviter les imports circulaires

### Le problème

Le package `pings` doit envoyer des XP au package `gamification`.
Le package `gamification` doit accorder des items au package `shop`.

Si `pings` importait directement `gamification`, et `gamification` importait directement `shop`,
tout fonctionnerait. Mais si un jour `shop` avait besoin d'un ping, on aurait :
`pings → gamification → shop → pings` = **import circulaire** = Go refuse de compiler.

### La solution : interfaces dans le paquet consommateur

Au lieu d'importer le package cible, on définit une **interface minimale** dans le package qui en a besoin.

```
pings/service.go définit :
    type Broadcaster interface { BroadcastPingCreated(ping) }
    type XPAwarder interface { AwardXP(ctx, userID, action) }

gamification/service.go définit :
    type ItemGranter interface { CheckQuestItems(ctx, userID) }
```

Puis dans `main.go` on "branche" les vraies implémentations :

```go
pingsService.SetBroadcaster(wsHub)         // wsHub implémente Broadcaster
pingsService.SetXPAwarder(gamifService)    // gamifService implémente XPAwarder
gamifService.SetItemGranter(shopService)   // shopService implémente ItemGranter
```

Résultat : les packages ne se connaissent pas entre eux — ils connaissent seulement des interfaces.
C'est `main.go` qui fait les connexions. Zéro import circulaire.

> Analogie : une prise électrique (interface) ne "sait" pas ce qui sera branché dessus.
> Une lampe, un chargeur, un aspirateur — tous peuvent s'y brancher tant qu'ils ont la bonne fiche.

---

## Les goroutines et la concurrence Go

### C'est quoi une goroutine ?

Une goroutine est un **fil d'exécution léger** géré par Go.
Tu peux en lancer des milliers sans problème (elles consomment ~2 Ko chacune, contre ~1 Mo pour un thread OS).

```go
go maFonction()  // Lance maFonction() en parallèle — le code continue sans attendre
```

Dans FeedThemAll, les goroutines servent à :
- Gérer chaque client WebSocket (2 goroutines par client : ReadPump + WritePump)
- Vérifier les badges après chaque action (sans bloquer la réponse HTTP)
- Maintenir le Hub WebSocket dans son propre fil d'exécution

### Les channels — communication entre goroutines

Les goroutines communiquent via des **channels** (tuyaux typés).
Envoyer dans un channel bloque jusqu'à ce que quelqu'un lise de l'autre côté.

```
goroutine A ──── channel ────► goroutine B
              "nouveau ping"
```

Dans le Hub WebSocket :
```go
hub.broadcast ← message  // "envoie ce message à tous les clients"
hub.register  ← client   // "enregistre ce nouveau client"
hub.unregister ← client  // "retire ce client"
```

### Mutex — protéger les données partagées

Si deux goroutines modifient la même variable en même temps, le résultat est imprévisible (**race condition**).

Un **mutex** (mutual exclusion) permet à une seule goroutine d'accéder à une ressource à la fois :

```go
mu.Lock()        // "je prends le verrou"
données = ...    // modification sécurisée
mu.Unlock()      // "je libère le verrou"
```

`sync.RWMutex` est une variante : plusieurs goroutines peuvent **lire** simultanément,
mais une seule peut **écrire**. Utilisé pour le cache du leaderboard.

---

## Migrations SQL — tableau complet

| Fichier | Ce qu'il crée |
|---|---|
| `000001_init.up.sql` | Tables initiales : users, pings, animal_profiles, badges, user_badges, xp_actions, subscriptions |
| `000002_refresh_tokens.up.sql` | Table `refresh_tokens` (connexion persistante) |
| `000003_ping_media.up.sql` | Table `ping_media` (photos uploadées) |
| `000004_ping_reports.up.sql` | Table `ping_reports` (signalements) |
| `000005_ping_report_votes.up.sql` | Table `ping_report_votes` (votes up/down) |
| `000006_gamification.up.sql` | Table `xp_log` (historique XP), seeds xp_actions + 10 badges |
| `000007_avatar_shop.up.sql` | Tables `avatar_items`, `user_avatar_items`, `shop_orders` + 9 items seedés |

---

## Variables d'environnement nécessaires au démarrage du serveur

```
DATABASE_URL        = postgres://fta:fta@localhost:5432/feedthemall?sslmode=disable
JWT_SECRET          = (clé secrète pour signer les access tokens)
JWT_REFRESH_SECRET  = (clé secrète pour signer les refresh tokens)
ENV                 = development
PORT                = 8080
STRIPE_SECRET_KEY   = sk_test_...   (clé Stripe — ne jamais committer)
STRIPE_WEBHOOK_SECRET = whsec_...   (clé Stripe webhook — ne jamais committer)
```

> Ces valeurs ne sont **jamais** écrites dans le code source.
> Sur un vrai serveur, elles seront dans des variables système ou un gestionnaire de secrets.

---

## Stripe CLI — outil de développement local

Stripe ne peut pas envoyer des webhooks vers `localhost` (ton ordinateur n'est pas accessible depuis Internet).
La **Stripe CLI** crée un tunnel pour recevoir ces webhooks localement :

```bash
stripe listen --forward-to localhost:8080/shop/webhook
```

Cette commande affiche un `whsec_...` temporaire à utiliser comme `STRIPE_WEBHOOK_SECRET`.
Elle reste active en arrière-plan et redirige tous les événements Stripe vers ton serveur local.

En production, Stripe enverra les webhooks directement à l'URL du serveur — la CLI n'est utile qu'en dev.

---

## Le Frontend React — `frontend-web/`

### C'est quoi React ?

React est une bibliothèque JavaScript créée par Meta (Facebook) pour construire des interfaces utilisateur.
L'idée centrale : au lieu de manipuler directement la page HTML, tu décris **à quoi doit ressembler
l'interface selon l'état des données**, et React se charge de mettre à jour la page automatiquement.

> Analogie : plutôt que de dire "va chercher le bouton, change sa couleur, mets à jour ce texte...",
> tu dis "si `isLogged = true`, affiche le bouton rouge avec ce texte". React s'occupe du reste.

### TypeScript — JavaScript avec types

TypeScript est JavaScript auquel on a ajouté des **types** : on précise ce qu'une variable contient.

```ts
// JavaScript (aucune indication)
function ajouterXP(user, montant) { ... }

// TypeScript (clair et vérifiable)
function ajouterXP(user: User, montant: number): void { ... }
```

Avantage : l'éditeur de code détecte les erreurs **avant** d'exécuter le programme.
Si tu passes un texte là où un nombre est attendu, VS Code te le signale immédiatement.

### JSX — HTML dans le JavaScript

React utilise une syntaxe spéciale appelée **JSX** qui permet d'écrire du HTML directement dans le code :

```tsx
// JSX (ce qu'on écrit)
function Bouton({ texte }: { texte: string }) {
  return <button className="btn-primary">{texte}</button>
}

// Ce que le navigateur reçoit après compilation
React.createElement("button", { className: "btn-primary" }, texte)
```

Le fichier a l'extension `.tsx` (TypeScript + JSX) au lieu de `.ts`.

### Composants — les briques de l'interface

Un **composant** React est une fonction qui retourne du JSX.
C'est comme un élément HTML personnalisé et réutilisable.

```tsx
// Composant PingMarker — une épingle sur la carte
interface PingMarkerProps {
  lat: number
  lon: number
  animalType: 'cat' | 'dog' | 'other'
  count: number
  onClick: () => void
}

function PingMarker({ lat, lon, animalType, count, onClick }: PingMarkerProps) {
  return (
    <Marker position={[lat, lon]} icon={getIcon(animalType)} eventHandlers={{ click: onClick }}>
      <Popup>{count} {animalType}(s) ici</Popup>
    </Marker>
  )
}
```

Les composants peuvent être imbriqués :
```tsx
<MapView>
  <PingMarker ... />
  <PingMarker ... />
  <AvatarMarker ... />
</MapView>
```

### Props — les paramètres d'un composant

Les **props** (propriétés) sont les données qu'on passe à un composant, comme des paramètres de fonction.

```tsx
// Utilisation
<PingMarker lat={48.856} lon={2.352} animalType="cat" count={3} onClick={handleClick} />

// Dans le composant, on les reçoit comme paramètres
function PingMarker({ lat, lon, animalType, count, onClick }: PingMarkerProps) {
  // ...
}
```

### Hooks — la mémoire et les effets des composants

Les **hooks** sont des fonctions spéciales qui ajoutent des capacités aux composants.

#### `useState` — mémoriser une valeur

```tsx
const [isOpen, setIsOpen] = useState(false)
// isOpen = valeur actuelle
// setIsOpen = fonction pour changer la valeur (déclenche un re-rendu)

<button onClick={() => setIsOpen(true)}>Ouvrir</button>
{isOpen && <Modal onClose={() => setIsOpen(false)} />}
```

Chaque fois que `setIsOpen` est appelé, React re-calcule l'affichage du composant.

#### `useEffect` — déclencher du code au bon moment

```tsx
useEffect(() => {
  // Ce code s'exécute après le premier affichage du composant
  fetchPings().then(data => setPings(data))

  return () => {
    // Ce code s'exécute quand le composant disparaît (nettoyage)
    closeWebSocket()
  }
}, []) // [] = ne s'exécute qu'une fois (au montage)
```

Cas d'usage dans FeedThemAll :
- Charger les pings quand la carte s'affiche
- Ouvrir la connexion WebSocket quand l'utilisateur se connecte
- Démarrer la géolocalisation GPS au chargement

#### `useCallback` — mémoriser une fonction

```tsx
// Sans useCallback : une nouvelle fonction est créée à chaque rendu → ralentit l'app
const handleClick = (id: string) => { fetchPingDetails(id) }

// Avec useCallback : la fonction est mémorisée, recréée seulement si fetchPingDetails change
const handleClick = useCallback((id: string) => {
  fetchPingDetails(id)
}, [fetchPingDetails])
```

### Hooks personnalisés — réutiliser de la logique

On peut créer ses propres hooks pour encapsuler de la logique réutilisable.

```tsx
// hooks/useGeolocation.ts
function useGeolocation() {
  const [position, setPosition] = useState<[number, number] | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!navigator.geolocation) {
      setError('Géolocalisation non supportée')
      return
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => setPosition([pos.coords.latitude, pos.coords.longitude]),
      () => setError('Accès refusé')
    )
  }, [])

  return { position, error }
}

// Utilisation dans un composant
function MapPage() {
  const { position, error } = useGeolocation()
  // position = [48.856, 2.352] ou null si en attente
}
```

---

## Zustand — la gestion d'état global

### C'est quoi "l'état global" ?

Dans React, chaque composant a son propre `useState` local.
Mais parfois plusieurs composants éloignés dans l'arbre ont besoin de la même information.

Exemple : quand l'utilisateur se connecte, le `Header` doit afficher son nom, la `MapPage` doit
charger ses pings, et le `ProfileButton` doit s'activer. Ces 3 composants n'ont aucun lien direct.

Plutôt que de "faire remonter l'état" jusqu'à un ancêtre commun (ce qui devient vite illisible),
on utilise un **store global** — une source de vérité accessible depuis n'importe quel composant.

**Zustand** est la bibliothèque de store global choisie pour FeedThemAll.
Elle est légère (1 Ko), simple à utiliser, et compatible React.

### Comment ça marche ?

```ts
// src/store/auth.ts
import { create } from 'zustand'

interface User { id: string; username: string; role: string }

interface AuthStore {
  user: User | null
  isLogged: boolean
  login: (user: User) => void   // action pour se connecter
  logout: () => void            // action pour se déconnecter
}

export const useAuthStore = create<AuthStore>((set) => ({
  user: null,
  isLogged: false,
  login: (user) => set({ user, isLogged: true }),   // set = modifier l'état
  logout: () => set({ user: null, isLogged: false }),
}))
```

### Utilisation dans les composants

```tsx
// Dans n'importe quel composant, n'importe où dans l'arbre
function Header() {
  const user = useAuthStore((s) => s.user)     // lire l'état
  const logout = useAuthStore((s) => s.logout) // lire une action

  return <div>{user?.username} <button onClick={logout}>Déconnexion</button></div>
}

function LoginPage() {
  const login = useAuthStore((s) => s.login)

  async function handleSubmit() {
    const data = await apiLogin(email, password)
    login(data.user)  // met à jour le store global → tous les composants se mettent à jour
  }
}
```

Le `(s) => s.user` est un **sélecteur** : il dit "abonne-moi uniquement aux changements de `user`".
Si `isLogged` change mais pas `user`, le composant ne se re-rend pas — optimisation automatique.

### Les stores de FeedThemAll

| Store | Fichier | Contient |
|---|---|---|
| `useAuthStore` | `store/auth.ts` | user connecté, login/logout |
| `useMapStore` | `store/map.ts` | liste des pings sur la carte, filtres actifs |
| `useWebSocketStore` | `store/websocket.ts` | connexion WS, état connecté/déconnecté |

### Zustand vs localStorage — où stocker les données ?

| Données | Où les mettre | Pourquoi |
|---|---|---|
| Access token JWT | Zustand (mémoire) | Ne jamais le mettre dans localStorage (XSS) |
| Refresh token | Cookie HttpOnly | Inaccessible à JavaScript |
| Préférences UI | localStorage | Peut survivre au rechargement de page |
| Liste des pings | Zustand | Temporaire, rechargée à chaque session |

> **XSS (Cross-Site Scripting)** : une attaque où un script malveillant injecté dans la page
> lit le localStorage et vole le token. Un cookie HttpOnly est inaccessible à JavaScript — même le script malveillant ne peut pas le lire.

---

## Axios — le client HTTP

### C'est quoi Axios ?

Axios est une bibliothèque JavaScript qui facilite les appels API HTTP.
Dans FeedThemAll, **tous** les appels au backend passent par un client Axios configuré
dans `src/api/client.ts`.

### L'instance configurée

```ts
// src/api/client.ts
const apiClient = axios.create({
  baseURL: '/api',           // préfixe ajouté à toutes les URLs
  withCredentials: true,     // envoie le cookie refresh token automatiquement
})
```

Le proxy Vite transforme `/api/pings` → `http://localhost:8080/pings` pendant le développement.
En production, c'est nginx ou le serveur qui fait ce travail.

### L'intercepteur de refresh automatique

Quand l'access token expire (après 15 min), le serveur répond `401 Unauthorized`.
L'intercepteur Axios attrape ce `401` et tente automatiquement un refresh **sans que l'utilisateur
ne s'en rende compte** :

```
[Composant]  →  GET /pings (access token expiré)
                    │
                    ▼
              [Serveur]  →  401 Unauthorized
                    │
                    ▼
         [Intercepteur Axios]
              - POST /auth/refresh (cookie HttpOnly envoyé automatiquement)
              - Nouveau access token reçu
              - Rejoue GET /pings avec le nouveau token
                    │
                    ▼
              [Composant]  →  reçoit les pings (l'utilisateur n'a rien vu)
```

Si le refresh échoue (token expiré depuis > 7 jours), l'intercepteur redirige vers `/login`.

### La couche API (`src/api/`)

Les appels réseau sont **toujours** dans des fichiers dédiés, jamais directement dans les composants.

```ts
// src/api/pings.ts
export async function getPingsNearby(lat: number, lon: number, radius: number): Promise<Ping[]> {
  const res = await apiClient.get(`/pings?lat=${lat}&lon=${lon}&radius=${radius}`)
  return res.data
}

export async function createPing(data: CreatePingInput): Promise<Ping> {
  const res = await apiClient.post('/pings', data)
  return res.data
}

export async function addFeedingEvent(pingId: string, data: FeedingEventInput): Promise<void> {
  await apiClient.post(`/pings/${pingId}/feedings`, data)
}
```

> Avantage : si l'URL du serveur change, on ne modifie qu'un seul fichier.
> Les composants ne savent pas comment les données arrivent — ils savent juste que `getPingsNearby(lat, lon, 500)` leur retourne une liste de pings.

---

## Leaflet & React-Leaflet — la carte interactive

### C'est quoi Leaflet ?

Leaflet est la bibliothèque JavaScript la plus populaire pour afficher des cartes interactives.
Elle affiche une carte OpenStreetMap (tuiles libres et gratuites, contributeurs volontaires).

**React-Leaflet** est un wrapper qui transforme les composants Leaflet en composants React.

### Structure de base

```tsx
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet'
import 'leaflet/dist/leaflet.css'

function MapPage() {
  return (
    <MapContainer
      center={[48.856, 2.352]}   // Paris par défaut
      zoom={15}
      style={{ height: '100vh', width: '100%' }}
    >
      <TileLayer
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        attribution="© OpenStreetMap contributors"
      />

      <Marker position={[48.856, 2.352]}>
        <Popup>Un animal ici !</Popup>
      </Marker>
    </MapContainer>
  )
}
```

### Marqueurs personnalisés (pixel art)

Leaflet permet de remplacer le marqueur par défaut par une image ou du HTML personnalisé.

```tsx
import L from 'leaflet'

// Icône image simple (patte, gamelle...)
const pawIcon = L.icon({
  iconUrl: '/assets/sprites/markers/paw.png',
  iconSize: [32, 32],      // taille en pixels
  iconAnchor: [16, 32],    // point d'ancrage (bas-centre pour une épingle)
  popupAnchor: [0, -32],   // où apparaît la popup par rapport à l'icône
})

// Icône HTML (sprite animé d'avatar)
const avatarIcon = L.divIcon({
  html: `<img src="/assets/sprites/characters/default.png"
              style="image-rendering:pixelated;width:32px;height:48px" />`,
  iconSize: [32, 48],
  iconAnchor: [16, 48],
  className: '',  // important : vide = pas de styles Leaflet par défaut
})

<Marker position={[lat, lon]} icon={pawIcon} />
```

### Réagir aux changements de vue (BoundingBox)

Quand l'utilisateur fait défiler la carte, on veut envoyer la nouvelle zone visible au WebSocket.
React-Leaflet expose un hook `useMapEvents` pour ça :

```tsx
function MapEventHandler() {
  const map = useMapEvents({
    moveend: () => {
      const bounds = map.getBounds()
      // Envoyer la nouvelle bounding box au WebSocket
      wsSend({
        type: 'subscribe_bbox',
        north: bounds.getNorth(),
        south: bounds.getSouth(),
        east: bounds.getEast(),
        west: bounds.getWest(),
      })
    }
  })
  return null  // ce composant ne rend rien visuellement
}
```

### Tuiles OpenStreetMap — gratuites mais respectueuses

OpenStreetMap est une carte libre, construite par des bénévoles (comme Wikipedia pour les cartes).
Ses tuiles sont gratuites mais il faut respecter leurs conditions :
- Afficher l'attribution `© OpenStreetMap contributors`
- Ne pas faire d'appels abusifs (cache les tuiles côté navigateur automatiquement)

Pour un usage intensif en production, on peut passer sur **Mapbox** (tuiles payantes mais plus belles)
ou **Stadia Maps** (gratuit jusqu'à 200k tuiles/mois).

---

## Les pings — type d'animal et nombre

### Pourquoi ajouter le type d'animal et le nombre ?

Un ping "animal vu ici" ne dit pas grand-chose. En ajoutant le type et le nombre :
- Un bénévole sait **combien de nourriture apporter** (1 chat = 80g, 1 chienne + 5 chiots = bien plus)
- La carte peut afficher des icônes différentes selon l'espèce (icône chat 🐱 vs chien 🐶)
- Les statistiques globales deviennent plus riches ("3847 chats signalés cette année")

### Ce qui a changé en base de données

Migration `000008_ping_animal_fields` :

```sql
ALTER TABLE pings
  ADD COLUMN animal_type VARCHAR(10) DEFAULT NULL,  -- 'cat', 'dog', 'other', NULL si type='food'
  ADD COLUMN animal_count INTEGER DEFAULT 1;         -- nombre d'animaux observés
```

Contrainte : `animal_type` n'est rempli que si `type = 'animal'`.

### Tableau des types supportés (V1)

| Valeur | Icône | Description |
|---|---|---|
| `cat` | 🐱 | Chat errant ou de gouttière |
| `dog` | 🐶 | Chien errant |
| `other` | 🐾 | Autre animal (pigeon, lapin...) |

D'autres types pourront être ajoutés en V2 sans modifier la structure.

---

## L'historique des nourrissages — Feeding Events

### Pourquoi un historique ?

Avant : `PATCH /pings/:id/fed` marquait le ping "nourri une fois" avec un timestamp unique.
Problème : si le même animal est nourri lundi, mercredi et vendredi par trois personnes différentes,
il n'y avait aucun moyen de le savoir.

L'historique permet de :
- Voir **qui a nourri**, **quand**, et **avec quelle photo**
- Construire un lien affectif ("Moustache le chat roux a été nourri 47 fois ce mois")
- Estimer la fréquence de nourrissage et alerter si un animal n'a pas été nourri depuis longtemps
- Afficher une galerie photo chronologique dans la popup d'un ping

### La nouvelle table `ping_feeding_events`

```sql
CREATE TABLE ping_feeding_events (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ping_id    UUID NOT NULL REFERENCES pings(id) ON DELETE CASCADE,
  fed_by     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  fed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  note       TEXT,                         -- note optionnelle ("elle avait faim !")
  animal_count_seen INTEGER               -- combien d'animaux vus ce jour-là
);
```

Les photos uploadées après un nourrissage sont liées à l'événement via `ping_media.feeding_event_id`.

### Les nouvelles routes

| Route | Méthode | Rôle |
|---|---|---|
| `POST /pings/:id/feedings` | Enregistrer un nourrissage (remplace `PATCH /pings/:id/fed`) |
| `GET /pings/:id/feedings` | Récupérer l'historique complet des nourrissages |

Le champ `pings.fed_at` reste présent — il est mis à jour automatiquement à chaque nourrissage
pour indiquer **la date du dernier nourrissage** (utile pour filtrer les pings récemment nourris sur la carte).

---

## Vite — l'outil de build React

### C'est quoi Vite ?

Vite (prononcé "vit") est l'outil qui :
1. **En développement** : démarre un serveur local ultra-rapide avec rechargement instantané (HMR — Hot Module Replacement). Chaque sauvegarde de fichier met à jour le navigateur en < 100ms.
2. **En production** : compile et optimise tout le code en fichiers minifiés pour le serveur.

### Le proxy Vite (développement uniquement)

Dans `vite.config.ts`, on configure un proxy qui redirige `/api/*` vers le serveur Go :

```
Navigateur → /api/pings → Vite proxy → http://localhost:8080/pings → Go
```

Sans ce proxy, le navigateur bloquerait la requête (CORS — deux origines différentes).
En production, c'est nginx qui joue ce rôle.

### `.vite/` — dossier de cache

Vite crée un dossier `.vite/` qui met en cache les dépendances pré-compilées (axios, leaflet...).
Ce dossier est ignoré par git (`.gitignore`) car il est regénéré automatiquement.

---

## Démarrer le projet en local — commandes complètes

### Backend Go (serveur API)

```powershell
# Ajouter Go au PATH si nécessaire
$env:Path += ";C:\Program Files\Go\bin"

# Dans le dossier backend
cd backend
$env:DATABASE_URL="postgres://fta:fta@localhost:5432/feedthemall?sslmode=disable"
$env:JWT_SECRET="dev-secret-change-in-prod"
$env:JWT_REFRESH_SECRET="dev-refresh-secret-change-in-prod"
$env:ENV="development"
$env:PORT="8080"
go run ./cmd/api
```

### Frontend React (interface web)

```bash
cd frontend-web
npm install   # première fois uniquement
npm run dev   # démarre sur http://localhost:5173
```

### Base de données PostgreSQL

```bash
docker compose up -d   # démarre PostgreSQL en arrière-plan
```

### Dashboard admin

Accessible sur `http://localhost:5173/login` avec `shoptest@test.com` / `Admin1234`.

---

## Tableau de bord des migrations SQL — état actuel

| # | Fichier | Ce qu'il crée | Statut |
|---|---|---|---|
| 1 | `000001_init` | Tables initiales (users, pings, animal_profiles...) | ✅ |
| 2 | `000002_refresh_tokens` | Tokens de reconnexion | ✅ |
| 3 | `000003_ping_media` | Photos de pings | ✅ |
| 4 | `000004_ping_reports` | Signalements | ✅ |
| 5 | `000005_ping_report_votes` | Votes sur signalements | ✅ |
| 6 | `000006_gamification` | XP, badges, leaderboard | ✅ |
| 7 | `000007_avatar_shop` | Boutique avatars + inventaire | ✅ |
| 8 | `000008_ping_animal_fields` | Type d'animal + nombre sur les pings | ✅ |
| 9 | `000009_ping_feeding_events` | Historique des nourrissages | ✅ |

