# API Documentation

The Oxmail REST API runs on port 8080 and provides full control over domains, users, aliases, and server health. Real-time log streaming is available over WebSocket.

Full OpenAPI 3.1 specification: [`../api-spec/openapi.yaml`](../api-spec/openapi.yaml)

## Base URL

```
http://localhost:8080
```

## Authentication

Most endpoints require a JWT token. Obtain one by logging in:

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "changeme123"}'
```

Response:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-01-16T12:00:00Z"
}
```

Use the token in subsequent requests:

```bash
curl http://localhost:8080/api/domains \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

## Endpoints Overview

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/auth/login` | Authenticate, get JWT |
| GET | `/api/domains` | List all domains |
| POST | `/api/domains` | Create a domain |
| GET | `/api/domains/{domain}` | Get domain details |
| DELETE | `/api/domains/{domain}` | Delete a domain |
| GET | `/api/users` | List all users |
| POST | `/api/users` | Create a user |
| GET | `/api/users/{email}` | Get user details |
| DELETE | `/api/users/{email}` | Delete a user |
| GET | `/api/aliases` | List all aliases |
| POST | `/api/aliases` | Create an alias |
| DELETE | `/api/aliases/{id}` | Delete an alias |
| GET | `/api/health` | Service health check |
| WS | `/api/logs/stream` | Real-time log streaming |

## Examples

### Health check

No authentication required.

```bash
curl http://localhost:8080/health
```

```json
{"status": "ok"}
```

### Create a domain

```bash
curl -X POST http://localhost:8080/api/domains \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "example.com"}'
```

```json
{
  "name": "example.com",
  "created_at": "2026-01-15T10:30:00Z"
}
```

### List domains

```bash
curl http://localhost:8080/api/domains \
  -H "Authorization: Bearer $TOKEN"
```

```json
{
  "data": [
    {"name": "example.com", "created_at": "2026-01-15T10:30:00Z"},
    {"name": "local.test", "created_at": "2026-01-14T08:00:00Z"}
  ],
  "total": 2,
  "page": 1,
  "limit": 20
}
```

### Create a user

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "securepass123", "display_name": "Alice"}'
```

```json
{
  "email": "alice@example.com",
  "display_name": "Alice",
  "domain": "example.com",
  "created_at": "2026-01-15T10:35:00Z"
}
```

### Create an alias

```bash
curl -X POST http://localhost:8080/api/aliases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"source": "info@example.com", "destination": "alice@example.com"}'
```

```json
{
  "id": 1,
  "source": "info@example.com",
  "destination": "alice@example.com",
  "created_at": "2026-01-15T10:40:00Z"
}
```

### Delete a domain

```bash
curl -X DELETE http://localhost:8080/api/domains/example.com \
  -H "Authorization: Bearer $TOKEN"
```

Returns `204 No Content` on success.

### Stream logs (WebSocket)

Connect with any WebSocket client:

```bash
websocat ws://localhost:8080/api/logs/stream?token=$TOKEN
```

Messages arrive as JSON:

```json
{"timestamp": "2026-01-15T10:45:00Z", "service": "postfix", "level": "info", "message": "connect from unknown[192.168.1.5]"}
```

## Pagination

List endpoints accept `page` and `limit` query parameters:

```bash
curl "http://localhost:8080/api/users?page=2&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

Responses include pagination metadata:

```json
{
  "data": [...],
  "total": 45,
  "page": 2,
  "limit": 10
}
```

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "validation_error",
    "message": "Domain name is required"
  }
}
```

Common HTTP status codes:

| Code | Meaning |
|------|---------|
| 400 | Bad request (validation failed) |
| 401 | Unauthorized (missing or invalid token) |
| 404 | Resource not found |
| 409 | Conflict (resource already exists) |
| 500 | Internal server error |
