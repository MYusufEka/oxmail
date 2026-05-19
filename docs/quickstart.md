# Quickstart

Get Oxmail running locally in under two minutes.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose (v2+)
- Git
- A terminal

That's it. No Go, Node.js, or other toolchains needed for running the server.

## Step 1: Clone the repository

```bash
git clone https://github.com/MYusufEka/oxmail.git && cd oxmail
```

Expected output:

```
Cloning into 'oxmail'...
remote: Enumerating objects: ...
Receiving objects: 100% ...
```

## Step 2: Create your environment file

```bash
cp .env.example .env
```

This creates a `.env` with sensible defaults for local development:

- Domain: `local.test`
- SMTP on port 1025
- IMAP on port 1143
- API on port 8080
- Web UI on port 3000
- Admin password: `changeme123`

Edit `.env` if you want to change the admin password or ports. For local dev, the defaults work fine.

## Step 3: Start the stack

```bash
docker compose up -d
```

Expected output:

```
[+] Running 6/6
 ✔ Container oxmail-redis    Healthy
 ✔ Container oxmail-rspamd   Healthy
 ✔ Container oxmail-postfix  Healthy
 ✔ Container oxmail-dovecot  Healthy
 ✔ Container oxmail-api      Healthy
 ✔ Container oxmail-web      Started
```

First run takes a few minutes while Docker builds the images. Subsequent starts are fast.

## Step 4: Verify everything is healthy

```bash
docker compose ps
```

All six containers should show `healthy` status:

```
NAME             STATUS
oxmail-redis     Up (healthy)
oxmail-rspamd    Up (healthy)
oxmail-postfix   Up (healthy)
oxmail-dovecot   Up (healthy)
oxmail-api       Up (healthy)
oxmail-web       Up (healthy)
```

You can also hit the API health endpoint:

```bash
curl http://localhost:8080/health
```

```json
{"status":"ok"}
```

## Step 5: Open the Web UI

Navigate to [http://localhost:3000](http://localhost:3000) in your browser.

Log in with:
- Username: `admin`
- Password: the value of `OXMAIL_ADMIN_PASSWORD` in your `.env` (default: `changeme123`)

## Step 6: Seed test data (optional)

To populate the server with sample domains, users, and emails:

```bash
make seed
```

This creates a test domain, a few mailboxes, and some sample messages so you can explore the UI immediately.

## What's running

| Port | Service | Purpose |
|------|---------|---------|
| 1025 | Postfix | SMTP (send/receive mail) |
| 1143 | Dovecot | IMAP (read mail) |
| 8080 | oxmail-api | REST API + WebSocket logs |
| 3000 | Web UI | Admin dashboard + webmail |

## Stopping the server

```bash
docker compose down
```

To stop and wipe all data (volumes):

```bash
docker compose down -v
```

## Next steps

- [Architecture overview](architecture.md) for how the pieces fit together
- [API documentation](api.md) for curl examples and endpoint reference
- The CLI tool (`oxmail`) for command-line administration
- Run `make dev` instead of `docker compose up` for development mode with hot-reload
