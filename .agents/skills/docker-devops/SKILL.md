---
name: docker-devops
description: Docker Compose setup, build workflow, volume management, and common issues for this project.
---

# Docker & DevOps

## Services (docker-compose.yml)

| Service | Container | Port | Image |
|---|---|---|---|
| `postgres_db` | `wa_postgres` | 5432 | postgres:15-alpine |
| `redis` | `wa_redis` | 6379 | redis:7-alpine |
| `backend` | `wa_backend` | 3001 | Custom (Node 20 Alpine) |
| `frontend` | `wa_frontend` | 5173 | Custom (Nginx Alpine) |

All on `saas_network` bridge. Backend depends on healthy postgres + redis.

## Commands

```bash
# Build & deploy (standard workflow after code changes)
docker compose up -d --build

# Start only DB & Redis (for local dev)
docker compose up -d postgres_db redis

# Tail logs
docker compose logs -f backend
docker compose logs -f postgres_db

# Check status
docker compose ps

# Shell into container
docker exec -it wa_backend sh

# Stop all
docker compose stop

# Full teardown (keeps volumes)
docker compose down
```

## Build Workflow

After ANY code change:
```bash
docker compose up -d --build
```

This rebuilds both backend and frontend images, then recreates containers. Takes ~15-20 seconds.

For frontend-only changes, `npm run build` locally first to verify, then docker build.

## Volumes

| Volume | Mount | Purpose |
|---|---|---|
| `db_data` | `./docker/db_data/` | PostgreSQL data |
| `redis_data` | Named volume | Redis persistence |
| `wa_sessions` | `./backend/sessions/` | WhatsApp session files |

## PostgreSQL Data Corruption

### Symptoms
```
FATAL: could not open file "global/pg_filenode.map": I/O error
```

### Cause
Docker Desktop on Windows — WSL2 ↔ NTFS volume mount I/O error. Usually after:
- Docker Desktop crash
- `docker compose down` while PostgreSQL is writing
- Power loss

### Fix (preserves data)
```bash
docker compose stop
wsl --shutdown          # Restart WSL2
# Wait 30 seconds
docker compose up -d    # PostgreSQL auto-recovers from WAL
```

### Fix (nuclear — loses data)
```bash
docker compose down
rm -rf docker/db_data/*
docker compose up -d --build   # init.sql recreates schema
```

### Prevention
- Always `docker compose stop` before `docker compose down`
- Never kill Docker Desktop while containers are running
- PostgreSQL WAL handles most crash recovery automatically

## Database Init

`docker/init.sql` contains the full schema + seed data. Only runs on first start (empty `db_data/`).

For schema changes, use **migrations** (`backend/migrations/`), NOT editing `init.sql`.

## Environment Variables

All in `.env` file (see `.env.example`). Key ones:

| Variable | Required | Notes |
|---|---|---|
| `POSTGRES_*` | Yes | DB connection |
| `REDIS_*` | Yes | Cache connection |
| `JWT_SECRET` | Yes | Min 32 chars |
| `ADMIN_NUMBER` | Yes | Admin WhatsApp (62xxx) |
| `ADMIN_EMAIL` | Yes | Admin email |
| `FRONTEND_URL` | Yes | CORS origin |
| `MIDTRANS_*` | No | Payment gateway |
| `SMTP_*` | No | Email for 2FA |

## Frontend Dockerfile

Multi-stage build:
1. `builder` stage: `npm install` + `npm run build` (Vite)
2. `runner` stage: Nginx Alpine serves `dist/`

Nginx config at `frontend/nginx.conf` — proxies `/api` and `/ws` to backend.

## Backend Dockerfile

Multi-stage build:
1. `deps` stage: `npm install --legacy-peer-deps`
2. `runner` stage: Copy `node_modules` + source, run as non-root user `expressjs`

## Health Checks

- PostgreSQL: `pg_isready` every 10s
- Redis: `redis-cli ping` every 10s
- Backend: HTTP health endpoint, depends on postgres + redis healthy
- Frontend: depends on backend healthy

## Troubleshooting

### Backend won't start
Check: `docker compose logs backend` — usually DB connection or missing env var.

### Frontend shows blank page
Check: `docker compose logs frontend` — usually Nginx config or build error.

### "Connection refused" on API calls
Backend not healthy yet. Wait for health check or check logs.

### WhatsApp session not restoring
Check `backend/sessions/` directory. If empty, user needs to re-scan QR.
