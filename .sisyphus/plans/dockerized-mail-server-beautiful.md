# Oxmail - Beautiful Dockerized Mail Server

## TL;DR

> **Quick Summary**: Build Oxmail — an open-source, dockerized local mail server with a stunning modern UI (clean minimal dark-first), wrapping Postfix+Dovecot+Rspamd, orchestrated by a Go control-plane API, with Next.js frontend.
> 
> **Deliverables**:
> - Docker Compose stack (Postfix, Dovecot, Rspamd, Go API, Next.js UI, Redis)
> - Go control-plane API — `cmd/oxmail-api` (domain/user/alias CRUD, config generation, health, logs)
> - Next.js admin dashboard (domains, users, aliases, DKIM, queue, logs)
> - Next.js webmail (inbox, compose, search, thread view)
> - Real-time log viewer + health monitoring
> - CLI tool (`oxmail`) for common operations
> - Dev mode: one-command setup, no DNS required
> - Production mode: TLS, outbound delivery, DNS helper
> 
> **Estimated Effort**: XL
> **Parallel Execution**: YES - 6 waves
> **Critical Path**: Scaffolding → Go API core → Mail engine config → UI shell → Integration → Production

---

## Context

### Original Request
User wants to build a dockerized mail server with beautiful design because existing local mail servers (Mailcow, docker-mailserver, Mailu, Poste.io) all have ugly or dated UIs. No compromise on any dimension.

### Interview Summary
**Key Discussions**:
- Scope: Full platform (SMTP+IMAP+webmail+admin+monitoring+DX) — no compromise
- Backend: Wrap Postfix+Dovecot+Rspamd + custom Go control-plane (recommended by Prometheus)
- Frontend: Next.js App Router + shadcn/ui + Tailwind (recommended by Prometheus)
- Visual: Clean minimal dark-first (Linear/Vercel-like) + subtle cyber accents
- Test: TDD confirmed
- Target: OSS public quality, developers + small teams

**Research Findings**:
- Mailcow: best features but 4-6GB RAM, 15+ containers, UI performance issues
- docker-mailserver: lightest (0.5-1GB) but CLI-only
- Mailu: balanced but dated Flask UI
- Poste.io: nice UI but proprietary
- Gap: no modern open-source mail server with beautiful UI + dev-first experience

### Metis Review
**Identified Gaps** (addressed):
- MVP phasing needed → structured as waves with clear phase gates
- Security guardrails → open-relay test, auth-required admin, no secrets in frontend
- Webmail scope creep risk → locked to read/compose/search/thread (no filters/rules/calendar)
- Resource budget undefined → target <1.5GB idle RAM, ≤6 containers
- Config source of truth → Go API owns declarative config, generates Postfix/Dovecot configs

---

## Work Objectives

### Core Objective
Build Oxmail — an open-source, dockerized local mail server with a modern beautiful UI, wrapping battle-tested Postfix+Dovecot+Rspamd, orchestrated by a Go control-plane API, with Next.js frontend — targeting developers and small teams who want self-hosted email that's actually enjoyable to use.

### Concrete Deliverables
- `docker-compose.yml` with full stack
- Go API binary (`cmd/oxmail-api`)
- Next.js app (`web/`)
- CLI tool (`cmd/oxmail/`)
- Postfix/Dovecot/Rspamd config templates
- Documentation: README + quickstart

### Definition of Done
- [ ] `docker compose up -d` → all services healthy within 120s
- [ ] Create domain + user via API and UI
- [ ] Send email via SMTP → appears in IMAP inbox + webmail
- [ ] Real-time logs show SMTP event within 3s
- [ ] Restart stack → all data persists
- [ ] Open-relay test fails (security)
- [ ] Idle RAM < 1.5GB
- [ ] All TDD tests pass
- [ ] Production mode: TLS valid, outbound delivery works, DNS helper generates correct records

### Must Have
- SMTP send/receive (Postfix) with submission port
- IMAP access (Dovecot)
- Spam filtering (Rspamd)
- Go control-plane API (REST, typed, documented) — `cmd/oxmail-api`
- Next.js admin dashboard (domains, users, aliases, DKIM, queue)
- Next.js webmail (inbox, compose, search, threads)
- Real-time log viewer via WebSocket
- Health monitoring dashboard
- Docker Compose one-command setup
- Dev mode (no DNS, self-signed TLS, seeded test data)
- Dark-first clean minimal UI with command palette
- TDD test suite
- CLI tool (`oxmail`) for common ops
- Persistence across restarts
- Production mode: TLS (Let's Encrypt), outbound internet delivery, DNS helper UI, production Docker Compose profile

### Must NOT Have (Guardrails)
- Custom SMTP/IMAP engine (wrap Postfix+Dovecot ONLY)
- Custom spam filter (configure Rspamd ONLY)
- Clustering / HA / multi-node
- Kubernetes manifests
- Plugin/extension system
- Migration/import from other servers
- Calendar/Contacts/Groupware (CalDAV/CardDAV)
- S/MIME encryption
- POP3 protocol
- Internationalization (i18n)
- Over-abstraction or premature generalization
- `as any` / `@ts-ignore` in TypeScript
- Console.log in production code
- Commented-out code
- Generic variable names (data, result, item, temp)
- Empty catch blocks

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: NO (greenfield)
- **Automated tests**: TDD (test-first)
- **Framework**: Go: `go test` + testify; Frontend: Vitest + Testing Library; E2E: Playwright
- **TDD flow**: Each task follows RED (failing test) → GREEN (minimal impl) → REFACTOR

### QA Policy
Every task MUST include agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Frontend/UI**: Playwright - navigate, interact, assert DOM, screenshot
- **API/Backend**: curl/httpie - send requests, assert status + response fields
- **SMTP**: swaks - send test emails, verify delivery
- **IMAP**: scripted IMAP client (Go test or curl-like)
- **Docker**: docker compose commands - health, logs, restart
- **CLI**: bash - run oxmail commands, validate output

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - scaffolding + config + types):
├── Task 1: Monorepo scaffolding + Docker Compose skeleton [quick]
├── Task 2: Go API project setup + test infrastructure [quick]
├── Task 3: Next.js project setup + design system tokens [visual-engineering]
├── Task 4: Shared types/schemas (API contracts, domain models) [quick]
├── Task 5: Postfix+Dovecot+Rspamd Docker images + base configs [unspecified-high]
└── Task 6: Dev environment tooling (Makefile, scripts, .env) [quick]

Wave 2 (Core engine - Go API + mail config generation):
├── Task 7: Go API - domain CRUD + config generation (depends: 2, 4) [deep]
├── Task 8: Go API - user/mailbox CRUD + password hashing (depends: 2, 4) [deep]
├── Task 9: Go API - alias management (depends: 2, 4) [unspecified-high]
├── Task 10: Go API - DKIM key generation + management (depends: 2, 4) [unspecified-high]
├── Task 11: Go API - health check aggregator (depends: 2, 5) [unspecified-high]
├── Task 12: Go API - real-time log streaming via WebSocket (depends: 2, 5) [deep]
├── Task 13: Postfix dynamic config integration (depends: 5, 7) [deep]
└── Task 14: Dovecot dynamic config integration (depends: 5, 8) [deep]

Wave 3 (UI shell + API client):
├── Task 15: UI shell - layout, navigation, command palette (depends: 3) [visual-engineering]
├── Task 16: UI - API client + TanStack Query hooks (depends: 3, 4) [unspecified-high]
├── Task 17: UI - domain management pages (depends: 15, 16) [visual-engineering]
├── Task 18: UI - user/mailbox management pages (depends: 15, 16) [visual-engineering]
├── Task 19: UI - real-time log viewer (depends: 15, 12) [visual-engineering]
├── Task 20: UI - health/monitoring dashboard (depends: 15, 11) [visual-engineering]
├── Task 21: UI - DKIM management page (depends: 15, 10) [visual-engineering]
└── Task 22: CLI tool - oxmail (depends: 4, 7, 8) [unspecified-high]

Wave 4 (Webmail + integration):
├── Task 23: Go API - IMAP proxy/bridge for webmail (depends: 14) [deep]
├── Task 24: Go API - SMTP submission endpoint for compose (depends: 13) [deep]
├── Task 25: UI - webmail inbox + thread view (depends: 15, 23) [visual-engineering]
├── Task 26: UI - webmail compose + send (depends: 15, 24) [visual-engineering]
├── Task 27: UI - webmail search (depends: 25) [visual-engineering]
├── Task 28: Docker Compose full integration + health checks (depends: 13, 14, 11) [unspecified-high]
└── Task 29: Dev mode: seed data + test email tools (depends: 28) [unspecified-high]

Wave 5 (Polish + security + docs):
├── Task 30: Security hardening - open relay test, auth enforcement (depends: 28) [deep]
├── Task 31: UI polish - animations, transitions, empty/error states (depends: 17-21, 25-27) [visual-engineering]
├── Task 32: Performance tuning - RAM budget, cold start optimization (depends: 28) [unspecified-high]
├── Task 33: README + quickstart documentation (depends: 28) [writing]
└── Task 34: End-to-end integration test suite (depends: 28, 29) [deep]

Wave 6 (Production mode - TLS + outbound + DNS):
├── Task 35: Production TLS - Let's Encrypt auto-renewal (depends: 28) [deep]
├── Task 36: Production outbound delivery + DNS helper (depends: 28, 10, 13) [deep]
├── Task 37: UI - DNS setup wizard + production settings page (depends: 15, 16, 36) [visual-engineering]
└── Task 38: Production integration test + docker-compose profiles (depends: 35, 36, 37) [unspecified-high]

Wave FINAL (After ALL tasks — 4 parallel reviews, then user okay):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
-> Present results -> Get explicit user okay

Critical Path: T1 → T2/T5 → T7/T8 → T13/T14 → T28 → T30 → T35/T36 → T38 → F1-F4
Parallel Speedup: ~65% faster than sequential
Max Concurrent: 8 (Wave 2)
```

### Dependency Matrix

| Task | Depends On | Blocks | Wave |
|------|-----------|--------|------|
| 1 | - | 2,3,4,5,6 | 1 |
| 2 | 1 | 7,8,9,10,11,12 | 1 |
| 3 | 1 | 15,16 | 1 |
| 4 | 1 | 7,8,9,10,16,22 | 1 |
| 5 | 1 | 11,12,13,14 | 1 |
| 6 | 1 | - | 1 |
| 7 | 2,4 | 13,22 | 2 |
| 8 | 2,4 | 14,22 | 2 |
| 9 | 2,4 | - | 2 |
| 10 | 2,4 | 21 | 2 |
| 11 | 2,5 | 20,28 | 2 |
| 12 | 2,5 | 19 | 2 |
| 13 | 5,7 | 24,28 | 2 |
| 14 | 5,8 | 23,28 | 2 |
| 15 | 3 | 17-21,25-27 | 3 |
| 16 | 3,4 | 17-21 | 3 |
| 17 | 15,16 | 31 | 3 |
| 18 | 15,16 | 31 | 3 |
| 19 | 15,12 | 31 | 3 |
| 20 | 15,11 | 31 | 3 |
| 21 | 15,10 | 31 | 3 |
| 22 | 4,7,8 | - | 3 |
| 23 | 14 | 25 | 4 |
| 24 | 13 | 26 | 4 |
| 25 | 15,23 | 27,31 | 4 |
| 26 | 15,24 | 31 | 4 |
| 27 | 25 | 31 | 4 |
| 28 | 13,14,11 | 29,30,32,33,34 | 4 |
| 29 | 28 | 34 | 4 |
| 30 | 28 | F1-F4 | 5 |
| 31 | 17-21,25-27 | F1-F4 | 5 |
| 32 | 28 | F1-F4 | 5 |
| 33 | 28 | F1-F4 | 5 |
| 34 | 28,29 | F1-F4 | 5 |

### Agent Dispatch Summary

- **Wave 1**: 6 tasks — T1,T4,T6 → `quick`; T2 → `quick`; T3 → `visual-engineering`; T5 → `unspecified-high`
- **Wave 2**: 8 tasks — T7,T8,T12,T13,T14 → `deep`; T9,T10,T11 → `unspecified-high`
- **Wave 3**: 8 tasks — T15,T17,T18,T19,T20,T21 → `visual-engineering`; T16,T22 → `unspecified-high`
- **Wave 4**: 7 tasks — T23,T24 → `deep`; T25,T26,T27 → `visual-engineering`; T28,T29 → `unspecified-high`
- **Wave 5**: 5 tasks — T30,T34 → `deep`; T31 → `visual-engineering`; T32 → `unspecified-high`; T33 → `writing`
- **Wave 6**: 4 tasks — T35,T36 → `deep`; T37 → `visual-engineering`; T38 → `unspecified-high`
- **FINAL**: 4 tasks — F1 → `oracle`; F2,F3 → `unspecified-high`; F4 → `deep`

---

## TODOs

- [x] 1. Monorepo scaffolding + Docker Compose skeleton

  **What to do**:
  - Rename project folder from `mailer` to `oxmail` (or init fresh in `oxmail/`)
  - Initialize git: `git init && git remote add origin https://github.com/MYusufEka/oxmail.git`
  - Create monorepo structure: `cmd/oxmail-api/`, `cmd/oxmail/`, `internal/`, `web/`, `docker/`, `configs/`, `scripts/`
  - Create `docker-compose.yml` with service definitions (placeholder images): postfix, dovecot, rspamd, redis, oxmail-api, web
  - Create `docker-compose.dev.yml` override for dev mode (exposed ports, volume mounts, hot-reload)
  - Create `.env.example` with all required env vars documented
  - Create `Makefile` with targets: `up`, `down`, `build`, `test`, `lint`, `logs`
  - Create `go.work` for Go workspace (api + cli modules)
  - Set port mapping strategy: SMTP=1025, IMAP=1143, HTTP API=8080, Web UI=3000

  **Must NOT do**:
  - No actual service implementation yet
  - No Kubernetes manifests
  - No CI/CD pipelines

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2,3,4,5,6)
  - **Blocks**: Tasks 2,3,4,5,6
  - **Blocked By**: None

  **References**:
  - Mailcow docker-compose structure (reference for service naming)
  - docker-mailserver single-container approach (reference for port mapping)

  **Acceptance Criteria**:
  - [ ] `docker-compose.yml` valid: `docker compose config` exits 0
  - [ ] All directories exist with `.gitkeep` or initial files
  - [ ] `go.work` valid: `go work sync` exits 0
  - [ ] `.env.example` documents all vars

  **QA Scenarios**:
  ```
  Scenario: Docker Compose validates
    Tool: Bash
    Preconditions: Docker installed
    Steps:
      1. Run `docker compose config --quiet`
      2. Assert exit code 0
    Expected Result: No errors, valid compose file
    Evidence: .sisyphus/evidence/task-1-compose-valid.txt

  Scenario: Directory structure complete
    Tool: Bash
    Preconditions: Repo cloned
    Steps:
      1. Run `find . -type d | sort` from project root
      2. Assert directories exist: cmd/oxmail-api, cmd/oxmail, internal, web, docker, configs, scripts
    Expected Result: All 7 directories present
    Evidence: .sisyphus/evidence/task-1-dirs.txt
  ```

  **Commit**: YES
  - Message: `feat(scaffold): monorepo structure + docker compose skeleton`
  - Files: `docker-compose.yml`, `docker-compose.dev.yml`, `.env.example`, `Makefile`, `go.work`, all directories
  - Pre-commit: `docker compose config --quiet`

- [ ] 2. Go API project setup + test infrastructure

  **What to do**:
  - Initialize Go module: `cmd/oxmail-api/` with `go.mod`
  - Set up project structure: `internal/api/`, `internal/domain/`, `internal/config/`, `internal/mail/`
  - Add dependencies: chi (router), slog (logging), testify (testing), sqlc or raw SQL
  - Create `internal/api/server.go` with basic HTTP server skeleton
  - Create `internal/api/server_test.go` with TDD test: server starts and responds to /health
  - Set up SQLite for control-plane state (lightweight, no extra container)
  - Create database migration framework (golang-migrate or custom)
  - Configure `golangci-lint` with strict settings

  **Must NOT do**:
  - No actual business logic yet
  - No Postfix/Dovecot integration yet
  - No authentication yet

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1,3,4,5,6)
  - **Blocks**: Tasks 7,8,9,10,11,12
  - **Blocked By**: Task 1 (needs directory structure)

  **References**:
  - Go project layout: `github.com/golang-standards/project-layout`
  - Chi router: `github.com/go-chi/chi/v5`
  - Testify: `github.com/stretchr/testify`

  **Acceptance Criteria**:
  - [ ] `go build ./cmd/oxmail-api/` exits 0
  - [ ] `go test ./...` passes with health check test GREEN
  - [ ] `golangci-lint run` exits 0
  - [ ] Server starts on :8080 and GET /health returns `{"status":"ok"}`

  **QA Scenarios**:
  ```
  Scenario: API server starts and responds to health
    Tool: Bash
    Preconditions: Go 1.22+ installed
    Steps:
      1. Run `go build ./cmd/oxmail-api/`
      2. Start server in background: `./oxmail-api &`
      3. Wait 2s
      4. Run `curl -s http://localhost:8080/health`
      5. Assert response contains `"status":"ok"`
      6. Kill background process
    Expected Result: HTTP 200 with health JSON
    Evidence: .sisyphus/evidence/task-2-health-response.txt

  Scenario: Tests pass
    Tool: Bash
    Preconditions: Go modules downloaded
    Steps:
      1. Run `go test ./... -v`
      2. Assert exit code 0
      3. Assert output contains "PASS"
    Expected Result: All tests green
    Evidence: .sisyphus/evidence/task-2-test-output.txt
  ```

  **Commit**: YES
  - Message: `feat(api): Go project setup + health endpoint + test infra`
  - Files: `cmd/oxmail-api/`, `internal/api/`, `go.mod`, `go.sum`, `.golangci.yml`
  - Pre-commit: `go test ./... && golangci-lint run`

- [ ] 3. Next.js project setup + design system tokens

  **What to do**:
  - Initialize Next.js 14+ App Router project in `web/`
  - Install and configure: shadcn/ui, Tailwind CSS, TanStack Query, TanStack Table, Lucide icons
  - Set up dark-first theme with CSS variables (clean minimal + subtle cyber accents)
  - Define design tokens: colors (dark palette, accent colors), spacing, typography, border-radius
  - Create base layout component with dark theme applied
  - Set up Vitest + Testing Library for component tests
  - Set up Playwright for E2E tests
  - Create `web/src/styles/tokens.css` with full design system
  - Create sample component test to verify TDD pipeline works

  **Must NOT do**:
  - No actual pages/features yet
  - No API integration yet
  - No glassmorphism or heavy effects
  - No mobile-first (desktop-first)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1,2,4,5,6)
  - **Blocks**: Tasks 15,16
  - **Blocked By**: Task 1 (needs directory structure)

  **References**:
  - shadcn/ui theming: CSS variables approach
  - Linear.app design: clean, minimal, dark-first
  - Vercel dashboard: typography, spacing, subtle borders

  **Acceptance Criteria**:
  - [ ] `cd web && npm run build` exits 0
  - [ ] `cd web && npx vitest run` passes with sample test GREEN
  - [ ] Dark theme renders by default (no flash of light theme)
  - [ ] Design tokens defined in CSS variables

  **QA Scenarios**:
  ```
  Scenario: Next.js builds successfully
    Tool: Bash
    Preconditions: Node 20+ installed
    Steps:
      1. Run `cd web && npm install`
      2. Run `npm run build`
      3. Assert exit code 0
    Expected Result: Build completes without errors
    Evidence: .sisyphus/evidence/task-3-build.txt

  Scenario: Dark theme renders
    Tool: Playwright
    Preconditions: Dev server running on localhost:3000
    Steps:
      1. Navigate to http://localhost:3000
      2. Assert `html` element has class `dark`
      3. Assert background color is dark (rgb value < 50 per channel)
      4. Screenshot
    Expected Result: Dark theme active by default
    Evidence: .sisyphus/evidence/task-3-dark-theme.png
  ```

  **Commit**: YES
  - Message: `feat(ui): Next.js setup + design system + dark theme`
  - Files: `web/`
  - Pre-commit: `cd web && npm run build && npx vitest run`

- [ ] 4. Shared types/schemas (API contracts, domain models)

  **What to do**:
  - Create `api-spec/openapi.yaml` with full API contract (domains, users, aliases, health, logs, DKIM)
  - Define Go domain models: `internal/domain/models.go` (Domain, User, Alias, DKIMKey, HealthStatus, LogEntry)
  - Define TypeScript types: `web/src/types/api.ts` generated from or matching OpenAPI spec
  - Define request/response schemas with Zod in `web/src/lib/schemas.ts`
  - Define error response format (consistent across all endpoints)
  - Include pagination schema for list endpoints

  **Must NOT do**:
  - No implementation of endpoints
  - No database queries
  - No over-engineering (keep models flat, no deep nesting)

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1,2,3,5,6)
  - **Blocks**: Tasks 7,8,9,10,16,22
  - **Blocked By**: Task 1 (needs directory structure)

  **References**:
  - OpenAPI 3.1 spec format
  - Zod schema patterns for API validation

  **Acceptance Criteria**:
  - [ ] OpenAPI spec validates: `npx @redocly/cli lint api-spec/openapi.yaml` exits 0
  - [ ] Go models compile: `go build ./internal/domain/`
  - [ ] TypeScript types compile: `cd web && npx tsc --noEmit`
  - [ ] Zod schemas match OpenAPI contract

  **QA Scenarios**:
  ```
  Scenario: OpenAPI spec is valid
    Tool: Bash
    Preconditions: Node installed
    Steps:
      1. Run `npx @redocly/cli lint api-spec/openapi.yaml`
      2. Assert exit code 0
      3. Assert no errors in output
    Expected Result: Valid OpenAPI 3.1 spec
    Evidence: .sisyphus/evidence/task-4-openapi-lint.txt

  Scenario: Go and TS types are consistent
    Tool: Bash
    Preconditions: Both Go and Node available
    Steps:
      1. Run `go build ./internal/domain/`
      2. Run `cd web && npx tsc --noEmit`
      3. Both exit 0
    Expected Result: Types compile in both languages
    Evidence: .sisyphus/evidence/task-4-types-compile.txt
  ```

  **Commit**: YES
  - Message: `feat(types): API contracts + domain models + Zod schemas`
  - Files: `api-spec/openapi.yaml`, `internal/domain/models.go`, `web/src/types/api.ts`, `web/src/lib/schemas.ts`
  - Pre-commit: `go build ./internal/domain/ && cd web && npx tsc --noEmit`

- [ ] 5. Postfix+Dovecot+Rspamd Docker images + base configs

  **What to do**:
  - Create `docker/postfix/Dockerfile` based on official Postfix image
  - Create `docker/dovecot/Dockerfile` based on official Dovecot image
  - Create `docker/rspamd/Dockerfile` based on official Rspamd image
  - Create base config templates:
    - `configs/postfix/main.cf.tmpl` (virtual domains, virtual mailboxes, submission port)
    - `configs/postfix/master.cf.tmpl` (services)
    - `configs/dovecot/dovecot.conf.tmpl` (IMAP, auth, mail location)
    - `configs/dovecot/10-auth.conf.tmpl` (passdb/userdb via SQL/file)
    - `configs/rspamd/local.d/` (basic spam config)
  - Configure Postfix for virtual mailbox delivery to Dovecot
  - Configure Dovecot LMTP for local delivery from Postfix
  - Set up shared volume strategy: mail storage, config, DKIM keys
  - Add health check scripts for each service
  - Target: containers build and start (no dynamic config yet, static test config)

  **Must NOT do**:
  - No custom SMTP/IMAP engine
  - No ClamAV (antivirus excluded for resource budget)
  - No Let's Encrypt integration
  - No production TLS (self-signed for dev)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1,2,3,4,6)
  - **Blocks**: Tasks 11,12,13,14
  - **Blocked By**: Task 1 (needs docker/ directory)

  **References**:
  - docker-mailserver config approach (single container, Postfix+Dovecot integration)
  - Mailu container separation pattern
  - Postfix virtual mailbox docs: http://www.postfix.org/VIRTUAL_README.html
  - Dovecot LMTP: https://doc.dovecot.org/configuration_manual/protocols/lmtp_server/

  **Acceptance Criteria**:
  - [ ] `docker build -t oxmail-postfix docker/postfix/` exits 0
  - [ ] `docker build -t oxmail-dovecot docker/dovecot/` exits 0
  - [ ] `docker build -t oxmail-rspamd docker/rspamd/` exits 0
  - [ ] Containers start with static test config and pass health checks
  - [ ] Postfix listens on port 25 and 587 inside container
  - [ ] Dovecot listens on port 143 and 993 inside container

  **QA Scenarios**:
  ```
  Scenario: All mail containers build and start
    Tool: Bash
    Preconditions: Docker installed
    Steps:
      1. Run `docker compose up -d postfix dovecot rspamd`
      2. Wait 30s
      3. Run `docker compose ps` and assert all 3 are "healthy" or "running"
      4. Run `docker compose exec postfix postconf mail_version` - assert output contains version
      5. Run `docker compose exec dovecot doveconf -n` - assert output contains "protocols"
    Expected Result: All 3 containers running with correct services
    Evidence: .sisyphus/evidence/task-5-containers-running.txt

  Scenario: Postfix rejects open relay
    Tool: Bash
    Preconditions: Postfix container running with static config
    Steps:
      1. Run `swaks --server localhost --port 1025 --from attacker@evil.com --to victim@external.com`
      2. Assert exit code != 0 or response contains "Relay access denied"
    Expected Result: Relay attempt rejected
    Evidence: .sisyphus/evidence/task-5-relay-denied.txt
  ```

  **Commit**: YES
  - Message: `feat(mail): Postfix+Dovecot+Rspamd Docker images + base configs`
  - Files: `docker/postfix/`, `docker/dovecot/`, `docker/rspamd/`, `configs/`
  - Pre-commit: `docker compose build postfix dovecot rspamd`

- [ ] 6. Dev environment tooling (Makefile, scripts, .env)

  **What to do**:
  - Enhance `Makefile` with comprehensive targets:
    - `make dev` - start full stack in dev mode with hot-reload
    - `make test` - run all tests (Go + Vitest + Playwright)
    - `make test-go` / `make test-web` / `make test-e2e` - individual test suites
    - `make lint` - run all linters
    - `make build` - build all images and binaries
    - `make clean` - remove volumes and built artifacts
    - `make logs` - tail all container logs
    - `make seed` - seed test data (domain + user + test email)
  - Create `scripts/seed.sh` - creates test domain `local.test`, user `alice@local.test`
  - Create `scripts/wait-healthy.sh` - polls health endpoint until ready
  - Create `.env.example` with documented variables:
    - `OXMAIL_DOMAIN=local.test`
    - `OXMAIL_ADMIN_PASSWORD=admin`
    - `OXMAIL_API_PORT=8080`
    - `OXMAIL_WEB_PORT=3000`
    - `OXMAIL_SMTP_PORT=1025`
    - `OXMAIL_IMAP_PORT=1143`
  - Create `.env` from `.env.example` in Makefile target
  - Add `.env` to `.gitignore`

  **Must NOT do**:
  - No CI/CD pipeline
  - No production deployment scripts
  - No Kubernetes tooling

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1,2,3,4,5)
  - **Blocks**: None (convenience tooling)
  - **Blocked By**: Task 1 (needs project structure)

  **References**:
  - Mailcow `generate_config.sh` pattern
  - docker-mailserver `setup.sh` approach

  **Acceptance Criteria**:
  - [ ] `make help` shows all available targets
  - [ ] `scripts/seed.sh` is executable and has correct shebang
  - [ ] `scripts/wait-healthy.sh` is executable
  - [ ] `.env.example` documents all required variables with comments
  - [ ] `.gitignore` includes `.env`

  **QA Scenarios**:
  ```
  Scenario: Makefile targets are valid
    Tool: Bash
    Preconditions: make installed
    Steps:
      1. Run `make -n dev` (dry-run)
      2. Assert exit code 0 (target exists and is valid)
      3. Run `make -n test`
      4. Assert exit code 0
    Expected Result: All Makefile targets parse correctly
    Evidence: .sisyphus/evidence/task-6-makefile-valid.txt

  Scenario: Scripts are executable
    Tool: Bash
    Preconditions: Scripts created
    Steps:
      1. Run `test -x scripts/seed.sh` - assert exit 0
      2. Run `test -x scripts/wait-healthy.sh` - assert exit 0
      3. Run `head -1 scripts/seed.sh` - assert contains "#!/bin/bash" or "#!/usr/bin/env bash"
    Expected Result: Scripts executable with proper shebang
    Evidence: .sisyphus/evidence/task-6-scripts-exec.txt
  ```

  **Commit**: YES (groups with Task 1)
  - Message: `feat(scaffold): monorepo structure + docker compose skeleton`
  - Files: `Makefile`, `scripts/`, `.env.example`, `.gitignore`
  - Pre-commit: `make -n dev && make -n test`

- [ ] 7. Go API - domain CRUD + config generation

  **What to do**:
  - TDD: Write tests first for domain CRUD operations
  - Implement `internal/domain/domain_service.go`: Create, Get, List, Delete domain
  - Implement `internal/api/domains_handler.go`: REST endpoints (POST/GET/DELETE /api/domains)
  - Implement `internal/config/postfix_generator.go`: Generate `virtual_domains` file from DB
  - Implement `internal/config/dovecot_generator.go`: Generate domain entries for Dovecot
  - SQLite migrations for domains table
  - Config generation triggered on domain CRUD (write to shared volume)
  - Validate domain format (RFC 5321 compliant)
  - Return proper HTTP status codes: 201 created, 200 list, 404 not found, 409 conflict, 400 invalid

  **Must NOT do**:
  - No authentication yet (added in Task 30)
  - No Postfix reload trigger yet (added in Task 13)
  - No rate limiting
  - No pagination yet (keep simple for now, add in integration)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 8,9,10,11,12)
  - **Blocks**: Tasks 13, 22
  - **Blocked By**: Tasks 2, 4

  **References**:
  - `internal/domain/models.go` (Task 4) - Domain model definition
  - `api-spec/openapi.yaml` (Task 4) - API contract for domains endpoints
  - `internal/api/server.go` (Task 2) - Router setup pattern
  - Postfix virtual_domains format: one domain per line

  **Acceptance Criteria**:
  - [ ] `go test ./internal/domain/ -run TestDomain` passes (TDD: written first)
  - [ ] `go test ./internal/api/ -run TestDomains` passes
  - [ ] `curl -X POST localhost:8080/api/domains -d '{"domain":"local.test"}'` → 201
  - [ ] `curl localhost:8080/api/domains` → 200 with list containing "local.test"
  - [ ] `curl -X POST localhost:8080/api/domains -d '{"domain":"local.test"}'` → 409 (duplicate)
  - [ ] `curl -X POST localhost:8080/api/domains -d '{"domain":"not valid!"}'` → 400
  - [ ] Config file generated at expected path after domain creation

  **QA Scenarios**:
  ```
  Scenario: Domain CRUD happy path
    Tool: Bash
    Preconditions: API server running on localhost:8080
    Steps:
      1. curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/api/domains -H "Content-Type: application/json" -d '{"domain":"local.test"}'
      2. Assert status code = 201
      3. curl -s http://localhost:8080/api/domains | jq '.domains[0].domain'
      4. Assert output = "local.test"
      5. curl -s -o /dev/null -w "%{http_code}" -X DELETE http://localhost:8080/api/domains/local.test
      6. Assert status code = 200
    Expected Result: Full CRUD lifecycle works
    Evidence: .sisyphus/evidence/task-7-domain-crud.txt

  Scenario: Domain validation rejects invalid input
    Tool: Bash
    Preconditions: API server running
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/domains -H "Content-Type: application/json" -d '{"domain":"not valid!"}'
      2. Assert status code = 400
      3. Assert response body contains "invalid"
    Expected Result: 400 with error message
    Evidence: .sisyphus/evidence/task-7-domain-validation.txt
  ```

  **Commit**: YES
  - Message: `feat(api): domain CRUD + Postfix config generation`
  - Files: `internal/domain/domain_service.go`, `internal/api/domains_handler.go`, `internal/config/postfix_generator.go`
  - Pre-commit: `go test ./...`

- [ ] 8. Go API - user/mailbox CRUD + password hashing

  **What to do**:
  - TDD: Write tests first for user CRUD operations
  - Implement `internal/domain/user_service.go`: Create, Get, List, Delete user/mailbox
  - Implement `internal/api/users_handler.go`: REST endpoints (POST/GET/DELETE /api/users)
  - Implement password hashing using bcrypt (Dovecot-compatible scheme: BLF-CRYPT)
  - Implement `internal/config/dovecot_users_generator.go`: Generate passdb/userdb files
  - SQLite migrations for users table (email, password_hash, domain_id, quota, active)
  - Validate email format, enforce user belongs to existing domain
  - Create maildir structure on user creation
  - Return proper HTTP status codes

  **Must NOT do**:
  - No password reset flow
  - No 2FA
  - No session management for mail users
  - No quota enforcement (just store the value)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7,9,10,11,12)
  - **Blocks**: Tasks 14, 22
  - **Blocked By**: Tasks 2, 4

  **References**:
  - `internal/domain/models.go` (Task 4) - User model definition
  - `api-spec/openapi.yaml` (Task 4) - API contract for users endpoints
  - Dovecot passdb format: `user:{BLF-CRYPT}$2b$...`
  - Dovecot userdb format: `user:uid:gid::home`

  **Acceptance Criteria**:
  - [ ] `go test ./internal/domain/ -run TestUser` passes (TDD)
  - [ ] `go test ./internal/api/ -run TestUsers` passes
  - [ ] `curl -X POST localhost:8080/api/users -d '{"email":"alice@local.test","password":"TestPass123!"}'` → 201
  - [ ] Password stored as bcrypt hash (not plaintext)
  - [ ] `curl -X POST localhost:8080/api/users -d '{"email":"alice@local.test","password":"x"}'` → 409
  - [ ] `curl -X POST localhost:8080/api/users -d '{"email":"bob@nonexistent.test","password":"x"}'` → 400 (domain not found)
  - [ ] Dovecot passdb/userdb files generated correctly

  **QA Scenarios**:
  ```
  Scenario: User creation with password hashing
    Tool: Bash
    Preconditions: API running, domain "local.test" exists
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email":"alice@local.test","password":"TestPass123!"}'
      2. Assert status code = 201
      3. Read generated passdb file
      4. Assert contains "alice@local.test:{BLF-CRYPT}$2" (bcrypt prefix)
      5. Assert does NOT contain "TestPass123!" in plaintext anywhere
    Expected Result: User created with hashed password
    Evidence: .sisyphus/evidence/task-8-user-create.txt

  Scenario: User creation fails for nonexistent domain
    Tool: Bash
    Preconditions: API running, domain "nonexistent.test" does NOT exist
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email":"bob@nonexistent.test","password":"TestPass123!"}'
      2. Assert status code = 400
      3. Assert response contains "domain" and "not found" or similar
    Expected Result: 400 error referencing missing domain
    Evidence: .sisyphus/evidence/task-8-user-no-domain.txt
  ```

  **Commit**: YES
  - Message: `feat(api): user/mailbox CRUD + bcrypt password hashing`
  - Files: `internal/domain/user_service.go`, `internal/api/users_handler.go`, `internal/config/dovecot_users_generator.go`
  - Pre-commit: `go test ./...`

- [ ] 9. Go API - alias management

  **What to do**:
  - TDD: Write tests first for alias CRUD
  - Implement `internal/domain/alias_service.go`: Create, Get, List, Delete alias
  - Implement `internal/api/aliases_handler.go`: REST endpoints (POST/GET/DELETE /api/aliases)
  - Implement `internal/config/postfix_aliases_generator.go`: Generate `virtual_alias_maps` file
  - SQLite migrations for aliases table (source_email, destination_emails, domain_id)
  - Support multiple destinations per alias (comma-separated in Postfix format)
  - Validate source belongs to existing domain
  - Prevent circular aliases

  **Must NOT do**:
  - No catch-all aliases
  - No regex aliases
  - No external forwarding (local only)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7,8,10,11,12)
  - **Blocks**: None directly
  - **Blocked By**: Tasks 2, 4

  **References**:
  - `internal/domain/models.go` (Task 4) - Alias model
  - `api-spec/openapi.yaml` (Task 4) - Alias endpoints
  - Postfix virtual_alias_maps format: `source destination1,destination2`

  **Acceptance Criteria**:
  - [ ] `go test ./internal/domain/ -run TestAlias` passes (TDD)
  - [ ] `curl -X POST localhost:8080/api/aliases -d '{"source":"info@local.test","destinations":["alice@local.test"]}'` → 201
  - [ ] Generated virtual_alias_maps contains correct mapping
  - [ ] Circular alias detection works

  **QA Scenarios**:
  ```
  Scenario: Alias creation and config generation
    Tool: Bash
    Preconditions: API running, domain "local.test" and user "alice@local.test" exist
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/aliases -H "Content-Type: application/json" -d '{"source":"info@local.test","destinations":["alice@local.test"]}'
      2. Assert status code = 201
      3. Read generated virtual_alias_maps file
      4. Assert contains "info@local.test alice@local.test"
    Expected Result: Alias created and Postfix config updated
    Evidence: .sisyphus/evidence/task-9-alias-create.txt

  Scenario: Circular alias rejected
    Tool: Bash
    Preconditions: Alias "info@local.test → alice@local.test" exists
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/aliases -H "Content-Type: application/json" -d '{"source":"alice@local.test","destinations":["info@local.test"]}'
      2. Assert status code = 400 or 409
      3. Assert response mentions "circular"
    Expected Result: Circular alias prevented
    Evidence: .sisyphus/evidence/task-9-alias-circular.txt
  ```

  **Commit**: YES
  - Message: `feat(api): alias management + virtual_alias_maps generation`
  - Files: `internal/domain/alias_service.go`, `internal/api/aliases_handler.go`, `internal/config/postfix_aliases_generator.go`
  - Pre-commit: `go test ./...`

- [ ] 10. Go API - DKIM key generation + management

  **What to do**:
  - TDD: Write tests first for DKIM operations
  - Implement `internal/domain/dkim_service.go`: Generate, Get, Rotate, Delete DKIM keys
  - Implement `internal/api/dkim_handler.go`: REST endpoints (POST/GET/DELETE /api/domains/{domain}/dkim)
  - Generate RSA 2048-bit DKIM keys using Go crypto
  - Store private key in secure volume, expose public key via API (for DNS TXT record)
  - Implement `internal/config/dkim_generator.go`: Generate OpenDKIM signing table + key table
  - Return DNS TXT record value in API response for easy copy-paste

  **Must NOT do**:
  - No automatic DNS configuration
  - No DKIM verification (Rspamd handles inbound)
  - No Ed25519 keys (RSA only for compatibility)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7,8,9,11,12)
  - **Blocks**: Task 21
  - **Blocked By**: Tasks 2, 4

  **References**:
  - `internal/domain/models.go` (Task 4) - DKIMKey model
  - OpenDKIM signing table format
  - RFC 6376 DKIM specification

  **Acceptance Criteria**:
  - [ ] `go test ./internal/domain/ -run TestDKIM` passes (TDD)
  - [ ] `curl -X POST localhost:8080/api/domains/local.test/dkim` → 201 with public key
  - [ ] Private key file created in secure volume path
  - [ ] Public key response includes DNS TXT record format
  - [ ] `curl localhost:8080/api/domains/local.test/dkim` → 200 with key info (no private key exposed)

  **QA Scenarios**:
  ```
  Scenario: DKIM key generation
    Tool: Bash
    Preconditions: API running, domain "local.test" exists
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/domains/local.test/dkim -H "Content-Type: application/json"
      2. Assert status code = 201
      3. Assert response contains "public_key" field
      4. Assert response contains "dns_record" field with "v=DKIM1"
      5. curl -s http://localhost:8080/api/domains/local.test/dkim | jq '.public_key'
      6. Assert public key is non-empty and starts with "MII" (base64 RSA)
    Expected Result: DKIM key generated with DNS record
    Evidence: .sisyphus/evidence/task-10-dkim-generate.txt

  Scenario: Private key not exposed via API
    Tool: Bash
    Preconditions: DKIM key exists for local.test
    Steps:
      1. curl -s http://localhost:8080/api/domains/local.test/dkim
      2. Assert response does NOT contain "private_key" or "PRIVATE KEY"
    Expected Result: Only public key exposed
    Evidence: .sisyphus/evidence/task-10-dkim-no-private.txt
  ```

  **Commit**: YES
  - Message: `feat(api): DKIM key generation + DNS record export`
  - Files: `internal/domain/dkim_service.go`, `internal/api/dkim_handler.go`, `internal/config/dkim_generator.go`
  - Pre-commit: `go test ./...`

- [ ] 11. Go API - health check aggregator

  **What to do**:
  - TDD: Write tests first for health aggregation
  - Implement `internal/domain/health_service.go`: Check health of all services (Postfix, Dovecot, Rspamd, Redis, API itself)
  - Implement `internal/api/health_handler.go`: GET /api/health returns aggregated status
  - Health checks: TCP connect to each service port, verify response
  - Include per-service status, response time, and overall status
  - Overall status: "healthy" only if ALL services healthy, "degraded" if some down, "unhealthy" if critical down
  - Include uptime, version info, container memory usage (from Docker API or /proc)
  - Cache health results for 5s to avoid hammering services

  **Must NOT do**:
  - No Prometheus metrics export (deferred)
  - No alerting
  - No historical health data storage
  - No Docker socket access (use TCP checks only)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7,8,9,10,12)
  - **Blocks**: Tasks 20, 28
  - **Blocked By**: Tasks 2, 5

  **References**:
  - `internal/domain/models.go` (Task 4) - HealthStatus model
  - `api-spec/openapi.yaml` (Task 4) - Health endpoint contract
  - Postfix: SMTP EHLO on port 25 = alive
  - Dovecot: IMAP capability on port 143 = alive
  - Rspamd: HTTP /ping on port 11333 = alive
  - Redis: PING command on port 6379 = alive

  **Acceptance Criteria**:
  - [ ] `go test ./internal/domain/ -run TestHealth` passes (TDD)
  - [ ] `curl localhost:8080/api/health` → 200 with all service statuses
  - [ ] Response includes: `status`, `services[]`, `uptime`, `version`
  - [ ] When a service is down, status changes to "degraded" or "unhealthy"

  **QA Scenarios**:
  ```
  Scenario: Health endpoint returns all services
    Tool: Bash
    Preconditions: Full stack running
    Steps:
      1. curl -s http://localhost:8080/api/health | jq '.'
      2. Assert .status = "healthy"
      3. Assert .services | length >= 4 (postfix, dovecot, rspamd, redis)
      4. Assert each service has "status", "response_time_ms" fields
    Expected Result: Aggregated health with all services healthy
    Evidence: .sisyphus/evidence/task-11-health-all.txt

  Scenario: Health degrades when service stops
    Tool: Bash
    Preconditions: Full stack running
    Steps:
      1. docker compose stop rspamd
      2. Wait 6s (cache expiry)
      3. curl -s http://localhost:8080/api/health | jq '.status'
      4. Assert status = "degraded"
      5. curl -s http://localhost:8080/api/health | jq '.services[] | select(.name=="rspamd") | .status'
      6. Assert = "unhealthy"
      7. docker compose start rspamd
    Expected Result: Health reflects stopped service
    Evidence: .sisyphus/evidence/task-11-health-degraded.txt
  ```

  **Commit**: YES
  - Message: `feat(api): health check aggregator for all services`
  - Files: `internal/domain/health_service.go`, `internal/api/health_handler.go`
  - Pre-commit: `go test ./...`

- [ ] 12. Go API - real-time log streaming via WebSocket

  **What to do**:
  - TDD: Write tests for log streaming
  - Implement `internal/logs/collector.go`: Tail logs from Postfix/Dovecot/Rspamd via Docker log files or syslog
  - Implement `internal/logs/parser.go`: Parse mail log lines into structured LogEntry (timestamp, service, level, message, metadata)
  - Implement `internal/api/logs_handler.go`: WebSocket endpoint at /api/logs/stream
  - Support filters: by service (postfix/dovecot/rspamd), by level (info/warn/error), by search term
  - Implement `GET /api/logs` for paginated historical logs (last N entries from buffer)
  - Ring buffer for last 10000 log entries in memory
  - Parse Postfix log format: extract from/to/status/queue-id
  - Parse Dovecot log format: extract user/action/status

  **Must NOT do**:
  - No persistent log storage (memory buffer only)
  - No log rotation management
  - No external log shipping (no ELK/Loki)
  - No log file writing (read-only from mail services)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 7,8,9,10,11)
  - **Blocks**: Task 19
  - **Blocked By**: Tasks 2, 5

  **References**:
  - `internal/domain/models.go` (Task 4) - LogEntry model
  - Postfix log format: `Mon DD HH:MM:SS hostname postfix/smtpd[PID]: ...`
  - Dovecot log format: `Mon DD HH:MM:SS hostname dovecot: ...`
  - gorilla/websocket for WebSocket handling

  **Acceptance Criteria**:
  - [ ] `go test ./internal/logs/ -run TestParser` passes (TDD)
  - [ ] `go test ./internal/api/ -run TestLogsStream` passes
  - [ ] WebSocket connection to ws://localhost:8080/api/logs/stream succeeds
  - [ ] Sending test email produces log entry in stream within 3s
  - [ ] Filter by service works (only postfix logs when filtered)
  - [ ] `GET /api/logs?limit=50` returns paginated historical entries

  **QA Scenarios**:
  ```
  Scenario: Log stream receives SMTP event
    Tool: Bash
    Preconditions: Full stack running, WebSocket client available (websocat)
    Steps:
      1. Start websocat ws://localhost:8080/api/logs/stream in background, pipe to file
      2. Send test email: swaks --server localhost --port 1025 --from test@local.test --to alice@local.test
      3. Wait 3s
      4. Read captured WebSocket output
      5. Assert contains "postfix" and "alice@local.test"
    Expected Result: SMTP event appears in log stream within 3s
    Evidence: .sisyphus/evidence/task-12-log-stream.txt

  Scenario: Log filtering by service
    Tool: Bash
    Preconditions: Stack running with some log history
    Steps:
      1. curl -s "http://localhost:8080/api/logs?service=postfix&limit=10" | jq '.[].service'
      2. Assert all entries = "postfix" (no dovecot/rspamd mixed in)
    Expected Result: Only postfix logs returned
    Evidence: .sisyphus/evidence/task-12-log-filter.txt
  ```

  **Commit**: YES
  - Message: `feat(api): real-time log streaming via WebSocket + structured parsing`
  - Files: `internal/logs/collector.go`, `internal/logs/parser.go`, `internal/api/logs_handler.go`
  - Pre-commit: `go test ./...`

- [ ] 13. Postfix dynamic config integration

  **What to do**:
  - TDD: Write integration tests for config reload
  - Implement `internal/mail/postfix_manager.go`: Apply generated configs to running Postfix
  - Implement config reload mechanism: `postfix reload` via Docker exec or signal
  - Implement `postmap` execution for hash-type lookup tables
  - Wire domain/alias CRUD → config generation → postmap → reload pipeline
  - Implement atomic config writes (write to temp, rename)
  - Add retry logic for reload failures
  - Verify Postfix accepts config after reload (check exit code + logs)
  - Integration test: create domain via API → verify Postfix accepts mail for that domain

  **Must NOT do**:
  - No direct Postfix config editing (always generate from source of truth)
  - No manual postfix restart (reload only)
  - No custom transport maps

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 14)
  - **Parallel Group**: Wave 2 (with Task 14, after 7 completes)
  - **Blocks**: Tasks 24, 28
  - **Blocked By**: Tasks 5, 7

  **References**:
  - `internal/config/postfix_generator.go` (Task 7) - Config generation
  - `docker/postfix/` (Task 5) - Postfix container
  - Postfix `postmap` command for hash tables
  - Postfix `postfix reload` for config reload

  **Acceptance Criteria**:
  - [ ] `go test ./internal/mail/ -run TestPostfix` passes (integration test)
  - [ ] Create domain via API → `swaks` to that domain succeeds
  - [ ] Delete domain via API → `swaks` to that domain fails (rejected)
  - [ ] Config writes are atomic (no partial writes on failure)
  - [ ] Reload failure doesn't corrupt running config

  **QA Scenarios**:
  ```
  Scenario: Dynamic domain addition works end-to-end
    Tool: Bash
    Preconditions: Full stack running, no domains configured
    Steps:
      1. curl -s -X POST http://localhost:8080/api/domains -H "Content-Type: application/json" -d '{"domain":"dynamic.test"}'
      2. curl -s -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email":"bob@dynamic.test","password":"TestPass123!"}'
      3. Wait 5s (config generation + reload)
      4. swaks --server localhost --port 1025 --from sender@local.test --to bob@dynamic.test --body "dynamic test"
      5. Assert swaks exit code = 0
    Expected Result: Dynamically added domain accepts mail
    Evidence: .sisyphus/evidence/task-13-dynamic-domain.txt

  Scenario: Deleted domain rejects mail
    Tool: Bash
    Preconditions: Domain "dynamic.test" exists and accepts mail
    Steps:
      1. curl -s -X DELETE http://localhost:8080/api/domains/dynamic.test
      2. Wait 5s
      3. swaks --server localhost --port 1025 --from sender@local.test --to bob@dynamic.test
      4. Assert exit code != 0 or output contains "rejected" or "User unknown"
    Expected Result: Deleted domain no longer accepts mail
    Evidence: .sisyphus/evidence/task-13-domain-deleted.txt
  ```

  **Commit**: YES
  - Message: `feat(mail): Postfix dynamic config integration + reload pipeline`
  - Files: `internal/mail/postfix_manager.go`, integration tests
  - Pre-commit: `go test ./...`

- [ ] 14. Dovecot dynamic config integration

  **What to do**:
  - TDD: Write integration tests for Dovecot user management
  - Implement `internal/mail/dovecot_manager.go`: Apply user configs to running Dovecot
  - Wire user CRUD → passdb/userdb generation → Dovecot reload
  - Implement maildir creation for new users (correct permissions: uid/gid)
  - Implement Dovecot reload via `doveadm reload` or signal
  - Integration test: create user via API → IMAP login succeeds
  - Handle Dovecot auth cache flush on user changes
  - Verify LMTP delivery path works (Postfix → Dovecot LMTP → maildir)

  **Must NOT do**:
  - No Dovecot Sieve filters
  - No shared folders
  - No quota enforcement (store value only)
  - No POP3

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 13)
  - **Parallel Group**: Wave 2 (with Task 13, after 8 completes)
  - **Blocks**: Tasks 23, 28
  - **Blocked By**: Tasks 5, 8

  **References**:
  - `internal/config/dovecot_users_generator.go` (Task 8) - User config generation
  - `docker/dovecot/` (Task 5) - Dovecot container
  - Dovecot passdb/userdb file format
  - Dovecot LMTP configuration
  - `doveadm` commands for management

  **Acceptance Criteria**:
  - [ ] `go test ./internal/mail/ -run TestDovecot` passes (integration test)
  - [ ] Create user via API → IMAP login with those credentials succeeds
  - [ ] Delete user via API → IMAP login fails
  - [ ] Maildir created with correct permissions (uid 5000, gid 5000 or configured)
  - [ ] LMTP delivery: send via Postfix → message appears in user's maildir

  **QA Scenarios**:
  ```
  Scenario: User creation enables IMAP login
    Tool: Bash
    Preconditions: Full stack running, domain "local.test" exists
    Steps:
      1. curl -s -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email":"imap-test@local.test","password":"TestPass123!"}'
      2. Wait 5s (config + reload)
      3. Test IMAP login: curl -s "imap://imap-test@local.test:TestPass123!@localhost:1143/INBOX" or use openssl s_client / doveadm
      4. Assert login succeeds (no auth failure)
    Expected Result: Dynamically created user can login via IMAP
    Evidence: .sisyphus/evidence/task-14-imap-login.txt

  Scenario: Full delivery path (SMTP → LMTP → Maildir → IMAP)
    Tool: Bash
    Preconditions: User "alice@local.test" exists, IMAP accessible
    Steps:
      1. swaks --server localhost --port 1025 --from bob@local.test --to alice@local.test --body "delivery test" --header "Subject: Test Delivery"
      2. Wait 5s
      3. Check IMAP inbox for alice@local.test (doveadm fetch or IMAP client)
      4. Assert message with subject "Test Delivery" exists
    Expected Result: Email delivered end-to-end
    Evidence: .sisyphus/evidence/task-14-full-delivery.txt
  ```

  **Commit**: YES
  - Message: `feat(mail): Dovecot dynamic config + LMTP delivery integration`
  - Files: `internal/mail/dovecot_manager.go`, integration tests
  - Pre-commit: `go test ./...`

- [ ] 15. UI shell - layout, navigation, command palette

  **What to do**:
  - TDD: Write component tests first
  - Create app layout: collapsible left sidebar + top bar + main content area
  - Sidebar: nav items (Dashboard, Domains, Users, Aliases, DKIM, Logs, Webmail) with Lucide icons
  - Top bar: command palette trigger (Cmd+K), health status indicator, dark/light toggle
  - Implement command palette (cmdk library): search across all entities and actions
  - Keyboard shortcuts: Cmd+K (palette), Cmd+1-7 (nav items), Escape (close modals)
  - Responsive: sidebar collapses to icons on smaller screens
  - Loading skeleton states for all content areas
  - 404 page with navigation back

  **Must NOT do**:
  - No actual data fetching (mock data only for layout testing)
  - No mobile hamburger menu (desktop-first)
  - No user auth UI yet

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 16-22)
  - **Blocks**: Tasks 17,18,19,20,21,25,26,27
  - **Blocked By**: Task 3

  **References**:
  - `web/src/styles/tokens.css` (Task 3) - Design tokens
  - Linear.app layout pattern: left nav, clean content area
  - cmdk library: https://cmdk.paco.me/
  - shadcn/ui command component

  **Acceptance Criteria**:
  - [ ] `cd web && npx vitest run` passes with layout component tests
  - [ ] Sidebar renders all 7 nav items with correct icons
  - [ ] Cmd+K opens command palette
  - [ ] Navigation between pages works without full reload
  - [ ] Dark theme consistent across all shell elements

  **QA Scenarios**:
  ```
  Scenario: Shell layout renders correctly
    Tool: Playwright
    Preconditions: Dev server on localhost:3000
    Steps:
      1. Navigate to http://localhost:3000
      2. Assert sidebar visible with selector `nav[data-testid="sidebar"]`
      3. Assert 7 nav items present: `nav a` count = 7
      4. Assert top bar visible with `header[data-testid="topbar"]`
      5. Screenshot full page
    Expected Result: Complete shell layout with sidebar + topbar
    Evidence: .sisyphus/evidence/task-15-shell-layout.png

  Scenario: Command palette opens and searches
    Tool: Playwright
    Preconditions: App loaded
    Steps:
      1. Press Cmd+K (or Ctrl+K on Windows)
      2. Assert command palette dialog visible: `[data-testid="command-palette"]`
      3. Type "domains"
      4. Assert filtered results contain "Domains" navigation item
      5. Press Escape
      6. Assert palette closed
    Expected Result: Command palette opens, filters, closes
    Evidence: .sisyphus/evidence/task-15-command-palette.png
  ```

  **Commit**: YES
  - Message: `feat(ui): shell layout + sidebar navigation + command palette`
  - Files: `web/src/components/layout/`, `web/src/app/layout.tsx`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 16. UI - API client + TanStack Query hooks

  **What to do**:
  - Create `web/src/lib/api-client.ts`: typed fetch wrapper matching OpenAPI spec
  - Create TanStack Query hooks for all API endpoints:
    - `useDomains()`, `useCreateDomain()`, `useDeleteDomain()`
    - `useUsers()`, `useCreateUser()`, `useDeleteUser()`
    - `useAliases()`, `useCreateAlias()`, `useDeleteAlias()`
    - `useDkim()`, `useGenerateDkim()`
    - `useHealth()`
    - `useLogs()`
  - Create WebSocket hook: `useLogStream()` for real-time logs
  - Configure QueryClient with sensible defaults (staleTime, retry, refetchOnWindowFocus)
  - Error handling: toast notifications on mutation failures
  - Optimistic updates for CRUD operations

  **Must NOT do**:
  - No authentication headers yet
  - No offline support/caching
  - No request cancellation

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 15,17-22)
  - **Blocks**: Tasks 17,18,19,20,21
  - **Blocked By**: Tasks 3, 4

  **References**:
  - `api-spec/openapi.yaml` (Task 4) - API contract
  - `web/src/types/api.ts` (Task 4) - TypeScript types
  - `web/src/lib/schemas.ts` (Task 4) - Zod schemas
  - TanStack Query docs: mutation + optimistic updates

  **Acceptance Criteria**:
  - [ ] `cd web && npx vitest run` passes with hook tests (mocked API)
  - [ ] All hooks typed correctly (no `any`)
  - [ ] WebSocket hook connects and receives messages
  - [ ] Error states handled with toast notifications
  - [ ] `npx tsc --noEmit` passes

  **QA Scenarios**:
  ```
  Scenario: API client types match contract
    Tool: Bash
    Preconditions: Web project built
    Steps:
      1. cd web && npx tsc --noEmit
      2. Assert exit code 0
      3. grep -r "as any" src/lib/api-client.ts src/hooks/
      4. Assert no matches (zero `as any` usage)
    Expected Result: Fully typed, no any casts
    Evidence: .sisyphus/evidence/task-16-types-clean.txt

  Scenario: Query hooks work with mocked API
    Tool: Bash
    Preconditions: Vitest configured with MSW or similar mock
    Steps:
      1. cd web && npx vitest run --reporter=verbose src/hooks/
      2. Assert all hook tests pass
      3. Assert test count >= 10 (minimum coverage)
    Expected Result: All hooks tested and passing
    Evidence: .sisyphus/evidence/task-16-hooks-test.txt
  ```

  **Commit**: YES
  - Message: `feat(ui): API client + TanStack Query hooks + WebSocket`
  - Files: `web/src/lib/api-client.ts`, `web/src/hooks/`
  - Pre-commit: `cd web && npx tsc --noEmit && npx vitest run`

- [ ] 17. UI - domain management pages

  **What to do**:
  - TDD: Write component tests first
  - Create domain list page: TanStack Table with columns (domain, users count, aliases count, DKIM status, created, actions)
  - Create domain dialog: form to add new domain (React Hook Form + Zod validation)
  - Inline actions: delete with confirmation dialog
  - Empty state: illustration + "Add your first domain" CTA
  - Loading state: skeleton rows
  - Error state: retry button with error message
  - Toast notifications on success/failure
  - Keyboard: Enter to submit form, Escape to close dialog

  **Must NOT do**:
  - No bulk operations
  - No domain verification flow
  - No DNS checker UI (deferred)

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 15,16,18-22)
  - **Blocks**: Task 31
  - **Blocked By**: Tasks 15, 16

  **References**:
  - `web/src/hooks/` (Task 16) - useDomains(), useCreateDomain(), useDeleteDomain()
  - `web/src/types/api.ts` (Task 4) - Domain type
  - shadcn/ui DataTable pattern
  - shadcn/ui Dialog + Form components

  **Acceptance Criteria**:
  - [ ] Domain list renders with correct columns
  - [ ] Add domain form validates input (rejects invalid domains)
  - [ ] Delete shows confirmation, removes from list on confirm
  - [ ] Empty state shown when no domains
  - [ ] All component tests pass

  **QA Scenarios**:
  ```
  Scenario: Domain CRUD via UI
    Tool: Playwright
    Preconditions: App running, API running, no domains exist
    Steps:
      1. Navigate to http://localhost:3000/domains
      2. Assert empty state visible: text "Add your first domain"
      3. Click "Add Domain" button
      4. Fill domain field with "local.test"
      5. Click Submit
      6. Assert table row appears with "local.test"
      7. Click delete icon on row
      8. Assert confirmation dialog appears
      9. Click "Confirm"
      10. Assert row removed, empty state returns
    Expected Result: Full CRUD lifecycle in UI
    Evidence: .sisyphus/evidence/task-17-domain-crud-ui.png

  Scenario: Domain form validation
    Tool: Playwright
    Preconditions: Add domain dialog open
    Steps:
      1. Leave domain field empty, click Submit
      2. Assert error message "Domain is required"
      3. Type "not valid!" in domain field
      4. Click Submit
      5. Assert error message contains "invalid"
    Expected Result: Form rejects invalid input
    Evidence: .sisyphus/evidence/task-17-domain-validation-ui.png
  ```

  **Commit**: YES
  - Message: `feat(ui): domain management pages + CRUD`
  - Files: `web/src/app/domains/`
  - Pre-commit: `cd web && npx vitest run && npx tsc --noEmit`

- [ ] 18. UI - user/mailbox management pages

  **What to do**:
  - TDD: Write component tests first
  - Create user list page: TanStack Table (email, domain, quota, status, created, actions)
  - Create user dialog: form with email, password, domain selector, quota input
  - Password field: show/hide toggle, strength indicator
  - Filter by domain (dropdown above table)
  - Inline actions: delete with confirmation
  - Empty state + loading skeleton + error state
  - Toast notifications

  **Must NOT do**:
  - No password reset UI
  - No quota usage display (just configured value)
  - No user profile editing beyond create/delete

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 15-17,19-22)
  - **Blocks**: Task 31
  - **Blocked By**: Tasks 15, 16

  **References**:
  - `web/src/hooks/` (Task 16) - useUsers(), useCreateUser()
  - `web/src/types/api.ts` (Task 4) - User type
  - shadcn/ui Form + Select + Input components

  **Acceptance Criteria**:
  - [ ] User list renders with correct columns
  - [ ] Add user form validates email format and password
  - [ ] Domain selector shows existing domains
  - [ ] Filter by domain works
  - [ ] All component tests pass

  **QA Scenarios**:
  ```
  Scenario: User creation via UI
    Tool: Playwright
    Preconditions: App running, domain "local.test" exists
    Steps:
      1. Navigate to http://localhost:3000/users
      2. Click "Add User"
      3. Fill email: "alice@local.test"
      4. Fill password: "TestPass123!"
      5. Select domain: "local.test"
      6. Click Submit
      7. Assert table row with "alice@local.test" appears
    Expected Result: User created and visible in table
    Evidence: .sisyphus/evidence/task-18-user-create-ui.png

  Scenario: User form rejects invalid email
    Tool: Playwright
    Preconditions: Add user dialog open
    Steps:
      1. Type "not-an-email" in email field
      2. Click Submit
      3. Assert error message about invalid email format
    Expected Result: Validation prevents bad input
    Evidence: .sisyphus/evidence/task-18-user-validation-ui.png
  ```

  **Commit**: YES
  - Message: `feat(ui): user/mailbox management pages`
  - Files: `web/src/app/users/`
  - Pre-commit: `cd web && npx vitest run && npx tsc --noEmit`

- [ ] 19. UI - real-time log viewer

  **What to do**:
  - TDD: Write component tests first
  - Create log viewer page with WebSocket connection to /api/logs/stream
  - Auto-scroll log entries (with pause on manual scroll-up)
  - Syntax highlighting: timestamps, log levels (color-coded), service names
  - Filter bar: service dropdown (postfix/dovecot/rspamd/all), level dropdown, search input
  - Log entry format: `[timestamp] [service] [level] message` with monospace font
  - "Clear" button to reset visible log buffer
  - Connection status indicator (connected/reconnecting/disconnected)
  - Virtualized list for performance (react-window or similar)

  **Must NOT do**:
  - No log export/download
  - No log persistence settings
  - No regex search

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 15-18,20-22)
  - **Blocks**: Task 31
  - **Blocked By**: Tasks 15, 12

  **References**:
  - `web/src/hooks/` (Task 16) - useLogStream() WebSocket hook
  - `internal/domain/models.go` (Task 4) - LogEntry model
  - Vercel log viewer pattern: monospace, color-coded, auto-scroll

  **Acceptance Criteria**:
  - [ ] WebSocket connects and displays incoming log entries
  - [ ] Filter by service shows only matching entries
  - [ ] Auto-scroll works, pauses on manual scroll
  - [ ] Connection indicator shows correct state
  - [ ] Component tests pass

  **QA Scenarios**:
  ```
  Scenario: Log viewer receives real-time events
    Tool: Playwright
    Preconditions: Full stack running, log viewer page loaded
    Steps:
      1. Navigate to http://localhost:3000/logs
      2. Assert connection indicator shows "Connected" (green dot)
      3. In separate terminal: swaks --server localhost --port 1025 --from test@local.test --to alice@local.test
      4. Wait 3s
      5. Assert new log entry appears containing "alice@local.test"
      6. Screenshot
    Expected Result: SMTP event appears in log viewer within 3s
    Evidence: .sisyphus/evidence/task-19-log-realtime.png

  Scenario: Log filtering works
    Tool: Playwright
    Preconditions: Log viewer with mixed entries
    Steps:
      1. Select "postfix" from service filter dropdown
      2. Assert all visible entries contain "postfix" service tag
      3. Assert no "dovecot" or "rspamd" entries visible
    Expected Result: Only filtered service shown
    Evidence: .sisyphus/evidence/task-19-log-filter.png
  ```

  **Commit**: YES
  - Message: `feat(ui): real-time log viewer with WebSocket + filters`
  - Files: `web/src/app/logs/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 20. UI - health/monitoring dashboard

  **What to do**:
  - TDD: Write component tests first
  - Create dashboard page (default landing page)
  - KPI cards row: total domains, total users, total emails today, uptime
  - Service health grid: card per service (Postfix, Dovecot, Rspamd, Redis, API) with status badge + response time
  - Auto-refresh health every 10s via TanStack Query refetchInterval
  - Mail queue summary card: queued count, deferred count
  - Recent activity feed: last 10 log entries (mini log viewer)
  - Visual status: green/yellow/red badges, pulse animation on healthy

  **Must NOT do**:
  - No historical charts/graphs (deferred)
  - No alerting configuration
  - No Prometheus metrics
  - No custom time ranges

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 15-19,21,22)
  - **Blocks**: Task 31
  - **Blocked By**: Tasks 15, 11

  **References**:
  - `web/src/hooks/` (Task 16) - useHealth()
  - `internal/domain/models.go` (Task 4) - HealthStatus model
  - Grafana dashboard card pattern
  - Coolify dashboard: KPI strip + service cards

  **Acceptance Criteria**:
  - [ ] Dashboard shows all 5 service health cards
  - [ ] Health auto-refreshes (status changes reflected within 10s)
  - [ ] KPI cards show correct counts
  - [ ] Recent activity shows last log entries
  - [ ] Component tests pass

  **QA Scenarios**:
  ```
  Scenario: Dashboard shows healthy services
    Tool: Playwright
    Preconditions: Full stack running and healthy
    Steps:
      1. Navigate to http://localhost:3000 (dashboard)
      2. Assert 5 service cards visible: `[data-testid="service-card"]` count = 5
      3. Assert all cards show green status badge
      4. Assert KPI cards visible with numeric values
      5. Screenshot
    Expected Result: All services healthy, KPIs populated
    Evidence: .sisyphus/evidence/task-20-dashboard-healthy.png

  Scenario: Dashboard reflects degraded service
    Tool: Playwright + Bash
    Preconditions: Full stack running
    Steps:
      1. docker compose stop rspamd
      2. Wait 12s (health refresh cycle)
      3. Navigate to http://localhost:3000
      4. Assert rspamd card shows red/yellow status badge
      5. docker compose start rspamd
      6. Screenshot
    Expected Result: Degraded service shown visually
    Evidence: .sisyphus/evidence/task-20-dashboard-degraded.png
  ```

  **Commit**: YES
  - Message: `feat(ui): health monitoring dashboard + KPI cards`
  - Files: `web/src/app/(dashboard)/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 21. UI - DKIM management page

  **What to do**:
  - TDD: Write component tests first
  - Create DKIM page: list of domains with DKIM key status (generated/not generated)
  - "Generate Key" button per domain → calls API → shows public key + DNS record
  - DNS record display: copyable code block with TXT record value
  - Key info: algorithm, bits, selector, created date
  - "Rotate Key" action with confirmation (warns about DNS propagation)
  - Visual indicator: checkmark if key exists, warning if not

  **Must NOT do**:
  - No DNS verification (checking if record is published)
  - No automatic DNS configuration
  - No Ed25519 keys

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 15-20,22)
  - **Blocks**: Task 31
  - **Blocked By**: Tasks 15, 10

  **References**:
  - `web/src/hooks/` (Task 16) - useDkim(), useGenerateDkim()
  - `internal/domain/models.go` (Task 4) - DKIMKey model
  - shadcn/ui Card + Badge + Copy button pattern

  **Acceptance Criteria**:
  - [ ] DKIM page lists all domains with key status
  - [ ] Generate key shows public key in copyable format
  - [ ] DNS TXT record displayed correctly
  - [ ] Rotate key shows confirmation warning
  - [ ] Component tests pass

  **QA Scenarios**:
  ```
  Scenario: Generate DKIM key via UI
    Tool: Playwright
    Preconditions: App running, domain "local.test" exists, no DKIM key yet
    Steps:
      1. Navigate to http://localhost:3000/dkim
      2. Assert "local.test" row shows "No key" status
      3. Click "Generate Key" button for local.test
      4. Wait for response
      5. Assert public key code block appears
      6. Assert DNS record contains "v=DKIM1"
      7. Click copy button
      8. Screenshot
    Expected Result: DKIM key generated and DNS record shown
    Evidence: .sisyphus/evidence/task-21-dkim-generate-ui.png

  Scenario: DKIM key rotation warning
    Tool: Playwright
    Preconditions: DKIM key exists for local.test
    Steps:
      1. Click "Rotate Key" for local.test
      2. Assert confirmation dialog with warning about DNS propagation
      3. Click Cancel
      4. Assert key unchanged
    Expected Result: Rotation requires explicit confirmation
    Evidence: .sisyphus/evidence/task-21-dkim-rotate-warning.png
  ```

  **Commit**: YES
  - Message: `feat(ui): DKIM key management page`
  - Files: `web/src/app/dkim/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 22. CLI tool - oxmail

  **What to do**:
  - TDD: Write tests first for CLI commands
  - Create `cmd/oxmail/main.go` with cobra CLI framework
  - Commands:
    - `oxmail domain add <domain>` - create domain via API
    - `oxmail domain list` - list domains
    - `oxmail domain rm <domain>` - delete domain
    - `oxmail user add <email> --password <pass>` - create user
    - `oxmail user list [--domain <domain>]` - list users
    - `oxmail user rm <email>` - delete user
    - `oxmail alias add <source> <dest1,dest2>` - create alias
    - `oxmail status` - show health of all services
    - `oxmail logs [-f] [--service <name>]` - tail logs (follow mode)
    - `oxmail send-test <from> <to>` - send test email via SMTP
  - Configure API base URL via env var `OXMAIL_API_URL` (default: http://localhost:8080)
  - Colored output: green for healthy, red for errors, yellow for warnings
  - Table output for list commands (tabwriter)
  - JSON output flag: `--json` for scripting

  **Must NOT do**:
  - No interactive prompts (all args via flags)
  - No config file (env vars only)
  - No shell completion generation

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 15-21)
  - **Blocks**: None
  - **Blocked By**: Tasks 4, 7, 8

  **References**:
  - `api-spec/openapi.yaml` (Task 4) - API endpoints to call
  - `internal/domain/models.go` (Task 4) - Response models
  - cobra CLI: github.com/spf13/cobra
  - docker-mailserver `setup.sh` (reference for UX)

  **Acceptance Criteria**:
  - [ ] `go build ./cmd/oxmail/` exits 0
  - [ ] `oxmail status` shows service health table
  - [ ] `oxmail domain add local.test` creates domain (exit 0)
  - [ ] `oxmail domain list` shows created domain
  - [ ] `oxmail user add alice@local.test --password TestPass123!` creates user
  - [ ] `oxmail --json domain list` outputs valid JSON
  - [ ] All CLI tests pass

  **QA Scenarios**:
  ```
  Scenario: CLI domain lifecycle
    Tool: Bash
    Preconditions: API running on localhost:8080
    Steps:
      1. ./oxmail domain add local.test
      2. Assert exit code 0, output contains "created"
      3. ./oxmail domain list
      4. Assert output contains "local.test"
      5. ./oxmail domain rm local.test
      6. Assert exit code 0
      7. ./oxmail domain list
      8. Assert output does NOT contain "local.test"
    Expected Result: Full CLI domain lifecycle
    Evidence: .sisyphus/evidence/task-22-cli-domain.txt

  Scenario: CLI JSON output
    Tool: Bash
    Preconditions: API running, domain exists
    Steps:
      1. ./oxmail --json domain list | jq '.'
      2. Assert exit code 0 (valid JSON)
      3. Assert .domains is array
    Expected Result: Valid JSON output for scripting
    Evidence: .sisyphus/evidence/task-22-cli-json.txt
  ```

  **Commit**: YES
  - Message: `feat(cli): oxmail tool for domain/user/alias management`
  - Files: `cmd/oxmail/`
  - Pre-commit: `go test ./cmd/oxmail/...`

- [ ] 23. Go API - IMAP proxy/bridge for webmail

  **What to do**:
  - TDD: Write tests first for IMAP bridge
  - Implement `internal/mail/imap_bridge.go`: Connect to Dovecot IMAP as proxy
  - REST endpoints for webmail:
    - `GET /api/mail/inbox?page=1&limit=50` - list messages (subject, from, date, read status, has attachment)
    - `GET /api/mail/messages/{id}` - get full message (headers + body + attachments)
    - `GET /api/mail/messages/{id}/attachments/{filename}` - download attachment
    - `DELETE /api/mail/messages/{id}` - delete/move to trash
    - `PATCH /api/mail/messages/{id}` - mark read/unread, flag/unflag
  - Parse MIME messages: extract plain text, HTML body, attachments list
  - Thread grouping by In-Reply-To / References headers
  - Sanitize HTML email bodies (prevent XSS)
  - Pagination with cursor-based approach

  **Must NOT do**:
  - No folder management (INBOX only for v1)
  - No Sieve filter management
  - No full-text search indexing (use IMAP SEARCH)
  - No caching layer (direct IMAP queries)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 24)
  - **Parallel Group**: Wave 4 (with Tasks 24-29)
  - **Blocks**: Task 25
  - **Blocked By**: Task 14

  **References**:
  - `internal/mail/dovecot_manager.go` (Task 14) - Dovecot connection patterns
  - `internal/domain/models.go` (Task 4) - Message/Thread models
  - go-imap library: github.com/emersion/go-imap
  - MIME parsing: github.com/emersion/go-message
  - bluemonday for HTML sanitization

  **Acceptance Criteria**:
  - [ ] `go test ./internal/mail/ -run TestIMAPBridge` passes
  - [ ] `GET /api/mail/inbox` returns message list after sending test email
  - [ ] `GET /api/mail/messages/{id}` returns full parsed message
  - [ ] HTML bodies sanitized (no script tags, no event handlers)
  - [ ] Thread grouping works for reply chains

  **QA Scenarios**:
  ```
  Scenario: Webmail inbox lists received emails
    Tool: Bash
    Preconditions: User alice@local.test exists, email sent to her
    Steps:
      1. swaks --server localhost --port 1025 --from bob@local.test --to alice@local.test --header "Subject: Bridge Test" --body "hello from bridge"
      2. Wait 5s
      3. curl -s http://localhost:8080/api/mail/inbox?user=alice@local.test | jq '.messages[0].subject'
      4. Assert = "Bridge Test"
      5. Extract message id from response
      6. curl -s http://localhost:8080/api/mail/messages/{id}?user=alice@local.test | jq '.body'
      7. Assert contains "hello from bridge"
    Expected Result: Email accessible via REST bridge
    Evidence: .sisyphus/evidence/task-23-imap-bridge.txt

  Scenario: HTML sanitization prevents XSS
    Tool: Bash
    Preconditions: User exists
    Steps:
      1. Send email with HTML body containing <script>alert('xss')</script>
      2. Fetch message via API
      3. Assert response body does NOT contain "<script>"
      4. Assert response body does NOT contain "alert("
    Expected Result: Dangerous HTML stripped
    Evidence: .sisyphus/evidence/task-23-html-sanitize.txt
  ```

  **Commit**: YES
  - Message: `feat(api): IMAP bridge for webmail + MIME parsing + HTML sanitization`
  - Files: `internal/mail/imap_bridge.go`, `internal/api/mail_handler.go`
  - Pre-commit: `go test ./...`

- [ ] 24. Go API - SMTP submission endpoint for compose

  **What to do**:
  - TDD: Write tests first for SMTP submission
  - Implement `internal/mail/smtp_sender.go`: Submit email via Postfix submission port (587)
  - REST endpoint: `POST /api/mail/send` with body: {from, to[], cc[], bcc[], subject, body_text, body_html, attachments[]}
  - Authenticate as sending user against Dovecot (verify sender owns the from address)
  - Build proper MIME message with headers (Date, Message-ID, MIME-Version, Content-Type)
  - Support multipart messages (text + HTML + attachments)
  - Attachment handling: base64 encoded in request, proper MIME encoding in message
  - Rate limit: max 50 emails per hour per user (in-memory counter)
  - Return message-id on success for tracking

  **Must NOT do**:
  - No draft saving (compose is fire-and-forget)
  - No scheduled sending
  - No template system
  - No external recipient validation (local delivery only)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Task 23)
  - **Parallel Group**: Wave 4 (with Tasks 23,25-29)
  - **Blocks**: Task 26
  - **Blocked By**: Task 13

  **References**:
  - `internal/mail/postfix_manager.go` (Task 13) - Postfix connection
  - Postfix submission port (587) with SASL auth
  - net/smtp Go stdlib for SMTP client
  - MIME message construction: github.com/emersion/go-message

  **Acceptance Criteria**:
  - [ ] `go test ./internal/mail/ -run TestSMTPSender` passes
  - [ ] `POST /api/mail/send` with valid payload → 200 + message-id
  - [ ] Sent email arrives in recipient's IMAP inbox
  - [ ] Sender address validated (can't send as another user)
  - [ ] Rate limit enforced (51st email in hour → 429)

  **QA Scenarios**:
  ```
  Scenario: Send email via API
    Tool: Bash
    Preconditions: Users alice@local.test and bob@local.test exist
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/mail/send -H "Content-Type: application/json" -d '{"from":"alice@local.test","to":["bob@local.test"],"subject":"API Send Test","body_text":"Hello from API"}'
      2. Assert status code = 200
      3. Assert response contains "message_id"
      4. Wait 5s
      5. curl -s http://localhost:8080/api/mail/inbox?user=bob@local.test | jq '.messages[0].subject'
      6. Assert = "API Send Test"
    Expected Result: Email sent and delivered via API
    Evidence: .sisyphus/evidence/task-24-smtp-send.txt

  Scenario: Sender spoofing rejected
    Tool: Bash
    Preconditions: alice@local.test exists, bob@local.test exists
    Steps:
      1. curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/mail/send -H "Content-Type: application/json" -d '{"from":"bob@local.test","to":["alice@local.test"],"subject":"Spoofed","body_text":"fake","auth_user":"alice@local.test"}'
      2. Assert status code = 403 or 400
      3. Assert response mentions "not authorized" or "sender mismatch"
    Expected Result: Cannot send as another user
    Evidence: .sisyphus/evidence/task-24-sender-spoof.txt
  ```

  **Commit**: YES
  - Message: `feat(api): SMTP submission endpoint for webmail compose`
  - Files: `internal/mail/smtp_sender.go`, `internal/api/send_handler.go`
  - Pre-commit: `go test ./...`

- [ ] 25. UI - webmail inbox + thread view

  **What to do**:
  - TDD: Write component tests first
  - Create webmail inbox page: three-pane layout (folder list | message list | message preview)
  - Message list: sender, subject, date, read/unread indicator, attachment icon
  - Thread view: grouped messages with inline expansion
  - Message preview: rendered HTML (sanitized), plain text fallback, attachment list
  - Read/unread toggle on click
  - Keyboard shortcuts: j/k (next/prev), Enter (open), u (mark unread), Delete (trash)
  - Virtualized message list for performance
  - Empty inbox state with illustration
  - Loading states: skeleton for list, spinner for message body

  **Must NOT do**:
  - No folder navigation (INBOX only)
  - No drag-and-drop
  - No message selection/bulk actions
  - No star/flag UI

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 23,24,26-29)
  - **Blocks**: Tasks 27, 31
  - **Blocked By**: Tasks 15, 23

  **References**:
  - `web/src/hooks/` (Task 16) - API hooks for mail endpoints
  - `internal/api/mail_handler.go` (Task 23) - Mail REST endpoints
  - Linear.app inbox pattern: clean list, preview pane
  - Gmail thread grouping UX

  **Acceptance Criteria**:
  - [ ] Inbox displays received messages with correct metadata
  - [ ] Clicking message shows preview in right pane
  - [ ] Thread grouping works for reply chains
  - [ ] Keyboard navigation (j/k) moves selection
  - [ ] Component tests pass

  **QA Scenarios**:
  ```
  Scenario: Webmail inbox displays emails
    Tool: Playwright
    Preconditions: alice@local.test has 3 emails in inbox
    Steps:
      1. Navigate to http://localhost:3000/mail
      2. Assert message list shows 3 items: `[data-testid="message-row"]` count = 3
      3. Assert first message shows sender, subject, date
      4. Click first message
      5. Assert preview pane shows message body
      6. Screenshot
    Expected Result: Three-pane webmail with message preview
    Evidence: .sisyphus/evidence/task-25-webmail-inbox.png

  Scenario: Keyboard navigation works
    Tool: Playwright
    Preconditions: Inbox loaded with messages
    Steps:
      1. Press "j" key
      2. Assert second message row has active/selected state
      3. Press "k" key
      4. Assert first message row has active/selected state
      5. Press "Enter"
      6. Assert preview pane updates to show first message content
    Expected Result: Keyboard shortcuts navigate messages
    Evidence: .sisyphus/evidence/task-25-webmail-keyboard.png
  ```

  **Commit**: YES
  - Message: `feat(ui): webmail inbox + three-pane layout + thread view`
  - Files: `web/src/app/mail/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 26. UI - webmail compose + send

  **What to do**:
  - TDD: Write component tests first
  - Create compose dialog/page: To, Cc, Bcc fields, Subject, Body editor
  - Rich text editor: bold, italic, links, lists (use tiptap or similar lightweight editor)
  - Attachment upload: drag-and-drop zone + file picker button
  - Recipient fields: email input with validation, multiple recipients
  - "Send" button with loading state
  - Success toast with "Message sent" confirmation
  - Discard confirmation if body has content
  - "New Message" button in inbox toolbar
  - Reply/Reply All buttons in message preview → pre-fill compose

  **Must NOT do**:
  - No draft auto-save
  - No scheduled send
  - No email templates
  - No signature management

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 23-25,27-29)
  - **Blocks**: Task 31
  - **Blocked By**: Tasks 15, 24

  **References**:
  - `internal/api/send_handler.go` (Task 24) - Send endpoint
  - `web/src/hooks/` (Task 16) - API hooks
  - tiptap editor: https://tiptap.dev/
  - shadcn/ui Dialog + Form patterns

  **Acceptance Criteria**:
  - [ ] Compose form renders with all fields (to, cc, bcc, subject, body)
  - [ ] Send button calls API and shows success toast
  - [ ] Sent email arrives in recipient inbox
  - [ ] Reply pre-fills To and Subject with "Re: "
  - [ ] Attachment upload works (file appears in attachment list)
  - [ ] Component tests pass

  **QA Scenarios**:
  ```
  Scenario: Compose and send email via webmail
    Tool: Playwright
    Preconditions: alice@local.test logged in, bob@local.test exists
    Steps:
      1. Navigate to http://localhost:3000/mail
      2. Click "New Message" button
      3. Fill To: "bob@local.test"
      4. Fill Subject: "Webmail Compose Test"
      5. Type body: "Hello from webmail compose"
      6. Click "Send"
      7. Assert success toast appears with "Message sent"
      8. Navigate to bob's inbox
      9. Assert message "Webmail Compose Test" appears
    Expected Result: Email composed and delivered via UI
    Evidence: .sisyphus/evidence/task-26-compose-send.png

  Scenario: Discard confirmation on non-empty compose
    Tool: Playwright
    Preconditions: Compose dialog open with text in body
    Steps:
      1. Type "some content" in body
      2. Click close/discard button
      3. Assert confirmation dialog: "Discard this message?"
      4. Click "Cancel"
      5. Assert compose still open with content preserved
    Expected Result: Unsaved compose protected from accidental discard
    Evidence: .sisyphus/evidence/task-26-compose-discard.png
  ```

  **Commit**: YES
  - Message: `feat(ui): webmail compose + rich editor + send`
  - Files: `web/src/app/mail/compose/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 27. UI - webmail search

  **What to do**:
  - TDD: Write component tests first
  - Create search bar in webmail toolbar with search chips pattern
  - Search filters: from:, to:, subject:, has:attachment, date range
  - Search results page: same message list format as inbox
  - Highlight matching terms in results
  - Search via API: `GET /api/mail/search?q=...&user=...`
  - Backend: use IMAP SEARCH command through bridge
  - Empty results state: "No messages match your search"
  - Recent searches dropdown (last 5, stored in localStorage)

  **Must NOT do**:
  - No full-text indexing (rely on IMAP SEARCH)
  - No saved searches
  - No search across multiple mailboxes
  - No regex search

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 23-26,28,29)
  - **Blocks**: Task 31
  - **Blocked By**: Task 25

  **References**:
  - `internal/mail/imap_bridge.go` (Task 23) - IMAP SEARCH support
  - Gmail search chip pattern
  - shadcn/ui Input + Badge components

  **Acceptance Criteria**:
  - [ ] Search bar accepts text and returns matching messages
  - [ ] Filter chips work (from:, subject:, has:attachment)
  - [ ] Results display in message list format
  - [ ] Empty state shown for no results
  - [ ] Component tests pass

  **QA Scenarios**:
  ```
  Scenario: Search finds matching email
    Tool: Playwright
    Preconditions: alice@local.test has emails with various subjects
    Steps:
      1. Navigate to http://localhost:3000/mail
      2. Click search bar
      3. Type "Bridge Test"
      4. Press Enter
      5. Assert results contain message with subject "Bridge Test"
      6. Assert results do NOT contain unrelated messages
    Expected Result: Search returns relevant results
    Evidence: .sisyphus/evidence/task-27-search-results.png

  Scenario: Search empty state
    Tool: Playwright
    Preconditions: Inbox loaded
    Steps:
      1. Search for "xyznonexistent12345"
      2. Assert empty state message: "No messages match your search"
    Expected Result: Graceful empty state
    Evidence: .sisyphus/evidence/task-27-search-empty.png
  ```

  **Commit**: YES
  - Message: `feat(ui): webmail search with filter chips`
  - Files: `web/src/app/mail/search/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 28. Docker Compose full integration + health checks

  **What to do**:
  - Wire all services together in docker-compose.yml with proper depends_on + healthcheck
  - Health checks for each service:
    - API: `curl -f http://localhost:8080/health`
    - Postfix: `postfix status` or SMTP EHLO check
    - Dovecot: `doveadm` or IMAP capability check
    - Rspamd: `curl -f http://localhost:11333/ping`
    - Redis: `redis-cli ping`
    - Web: `curl -f http://localhost:3000`
  - Startup order: redis → postfix/dovecot/rspamd → api → web
  - Shared volumes: mail-data, config-data, dkim-keys
  - Network: internal network for inter-service, expose only API/Web/SMTP/IMAP to host
  - Environment variable passthrough from .env
  - Resource limits: memory limits per container
  - Restart policy: unless-stopped
  - Full stack test: `docker compose up -d` → all healthy within 120s

  **Must NOT do**:
  - No Kubernetes
  - No Docker Swarm
  - No external reverse proxy config
  - No production TLS termination

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (integration point)
  - **Parallel Group**: Wave 4 (after Tasks 13, 14, 11 complete)
  - **Blocks**: Tasks 29, 30, 32, 33, 34
  - **Blocked By**: Tasks 13, 14, 11

  **References**:
  - `docker-compose.yml` (Task 1) - Skeleton to complete
  - `docker/` (Task 5) - Service Dockerfiles
  - Mailcow docker-compose pattern (reference for health checks)
  - Docker Compose healthcheck docs

  **Acceptance Criteria**:
  - [ ] `docker compose up -d` starts all services
  - [ ] `docker compose ps` shows all healthy within 120s
  - [ ] Send email end-to-end: SMTP → IMAP → API → works
  - [ ] `docker stats --no-stream` shows total < 1.5GB idle
  - [ ] Services restart automatically on failure
  - [ ] Internal services not exposed to host (only API/Web/SMTP/IMAP)

  **QA Scenarios**:
  ```
  Scenario: Full stack cold start
    Tool: Bash
    Preconditions: Docker installed, no running containers
    Steps:
      1. docker compose down -v (clean slate)
      2. docker compose up -d
      3. Run scripts/wait-healthy.sh (timeout 120s)
      4. Assert exit code 0
      5. docker compose ps --format json | jq '.[].Health'
      6. Assert all services show "healthy"
      7. docker stats --no-stream --format "{{.MemUsage}}"
      8. Assert total < 1.5GB
    Expected Result: All services healthy within 120s, under RAM budget
    Evidence: .sisyphus/evidence/task-28-cold-start.txt

  Scenario: End-to-end email flow
    Tool: Bash
    Preconditions: Full stack healthy
    Steps:
      1. curl -s -X POST http://localhost:8080/api/domains -H "Content-Type: application/json" -d '{"domain":"local.test"}'
      2. curl -s -X POST http://localhost:8080/api/users -H "Content-Type: application/json" -d '{"email":"alice@local.test","password":"TestPass123!"}'
      3. Wait 5s
      4. swaks --server localhost --port 1025 --from bob@local.test --to alice@local.test --header "Subject: E2E Test" --body "end to end"
      5. Wait 5s
      6. curl -s http://localhost:8080/api/mail/inbox?user=alice@local.test | jq '.messages[0].subject'
      7. Assert = "E2E Test"
    Expected Result: Full email lifecycle works
    Evidence: .sisyphus/evidence/task-28-e2e-flow.txt
  ```

  **Commit**: YES
  - Message: `feat(docker): full stack integration + health checks + resource limits`
  - Files: `docker-compose.yml`, `docker-compose.dev.yml`
  - Pre-commit: `docker compose config --quiet`

- [ ] 29. Dev mode: seed data + test email tools

  **What to do**:
  - Implement `scripts/seed.sh` fully: create domain, users, send test emails
  - Seed data:
    - Domain: `local.test`
    - Users: `alice@local.test`, `bob@local.test` (password: `TestPass123!`)
    - Send 5 test emails between alice and bob (varied subjects, one with attachment)
    - Generate DKIM key for local.test
  - Create `GET /api/dev/send-test` endpoint (dev mode only): quick test email form
  - Add dev mode detection: `OXMAIL_MODE=dev` env var
  - Dev mode features: no auth required on admin API, relaxed rate limits, seed endpoint
  - Update `make seed` target to call seed script after stack is healthy
  - Add `make reset` target: wipe all data and re-seed

  **Must NOT do**:
  - No production seed data
  - No fake email generation (just simple test emails)
  - No load testing tools

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 23-28)
  - **Blocks**: Task 34
  - **Blocked By**: Task 28

  **References**:
  - `scripts/seed.sh` (Task 6) - Skeleton to complete
  - `cmd/oxmail/` (Task 22) - Use CLI for seeding
  - docker-mailserver setup.sh pattern

  **Acceptance Criteria**:
  - [ ] `make seed` creates domain + users + test emails (exit 0)
  - [ ] After seed: alice has 3+ emails in inbox
  - [ ] After seed: DKIM key exists for local.test
  - [ ] `make reset` wipes and re-seeds cleanly
  - [ ] Dev mode endpoint accessible without auth

  **QA Scenarios**:
  ```
  Scenario: Seed script populates test data
    Tool: Bash
    Preconditions: Full stack healthy, no data
    Steps:
      1. make seed
      2. Assert exit code 0
      3. curl -s http://localhost:8080/api/domains | jq '.domains | length'
      4. Assert >= 1
      5. curl -s http://localhost:8080/api/users | jq '.users | length'
      6. Assert >= 2
      7. curl -s http://localhost:8080/api/mail/inbox?user=alice@local.test | jq '.messages | length'
      8. Assert >= 3
    Expected Result: Seed creates full test environment
    Evidence: .sisyphus/evidence/task-29-seed.txt

  Scenario: Reset wipes and re-seeds
    Tool: Bash
    Preconditions: Seeded data exists
    Steps:
      1. make reset
      2. Assert exit code 0
      3. curl -s http://localhost:8080/api/domains | jq '.domains | length'
      4. Assert = 1 (fresh seed, not accumulated)
    Expected Result: Clean reset to known state
    Evidence: .sisyphus/evidence/task-29-reset.txt
  ```

  **Commit**: YES
  - Message: `feat(dev): seed data + test email tools + dev mode`
  - Files: `scripts/seed.sh`, `internal/api/dev_handler.go`
  - Pre-commit: `make seed`

- [ ] 30. Security hardening - open relay test, auth enforcement

  **What to do**:
  - TDD: Write security tests first (these MUST fail initially)
  - Implement admin API authentication: JWT or session-based auth
  - Create admin user on first run (from env var OXMAIL_ADMIN_PASSWORD)
  - Login endpoint: `POST /api/auth/login` → returns token
  - Protect all admin endpoints (domains, users, aliases, DKIM) with auth middleware
  - Webmail auth: user authenticates with their email/password
  - Open relay prevention tests:
    - Reject relay to external domains
    - Reject unauthenticated submission on port 587
    - Accept only for configured virtual domains
  - Security headers on API: CORS, X-Frame-Options, CSP
  - Rate limiting on login endpoint (5 attempts per minute)
  - No secrets in frontend bundle (verify with grep)
  - Audit log: record admin actions (who, what, when) to SQLite

  **Must NOT do**:
  - No OAuth/OIDC
  - No 2FA
  - No API keys (JWT only)
  - No IP-based access control

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 31-34)
  - **Parallel Group**: Wave 5
  - **Blocks**: F1-F4
  - **Blocked By**: Task 28

  **References**:
  - `internal/api/server.go` (Task 2) - Middleware chain
  - Postfix relay restrictions: smtpd_relay_restrictions
  - JWT: github.com/golang-jwt/jwt/v5
  - OWASP security headers

  **Acceptance Criteria**:
  - [ ] Unauthenticated `GET /api/domains` → 401
  - [ ] `POST /api/auth/login` with correct creds → 200 + token
  - [ ] Authenticated request with token → 200
  - [ ] Open relay test: `swaks --to external@gmail.com` → rejected
  - [ ] Submission without auth on 587 → rejected
  - [ ] No secrets in `web/.next/` bundle
  - [ ] Login rate limit: 6th attempt within minute → 429

  **QA Scenarios**:
  ```
  Scenario: Admin API requires authentication
    Tool: Bash
    Preconditions: Stack running in production mode (OXMAIL_MODE=production)
    Steps:
      1. curl -s -w "\n%{http_code}" http://localhost:8080/api/domains
      2. Assert status = 401
      3. curl -s -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"${OXMAIL_ADMIN_PASSWORD}"}'
      4. Extract token from response
      5. curl -s -w "\n%{http_code}" -H "Authorization: Bearer ${token}" http://localhost:8080/api/domains
      6. Assert status = 200
    Expected Result: Auth enforced, valid token grants access
    Evidence: .sisyphus/evidence/task-30-auth-enforce.txt

  Scenario: Open relay prevention
    Tool: Bash
    Preconditions: Stack running
    Steps:
      1. swaks --server localhost --port 1025 --from attacker@evil.com --to victim@gmail.com
      2. Assert exit code != 0
      3. Assert output contains "Relay access denied" or "Recipient address rejected"
      4. swaks --server localhost --port 1025 --from alice@local.test --to bob@local.test --body "local ok"
      5. Assert exit code = 0 (local delivery works)
    Expected Result: External relay blocked, local delivery allowed
    Evidence: .sisyphus/evidence/task-30-open-relay.txt
  ```

  **Commit**: YES
  - Message: `feat(security): auth enforcement + open relay protection + rate limiting`
  - Files: `internal/api/auth_handler.go`, `internal/api/middleware/auth.go`
  - Pre-commit: `go test ./... -run TestSecurity`

- [ ] 31. UI polish - animations, transitions, empty/error states

  **What to do**:
  - Add page transition animations (framer-motion or CSS transitions)
  - Add micro-interactions: button hover states, focus rings, card hover lift
  - Polish all empty states with illustrations or icons + helpful CTAs
  - Polish all error states: clear error message, retry button, support context
  - Loading states: consistent skeleton patterns across all pages
  - Toast notifications: consistent styling, auto-dismiss, action buttons
  - Sidebar active state animation (subtle indicator slide)
  - Table row hover highlight
  - Dialog open/close animations (scale + fade)
  - Command palette: smooth open animation, result highlight transitions
  - Ensure all interactive elements have visible focus states (accessibility)
  - Dark mode polish: verify contrast ratios meet WCAG AA (4.5:1 text, 3:1 UI)

  **Must NOT do**:
  - No heavy animations that affect performance
  - No animation library larger than framer-motion
  - No custom cursor effects
  - No parallax or scroll-based animations

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 30,32-34)
  - **Parallel Group**: Wave 5
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 17-21, 25-27

  **References**:
  - All UI pages from Tasks 17-21, 25-27
  - Linear.app transitions: subtle, fast (150-200ms)
  - Vercel dashboard: focus states, hover effects
  - WCAG 2.1 AA contrast requirements

  **Acceptance Criteria**:
  - [ ] All pages have loading/empty/error states
  - [ ] Page transitions smooth (no layout shift)
  - [ ] All interactive elements have focus-visible styles
  - [ ] Contrast ratios pass WCAG AA (verified with axe-core)
  - [ ] Animations complete within 200ms (no sluggish feel)

  **QA Scenarios**:
  ```
  Scenario: Empty states render correctly
    Tool: Playwright
    Preconditions: Fresh stack with no data
    Steps:
      1. Navigate to http://localhost:3000/domains
      2. Assert empty state visible with CTA button
      3. Navigate to http://localhost:3000/users
      4. Assert empty state visible
      5. Navigate to http://localhost:3000/mail
      6. Assert empty inbox state
      7. Screenshot each page
    Expected Result: All pages show helpful empty states
    Evidence: .sisyphus/evidence/task-31-empty-states.png

  Scenario: Accessibility contrast check
    Tool: Playwright
    Preconditions: App loaded
    Steps:
      1. Navigate to http://localhost:3000
      2. Run axe-core accessibility audit via page.evaluate
      3. Filter results for "color-contrast" violations
      4. Assert zero contrast violations
    Expected Result: All text meets WCAG AA contrast
    Evidence: .sisyphus/evidence/task-31-contrast-audit.txt
  ```

  **Commit**: YES
  - Message: `feat(ui): polish animations + transitions + empty/error states`
  - Files: `web/src/components/`, `web/src/app/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 32. Performance tuning - RAM budget, cold start optimization

  **What to do**:
  - Profile idle RAM usage per container, identify optimization targets
  - Postfix: disable unused features, minimize lookup tables loaded
  - Dovecot: tune memory allocation, disable unused plugins
  - Rspamd: reduce worker count for local use, disable heavy modules
  - Redis: set maxmemory to 64MB with eviction policy
  - Go API: tune GC, connection pool sizes
  - Next.js: analyze bundle size, remove unused dependencies, enable compression
  - Docker: use multi-stage builds, minimize image sizes (alpine base)
  - Measure and document: cold start time, idle RAM, peak RAM during email send
  - Set container memory limits in docker-compose.yml
  - Target: idle < 1.5GB total, cold start < 120s, individual images < 200MB

  **Must NOT do**:
  - No premature optimization of code paths
  - No custom memory allocators
  - No swap configuration

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 30,31,33,34)
  - **Parallel Group**: Wave 5
  - **Blocks**: F1-F4
  - **Blocked By**: Task 28

  **References**:
  - `docker-compose.yml` (Task 28) - Resource limits
  - docker-mailserver resource optimization docs
  - Postfix performance tuning guide
  - Next.js bundle analysis: `@next/bundle-analyzer`

  **Acceptance Criteria**:
  - [ ] `docker stats --no-stream` total idle < 1.5GB
  - [ ] Cold start (docker compose up to all healthy) < 120s
  - [ ] Each Docker image < 200MB
  - [ ] Next.js bundle size < 500KB gzipped (first load)
  - [ ] Memory limits set in docker-compose.yml

  **QA Scenarios**:
  ```
  Scenario: RAM budget met
    Tool: Bash
    Preconditions: Full stack running idle for 60s
    Steps:
      1. docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}"
      2. Sum all container memory usage
      3. Assert total < 1536MB (1.5GB)
      4. Assert no single container > 512MB
    Expected Result: Under RAM budget
    Evidence: .sisyphus/evidence/task-32-ram-usage.txt

  Scenario: Cold start time
    Tool: Bash
    Preconditions: No running containers
    Steps:
      1. Record start time
      2. docker compose up -d
      3. Run scripts/wait-healthy.sh
      4. Record end time
      5. Assert elapsed < 120s
    Expected Result: Stack healthy within 2 minutes
    Evidence: .sisyphus/evidence/task-32-cold-start-time.txt
  ```

  **Commit**: YES
  - Message: `perf: optimize RAM usage + cold start + image sizes`
  - Files: `docker-compose.yml`, Dockerfiles, configs
  - Pre-commit: `docker compose build`

- [ ] 33. README + quickstart documentation

  **What to do**:
  - Create `README.md` with:
    - Project name + tagline + hero screenshot
    - Feature highlights (bullet list with icons)
    - Quick comparison table vs Mailcow/docker-mailserver/Mailu
    - Quickstart: 3 commands to running mail server
    - Architecture diagram (ASCII or mermaid)
    - Tech stack list
    - Configuration reference (.env variables)
    - Development setup guide
    - Contributing guidelines (brief)
    - License (MIT)
  - Create `docs/quickstart.md`: detailed first-run walkthrough
  - Create `docs/architecture.md`: component diagram, data flow, config flow
  - Create `docs/api.md`: link to OpenAPI spec + key endpoint examples
  - Screenshots: dashboard, webmail, logs (captured via Playwright)

  **Must NOT do**:
  - No full API reference (OpenAPI spec serves that)
  - No deployment guides (local-first only)
  - No video tutorials
  - No translated docs

  **Recommended Agent Profile**:
  - **Category**: `writing`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 30-32,34)
  - **Parallel Group**: Wave 5
  - **Blocks**: F1-F4
  - **Blocked By**: Task 28

  **References**:
  - All previous tasks for feature descriptions
  - docker-mailserver README (reference for structure)
  - Coolify README (reference for visual appeal)

  **Acceptance Criteria**:
  - [ ] README.md exists with all required sections
  - [ ] Quickstart commands actually work (tested)
  - [ ] Architecture diagram present
  - [ ] Screenshots captured and embedded
  - [ ] No broken links

  **QA Scenarios**:
  ```
  Scenario: Quickstart commands work
    Tool: Bash
    Preconditions: Clean environment, Docker installed
    Steps:
      1. Follow exact commands from README quickstart section
      2. Assert each command exits 0
      3. Assert stack becomes healthy
      4. Assert can access web UI at localhost:3000
    Expected Result: README quickstart is accurate and functional
    Evidence: .sisyphus/evidence/task-33-quickstart.txt

  Scenario: No broken links in docs
    Tool: Bash
    Preconditions: Docs written
    Steps:
      1. grep -r "](http" README.md docs/ | extract URLs
      2. curl each URL, assert 200 or 301
      3. grep -r "](/" README.md docs/ | verify internal paths exist
    Expected Result: All links resolve
    Evidence: .sisyphus/evidence/task-33-links.txt
  ```

  **Commit**: YES
  - Message: `docs: README + quickstart + architecture documentation`
  - Files: `README.md`, `docs/`
  - Pre-commit: N/A

- [ ] 34. End-to-end integration test suite

  **What to do**:
  - Create comprehensive E2E test suite using Playwright + bash scripts
  - Test scenarios (all automated, zero human intervention):
    - Full lifecycle: start stack → seed → send email → read via webmail → verify logs
    - Persistence: send email → restart stack → email still accessible
    - Security: open relay rejected, auth required, no secrets exposed
    - Performance: cold start < 120s, idle RAM < 1.5GB
    - UI flows: domain CRUD, user CRUD, compose+send, search
    - API flows: all CRUD endpoints, health, logs
    - Error handling: invalid inputs rejected, graceful degradation
  - Create `make test-e2e` target that runs full suite
  - Create `scripts/test-e2e.sh` orchestrator
  - Generate test report: pass/fail per scenario, total time, evidence links
  - All evidence saved to `.sisyphus/evidence/e2e/`

  **Must NOT do**:
  - No load/stress testing
  - No chaos engineering
  - No cross-browser testing (Chromium only)
  - No mobile viewport testing

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 30-33)
  - **Parallel Group**: Wave 5
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 28, 29

  **References**:
  - All QA scenarios from Tasks 1-33 (consolidated)
  - Playwright test patterns
  - `scripts/seed.sh` (Task 29) - Test data setup

  **Acceptance Criteria**:
  - [ ] `make test-e2e` runs full suite and exits 0
  - [ ] All scenarios pass (lifecycle, persistence, security, performance, UI, API)
  - [ ] Test report generated with pass/fail counts
  - [ ] Evidence files created for each scenario
  - [ ] Total E2E suite completes within 10 minutes

  **QA Scenarios**:
  ```
  Scenario: Full E2E suite passes
    Tool: Bash
    Preconditions: Full stack healthy + seeded
    Steps:
      1. make test-e2e
      2. Assert exit code 0
      3. Read test report
      4. Assert 0 failures
      5. Assert evidence directory populated
    Expected Result: All E2E tests green
    Evidence: .sisyphus/evidence/task-34-e2e-report.txt

  Scenario: Persistence survives restart
    Tool: Bash
    Preconditions: Stack running with seeded data
    Steps:
      1. curl -s http://localhost:8080/api/mail/inbox?user=alice@local.test | jq '.messages | length'
      2. Record count as N
      3. docker compose restart
      4. scripts/wait-healthy.sh
      5. curl -s http://localhost:8080/api/mail/inbox?user=alice@local.test | jq '.messages | length'
      6. Assert count = N (unchanged)
    Expected Result: Data persists across restart
    Evidence: .sisyphus/evidence/task-34-persistence.txt
  ```

  **Commit**: YES
  - Message: `test: end-to-end integration test suite`
  - Files: `tests/e2e/`, `scripts/test-e2e.sh`
  - Pre-commit: `make test-e2e`

- [ ] 35. Production TLS - Let's Encrypt auto-renewal

  **What to do**:
  - TDD: Write tests for TLS configuration
  - Add Traefik or Caddy as reverse proxy container for automatic TLS
  - Create `docker-compose.prod.yml` override for production profile
  - Configure Let's Encrypt ACME challenge (HTTP-01 for simplicity)
  - TLS termination at reverse proxy, forward to internal services
  - Postfix TLS: configure for STARTTLS on port 25 and implicit TLS on 465
  - Dovecot TLS: configure for STARTTLS on 143 and implicit TLS on 993
  - Certificate auto-renewal (Traefik/Caddy handles this)
  - Env vars: `OXMAIL_DOMAIN`, `OXMAIL_ACME_EMAIL` for Let's Encrypt registration
  - Redirect HTTP → HTTPS for web UI
  - Production ports: 25, 465, 587, 143, 993, 80, 443

  **Must NOT do**:
  - No manual certificate management
  - No self-signed certs in production mode
  - No wildcard certificates
  - No DNS-01 challenge (HTTP-01 only)

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 36, 37, 38)
  - **Parallel Group**: Wave 6 (production)
  - **Blocks**: Task 38
  - **Blocked By**: Task 28

  **References**:
  - Traefik Docker provider docs
  - Postfix TLS configuration: smtpd_tls_cert_file, smtpd_tls_key_file
  - Dovecot SSL configuration: ssl_cert, ssl_key
  - Let's Encrypt rate limits

  **Acceptance Criteria**:
  - [ ] `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d` starts with TLS
  - [ ] HTTPS on port 443 serves valid certificate (or staging cert in test)
  - [ ] Postfix STARTTLS works: `openssl s_client -starttls smtp -connect localhost:25`
  - [ ] Dovecot IMAPS works: `openssl s_client -connect localhost:993`
  - [ ] HTTP redirects to HTTPS
  - [ ] Certificate auto-renewal configured

  **QA Scenarios**:
  ```
  Scenario: TLS termination works for web UI
    Tool: Bash
    Preconditions: Production stack running with valid domain
    Steps:
      1. curl -s -o /dev/null -w "%{http_code}" http://${OXMAIL_DOMAIN}
      2. Assert status = 301 or 308 (redirect to HTTPS)
      3. curl -s -o /dev/null -w "%{http_code}" https://${OXMAIL_DOMAIN}
      4. Assert status = 200
      5. openssl s_client -connect ${OXMAIL_DOMAIN}:443 </dev/null 2>/dev/null | grep "Verify return code"
      6. Assert "Verify return code: 0 (ok)"
    Expected Result: Valid TLS on HTTPS
    Evidence: .sisyphus/evidence/task-35-tls-web.txt

  Scenario: Mail TLS works
    Tool: Bash
    Preconditions: Production stack running
    Steps:
      1. openssl s_client -starttls smtp -connect ${OXMAIL_DOMAIN}:587 </dev/null 2>/dev/null | grep "Verify return code"
      2. Assert certificate presented
      3. openssl s_client -connect ${OXMAIL_DOMAIN}:993 </dev/null 2>/dev/null | grep "Verify return code"
      4. Assert certificate presented
    Expected Result: TLS on SMTP submission and IMAPS
    Evidence: .sisyphus/evidence/task-35-tls-mail.txt
  ```

  **Commit**: YES
  - Message: `feat(prod): Let's Encrypt TLS + reverse proxy + production compose`
  - Files: `docker-compose.prod.yml`, `docker/traefik/`, configs
  - Pre-commit: `docker compose -f docker-compose.yml -f docker-compose.prod.yml config --quiet`

- [ ] 36. Production outbound delivery + DNS helper

  **What to do**:
  - TDD: Write tests for outbound delivery configuration
  - Enable Postfix outbound delivery in production mode (relay to internet)
  - Configure proper HELO/EHLO hostname from `OXMAIL_DOMAIN`
  - Configure Postfix transport maps for outbound
  - Create `GET /api/dns/check` endpoint: verify MX, SPF, DKIM, DMARC, rDNS for configured domain
  - Create `GET /api/dns/records` endpoint: generate required DNS records for copy-paste
    - MX record: `@ IN MX 10 mail.{domain}`
    - SPF record: `@ IN TXT "v=spf1 ip4:{IP} -all"`
    - DKIM record: from Task 10 key
    - DMARC record: `_dmarc IN TXT "v=DMARC1; p=quarantine; rua=mailto:postmaster@{domain}"`
    - rDNS/PTR: instruction for ISP/VPS provider
  - UI page: DNS setup wizard showing all required records + live verification status
  - Production env vars: `OXMAIL_PUBLIC_IP`, `OXMAIL_DOMAIN`, `OXMAIL_HOSTNAME`
  - Outbound rate limiting: 100 emails/hour default (configurable)

  **Must NOT do**:
  - No automatic DNS configuration (user must set records manually)
  - No IP warmup automation
  - No blacklist monitoring service
  - No bounce processing beyond basic Postfix handling

  **Recommended Agent Profile**:
  - **Category**: `deep`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 35, 37, 38)
  - **Parallel Group**: Wave 6 (production)
  - **Blocks**: Task 38
  - **Blocked By**: Tasks 28, 10, 13

  **References**:
  - `internal/config/postfix_generator.go` (Task 7) - Extend for outbound
  - `internal/api/dkim_handler.go` (Task 10) - DKIM public key for DNS
  - Postfix transport maps documentation
  - SPF/DKIM/DMARC record formats (RFCs 7208, 6376, 7489)

  **Acceptance Criteria**:
  - [ ] `GET /api/dns/records` returns all required DNS records
  - [ ] `GET /api/dns/check` verifies each record (pass/fail per record)
  - [ ] Outbound delivery works: send to external address (test with another local domain)
  - [ ] HELO hostname matches configured domain
  - [ ] Rate limiting enforced on outbound

  **QA Scenarios**:
  ```
  Scenario: DNS records endpoint generates correct records
    Tool: Bash
    Preconditions: Production stack running with OXMAIL_DOMAIN=mail.example.com, OXMAIL_PUBLIC_IP=1.2.3.4
    Steps:
      1. curl -s http://localhost:8080/api/dns/records | jq '.'
      2. Assert .mx contains "mail.example.com"
      3. Assert .spf contains "ip4:1.2.3.4"
      4. Assert .dkim contains "v=DKIM1"
      5. Assert .dmarc contains "v=DMARC1"
    Expected Result: All DNS records generated correctly
    Evidence: .sisyphus/evidence/task-36-dns-records.txt

  Scenario: Outbound delivery works
    Tool: Bash
    Preconditions: Production mode, DNS configured, second test domain available
    Steps:
      1. swaks --server localhost --port 587 --auth --auth-user alice@local.test --auth-password TestPass123! --from alice@local.test --to bob@second.test --body "outbound test"
      2. Assert exit code 0
      3. Verify delivery on receiving end
    Expected Result: Email delivered to external domain
    Evidence: .sisyphus/evidence/task-36-outbound.txt
  ```

  **Commit**: YES
  - Message: `feat(prod): outbound delivery + DNS helper + record generator`
  - Files: `internal/api/dns_handler.go`, `internal/mail/outbound.go`
  - Pre-commit: `go test ./...`

- [ ] 37. UI - DNS setup wizard + production settings page

  **What to do**:
  - TDD: Write component tests first
  - Create "Production Setup" page in admin UI (new nav item)
  - DNS wizard: step-by-step guide showing each required record
    - Step 1: MX record (with copy button + verification status)
    - Step 2: SPF record (with copy button + verification status)
    - Step 3: DKIM record (with copy button + verification status)
    - Step 4: DMARC record (with copy button + verification status)
    - Step 5: rDNS/PTR instruction
  - Live verification: "Check DNS" button per record, shows green/red status
  - Overall readiness indicator: "Ready for production" when all records verified
  - Production settings: hostname, public IP, TLS status, outbound rate limit
  - Warning banner if running in dev mode on production page

  **Must NOT do**:
  - No automatic DNS API integration (Cloudflare/Route53 etc.)
  - No domain registrar integration
  - No email deliverability scoring

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES (with Tasks 35, 36, 38)
  - **Parallel Group**: Wave 6 (production)
  - **Blocks**: Task 38
  - **Blocked By**: Tasks 15, 16, 36

  **References**:
  - `internal/api/dns_handler.go` (Task 36) - DNS check/records endpoints
  - `web/src/hooks/` (Task 16) - API hooks pattern
  - Mailcow DNS check UI (reference)
  - shadcn/ui Stepper/Steps pattern

  **Acceptance Criteria**:
  - [ ] DNS wizard shows all 5 steps with correct records
  - [ ] Copy button works for each record
  - [ ] "Check DNS" shows live verification result
  - [ ] Overall readiness indicator reflects all checks
  - [ ] Component tests pass

  **QA Scenarios**:
  ```
  Scenario: DNS wizard displays records and verifies
    Tool: Playwright
    Preconditions: Production stack running, domain configured
    Steps:
      1. Navigate to http://localhost:3000/production
      2. Assert DNS wizard visible with 5 steps
      3. Assert MX record displayed with copy button
      4. Click "Check DNS" for MX
      5. Assert verification result shown (green check or red X)
      6. Screenshot
    Expected Result: DNS wizard functional with live verification
    Evidence: .sisyphus/evidence/task-37-dns-wizard.png

  Scenario: Production readiness indicator
    Tool: Playwright
    Preconditions: Some DNS records configured, some not
    Steps:
      1. Navigate to http://localhost:3000/production
      2. Assert readiness indicator shows partial (e.g., "3/5 records verified")
      3. Assert indicator is NOT green (not fully ready)
    Expected Result: Accurate readiness status
    Evidence: .sisyphus/evidence/task-37-readiness.png
  ```

  **Commit**: YES
  - Message: `feat(ui): DNS setup wizard + production settings page`
  - Files: `web/src/app/production/`
  - Pre-commit: `cd web && npx vitest run`

- [ ] 38. Production integration test + docker-compose profiles

  **What to do**:
  - Create Docker Compose profiles: `dev` (default) and `prod`
  - `docker compose --profile prod up -d` starts with TLS + outbound + production ports
  - `docker compose up -d` (no profile) starts in dev mode (current behavior)
  - Production profile includes: Traefik/Caddy, production ports, TLS, outbound enabled
  - Dev profile: no TLS, local ports (1025/1143), no outbound, relaxed auth
  - Create production E2E test: TLS handshake, outbound delivery, DNS check, auth enforcement
  - Document production deployment in `docs/production.md`
  - Add `make prod` target for production mode
  - Add `make prod-test` target for production integration tests

  **Must NOT do**:
  - No cloud-specific deployment (AWS/GCP/DO)
  - No Ansible/Terraform
  - No monitoring stack (Prometheus/Grafana)

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO (integration point for production)
  - **Parallel Group**: Wave 6 (after Tasks 35, 36, 37)
  - **Blocks**: F1-F4
  - **Blocked By**: Tasks 35, 36, 37

  **References**:
  - `docker-compose.yml` (Task 28) - Base compose
  - `docker-compose.prod.yml` (Task 35) - Production override
  - Docker Compose profiles documentation

  **Acceptance Criteria**:
  - [ ] `docker compose --profile prod config --quiet` exits 0
  - [ ] `make prod` starts production stack
  - [ ] Production stack has TLS on 443, SMTP on 25/465/587, IMAPS on 993
  - [ ] Dev mode still works without profile flag
  - [ ] `make prod-test` runs production tests
  - [ ] `docs/production.md` documents full setup

  **QA Scenarios**:
  ```
  Scenario: Production profile starts correctly
    Tool: Bash
    Preconditions: Docker installed, domain configured
    Steps:
      1. docker compose --profile prod up -d
      2. scripts/wait-healthy.sh
      3. docker compose ps
      4. Assert traefik/caddy container running
      5. Assert ports 25, 443, 587, 993 exposed
      6. docker compose --profile prod down
    Expected Result: Production stack with all services
    Evidence: .sisyphus/evidence/task-38-prod-profile.txt

  Scenario: Dev mode unaffected
    Tool: Bash
    Preconditions: No profile flag
    Steps:
      1. docker compose up -d
      2. scripts/wait-healthy.sh
      3. Assert NO traefik container
      4. Assert ports 1025, 1143, 3000, 8080 exposed (dev ports)
      5. docker compose down
    Expected Result: Dev mode unchanged
    Evidence: .sisyphus/evidence/task-38-dev-mode.txt
  ```

  **Commit**: YES
  - Message: `feat(prod): Docker Compose profiles + production integration tests`
  - Files: `docker-compose.yml`, `docs/production.md`, `Makefile`
  - Pre-commit: `docker compose --profile prod config --quiet`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [ ] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [ ] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `golangci-lint` + `tsc --noEmit` + `eslint` + `vitest run` + `go test ./...`. Review all changed files for: `as any`/`@ts-ignore`, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [ ] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` skill)
  Start from clean state (`docker compose down -v && docker compose up -d`). Execute EVERY QA scenario from EVERY task. Test cross-task integration. Test edge cases: empty state, invalid input, rapid actions. Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [ ] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- **Repository**: `https://github.com/MYusufEka/oxmail.git`
- **Project directory**: `oxmail/` (rename from `mailer/` before starting)
- **Wave 1**: `feat(scaffold): monorepo structure + docker compose skeleton`
- **Wave 2 (per task)**: `feat(api): domain CRUD + config generation`, `feat(api): user/mailbox CRUD`, etc.
- **Wave 3 (per task)**: `feat(ui): shell layout + command palette`, `feat(ui): domain management`, etc.
- **Wave 4 (per task)**: `feat(api): IMAP proxy for webmail`, `feat(ui): webmail inbox`, etc.
- **Wave 5**: `feat(security): open relay protection + auth`, `feat(ui): polish + animations`, `docs: README + quickstart`
- **Wave 6 (per task)**: `feat(prod): Let's Encrypt TLS + reverse proxy`, `feat(prod): outbound delivery + DNS helper`, `feat(ui): DNS setup wizard`, `feat(prod): Docker Compose profiles`

---

## Success Criteria

### Verification Commands
```bash
docker compose up -d                    # Expected: all services start
docker compose ps                       # Expected: all healthy within 120s
curl http://localhost:8080/api/health    # Expected: {"status":"healthy","services":{...}}
curl -X POST http://localhost:8080/api/domains -d '{"domain":"local.test"}'  # Expected: 201
curl -X POST http://localhost:8080/api/users -d '{"email":"alice@local.test","password":"TestPass123!"}'  # Expected: 201
swaks --server localhost --port 1025 --from bob@local.test --to alice@local.test --body "hello"  # Expected: exit 0
# IMAP check: message in inbox
docker compose restart                  # Expected: data persists
docker stats --no-stream                # Expected: total < 1.5GB idle
# Production mode:
docker compose --profile prod up -d     # Expected: TLS + outbound enabled
curl -s http://localhost:8080/api/dns/records  # Expected: MX, SPF, DKIM, DMARC records
openssl s_client -connect localhost:993 </dev/null  # Expected: valid certificate
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All Go tests pass
- [ ] All Vitest tests pass
- [ ] All Playwright E2E tests pass
- [ ] Open-relay test fails (security verified)
- [ ] RAM < 1.5GB idle
- [ ] Cold start < 120s
- [ ] Data persists across restart
