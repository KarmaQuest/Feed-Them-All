---
name: go-backend
description: Expert Go backend development for FeedThemAll. Use this skill for all Go code in backend/ — HTTP handlers, middleware, WebSockets, database queries, migrations, and service logic.
---

# Go Backend — FeedThemAll

## When to Use This Skill
- Writing or reviewing any Go file in `backend/`
- Creating HTTP handlers, middleware, or routing
- Implementing WebSocket logic (`gorilla/websocket`)
- Writing PostgreSQL queries with `pgx`
- Creating or running database migrations with `golang-migrate`
- Structuring Go packages and interfaces

## Project Structure
```
backend/
├── cmd/api/          # main.go — entry point, wires everything together
├── internal/
│   ├── auth/         # JWT issue/validate, bcrypt, refresh tokens
│   ├── pings/        # CRUD pings, ST_DWithin queries, media upload
│   ├── users/        # user profiles, avatar config, XP
│   ├── gamification/ # award_xp(), badge unlock jobs
│   └── websocket/    # hub, clients, bounding box subscriptions
├── migrations/       # numbered SQL files (golang-migrate format)
└── uploads/          # served as static files locally
```

## Code Conventions

### Formatting & Linting
- All code must pass `gofmt` — run before every commit
- Use `golangci-lint` for linting (config at root `.golangci.yml`)
- Never ignore errors: `_ = err` is forbidden unless explicitly justified with a comment

### Package Structure
- **Thin handlers**: HTTP handlers only parse input, call a service, return response
- **Service layer**: all business logic lives in `internal/<package>/service.go`
- **Repository layer**: all SQL queries in `internal/<package>/repository.go`
- Interfaces defined in the package that uses them (not where implemented)

### Error Handling
```go
// ✅ Correct
if err != nil {
    return fmt.Errorf("pings.Create: %w", err)
}

// ❌ Wrong
if err != nil {
    log.Fatal(err)
}
```

### Logging
Use `log/slog` (stdlib Go 1.21+):
```go
slog.Info("ping created", "id", ping.ID, "user_id", userID)
slog.Error("db query failed", "err", err)
```
Never log personal data (email, GPS coordinates at full precision).

### Environment Variables
```go
// Load with os.Getenv — never hardcode secrets
dbURL := os.Getenv("DATABASE_URL")
if dbURL == "" {
    log.Fatal("DATABASE_URL is required")
}
```

### HTTP Handlers
```go
// Pattern: parse → validate → call service → respond
func (h *Handler) CreatePing(w http.ResponseWriter, r *http.Request) {
    var req CreatePingRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    ping, err := h.svc.Create(r.Context(), req)
    if err != nil {
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(ping)
}
```

### JWT Auth
- Access token: 15 min expiry, signed with `JWT_SECRET`
- Refresh token: 7 days, HttpOnly cookie, signed with `JWT_REFRESH_SECRET`
- Middleware adds `userID` to context: `r.Context().Value(ctxKeyUserID)`

### Rate Limiting
Use `golang.org/x/time/rate` — apply to `/auth/register`, `/auth/login`, and ping creation:
```go
limiter := rate.NewLimiter(rate.Every(time.Second), 5) // 5 req/sec
```

### WebSocket (gorilla/websocket)
- Each client registers with a bounding box `{minLat, maxLat, minLon, maxLon}`
- Hub broadcasts only to clients whose bounding box contains the event location
- Disconnect: always call `hub.Unregister(client)` in a deferred function
- Max GPS push rate per client: 1 message/second (server-side enforcement)

### Database (pgx + PostGIS)
```go
// Proximity query — always use GEOGRAPHY type for accuracy
const nearbyPings = `
    SELECT id, type, ST_Y(location::geometry) as lat, ST_X(location::geometry) as lon
    FROM pings
    WHERE is_active = true
    AND ST_DWithin(location, ST_MakePoint($1, $2)::geography, $3)
`
// $1 = longitude, $2 = latitude, $3 = radius in meters
```

### Migrations (golang-migrate)
```
migrations/
├── 000001_init.up.sql
├── 000001_init.down.sql
├── 000002_add_xp_actions.up.sql
└── 000002_add_xp_actions.down.sql
```
Never modify an existing migration — always create a new one.

## Security Rules
- Validate all inputs server-side (never trust client)
- MIME-type check on uploads: read first 512 bytes with `http.DetectContentType`
- Max upload size: 10 MB (`http.MaxBytesReader`)
- Round GPS coordinates to 4 decimal places (~11m precision) before storing/returning publicly
- UUID primary keys only — never expose sequential integer IDs

## Testing
```bash
go test ./...                    # Run all tests
go test ./internal/auth/... -v   # Specific package
```
- Unit tests: `_test.go` files alongside source
- Integration tests: tag with `//go:build integration`, use Docker PostgreSQL
- Use `github.com/stretchr/testify` for assertions
