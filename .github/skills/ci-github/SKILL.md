---
name: ci-github
description: GitHub Actions CI/CD and Git workflow for FeedThemAll. Use for writing or fixing workflow YAML files, branch strategy, PR rules, and commit conventions.
---

# CI / GitHub Workflow — FeedThemAll

## When to Use This Skill
- Writing or fixing `.github/workflows/` YAML files
- Debugging failing GitHub Actions pipelines
- Following the branch and PR workflow
- Writing commit messages
- Setting up branch protection rules

## Branch Strategy
```
main          ← Production (protected — merge via PR only, requires CI green)
dev           ← Integration (protected — merge via PR only)
feature/*     ← New features (ex: feature/ping-creation)
fix/*         ← Bug fixes   (ex: fix/websocket-disconnect)
chore/*       ← Tooling, deps, config
```

**Rules:**
- Never push directly to `main` or `dev`
- Branch from `dev`, PR back to `dev`
- `dev` → `main` only when a milestone is complete and all tests pass

## Commit Messages (Conventional Commits)
```
<type>(<scope>): <short description>

Types: feat | fix | chore | docs | refactor | test | style
Scope: backend | frontend | mobile | db | ci | design

Examples:
feat(backend): add ST_DWithin proximity endpoint
fix(frontend): correct avatar sprite direction mapping
chore(ci): add golangci-lint step to workflow
test(backend): add integration tests for auth package
```

## GitHub Actions Workflows

### CI — Backend (Go)
```yaml
# .github/workflows/ci-backend.yml
name: CI Backend
on:
  push:
    branches: [dev]
    paths: ['backend/**']
  pull_request:
    branches: [dev, main]
    paths: ['backend/**']

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgis/postgis:15-3.3
        env:
          POSTGRES_USER: fta
          POSTGRES_PASSWORD: fta
          POSTGRES_DB: feedthemall_test
        ports: ['5432:5432']
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Install golangci-lint
        uses: golangci/golangci-lint-action@v6

      - name: Run tests
        working-directory: backend
        env:
          DATABASE_URL: postgres://fta:fta@localhost:5432/feedthemall_test?sslmode=disable
        run: go test ./...
```

### CI — Frontend (React)
```yaml
# .github/workflows/ci-frontend.yml
name: CI Frontend
on:
  push:
    branches: [dev]
    paths: ['frontend-web/**']
  pull_request:
    branches: [dev, main]
    paths: ['frontend-web/**']

jobs:
  lint-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend-web/package-lock.json

      - run: npm ci
        working-directory: frontend-web

      - run: npm run lint
        working-directory: frontend-web

      - run: npm run test -- --run
        working-directory: frontend-web
```

## PR Rules
- PR title must follow Conventional Commits format
- PRs must include a description of what changed and how to test
- At least one reviewer required before merge (when collaborating)
- CI must be green before merge — no exceptions

## Debugging Failing CI

### Go test failures
1. Check if the `postgres` service started: look for `pg_isready` in health check logs
2. Ensure `DATABASE_URL` env var is set in the job's `env:` block
3. Run `go test -v ./...` locally first to reproduce

### golangci-lint failures
```bash
# Run locally to see exact errors
golangci-lint run ./...
# Auto-fix what's possible
golangci-lint run --fix ./...
```

### Node/npm failures
```bash
# Use 'npm ci' (not 'npm install') in CI — it uses package-lock.json exactly
# If cache is stale, delete the cache key and re-run
```

## .gitignore (critical entries)
```
# Backend
backend/.env
backend/uploads/
backend/feedthemall   # compiled binary

# Frontend
frontend-web/node_modules/
frontend-web/dist/
frontend-mobile/node_modules/

# General
*.env
*.env.local
.DS_Store
```
