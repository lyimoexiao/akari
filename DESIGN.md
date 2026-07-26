# Akari Design System

## 1. Atmosphere & Identity

Akari is a quiet, practical control surface for account and Minecraft identity management. It should feel light, direct, and trustworthy. The signature is a restrained Naive UI surface system paired with small duotone icons: content remains calm while navigation and status affordances stay immediately legible.

## 2. Color

### Palette

Akari delegates light/dark color resolution to `NConfigProvider`; application code uses Naive UI theme variables and semantic component types instead of raw colors.

| Role | Token | Usage |
|---|---|---|
| Page surface | `--n-color-body` | Route backgrounds |
| Primary surface | `--n-color` | Cards, headers, footers |
| Text primary | `--n-text-color` | Headings and body |
| Text secondary | `--n-text-color-2` | Supporting copy |
| Text tertiary | `--n-text-color-3` | Captions and metadata |
| Border | `--n-border-color` | Dividers and outlined surfaces |
| Accent | `--n-primary-color` | Primary actions, links, focus |
| Accent hover | `--n-primary-color-hover` | Hover state |
| Success | Naive `success` semantic | Verified and healthy states |
| Warning | Naive `warning` semantic | Unverified and caution states |
| Error | Naive `error` semantic | Failures and destructive actions |
| Information | Naive `info` semantic | Staff and informational states |

### Rules

- Use Naive UI semantic component types for status colors.
- Use UnoCSS opacity utilities only for subdued text or icons, never to invent a new semantic color.
- New colors must first be represented as a Naive theme override or documented token here.

## 3. Typography

### Scale

| Level | Size | Weight | Line Height | Usage |
|---|---:|---:|---:|---|
| H1 | 24px / 1.5rem | 700 | 1.3 | Authentication titles |
| H2 | 20px / 1.25rem | 600 | 1.4 | Route titles |
| H3 | 18px / 1.125rem | 600 | 1.4 | Section titles |
| Body | 16px / 1rem | 400 | 1.6 | Default content |
| Body/sm | 14px / 0.875rem | 400 | 1.5 | Controls and supporting copy |
| Caption | 12px / 0.75rem | 400 | 1.4 | Metadata |

### Font Stack

- Primary: Naive UI system sans-serif stack.
- Mono: Naive UI system monospace stack, used only for IDs and request data.

## 4. Spacing & Layout

### Base Unit

All spacing derives from 4px.

| Token | Value | Usage |
|---|---:|---|
| `--space-1` | 4px | Tight inline gaps |
| `--space-2` | 8px | Compact controls |
| `--space-3` | 12px | Form rhythm |
| `--space-4` | 16px | Standard padding |
| `--space-5` | 20px | Page heading rhythm |
| `--space-6` | 24px | Card and route padding |
| `--space-8` | 32px | Major inner separation |
| `--space-12` | 48px | Header/footer geometry and page breathing room |

### Grid

- Public content width: 1280px maximum.
- Account and admin content width: 1120px maximum.
- Authentication card width: 400px maximum, fluid below 432px.
- Breakpoints follow UnoCSS/Wind: `sm` 640px, `md` 768px, `lg` 1024px, `xl` 1280px.
- At 375px all primary content becomes a single readable column without horizontal page scrolling.

## 5. Components

### Application Header

- **Structure**: sticky header, brand, primary navigation, account dropdown.
- **Variants**: public, authenticated, administration-aware.
- **Spacing**: `--space-2`, `--space-4`; 48px height.
- **States**: transparent at top, elevated/blurred while scrolled; hover, focus and active navigation states.
- **Accessibility**: semantic navigation, buttons retain visible focus, text labels accompany icons where space permits.
- **Motion**: 200ms opacity/background transition.
- **Layout**: three-part cluster; document remains scroll owner.

### Authentication Shell

- **Structure**: centered landmark, heading, description, one card slot, footer links.
- **Variants**: form, progress, success.
- **Spacing**: `--space-4`, `--space-6`, `--space-8`.
- **States**: loading, validation error, server error, success.
- **Accessibility**: one `h1`, labelled inputs, keyboard-submit forms, live message feedback.
- **Motion**: no decorative entry motion; form controls use Naive transitions.
- **Layout**: centered stack; page scrolls when form height exceeds viewport.

### Workspace Page

- **Structure**: title/description row followed by responsive cards, tables or forms.
- **Variants**: dashboard, list, detail, settings.
- **Spacing**: `--space-4`, `--space-6`, `--space-8`.
- **States**: loading, empty, error, populated.
- **Accessibility**: semantic headings, table actions named with visible text, destructive actions confirmed.
- **Motion**: state changes use built-in Naive transitions only.
- **Layout**: centered 1120px stack; document remains scroll owner.

### Status Tag

- **Structure**: Naive `NTag` with concise text.
- **Variants**: success, warning, error, info, neutral.
- **Spacing**: compact Naive UI sizing.
- **States**: static; never used as an unlabeled icon.
- **Accessibility**: status meaning is repeated in text, not color alone.
- **Motion**: none.
- **Layout**: inline cluster.

### Data Table Toolbar

- **Structure**: search input, primary action, reset/secondary actions, result count.
- **Variants**: users, roles, request-log lookup.
- **Spacing**: `--space-2`, `--space-3`, `--space-4`.
- **States**: default, searching, empty query, loading.
- **Accessibility**: input has an accessible label/placeholder; Enter submits search.
- **Motion**: none beyond control state transitions.
- **Layout**: wrapping cluster; table owns horizontal overflow on narrow screens.

## 6. Motion & Interaction

| Type | Duration | Easing | Usage |
|---|---:|---|---|
| Micro | 120ms | ease-out | Button and focus feedback |
| Standard | 200ms | ease-in-out | Header surface and panel state |

- Animate only `transform`, `opacity`, or composited filters.
- Respect `prefers-reduced-motion` through platform/Naive UI defaults.
- Every interactive element exposes hover, active, focus, disabled, and loading states where relevant.

## 7. Depth & Surface

### Strategy

Mixed, restrained depth: Naive UI borders establish most hierarchy; the sticky header may use a small shadow and backdrop blur after scrolling. Cards, modals, dropdowns, and popovers use Naive UI's theme-provided elevation. No decorative shadows are introduced in route pages.

## 8. Accessibility Constraints & Accepted Debt

### Constraints

- Target WCAG 2.2 AA: 4.5:1 body contrast and 3:1 large-text contrast.
- All functionality is keyboard reachable and visible focus is retained.
- Color never carries status alone.
- Primary content does not horizontally overflow at 375px; data tables may use their own labelled horizontal scroll region.

### Accepted Debt

| Item | Location | Why accepted | Owner / Exit |
|---|---|---|---|
| Third-party go-captcha canvas semantics | `src/views/auth/CaptchaWidget.vue` | Accessibility depends on the upstream widget API. | Replace or augment when the widget exposes an accessible challenge mode. |
