# Client structure

React 19 SPA structure for `client/` (Vite 7, Tailwind 4, TypeScript 5.8,
react-router-dom 7, react-hook-form + zod, i18next, motion).

## Root (`client/src/`)

- `main.tsx` — `createRoot` under `StrictMode`; imports `./shared/i18n`,
  `./styles.css`, and `App`.
- `styles.css` — global styles + Tailwind tokens (plus `shared/design-system/`).
- `app/` — application wiring (router, auth, routes, features).
- `shared/` — cross-feature primitives (design system, i18n, motion, UI).

## `app/`

```text
app/
├─ App.tsx          # AppErrorBoundary → AuthProvider → RouterProvider
├─ router.tsx       # createBrowserRouter tree + lazy route wrappers
├─ auth/            # AuthContext, RequireAuth (+ PublicOnly)
├─ lib/             # api.ts (fetchMe/loginRequest/fetchProperties + parseApiError),
│                   # returnUrl.ts (sanitizeReturnUrl + helpers)
├─ shell/           # AuthenticatedShell (header, nav, skip-link, Outlet)
├─ routes/          # RootRedirect, ZeroProperties, PortfolioPlaceholder,
│                   # PropertyHomePlaceholder, InvitationAcceptPlaceholder, NotFound
└─ features/        # feature-owned slices (authentication/)
    └─ authentication/
        ├─ pages/        # LoginExperience
        └─ components/   # LoginForm, LoginArtwork, BrandSignature
```

## `shared/`

```text
shared/
├─ design-system/
│   └─ tokens/tokens.css   # Tailwind v4 design tokens
├─ i18n/                   # index.ts (init + detection), en.ts, es.ts
├─ motion/                 # calmTransition / revealUp variants
└─ ui/                     # ErrorBoundary (AppErrorBoundary), RouteError
```

## Routing flow

```text
/login          → PublicOnly → LoginExperience        (unauthenticated only)
/               → RequireAuth → AuthenticatedShell
                    ├─ index       → RootRedirect      (0→/welcome, 1→/properties/:id, n→/portfolio)
                    ├─ welcome      → ZeroProperties
                    ├─ portfolio    → PortfolioPlaceholder
                    ├─ properties   → Navigate /portfolio
                    ├─ properties/:id → PropertyHomePlaceholder
                    ├─ invitations/:token/accept → InvitationAcceptPlaceholder
                    └─ *            → NotFound
* (any)         → NotFound / RouteError
```

- `AuthProvider` resolves `/api/auth/me` (`credentials: 'include'`) into
  `{ status, principal, error, refresh, setPrincipal }`.
- `RequireAuth` guards the authenticated tree and redirects unauthenticated
  access to `/login?returnUrl=<encoded>`.
- `RootRedirect` consults `GET /api/properties` for the 0/1/n decision; a 401
  propagates back to the auth lifecycle (see ADR-003).
- Lazy-loaded routes (`React.lazy` + `Suspense`) split Property Home / Portfolio
  / Invitation / NotFound into separate chunks.
