# AGENTS.md — Oxmail

## Project

Oxmail is a dockerized mail server with a modern beautiful UI. Open-source, targeting developers and small teams.

- **Repo**: https://github.com/MYusufEka/oxmail.git
- **Plan**: `.sisyphus/plans/dockerized-mail-server-beautiful.md` (38 tasks, 6 waves)
- **Status**: Greenfield — no code yet, plan approved

## Architecture

```
┌─────────────────────────────────────────────────────┐
│ Docker Compose                                       │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐          │
│  │ Postfix  │  │ Dovecot  │  │ Rspamd   │  ← Mail  │
│  └────┬─────┘  └────┬─────┘  └──────────┘          │
│       │              │                               │
│  ┌────┴──────────────┴─────┐                        │
│  │  oxmail-api (Go)        │  ← Control plane       │
│  │  - REST API             │                        │
│  │  - Config generation    │                        │
│  │  - WebSocket logs       │                        │
│  └────────────┬────────────┘                        │
│               │                                      │
│  ┌────────────┴────────────┐                        │
│  │  web (Next.js)          │  ← UI                  │
│  │  - Admin dashboard      │                        │
│  │  - Webmail              │                        │
│  └─────────────────────────┘                        │
│                                                      │
│  ┌──────────┐                                       │
│  │  Redis   │  ← Cache/session                      │
│  └──────────┘                                       │
└─────────────────────────────────────────────────────┘
```

- Go API is the **config source of truth** — generates Postfix/Dovecot configs, never edit them directly
- Mail engine: wrap Postfix+Dovecot+Rspamd ONLY — no custom SMTP/IMAP engine
- State: SQLite (no extra container)
- Two modes: `dev` (no DNS, local ports, relaxed auth) and `prod` (TLS, outbound, real ports)

## Directory Structure (planned)

```
cmd/oxmail-api/    Go API server binary
cmd/oxmail/        CLI tool (cobra)
internal/          Go packages (api, domain, config, mail, logs)
web/               Next.js App Router (shadcn/ui + Tailwind)
docker/            Dockerfiles (postfix, dovecot, rspamd, traefik)
configs/           Config templates (.tmpl)
scripts/           Shell scripts (seed, wait-healthy, test-e2e)
api-spec/          OpenAPI 3.1 spec
docs/              Documentation
tests/e2e/         Playwright E2E tests
```

## Tech Stack

| Layer | Tech |
|-------|------|
| Backend API | Go 1.22+, chi router, slog, testify, SQLite |
| Frontend | Next.js 14+ App Router, TypeScript, shadcn/ui, Tailwind, TanStack Query/Table |
| Mail | Postfix (SMTP), Dovecot (IMAP/LMTP), Rspamd (spam) |
| Infra | Docker Compose, Redis |
| TLS (prod) | Traefik or Caddy (Let's Encrypt) |
| Test | Go test, Vitest, Playwright |
| Lint | golangci-lint, ESLint, tsc --noEmit |

## Commands (planned)

```bash
make dev          # Start full stack in dev mode
make test         # Run all tests (Go + Vitest + Playwright)
make test-go      # Go tests only
make test-web     # Vitest only
make test-e2e     # Playwright E2E
make lint         # All linters
make build        # Build all images + binaries
make seed         # Seed test data (domain + users + emails)
make reset        # Wipe + re-seed
make prod         # Start in production mode (TLS + outbound)
make logs         # Tail all container logs
```

## Constraints (hard rules)

- No `as any`, `@ts-ignore`, `@ts-expect-error` in TypeScript
- No empty catch blocks
- No console.log in production code
- No commented-out code
- No generic variable names (data, result, item, temp)
- No custom SMTP/IMAP engine — wrap Postfix+Dovecot only
- No custom spam filter — configure Rspamd only
- TDD: write tests first, then implement

## Design

- **Visual**: Clean minimal dark-first (Linear/Vercel-like) + subtle cyber accents
- **Layout**: Collapsible left sidebar, top command palette (Cmd+K), main content
- **Components**: shadcn/ui, Lucide icons, TanStack Table for data grids
- **Patterns**: Command palette, keyboard shortcuts, inline row actions, real-time WebSocket logs, toast notifications, optimistic updates

## Environment Variables

All prefixed with `OXMAIL_`:
- `OXMAIL_MODE` — `dev` or `production`
- `OXMAIL_DOMAIN` — mail domain (e.g., `local.test`)
- `OXMAIL_ADMIN_PASSWORD` — admin password (first-run setup)
- `OXMAIL_API_PORT` — API port (default: 8080)
- `OXMAIL_WEB_PORT` — Web UI port (default: 3000)
- `OXMAIL_SMTP_PORT` — SMTP port (default: 1025 dev, 25 prod)
- `OXMAIL_IMAP_PORT` — IMAP port (default: 1143 dev, 993 prod)
- `OXMAIL_PUBLIC_IP` — public IP (production only)
- `OXMAIL_ACME_EMAIL` — Let's Encrypt email (production only)

## Ports (dev mode)

| Port | Service |
|------|---------|
| 1025 | SMTP (Postfix) |
| 1143 | IMAP (Dovecot) |
| 8080 | API (Go) |
| 3000 | Web UI (Next.js) |

## Working with the Plan

The plan at `.sisyphus/plans/dockerized-mail-server-beautiful.md` contains:
- 38 detailed tasks with acceptance criteria and QA scenarios
- Dependency matrix and parallel execution waves
- Each task specifies: what to do, must NOT do, agent category, references, commit message
- Follow task order within waves; tasks in same wave can run in parallel
- Every task requires TDD: write failing test → implement → verify
