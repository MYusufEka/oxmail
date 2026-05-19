---
name: i18n-localization
description: Internationalization patterns — bilingual ID/EN translations, key naming, and common pitfalls.
---

# i18n / Localization

## Setup

- Library: `i18next` + `react-i18next`
- Config: `frontend/src/i18n.js`
- Languages: `id` (Indonesian), `en` (English)
- Translation files: `frontend/src/locales/{id,en}/translation.json`

## Usage

```jsx
import { useTranslation } from "react-i18next";

function MyComponent() {
  const { t, i18n } = useTranslation();

  return <p>{t('debts.title')}</p>;
  // With interpolation:
  // t('debts.toast.exceeds_remaining', { max: '100.000' })
  // With default value:
  // t('common.load_more', { defaultValue: 'Muat Lebih Banyak' })
}
```

## Key Naming Convention

```
module.section.key

Examples:
  nav.dashboard
  nav.group_finance
  dashboard.stats.income
  dashboard.charts.period_1d
  transactions.table.note
  debts.toast.fetch_failed
  debts.labels.wallet
  debts.modal.edit_title
  common.save
  common.cancel
  common.delete
```

### Structure

```json
{
  "nav": { "dashboard": "...", "group_finance": "..." },
  "common": { "save": "...", "cancel": "...", "delete": "..." },
  "dashboard": {
    "stats": { "income": "...", "expense": "..." },
    "charts": { "history": "...", "period_1d": "..." },
    "insight_surplus_1m": "..."
  },
  "transactions": {
    "title": "...",
    "table": { "note": "...", "category": "..." },
    "modal": { "add_title": "..." },
    "delete_confirm": "..."
  },
  "debts": {
    "title": "...",
    "labels": { "wallet": "...", "amount": "..." },
    "toast": { "fetch_failed": "...", "create_success": "..." },
    "modal": { "edit_title": "...", "pay_title": "..." }
  }
}
```

## Rules

1. **All user-facing text must use `t()` keys** — no hardcoded strings in JSX
2. **Error messages in backend are Indonesian** — these are returned via API, not translated client-side
3. **Both ID and EN files must have the same keys** — missing keys show the key string as fallback
4. **Use `defaultValue` for new keys** — so the UI works even before translation is added:
   ```jsx
   t('new.key', { defaultValue: 'Fallback text' })
   ```
5. **Interpolation** uses `{{variable}}` syntax:
   ```json
   "exceeds_remaining": "Maks: Rp{{max}}"
   ```

## Locale-Aware Formatting

```javascript
// Dates
const locale = i18n.language === 'id' ? 'id-ID' : 'en-US';
new Date(val).toLocaleDateString(locale, { day: 'numeric', month: 'long', year: 'numeric' });
// id-ID: "9 April 2026"
// en-US: "April 9, 2026"

// Currency — use formatRupiah() from utils/api.js
import { formatRupiah } from "@/utils/api";
formatRupiah(25000); // "Rp 25.000"
```

## Common Pitfalls

### 1. Key Mismatch
Code uses `t('debts.toast.fetch_error')` but translation has `fetch_failed`. Always check both files.

### 2. Missing Keys
When adding new features, add keys to BOTH `id/translation.json` AND `en/translation.json`.

### 3. Dynamic Period Keys
Dashboard uses dynamic keys like `t(\`dashboard.insight_surplus_${chartPeriod}\`)`. All variants must exist:
- `insight_surplus_1d`, `insight_surplus_1m`, `insight_surplus_1y`, `insight_surplus_3y`
- `insight_deficit_1d`, `insight_deficit_1m`, `insight_deficit_1y`, `insight_deficit_3y`

### 4. Sidebar Group Labels
Nav group labels use `nav.group_*` keys:
- `nav.group_finance` = "Keuangan" / "Finance"
- `nav.group_planning` = "Perencanaan" / "Planning"
- `nav.group_settings` = "Pengaturan" / "Settings"
- `nav.group_whatsapp` = "WhatsApp"
