# Design System Master File

> **LOGIC:** When building a specific page, first check `design-system/pages/[page-name].md`.
> If that file exists, its rules **override** this Master file.
> If not, strictly follow the rules below.

---

**Project:** WhatsApp Finance Bot
**Generated:** 2026-05-11
**Category:** Fintech SaaS (WhatsApp-integrated)
**Theme:** Light + Dark (Zinc + Purple Accent)

---

## Global Rules

### Color Palette (Light Theme — Default)

| Role | RGB Channels | Hex |
|------|-------------|-----|
| Background | `250 250 250` | `#FAFAFA` |
| Surface | `255 255 255` | `#FFFFFF` |
| Surface Elevated | `244 244 245` | `#F4F4F5` |
| Border | `228 228 231` | `#E4E4E7` |
| Text Primary | `9 9 11` | `#09090B` |
| Text Secondary | `113 113 122` | `#71717A` |
| Text Dim | `161 161 170` | `#A1A1AA` |
| Accent/Primary | `124 106 255` | `#7C6AFF` |
| Success (Income) | `16 163 127` | `#10A37F` |
| Danger (Expense) | `220 60 90` | `#DC3C5A` |
| Warning | `200 130 0` | `#C88200` |

### Color Palette (Dark Theme)

| Role | RGB Channels | Hex |
|------|-------------|-----|
| Background | `12 14 28` | `#0C0E1C` |
| Surface | `17 20 38` | `#111426` |
| Surface Elevated | `26 30 52` | `#1A1E34` |
| Border | `38 42 68` | `#262A44` |
| Text Primary | `240 242 255` | `#F0F2FF` |
| Text Secondary | `156 163 186` | `#9CA3BA` |
| Text Dim | `75 82 108` | `#4B526C` |

### Typography

- **Font:** Plus Jakarta Sans (Google Fonts)
- **Mono:** JetBrains Mono
- **CSS:** `@import url('https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&display=swap')`
- **Mood:** Professional, clean, modern fintech, trustworthy

### Spacing Scale (8pt grid)

| Token | Value | Usage |
|-------|-------|-------|
| `--space-xs` | `4px` | Icon gaps, tight spacing |
| `--space-sm` | `8px` | Small gaps |
| `--space-md` | `16px` | Standard padding |
| `--space-lg` | `24px` | Section padding |
| `--space-xl` | `32px` | Large gaps |
| `--space-2xl` | `48px` | Page margins |
| `--space-3xl` | `64px` | Hero/major sections |

### Border Radius

| Token | Value |
|-------|-------|
| `--radius-sm` | `8px` |
| `--radius-md` | `12px` |
| `--radius-lg` | `16px` |

### Shadow Depths

| Level | Value |
|-------|-------|
| `--shadow-sm` | `0 1px 3px rgba(0,0,0,0.04)` |
| `--shadow-md` | `0 4px 6px rgba(0,0,0,0.1)` |
| `--shadow-lg` | `0 10px 15px rgba(0,0,0,0.1)` |

### Animation Durations

| Type | Duration |
|------|-----------|
| Micro-interactions | `150-200ms` |
| State changes | `200-300ms` |
| Page transitions | `300-400ms` |
| Never exceed | `500ms` |

---

## Component Specs

### Buttons

**Sizes (use `size` prop):** `xs`, `sm`, `default`, `lg`, `xl`, `icon`, `icon-sm`, `icon-lg`

**Variants:** `default`, `secondary`, `destructive`, `outline`, `ghost`, `success`, `warning`, `tab`

**Text Styles:** `default`, `uppercase`, `uppercase-sm`, `mono`, `normal`

**CRITICAL:** Never add custom padding. Always use `size` prop. Icon buttons use `::after` pseudo-element for mobile touch expansion (44px min touch area).

### Cards

```css
.card {
  background: rgb(var(--ui-surface));
  border: 1px solid rgb(var(--ui-border));
  border-radius: 1rem;
  padding: 1.5rem;
  transition: border-color 0.2s, box-shadow 0.2s, transform 0.2s;
}
.card:hover { box-shadow: var(--shadow-lg); transform: translateY(-2px); }
```

### Inputs

```css
.input {
  padding: 0.625rem 1rem;
  border: 1px solid rgb(var(--ui-border));
  border-radius: 0.75rem;
  font-size: 0.875rem;
}
.input:focus { border-color: rgb(var(--ui-accent)); box-shadow: 0 0 0 3px rgba(124,106,255,0.15); }
```

### Dialog/Modal

Mobile: full-screen from bottom. Desktop: centered, max-w-lg, rounded-xl.

---

## Anti-Patterns (Do NOT Use)

- ❌ **Custom button padding** — always use `size` prop
- ❌ **Tiny touch targets** — icon/icon-sm/icon require `::after` pseudo expansion on mobile
- ❌ **Horizontal scroll on mobile** — use `overflow-x-auto` wrapper
- ❌ **No cursor-pointer** on clickable elements
- ❌ **Instant state changes** — always transition 150-300ms
- ❌ **Invisible focus states** — keyboard nav harus visible
- ❌ **Emojis as icons** — use Lucide React SVG icons

---

## Pre-Delivery Checklist

- [ ] No emojis used as icons (Lucide SVG only)
- [ ] All icon buttons have aria-label
- [ ] Touch targets ≥44px on mobile (::after pseudo expansion)
- [ ] prefers-reduced-motion respected (animation-duration: 0.01ms)
- [ ] Focus states visible for keyboard nav
- [ ] Mobile: 375px, 768px, 1024px tested
- [ ] No horizontal scroll on mobile
- [ ] cursor-pointer on all interactive elements
