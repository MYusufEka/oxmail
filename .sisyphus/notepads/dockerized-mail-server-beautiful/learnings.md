## Task 10 — Mail Docker Images (2026-05-19)

### Alpine package gotchas
- `cyrus-sasl-plain` and `cyrus-sasl-login` don't exist as separate packages in Alpine 3.19. The base `cyrus-sasl` package includes all mechanisms.
- Postfix's Alpine package creates a `vmail` user/group automatically. Use `|| true` guards when creating vmail user to avoid build failures.
- `rspamd-controller` and `rspamd-proxy` are separate packages from `rspamd` in Alpine.

### Image sizes (approximate)
- oxmail-postfix: ~27 MiB (19 packages)
- oxmail-dovecot: ~30 MiB (15 packages)
- oxmail-rspamd: ~77 MiB (33 packages) — largest due to vectorscan, icu-data-full, fasttext

### Architecture decisions
- Postfix and Dovecot share a `postfix-spool` volume for LMTP unix socket communication
- Health checks: postfix uses `postfix status`, dovecot uses `doveadm process status`, rspamd uses `curl localhost:11333/ping`
- Config templates use Go template syntax ({{.Variable}}) for later rendering by oxmail-api
- Dev mode exposes SMTP on ${OXMAIL_SMTP_PORT:-1025} and IMAP on ${OXMAIL_IMAP_PORT:-1143}
- Rspamd milter integration via inet:rspamd:11332 (docker network DNS)

## Task 2 — Go API Project Setup (2026-05-19)

### Go workspace setup
- Root `go.mod` (module github.com/MYusufEka/oxmail) owns `internal/` packages
- `cmd/oxmail-api/go.mod` is a separate module; imports root module's internal packages via go.work
- go.work must include root `.` in `use` block for workspace resolution to work
- `go mod tidy` in sub-modules tries to fetch from remote if go.work isn't used — always build/test from workspace root

### Go version
- Go 1.26.3 installed via winget; go.mod auto-upgraded to `go 1.25.0` minimum
- go.work version must be >= all module versions

### Dependencies chosen
- chi v5.2.5 — router
- modernc.org/sqlite v1.50.1 — CGO-free SQLite
- testify v1.11.1 — test assertions

### Patterns established
- Server struct with chi.Mux + slog.Logger
- NewServer() constructor wires middleware + routes
- Router() method exposes chi.Mux for httptest usage
- Graceful shutdown via SIGINT/SIGTERM with 10s timeout
- Health endpoint: GET /health → 200 {"status":"ok","version":"0.1.0"}
- Database: embed migrations via go:embed, apply in order on Open()
- WAL mode enabled by default for SQLite

## Task — CLI Tool oxmail (2026-05-19)

### Go file naming
- Files named `*_test.go` are treated as test files by Go toolchain — never use underscores before "test" for production command files
- Used `sendtest.go` instead of `send_test.go` for the send-test command

### Cobra patterns
- `CompletionOptions.DisableDefaultCmd = true` removes the auto-generated `completion` subcommand
- PersistentFlags on rootCmd propagate to all subcommands (used for --json, --api-url)
- Env var defaults: read os.Getenv in init(), use as default value for flag

### CLI structure
- cmd/oxmail/main.go → entry point, calls cmd.Execute()
- cmd/oxmail/cmd/root.go → root command + global flags
- cmd/oxmail/cmd/client.go → shared HTTP client + apiRequest helper
- cmd/oxmail/cmd/output.go → shared output helpers (printJSON, printSuccess, printError, newTabWriter)
- One file per command group: domain.go, user.go, alias.go, status.go, logs.go, sendtest.go

### Dependencies added
- github.com/spf13/cobra v1.10.2
- github.com/fatih/color v1.19.0

### Testing approach
- httptest.NewServer to mock API responses
- Override package-level `apiURL` var to point at test server
- Tests verify HTTP method, path, request body, and response parsing

### Go path on this machine
- Go binary at `C:\Program Files\Go\bin\go.exe` — not in default PATH for pwsh

## Task — README + Documentation (2026-05-19)

### Documentation structure
- README.md: project overview, quickstart (3 commands), architecture mermaid diagram, tech stack, config reference, dev setup, CLI usage, API summary, contributing, license
- docs/quickstart.md: step-by-step with expected output at each step
- docs/architecture.md: mermaid diagram + data flow explanations (inbound, outbound, config change, log streaming)
- docs/api.md: curl examples for key endpoints, links to OpenAPI spec

### Writing conventions
- Used "..." instead of em dashes for list item descriptions in README
- Kept quickstart to exactly 3 commands (clone, cp .env, docker compose up)
- Comparison table: Oxmail vs Mailcow vs docker-mailserver vs Mailu on RAM, containers, UI, setup time
- All port numbers reference actual docker-compose.yml values (1025, 1143, 8080, 3000)
- Architecture doc covers volumes, networks, resource limits, and key design decisions

## Task — Performance Tuning (2026-05-19)

### Memory budget (total limits: 1,344M < 1.5GB target)
- Redis: 64M
- Postfix: 256M
- Dovecot: 256M
- Rspamd: 384M (reduced from 512M)
- API: 128M
- Web: 256M

### Postfix optimizations
- `default_process_limit = 10` — sufficient for small team
- `smtpd_client_connection_count_limit = 10` / `rate_limit = 30`
- Disabled TLS session cache databases (saves memory, negligible perf impact for small scale)
- `rm -rf /var/cache/apk/* /tmp/*` in Dockerfile

### Dovecot optimizations
- `auth_cache_size = 100` with 1 hour TTL — reduces auth lookups
- `mail_max_userip_connections = 10` — prevents runaway connections
- `login_max_processes_count = 16` / `default_process_limit = 20`

### Rspamd optimizations (biggest win)
- Workers = 1 (normal + controller) — single worker is fine for < 50 users
- Disabled heavy modules: fuzzy_check, neural, chartable (via override.d)
- `dns_max_requests = 16`, `dns_retransmits = 2`
- `history_rows = 100` (reduced from default 200)
- `task_timeout = 20s` on normal worker

### .dockerignore files
- Root .dockerignore for API build context (excludes .git, node_modules, .next, docs, tests)
- web/.dockerignore (excludes node_modules, .next, coverage)
- Per-service .dockerignore for postfix/dovecot/rspamd (minimal, excludes .md files)

### Image sizes (unchanged from Task 10)
- All Alpine-based, already < 200MB each
- Rspamd is largest at ~77MB due to vectorscan/icu-data-full dependencies
- `apk add --no-cache` already used; added explicit `rm -rf /var/cache/apk/*` for safety

## Production TLS (Task 35)
- Traefik v3 static config uses env var interpolation for ACME email
- docker compose config --quiet shows warnings for unset env vars but exits 0 (valid)
- Postfix TLS: smtpd_tls_security_level=may for opportunistic STARTTLS on 25/587, wrappermode for 465
- Dovecot TLS: ssl=yes with service imap-login block for imaps on 993
- Let's Encrypt certs shared via letsencrypt-data volume mounted read-only in mail services
- Production override uses Docker labels for Traefik routing + file provider for dynamic config
- Base docker-compose.yml dev behavior unchanged — prod override only adds/overrides

