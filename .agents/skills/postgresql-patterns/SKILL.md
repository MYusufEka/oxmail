---
name: postgresql-patterns
description: PostgreSQL best practices for this project — raw SQL, triggers, transactions, pagination, and common pitfalls.
---

# PostgreSQL Patterns

## Connection

```javascript
const db = require("./src/db");
const result = await db.query("SELECT * FROM users WHERE id = $1", [userId]);
// For transactions:
const client = await db.pool.connect();
try {
  await client.query("BEGIN");
  // ... queries ...
  await client.query("COMMIT");
} catch (err) {
  await client.query("ROLLBACK");
  throw err;
} finally {
  client.release();
}
```

## Rules

1. **No ORM** — Always raw SQL via `db.query(sql, params)`
2. **Parameterized queries** — Always `$1, $2, ...` placeholders. NEVER string interpolation.
3. **UUID primary keys** — All tables use `uuid_generate_v4()`
4. **Money** — `NUMERIC(15,2)` — never use float
5. **Timestamps** — `TIMESTAMPTZ` with `NOW()` defaults
6. **Naming** — `snake_case` for columns, `camelCase` for JS variables

## Critical Trigger: `update_wallet_balance()`

This trigger fires AFTER INSERT/UPDATE/DELETE on `transactions` table:
- **INSERT**: income → `+amount`, expense → `-amount` on `wallet_id`
- **DELETE**: reverses the INSERT
- **UPDATE**: reverses old values, applies new values

### Implications
- **NEVER manually UPDATE wallet balance** — the trigger handles it
- When inserting transfer transactions (expense + income legs), the trigger automatically adjusts both wallets
- When deleting a transaction, the trigger automatically reverts the wallet balance
- Double-counting bug: if you manually update balance AND insert a transaction, balance changes 2x

## Pagination Pattern

```sql
-- Always use LIMIT/OFFSET with COUNT
const { page = 1, limit = 20 } = req.query;
const offset = (parseInt(page) - 1) * parseInt(limit);

const [dataResult, countResult] = await Promise.all([
  db.query(`SELECT ... FROM table WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, [userId, limit, offset]),
  db.query(`SELECT COUNT(*) AS total FROM table WHERE user_id = $1`, [userId])
]);

res.json({
  data: dataResult.rows,
  total: parseInt(countResult.rows[0].total),
  page: parseInt(page),
  limit: parseInt(limit),
  totalPages: Math.ceil(total / parseInt(limit))
});
```

## Common Pitfalls

### 1. Transfer Double Balance
**WRONG**: Manual `UPDATE wallets SET balance` + INSERT transaction (trigger fires = 2x)
**RIGHT**: Only INSERT transactions — trigger handles balance

### 2. Cascade Delete
When deleting a transaction with `transfer_id`:
- Must delete BOTH transaction legs (`WHERE transfer_id = $1`)
- Must delete `wallet_transfers` record
- Trigger reverts balance for each deleted transaction

When deleting a transaction with `debt_id` (payment):
- Must revert `debts.paid_amount` and recalculate `status`

### 3. PostgreSQL SUM Returns String
`SUM(amount)` returns a string in node-pg. Always `parseFloat()`:
```javascript
const total = parseFloat(result.rows[0].total) || 0;
```

### 4. Timezone
- Store: `TIMESTAMPTZ` (always UTC in DB)
- Display: Convert to `Asia/Jakarta` in queries with `AT TIME ZONE 'Asia/Jakarta'`
- Frontend: Format with `Intl.DateTimeFormat` using user's locale

## Tables Reference

| Table | Purpose | Key Relations |
|---|---|---|
| `users` | User accounts | `instance_id` → WhatsApp session |
| `wallets` | User wallets | `user_id`, `is_default`, `keyword` |
| `transactions` | Income/expense | `wallet_id`, `transfer_id`, `debt_id` |
| `wallet_transfers` | Transfer records | `from_wallet_id`, `to_wallet_id` |
| `debts` | Hutang/piutang | `wallet_id`, `paid_amount`, `status` |
| `debt_payments` | Payment records | `debt_id` |
| `goals` | Savings goals | `current_amount`, `target_amount` |
| `goal_contributions` | Goal deposits | `goal_id` |
| `budgets` | Category budgets | `category`, `period`, `alert_threshold` |
| `categories` | User categories | `type` (income/expense/both) |
| `custom_keywords` | Keyword→category | `keyword`, `category`, `type` |
| `registered_groups` | WhatsApp groups | `group_jid`, `user_id` |
| `split_bills` | Split bill headers | `group_jid` |
| `split_bill_participants` | Participants | `split_bill_id` |

## Migration

```bash
cd backend
npm run migrate:create -- <name>    # Create new migration
npm run migrate:up                   # Run pending
npm run migrate:down                 # Rollback last
```

Files in `backend/migrations/` with UTC timestamp filenames. Do NOT modify `docker/init.sql` for schema changes.
