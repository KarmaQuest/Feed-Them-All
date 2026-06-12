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

| Phase | Ce qui sera créé |
|---|---|
| P3 — WebSocket | Mise à jour de la carte en temps réel |
| P4 — Gamification | Système XP, badges, leaderboard |
| P5/6 — Frontend Web | Interface React avec la carte Leaflet |
| P7 — Avatars | Sprites pixel art des personnages |
| P8 — Design | Assets Aseprite / PixelLab.ai |
| P9 — Mobile | Application React Native |

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
| **Hetzner VPS CX21** | ~5€/mois | Moyenne (Linux) | ✅ | ✅ via Docker | ✅ disque local |
| **Fly.io** | ~5-10€/mois | Facile | ✅ | ✅ | ⚠️ disparaissent au redéploiement |
| **Railway** | ~5€/mois | Très facile | ✅ | ⚠️ non garanti | ⚠️ disparaissent |
| **Render** | ~7€/mois | Facile | ✅ | ✅ | ⚠️ disparaissent |

**"Photos disparaissent"** : sur les hébergements cloud modernes, chaque mise à jour du serveur repart d'une image vierge. Les photos uploadées par les utilisateurs sont effacées. Solution : les stocker sur un service externe comme **Cloudflare R2** ou **AWS S3** qui garde les fichiers séparément du serveur.

### Recommandation selon le stade du projet

**Phase actuelle (dev local) → tout tourne sur ta machine.** Aucun serveur requis.

**MVP / premiers utilisateurs → Hetzner VPS CX21 (~5€/mois)**
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
