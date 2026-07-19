# AGENTS.md — Oxmail

Dockerized mail server: Postfix + Dovecot + Rspamd + Redis + Go API + Next.js UI. Code exists; extend current patterns, do not rebuild from scratch.

## Agent workflow

- Do not use more than 2 subagents for one user request.
- Prefer verified executable sources (`Makefile`, `go.work`, `web/package.json`, compose files) over README prose when they conflict.
- For frontend work under `web/`, also read `web/AGENTS.md`: Next.js 16 has breaking changes; check `node_modules/next/dist/docs/` before changing Next APIs.

## Commands

```bash
make dev          # creates .env, then references docker-compose.dev.yml; verify file exists before relying on it
make test         # test-go + test-web + test-e2e
make test-go      # cd cmd/oxmail-api && go test ./...; cd cmd/oxmail && go test ./...; go test ./internal/...
make test-web     # cd web && npm test (Vitest run, jsdom)
make test-e2e     # ./scripts/test-e2e.sh -> tests/e2e/run-all.sh
make lint         # golangci-lint run ./...; cd web && npm run lint
make reset        # docker compose down -v; dev stack up --build; ./scripts/seed.sh
make prod         # prod profile + docker-compose.prod.yml + TLS/Traefik
```

Focused checks:

```bash
cd web && npm run lint
cd web && npm run build
cd web && npm test -- path/to/file.test.tsx
go test ./internal/api -run TestName
cd cmd/oxmail && go test ./...
```

E2E writes evidence to `.sisyphus/evidence/e2e/` and runs `lifecycle_test.sh`, `persistence_test.sh`, `security_test.sh`, `api_test.sh`.

## Repo boundaries

- Go workspace has 3 modules (`go.work`): root `github.com/MYusufEka/oxmail`, `cmd/oxmail-api`, `cmd/oxmail`. Run Go package tests from module dirs as Makefile does.
- API entrypoint: `cmd/oxmail-api/main.go`; it opens SQLite from `OXMAIL_DB_PATH` (default `oxmail.db`) and calls `api.NewServer(db.Conn)`.
- Router wiring lives in `internal/api/server.go`; handlers live one resource per file under `internal/api/`.
- CLI lives in `cmd/oxmail` (Cobra) and talks to API via HTTP; global flags include `--json` and `--api-url` / `OXMAIL_API_URL`.
- Web app lives in `web/src` (App Router). API client is `web/src/lib/api-client.ts`; query hooks are in `web/src/hooks`; auth hook `useAuth()` lives in `web/src/contexts/auth.tsx` and exposes `login`/`logout`/`refresh`/`mustChangePassword`.

## Backend gotchas

- Go API is config source of truth. Do not edit live Postfix/Dovecot config directly; update domain/user/alias services or generators under `internal/config` / `internal/mail`.
- Startup renders templates to `/etc/oxmail` via `config.RenderAll` in `internal/api/server.go`.
- Domain/user/alias mutations trigger Postfix/Dovecot config regeneration and service reloads.
- `OXMAIL_MODE=dev` bypasses protected-route JWT middleware and enables `/api/dev/send-test`; production requires JWT.
- If `OXMAIL_JWT_SECRET` is unset, server generates random secret and tokens break after restart.
- Domain routes have a Chi ordering conflict: keep collection routes (`RegisterRoutes`), domain-scoped routes (`RegisterDomainScopedRoutes`), then flat name routes (`RegisterNameRoutes`) as wired in `internal/api/server.go`.
- Database migrations are embedded from `internal/database/migrations/*.sql`; current set is `001` through `012`.

## Database tables

Migrations are embedded from `internal/database/migrations/*.sql`. Notable tables and columns:

- `dkim_keys` ... persisted DKIM key material per domain/selector (migration `005`); written by `internal/domain/dkim_service.go` via `INSERT OR REPLACE INTO dkim_keys`.
- `signatures` ... per-user mail signatures (migration `010`).
- `users.must_change_password` ... forces password change on next login (migration `011`).
- `contacts` ... unique index on owner + email (migration `012`).

## API routes (notable)

- `POST /api/auth/login`, `POST /api/auth/logout`, `GET /api/auth/me` ... session lifecycle; `/auth/me` returns `mustChangePassword`.
- `GET|POST|DELETE /api/mail/signature/{email}` ... per-user signature CRUD.
- `GET /api/domains`, `POST /api/domains`, `GET /api/users`, `POST /api/users`, `GET /api/aliases` ... resource collections.
- `WS /api/logs/stream`, `GET /api/health`, `POST /api/dev/send-test` (dev only).

## Frontend gotchas

- Next.js is `16.2.6`, React `19.2.4`, Tailwind v4, Vitest v4. Do not assume older Next.js behavior.
- `web/next.config.ts` uses `output: "standalone"` for Docker.
- `NEXT_PUBLIC_API_URL` defaults to `http://localhost:8080`; log streaming derives WebSocket URL by replacing `http` with `ws`.
- `api-client.ts` sends every request with `credentials: "include"`; the API sets the auth token in an httpOnly cookie, never in localStorage.
- Vitest uses jsdom and `web/src/__tests__/setup.ts`; `@/` aliases to `web/src`.
- UI conventions: dark-first, shadcn/Radix components, TanStack Query/Table, `sonner`, Lucide icons, interactive elements need `data-testid`.

## Docker/runtime

- Main compose exposes SMTP `${OXMAIL_SMTP_PORT:-1025}:25`, IMAP `${OXMAIL_IMAP_PORT:-1143}:143`, API `${OXMAIL_API_PORT:-8080}:8080`, web `${OXMAIL_WEB_PORT:-3000}:3000`.
- `docker-compose.yml` forces `OXMAIL_MODE=dev` for `oxmail-api`; production uses `docker-compose.prod.yml` and `OXMAIL_MODE=production`.
- API container mounts `/var/run/docker.sock:ro` because managers reload Postfix/Dovecot via Docker exec.
- Internal network `oxmail-internal` is marked `internal: true`; web/API/mail exposure goes through `oxmail-external`.

## Style constraints

- TypeScript: no `as any`, `@ts-ignore`, or `@ts-expect-error`.
- No empty catch blocks; if swallowing non-JSON responses, leave explanatory comment like existing `api-client.ts`.
- No `console.log` in frontend production code.
- No commented-out code or placeholder TODOs in committed changes.
- Avoid vague variable names (`data`, `result`, `item`, `temp`) when domain names exist.
- Do not build custom SMTP/IMAP/spam engines; wrap/configure Postfix, Dovecot, and Rspamd.
- No localStorage for auth tokens — httpOnly cookie only.
- Tests are expected with changes; existing Go tests use `testify` and in-memory SQLite (`database.Open(":memory:", logger)`).
