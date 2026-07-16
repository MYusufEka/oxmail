# AGENTS.md — Oxmail

Dockerized mail server (Postfix + Dovecot + Rspamd) with Go API + Next.js UI.
6 containers, ~1.2 GB idle. **Code exists** — build on what's here, don't restart from zero.

## Quick Reference

```bash
make dev          # Full stack in dev mode (builds + starts)
make test         # All tests (Go + Vitest + E2E)
make test-go      # cd cmd/oxmail-api && go test ./... && cd cmd/oxmail && go test ./... && go test ./internal/...
make test-web     # cd web && npm test (Vitest)
make lint         # golangci-lint && cd web && npm run lint
make seed         # ./scripts/seed.sh (creates domain + 2 users + 5 test emails + DKIM)
make reset        # docker compose down -v + dev up + seed
make logs         # docker compose logs -f
```

## Architecture

```
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Postfix  │  │ Dovecot  │  │ Rspamd   │ ← Mail layer
└────┬─────┘  └────┬─────┘  └──────────┘
     │              │
┌────┴──────────────┴─────┐
│  oxmail-api (Go :8080)  │ ← Config source of truth (SQLite)
│  - chi router + JWT     │   Generates Postfix/Dovecot configs
│  - WebSocket log stream │   Never edit mail configs directly
└────────────┬────────────┘
             │
┌────────────┴────────────┐
│  web (Next.js :3000)    │ ← Dark-first UI (shadcn/ui + Tailwind v4)
│  - Dashboard            │
│  - Domain/User/Alias    │
│  - Webmail              │
│  - Log streaming        │
└─────────────────────────┘
```

## Project Structure

```
cmd/oxmail-api/main.go     Go API entrypoint (slog JSON handler, port from OXMAIL_API_PORT)
cmd/oxmail/                CLI tool (Cobra), talks to API via HTTP
internal/api/              HTTP handlers (chi), one file per resource + server.go
internal/api/middleware/   JWTAuth + SecurityHeaders middleware
internal/config/           Postfix/Dovecot config generators
internal/database/         SQLite via modernc.org/sqlite, embed.FS migrations
internal/domain/           Business logic (services + models + error vars)
internal/health/           Service health checkers (Postfix/Dovecot/Rspamd/Redis)
internal/logs/             Log collection (buffer, collector, parser)
internal/mail/             Postfix/Dovecot management + IMAP bridge + SMTP sender
web/src/                   Next.js App Router (src/ layout)
web/src/app/               Route pages (domains, users, aliases, mail, logs, dkim, production)
web/src/hooks/             TanStack Query hooks (use-domains, use-users, etc. with optimistic updates)
web/src/lib/               api-client, query-client, schemas (Zod), utils (cn)
web/src/types/             TypeScript API types
web/src/components/        UI (shadcn/ui) + layout (app-shell, sidebar, topbar) + command-palette
docker/                    Dockerfiles (postfix, dovecot, rspamd, api) + traefik configs
configs/                   Go template files (.tmpl) for Postfix, Dovecot, Rspamd
configs/postfix/           main.cf.tmpl, master.cf.tmpl
configs/dovecot/           dovecot.conf.tmpl, 10-auth.conf.tmpl, 10-mail.conf.tmpl
configs/rspamd/local.d/    worker-normal.inc, worker-controller.inc
scripts/                   seed.sh, wait-healthy.sh, test-e2e.sh
tests/e2e/                 Shell-based E2E tests (api, lifecycle, persistence, security)
```

## Go Backend

### Entrypoint
- `cmd/oxmail-api/main.go` — opens SQLite, calls `api.NewServer(db.Conn)`, then `ListenAndServe(port)`
- Logger: `slog.NewJSONHandler(os.Stdout, slog.LevelInfo)` — always JSON, always stdout

### Dependencies
- `github.com/go-chi/chi/v5` — router
- `github.com/golang-jwt/jwt/v5` — JWT auth
- `modernc.org/sqlite` — pure Go SQLite (no CGO)
- `github.com/emersion/go-imap/v2` — IMAP bridge to Dovecot
- `github.com/gorilla/websocket` — log streaming
- `github.com/microcosm-cc/bluemonday` — HTML sanitization
- `github.com/stretchr/testify` — tests (assert, require)

### Go workspace
Go 1.25 workspace at root with 3 modules:
- `github.com/MYusufEka/oxmail` (root — internal packages)
- `github.com/MYusufEka/oxmail/cmd/oxmail` (CLI, deps: cobra + fatih/color)
- `github.com/MYusufEka/oxmail/cmd/oxmail-api` (API server)

**Test commands must run from specific dirs** (Makefile does this): `cd cmd/oxmail-api && go test ./...`, `cd cmd/oxmail && go test ./...`, `go test ./internal/...`

### Router setup (`internal/api/server.go`)
1. Global middleware: RequestID, RealIP, Recoverer, Timeout(30s), SecurityHeaders(CORS)
2. Public routes: `GET /health`, `POST /api/auth/login`
3. Protected group (JWT required unless dev mode):
   - `POST/GET /api/domains`, `GET/DELETE /api/domains/{name}`, `GET /api/domains/{name}/health`
   - `POST/GET /api/domains/{domainID}/users`, `DELETE .../{userID}`
   - `POST/GET /api/domains/{domainID}/aliases`, `DELETE .../{aliasID}`
   - `GET /api/mail/{userID}/inbox`, `GET/PATCH/DELETE .../messages/{msgID}`
   - `POST /api/mail/send`, `GET /api/mail/search?q=&user=`
   - `GET /api/logs`, `WS /api/logs/stream`
   - `GET /api/health`
   - `GET/POST /api/dkim/{domain}`
   - `GET /api/dns/records`, `GET /api/dns/check`
   - `GET/POST /api/contacts`, bounces, stats, audit routes
   - `POST /api/sieve/{email}`, autodiscover routes
4. Dev-only routes (when `OXMAIL_MODE=dev`): test endpoints via devHandler

**Domain routes routing quirk**: `GET /api/domains/{name}` and `GET /api/domains/{name}/health` conflict with `GET /api/domains/{domainID}/users` when both use `r.Route("/api/domains", ...)`. Fix: split into `RegisterRoutes` (collection: `POST/GET /`) and `RegisterNameRoutes` (named: flat `r.Get("/api/domains/{name}", ...)` without nesting). Both called in server.go after each other.

### Dev mode auth
JWT middleware checks `os.Getenv("OXMAIL_MODE") == "dev"` — if dev, **all requests pass through without token**. Prod mode requires `Authorization: Bearer <token>`.

### Error response format
Handlers return `{"error": {"code": "ERROR_CODE", "message": "Human message"}}` with appropriate HTTP status. Error variables defined per package (e.g. `domain.ErrDomainExists`).

### Handler pattern
Each resource has a handler struct with `RegisterRoutes(r chi.Router)` method. Tests use `handler.Router()` to get sub-router for httptest. Domain/User/Alias handlers trigger Postfix/Dovecot config regeneration on mutations.

### Database
- SQLite with WAL mode (`PRAGMA journal_mode=WAL`)
- 9 migration files embedded via `//go:embed migrations/*.sql`
- Tables: `schema_migrations`, `domains`, `users`, `aliases`, `dkim_keys`, `contacts`, `bounces`, `stats`, `audit_log`
- Foreign keys: `users.domain_id → domains.id` (ON DELETE RESTRICT), `aliases.domain_id → domains.id` (ON DELETE CASCADE)
- `database.DB` struct wraps `*sql.DB` with logger

### Config generation
Go API **is the config source of truth**. On domain/user/alias mutations:
- `PostfixDomainsGenerator` writes `/etc/postfix/virtual_domains`
- `PostfixAliasesGenerator` writes `/etc/postfix/virtual_aliases`
- `DovecotUsersGenerator` writes `/etc/dovecot/passwd`
- Then `PostfixManager.ApplyDomainConfig()` / `DovecotManager.ApplyUserConfig()` reloads services
- Config templates use `text/template` (`.tmpl` files in `configs/`)

**DKIM keys ARE persisted**: `DKIMService.Generate()` does `INSERT OR REPLACE INTO dkim_keys` (line ~103). Keys survive restart. `loadFromDB()` re-hydrates in-memory cache on startup. ~~Stale claim~~ corrected.

### Tests
- Table-driven tests with testify
- Every service has test file alongside it (no separate `_test` dirs)
- API handlers use `httptest` with handler's `Router()` method
- Config generators test output content
- Log parser tests parsing patterns
- All use in-memory SQLite (`database.Open(":memory:", logger)`) — no mocking of DB layer
- `AliasService` takes `*sql.DB` directly; `DomainService`/`UserService` take `*database.DB`

## Web Frontend

### Stack
- **Next.js 16.2.6** — App Router, output: standalone (Docker)
- **React 19.2.4** — client components use `"use client"` directive
- **Tailwind v4** — CSS config in `src/app/globals.css`, PostCSS via `@tailwindcss/postcss`
- **shadcn/ui** — new-york style, components in `src/components/ui/`
- **TanStack Query v5** — hooks with optimistic updates
- **Zod v4** — schemas in `src/lib/schemas.ts`
- **Vitest v4** — jsdom environment, setup: `src/__tests__/setup.ts`
- **TypeScript** — strict mode, path alias `@/*` → `./src/*`

### Route pages
| Route | Component | Data |
|-------|-----------|------|
| `/` | Dashboard | KPI cards + ServiceHealthGrid + RecentActivity |
| `/domains` | DomainsPage | DomainTable, Add/Delete dialogs |
| `/users` | UsersPage | Domain selector + UserTable |
| `/aliases` | AliasesPage | Static placeholder (not fully implemented) |
| `/mail` | MailPage | MessageList, MessagePreview, Compose dialog |
| `/logs` | LogsPage | Log entries + WebSocket stream |
| `/dkim` | DKIMPage | Domain DKIM cards |
| `/production` | ProductionPage | DNS wizard, production settings |

### API client pattern
`src/lib/api-client.ts` exports `apiClient` object with typed methods. Each returns typed Promise. Uses `NEXT_PUBLIC_API_URL` env var (default `http://localhost:8080`). Custom `ApiError` class with status, code, message.

**No auth headers** in `apiClient` — frontend auth not wired (T30 pending). Some mail sub-pages (`filters`, `vacation`, `signature`) still use `DEFAULT_EMAIL = "alice@local.test"` as fallback. Replace when auth lands.

### Hook pattern (optimistic updates)
Each CRUD resource has 3 hooks: `useX()`, `useCreateX()`, `useDeleteX()`. Create/Delete hooks:
1. `onMutate`: cancel queries, snapshot previous data, apply optimistic update
2. `onError`: rollback to snapshot
3. `onSettled`: invalidate queries

Query keys: `["domains", params]`, `["users", domainId, params]`, `["aliases", domainId, params]`, `["inbox", userId, params]`, `["logs", params]`, `["health"]`, `["dkim", domain]`, `["dns", "records"]`, `["dns", "check"]`.

Notable: `useDnsCheck` has `enabled: false` (manual trigger). `useHealth` polls every 10s.

### WebSocket logs
`useLogStream()` hook connects to `ws://<api>/api/logs/stream`. Auto-reconnects every 3s. Returns `{ entries, connected, clearEntries }`.

### Aliases page
`/aliases` is a **static placeholder** — Add Alias button is disabled, no hooks wired. Not fully implemented.

### Mail rich editor
`rich-editor.tsx` uses `document.execCommand` (deprecated by browsers). Functional but may break in future browser versions.

### Default dark mode
HTML tag has class `dark` hardcoded in root layout. Font: Geist (sans) + Geist Mono. No theme toggle.

## Docker Infrastructure

### Services (6+1)
| Service | Image | Exposed | Memory |
|---------|-------|---------|--------|
| redis | redis:7-alpine | — | 64M |
| postfix | custom (alpine) | 25, 587 | 256M |
| dovecot | custom (alpine) | 143 | 256M |
| rspamd | custom (alpine) | — | 384M |
| oxmail-api | custom (golang:1.25 → alpine) | 8080 | 128M |
| web | custom (node:20 → alpine) | 3000 | 256M |
| traefik | traefik:v3.0 | 80, 443 | 128M (prod only) |

### Networks
- `oxmail-internal`: `internal: true` — no external access
- `oxmail-external`: for web, API, and mail services that need inbound

### Volumes
- `mail-data`, `config-data`, `dkim-keys`, `redis-data`, `postfix-spool`, `rspamd-data`
- Prod adds `letsencrypt-data`

### Dev mode overrides (`docker-compose.dev.yml`)
- Mounts source code as volumes for hot reload (Go: `./cmd` + `./internal`, Next.js: `./web/src`)
- `NEXT_PUBLIC_API_URL=http://localhost:8080` (points to host port)

### Dockerfiles
- **api**: Multi-stage Go build (CGO_ENABLED=0, `-ldflags="-s -w"`), distroless alpine runtime
- **web**: Node 20-alpine deps → Next.js standalone build → node:20-alpine runner with `nextjs` user
- **Postfix/Dovecot/Rspamd**: Alpine-based, each has healthcheck.sh

## CLI (`cmd/oxmail`)

Cobra CLI with subcommands: `domain`, `user`, `alias`, `status`, `logs`, `send-test`.
Global flags: `--json` (machine output), `--api-url` (env: `OXMAIL_API_URL`).
HTTP client talks to API at `oxmail-api:8080` inside Docker network.

## Environment Variables

All prefixed `OXMAIL_`. Defaults in `.env.example`:
- `OXMAIL_MODE` — `dev` (relaxed auth, local ports) or `production` (TLS, real ports)
- `OXMAIL_DOMAIN` — primary mail domain (default `local.test`)
- `OXMAIL_ADMIN_PASSWORD` — min 8 chars (default `changeme123`)
- `OXMAIL_API_PORT` / `OXMAIL_WEB_PORT` / `OXMAIL_SMTP_PORT` / `OXMAIL_IMAP_PORT`
- `OXMAIL_REDIS_URL` — default `redis://redis:6379`
- `OXMAIL_PUBLIC_IP` / `OXMAIL_ACME_EMAIL` — production only
- `OXMAIL_JWT_SECRET` — if unset, random on each restart (tokens invalidated)
- `OXMAIL_WEB_URL` — CORS origin (defaults to `*` in dev)
- `OXMAIL_DB_PATH` — SQLite path (default `oxmail.db`)

## Testing

### Go tests
Run from project root: `go test ./internal/...` (handlers run via sub-routers).
Table-driven with `testify/assert` and `testify/require`.
API handler tests use `httptest.NewServer(handler.Router())`.

### Frontend tests (Vitest)
```bash
cd web && npm test
```
jsdom environment, setup in `src/__tests__/setup.ts`. `@/` alias configured.

### E2E tests (shell scripts)
```bash
./scripts/test-e2e.sh  # runs tests/e2e/run-all.sh
```
Shell-based, test API endpoints directly with curl.
Separate suites: api_test, lifecycle_test, persistence_test, security_test.

## Docs

Key files in `docs/`:
- `docs/architecture.md` — system data flows, volume table, network zones, design decisions
- `docs/api.md` — curl examples for every endpoint, pagination, error format, WebSocket usage
- `docs/quickstart.md` — 6-step local setup walkthrough with expected output
- `docs/production.md` — VPS deployment (DNS records, backup/restore, DKIM rotation, troubleshooting)

## Constraints (hard rules)

- **No `as any`, `@ts-ignore`, `@ts-expect-error`** in TypeScript
- **No empty catch blocks**
- **No `console.log`** in production code (frontend)
- **No commented-out code**
- **No generic variable names** (data, result, item, temp)
- **No custom SMTP/IMAP engine** — wrap Postfix+Dovecot only
- **No custom spam filter** — configure Rspamd only
- **TDD**: write tests first, then implement
- **Go API is config source of truth** — never edit Postfix/Dovecot configs directly

## Design conventions

- Dark-first, Linear/Vercel-inspired with subtle cyber accents
- Collapsible left sidebar, top Cmd+K command palette, main content area
- shadcn/ui components, Lucide icons, TanStack Table for data grids
- Toast notifications via `sonner`, optimistic updates via TanStack Query
- Every list page has 4 states: loading (skeleton), empty, error, data
- Test IDs on interactive elements (`data-testid="..."`)
