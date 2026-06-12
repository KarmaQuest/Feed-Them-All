---
name: postgresql-postgis
description: Expert PostgreSQL + PostGIS database work for FeedThemAll. Use for schema design, migrations, spatial queries, indexes, and Docker local setup.
---

# PostgreSQL + PostGIS — FeedThemAll

## When to Use This Skill
- Writing or reviewing SQL migrations in `backend/migrations/`
- Writing spatial queries (proximity, bounding box)
- Designing or modifying the database schema
- Debugging slow queries with EXPLAIN ANALYZE
- Setting up or troubleshooting the local Docker PostgreSQL instance

## Local Setup (Docker)
```yaml
# docker-compose.yml
services:
  db:
    image: postgis/postgis:15-3.3
    environment:
      POSTGRES_USER: fta
      POSTGRES_PASSWORD: fta
      POSTGRES_DB: feedthemall
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
volumes:
  pgdata:
```
```bash
docker compose up -d      # Start
docker compose down       # Stop (data persists in volume)
docker compose down -v    # Stop + DELETE data
```

Connect:
```bash
psql postgresql://fta:fta@localhost:5432/feedthemall
```

Enable PostGIS (done once in the init migration):
```sql
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
```

## Schema

### users
```sql
CREATE TABLE users (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  username      TEXT NOT NULL UNIQUE,
  role          VARCHAR(10) NOT NULL DEFAULT 'feeder' CHECK (role IN ('feeder','giver')),
  xp            INTEGER NOT NULL DEFAULT 0,
  avatar_config JSONB NOT NULL DEFAULT '{}',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### pings
```sql
CREATE TABLE pings (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type        VARCHAR(10) NOT NULL CHECK (type IN ('animal','food')),
  location    GEOGRAPHY(POINT, 4326) NOT NULL,  -- lon, lat in WGS84
  created_by  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  is_active   BOOLEAN NOT NULL DEFAULT TRUE,
  fed_at      TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- MANDATORY spatial index
CREATE INDEX idx_pings_location ON pings USING GIST(location);
CREATE INDEX idx_pings_active ON pings(is_active) WHERE is_active = TRUE;
```

### xp_actions
```sql
CREATE TABLE xp_actions (
  action      VARCHAR(50) PRIMARY KEY,  -- 'signal_animal','feed','upload_photo','confirm_presence'
  xp_value    INTEGER NOT NULL,
  daily_limit INTEGER NOT NULL DEFAULT 10  -- anti-cheat: max rewards/day per user
);

INSERT INTO xp_actions VALUES
  ('signal_animal',   10, 20),
  ('feed',            25, 10),
  ('upload_photo',    15, 15),
  ('confirm_presence', 5, 30);
```

### badges
```sql
CREATE TABLE badges (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  slug        VARCHAR(50) NOT NULL UNIQUE,
  label       TEXT NOT NULL,
  description TEXT,
  condition   JSONB NOT NULL  -- e.g. {"type":"xp_threshold","value":100}
);

CREATE TABLE user_badges (
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  badge_id    UUID NOT NULL REFERENCES badges(id) ON DELETE CASCADE,
  earned_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, badge_id)
);
```

## Spatial Queries

### Pings within radius (main map query)
```sql
SELECT
  id,
  type,
  ST_Y(location::geometry) AS lat,
  ST_X(location::geometry) AS lon,
  created_at,
  fed_at
FROM pings
WHERE is_active = TRUE
  AND ST_DWithin(
    location,
    ST_MakePoint($1, $2)::geography,  -- $1=lon, $2=lat
    $3                                 -- $3=radius in meters
  )
ORDER BY created_at DESC
LIMIT 200;
```

> ⚠️ Parameter order: **longitude first, then latitude** in `ST_MakePoint`.

### Insert a ping
```sql
INSERT INTO pings (type, location, created_by)
VALUES ($1, ST_MakePoint($2, $3)::geography, $4)
-- $2=lon, $3=lat
RETURNING id, created_at;
```

### Bounding box query (WebSocket subscriptions)
```sql
SELECT id, type,
  ST_Y(location::geometry) AS lat,
  ST_X(location::geometry) AS lon
FROM pings
WHERE is_active = TRUE
  AND location && ST_MakeEnvelope($1, $2, $3, $4, 4326)::geography
-- $1=minLon, $2=minLat, $3=maxLon, $4=maxLat
```

## Migrations (golang-migrate)

### Naming convention
```
000001_init.up.sql         ← Apply
000001_init.down.sql       ← Rollback
000002_add_badges.up.sql
000002_add_badges.down.sql
```

### Running migrations
```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Apply all pending
migrate -path ./migrations -database $DATABASE_URL up

# Rollback last
migrate -path ./migrations -database $DATABASE_URL down 1
```

### Rules
- Never modify an existing migration file after it has been applied
- Always write a matching `.down.sql` that perfectly reverses the `.up.sql`
- Each migration does ONE thing (one table, one index, etc.)

## Performance

### Checking query performance
```sql
EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
SELECT ... FROM pings WHERE ST_DWithin(...);
```
Target: spatial queries should complete in **< 100ms**.

### Common optimizations
- Partial index on active pings only: `WHERE is_active = TRUE`
- JSONB index for avatar_config if queried: `CREATE INDEX ON users USING GIN(avatar_config)`
- Use `LIMIT` on all queries returning map data

## Rules
- Primary keys: always `UUID` with `gen_random_uuid()` — never SERIAL
- Timestamps: always `TIMESTAMPTZ` (with timezone), never `TIMESTAMP`
- Soft deletes only: use `is_active = FALSE`, never `DELETE` on user-visible data
- Foreign keys: always include `ON DELETE CASCADE` or `ON DELETE SET NULL` explicitly
