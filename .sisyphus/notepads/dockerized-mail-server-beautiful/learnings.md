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
