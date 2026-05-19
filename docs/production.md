# Production Deployment

This guide covers deploying Oxmail to a production server with TLS, outbound mail, and proper DNS configuration.

## Prerequisites

- **VPS/Dedicated server** with at least 2 GB RAM and 20 GB disk
- **Domain name** with access to DNS management
- **Public IP** with ports 25, 80, 443, 465, 587, 993 open (not blocked by ISP)
- **Reverse DNS (rDNS/PTR)** set to your mail domain (contact your VPS provider)
- Docker and Docker Compose installed
- `make` installed

### Port Requirements

| Port | Protocol | Purpose |
|------|----------|---------|
| 25 | TCP | SMTP (inbound mail) |
| 80 | TCP | HTTP (Let's Encrypt + redirect) |
| 443 | TCP | HTTPS (Web UI + API) |
| 465 | TCP | SMTPS (implicit TLS) |
| 587 | TCP | SMTP Submission (STARTTLS) |
| 993 | TCP | IMAPS (implicit TLS) |

> Many residential ISPs and some cloud providers (AWS EC2, GCP) block port 25 outbound by default. Verify with your provider before deploying.

## DNS Setup

Configure the following DNS records for your domain (e.g., `mail.example.com`):

### Required Records

| Type | Name | Value | TTL |
|------|------|-------|-----|
| A | `mail.example.com` | `YOUR_PUBLIC_IP` | 300 |
| MX | `example.com` | `mail.example.com` (priority 10) | 3600 |
| TXT | `example.com` | `v=spf1 ip4:YOUR_PUBLIC_IP -all` | 3600 |

### Recommended Records

| Type | Name | Value | TTL |
|------|------|-------|-----|
| TXT | `_dmarc.example.com` | `v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com` | 3600 |
| TXT | `mail._domainkey.example.com` | *(DKIM key — generated after first start)* | 3600 |
| PTR | `YOUR_PUBLIC_IP` | `mail.example.com` | 3600 |

> After starting Oxmail, use the **DNS Wizard** in the admin UI to verify all records and get the exact DKIM value.

## Step-by-Step Deployment

### 1. Clone the Repository

```bash
git clone https://github.com/MYusufEka/oxmail.git
cd oxmail
```

### 2. Configure Environment

```bash
cp .env.example .env
```

Edit `.env` with your production values:

```bash
# Required for production
OXMAIL_MODE=production
OXMAIL_DOMAIN=mail.example.com
OXMAIL_ADMIN_PASSWORD=your-secure-password-here
OXMAIL_PUBLIC_IP=203.0.113.10
OXMAIL_ACME_EMAIL=admin@example.com

# Ports (production defaults)
OXMAIL_SMTP_PORT=25
OXMAIL_IMAP_PORT=993
OXMAIL_API_PORT=8080
OXMAIL_WEB_PORT=3000
```

| Variable | Description |
|----------|-------------|
| `OXMAIL_MODE` | Must be `production` for TLS and outbound mail |
| `OXMAIL_DOMAIN` | Your mail server FQDN (must match DNS A record) |
| `OXMAIL_ADMIN_PASSWORD` | Admin login password (min 8 characters, use a strong one) |
| `OXMAIL_PUBLIC_IP` | Server's public IPv4 address (used for SPF/rDNS checks) |
| `OXMAIL_ACME_EMAIL` | Email for Let's Encrypt certificate notifications |

### 3. Start Production Stack

```bash
make prod
```

This runs:
```bash
docker compose --profile prod -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

It builds all images, starts the mail services, and launches Traefik for automatic TLS via Let's Encrypt.

### 4. Verify Deployment

Wait 30–60 seconds for Let's Encrypt to issue certificates, then:

```bash
make prod-test
```

This runs the production integration test suite that checks:
- HTTPS TLS on port 443
- SMTP STARTTLS on port 587
- IMAPS on port 993
- Auth enforcement (unauthenticated requests → 401)
- HTTP → HTTPS redirect

### 5. Access the Admin UI

Open `https://mail.example.com` in your browser and log in with:
- **Email:** `admin@mail.example.com`
- **Password:** The value of `OXMAIL_ADMIN_PASSWORD`

### 6. Configure DNS (via UI)

Navigate to **Settings → DNS Wizard** in the admin panel. It will:
1. Show you the exact DNS records needed
2. Verify each record in real-time
3. Provide the DKIM public key to add as a TXT record

### 7. Send a Test Email

From the admin UI or CLI:
```bash
docker exec oxmail-api oxmail send-test your-external@gmail.com
```

Check the recipient's inbox (and spam folder). If it lands in spam, verify:
- SPF record is correct
- DKIM record is published
- rDNS/PTR matches your domain

## Monitoring

### View Logs

```bash
# All services
make logs

# Specific service
docker compose logs -f postfix
docker compose logs -f dovecot
docker compose logs -f traefik
```

### Real-time Log Streaming

The admin UI provides real-time WebSocket log streaming at **Logs** in the sidebar. Filter by service, severity, or search terms.

### Health Check

```bash
curl -s https://mail.example.com/api/health | jq .
```

Expected response:
```json
{
  "status": "healthy",
  "services": {
    "postfix": "running",
    "dovecot": "running",
    "rspamd": "running",
    "redis": "connected"
  }
}
```

### Container Status

```bash
docker compose --profile prod -f docker-compose.yml -f docker-compose.prod.yml ps
```

## Backup Strategy

### What to Back Up

| Data | Location | Method |
|------|----------|--------|
| Mail data | `mail-data` volume | Volume snapshot or `docker cp` |
| Configuration | `config-data` volume | Volume snapshot |
| DKIM keys | `dkim-keys` volume | Volume snapshot (critical!) |
| TLS certificates | `letsencrypt-data` volume | Volume snapshot |
| SQLite database | Inside `config-data` | File copy |
| Environment | `.env` file | Secure file copy |

### Backup Script Example

```bash
#!/bin/bash
BACKUP_DIR="/backups/oxmail/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Stop services briefly for consistent backup
docker compose --profile prod -f docker-compose.yml -f docker-compose.prod.yml stop

# Backup volumes
for vol in mail-data config-data dkim-keys letsencrypt-data redis-data; do
  docker run --rm -v "oxmail_${vol}:/source:ro" -v "${BACKUP_DIR}:/backup" \
    alpine tar czf "/backup/${vol}.tar.gz" -C /source .
done

# Backup env
cp .env "${BACKUP_DIR}/.env"

# Restart
docker compose --profile prod -f docker-compose.yml -f docker-compose.prod.yml up -d

echo "Backup complete: ${BACKUP_DIR}"
```

### Restore

```bash
# Stop services
make prod-down

# Restore volumes
for vol in mail-data config-data dkim-keys letsencrypt-data redis-data; do
  docker volume rm "oxmail_${vol}" 2>/dev/null || true
  docker volume create "oxmail_${vol}"
  docker run --rm -v "oxmail_${vol}:/target" -v "/backups/oxmail/20260101:/backup:ro" \
    alpine tar xzf "/backup/${vol}.tar.gz" -C /target
done

# Restart
make prod
```

## Maintenance

### Update Oxmail

```bash
cd oxmail
git pull origin main
make prod
```

Docker Compose will rebuild only changed images and recreate affected containers.

### Renew TLS Certificates

Traefik handles Let's Encrypt renewal automatically. Certificates renew 30 days before expiry. No manual action needed.

If renewal fails, check:
- Port 80 is accessible from the internet (HTTP-01 challenge)
- DNS A record points to this server
- `docker compose logs traefik` for errors

### Rotate DKIM Keys

```bash
# Generate new DKIM key via API
curl -X POST https://mail.example.com/api/domains/example.com/dkim/rotate \
  -H "Authorization: Bearer YOUR_TOKEN"

# Update DNS TXT record with the new public key (shown in response)
# Wait for DNS propagation (check via DNS Wizard in UI)
```

## Stopping Production

```bash
make prod-down
```

This stops all containers but preserves volumes (data intact).

## Troubleshooting

### TLS Certificate Not Issued

**Symptoms:** Browser shows "connection not secure", `make prod-test` fails TLS checks.

**Causes:**
- Port 80 blocked (Let's Encrypt HTTP-01 needs it)
- DNS A record not pointing to this server
- Rate limit hit (5 certs per domain per week)

**Fix:**
```bash
# Check Traefik logs
docker compose logs traefik | grep -i "acme\|certificate\|error"

# Verify port 80 is open
curl -I http://mail.example.com
```

### Mail Going to Spam

**Causes:**
- Missing or incorrect SPF record
- DKIM not configured
- No rDNS/PTR record
- IP on a blocklist

**Fix:**
1. Use the DNS Wizard in the admin UI to verify all records
2. Check blocklists: [MXToolbox](https://mxtoolbox.com/blacklists.aspx)
3. Start with low volume and build reputation

### Cannot Receive Mail

**Causes:**
- Port 25 blocked inbound
- MX record not set or wrong priority
- Postfix not running

**Fix:**
```bash
# Check if port 25 is listening
docker compose exec postfix ss -tlnp | grep :25

# Check MX record
dig MX example.com +short

# Check Postfix logs
docker compose logs postfix | tail -50
```

### Cannot Send Mail

**Causes:**
- Port 25 blocked outbound (common on cloud providers)
- SPF/DKIM misconfigured
- Recipient server rejecting

**Fix:**
```bash
# Test outbound connectivity
docker compose exec postfix bash -c "echo test | nc -w5 gmail-smtp-in.l.google.com 25"

# Check mail queue
docker compose exec postfix postqueue -p

# Flush stuck mail
docker compose exec postfix postqueue -f
```

### High Memory Usage

Oxmail targets ~1.2 GB idle. If usage is higher:

```bash
# Check per-container usage
docker stats --no-stream

# Rspamd can grow with large bayes database
docker compose restart rspamd
```

### Container Won't Start

```bash
# Check specific container logs
docker compose logs <service-name>

# Check if port is already in use
ss -tlnp | grep -E ":(25|80|443|465|587|993)"

# Rebuild from scratch
docker compose --profile prod -f docker-compose.yml -f docker-compose.prod.yml up -d --build --force-recreate
```
