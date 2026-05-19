---
name: expressjs-backend
description: Express.js 5 backend patterns — service layer, auth, error handling, logging, and API conventions for this project.
---

# Express.js Backend Patterns

## Architecture

```
backend/
├── server.js                 # Entry point — Express app, WebSocket, startup
├── src/
│   ├── config/env.js         # Zod-validated environment variables
│   ├── db.js                 # PostgreSQL pool (pg)
│   ├── redis.js              # Redis client (ioredis)
│   ├── ws.js                 # WebSocket manager
│   ├── parser.js             # Natural language message parser
│   ├── whatsapp.js           # Baileys WhatsApp session manager
│   ├── quota.js              # Quota enforcement
│   ├── scheduler.js          # node-cron scheduled jobs
│   ├── middleware/
│   │   └── auth.js           # JWT auth + adminOnly guard
│   ├── routes/               # Express routers (18 files)
│   ├── services/             # Business logic layer (shared between routes + WhatsApp)
│   │   ├── transaction.service.js
│   │   └── debt.service.js
│   └── utils/
│       ├── logger.js         # Structured logger
│       ├── mailer.js         # Nodemailer SMTP
│       ├── otp.js            # OTP generation
│       └── currency.js       # Exchange rates
```

## Service Layer Pattern

Business logic lives in `services/` — shared between HTTP routes and WhatsApp handler.

```javascript
// services/example.service.js
const db = require("../db");

async function createThing({ userId, name, amount }) {
  // Validation
  if (!name) throw Object.assign(new Error("Name required."), { status: 400 });

  // Business logic with transaction
  const client = await db.pool.connect();
  try {
    await client.query("BEGIN");
    // ... queries ...
    await client.query("COMMIT");
    return result;
  } catch (err) {
    await client.query("ROLLBACK");
    throw err;
  } finally {
    client.release();
  }
}

module.exports = { createThing };
```

```javascript
// routes/example.js — thin route handler
const service = require("../services/example.service");

router.post("/", authMiddleware, async (req, res) => {
  try {
    const result = await service.createThing({ userId: req.user.id, ...req.body });
    res.status(201).json(result);
  } catch (err) {
    res.status(err.status || 500).json({ error: err.message });
  }
});
```

## Existing Services

### transaction.service.js
- `createTransaction()` — validate + insert + currency conversion
- `updateTransaction()` — block transfer edits + update
- `deleteTransaction()` — cascade transfer delete + revert debt payments
- `createTransfer()` — wallet check + balance check + double-entry ledger
- `deleteTransfer()` — cascade delete legs + transfer record

### debt.service.js
- `createDebt()` — wallet_id required + ledger transaction
- `updateDebt()` — metadata only (contact, description, due date)
- `payDebt()` — overpay protection + wallet selection + ledger
- `deleteDebt()` — only unpaid + cascade transactions

## Logging

```javascript
const { logger } = require("../utils/logger");
const log = logger.child("ModuleName");

log.info("message", data);
log.error("message", err);
log.warn("message");
```

Do NOT use `console.log` — only `server.js` entry point uses raw console.

## Authentication

```javascript
const { authMiddleware } = require("../middleware/auth");

// Protected route
router.get("/", authMiddleware, async (req, res) => {
  const userId = req.user.id;  // Available after auth
  const userTier = req.user.tier;  // 'free', 'pro', 'admin'
});

// Admin only
const { adminOnly } = require("../middleware/auth");
router.get("/admin", authMiddleware, adminOnly, handler);
```

## API Response Format

```javascript
// List with pagination
res.json({ data: [], total, page, limit, totalPages });

// Single item
res.json({ transaction: {} });
res.json({ debt: {} });

// Success action
res.json({ ok: true, message: "..." });

// Error
res.status(400).json({ error: "Indonesian error message" });
```

## Error Handling

- Service throws errors with `status` property: `Object.assign(new Error("msg"), { status: 400 })`
- Route catches and uses `err.status || 500`
- Error messages in **Indonesian** (Bahasa)
- Express 5 — async errors auto-propagate (no need for try/catch in middleware)

## WebSocket

```javascript
const { wsManager } = require("../ws");

// Broadcast to specific user
wsManager.broadcast(userId, "event_name", { data });

// Broadcast to all connected clients
wsManager.broadcastAll("event_name", { data });
```

## Rate Limiting (already configured in server.js)

| Scope | Limit |
|---|---|
| Login/Register | 20 req / 15 min |
| Auth routes | 100 req / 15 min |
| General API | 300 req / 1 min |
| Exports | 20 req / 5 min |

## CommonJS

Backend uses `require()` / `module.exports`. NOT `import`/`export`.
