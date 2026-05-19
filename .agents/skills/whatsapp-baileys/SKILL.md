---
name: whatsapp-baileys
description: WhatsApp integration via @whiskeysockets/baileys — session management, message handling, and bot command patterns.
---

# WhatsApp Baileys Integration

## Architecture

`backend/src/whatsapp.js` — `WhatsAppManager` class manages all WhatsApp sessions.

```
WhatsAppManager
├── sessions: Map<instanceName, { sock, userId }>
├── qrs: Map<instanceName, qrBase64>
├── initSession(userId, instanceName)    # Start/restore session
├── closeSession(instanceName)           # Logout + delete files
├── disconnectSession(instanceName)      # Graceful disconnect (keep files)
├── handleMessage(instanceName, userId, sock, msg)  # Process incoming
├── sendMessage(instanceName, jid, text) # Send to JID
└── getQR(instanceName)                  # Get QR code
```

## Session Lifecycle

1. **Init**: `initSession(userId, instanceName)` → creates Baileys socket
2. **QR**: Socket emits QR → converted to base64 → broadcast via WebSocket to frontend
3. **Connected**: Update `users.instance_status = 'connected'`
4. **Disconnected**: Auto-reconnect after 5s (unless logged out)
5. **Logout**: Delete session files + update status to 'disconnected'

Session files stored in `backend/sessions/{instanceName}/`

## Message Flow

```
Incoming message
  → Skip if from self (except ".jid" command)
  → Skip if not from group (@g.us)
  → Find user by instanceName
  → Check registered group
  → Check chat quota (free tier limit)
  → Parse message via parser.js
  → Check module toggle (active_modules)
  → Deduplication check (5 second window)
  → Route to handler:
      ├── Transfer → txService.createTransfer()
      ├── Debt/Payment → debtService.createDebt() / payDebt()
      ├── Goal → inline (goal_contributions + goals update)
      ├── Split Bill → inline (split_bills + participants)
      └── Regular TX → txService.createTransaction()
  → Send reply to group
  → Mark message as read
  → Broadcast via WebSocket
```

## Parser (`backend/src/parser.js`)

`parseFinanceMessage(text, userKeywords)` returns:

```javascript
{
  type: 'income' | 'expense',
  amount: 25000,
  category: 'Makanan',
  note: 'makan siang',
  hashtags: ['gopay'],
  currency: 'IDR',
  // Transfer:
  isTransfer: true,
  fromKeyword: 'gopay',
  toKeyword: 'bca',
  // Debt:
  isDebt: true,
  debtType: 'owe' | 'lend' | 'payment',
  contactName: 'Budi',
  // Goal:
  isGoal: true,
  goalKeyword: 'laptop',
  // Split:
  isSplit: true,
  title: 'makan',
  totalAmount: 150000,
  participants: ['Budi', 'Andi', 'Cici'],
}
```

## Service Integration

WhatsApp handler uses the SAME service layer as HTTP routes:

```javascript
const txService = require("./services/transaction.service");
const debtService = require("./services/debt.service");

// Transfer
await txService.createTransfer({ userId, fromWalletId, toWalletId, amount, note, recordedBy: senderJid });

// Regular transaction
await txService.createTransaction({ userId, type, amount, category, note, walletId, tags, recordedBy: senderJid, groupJid, rawMessage: text });

// Debt
await debtService.createDebt({ userId, contactName, type, amount, walletId, description });
await debtService.payDebt(debtId, userId, amount, note, walletId);
```

## Wallet Resolution (WhatsApp)

1. Check hashtags for wallet keyword: `#gopay` → lookup `wallets.keyword`
2. If no hashtag match → use default wallet (`is_default = TRUE`)
3. If no default wallet → `walletId = null` (for regular TX) or error (for transfer/debt)

## Reply Templates

Bilingual templates in `REPLY_TEMPLATES.id` and `REPLY_TEMPLATES.en`. User language from `users.language`.

## Module Toggle

System settings `active_modules` (JSONB) can disable features:
```javascript
if (activeModules[moduleKey] === false) {
  // Skip command, send "unrecognized" reply
}
```

## Common Pitfalls

1. **Group-only**: Bot only responds in registered groups, not DMs
2. **Deduplication**: Same message within 5 seconds is skipped (prevents double-processing)
3. **Session restore**: On server restart, `server.js` auto-restores all sessions with `instance_status = 'connected'`
4. **QR timeout**: If QR not scanned, Baileys retries automatically
5. **Graceful shutdown**: Use `disconnectSession()` (keeps files) not `closeSession()` (deletes files)
