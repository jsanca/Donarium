# ADR-003 — Client Routing

- **Estado:** Aceptado
- **Fecha:** 2026-09-02
- **Relacionado con:** [Plan EP-001 — First Property Experience](../engineering/plans/DONARIUM-PA-EXP-001-first-property-experience-engineering-plan-proposal.md) (EP-001.01), ADRs 001/002
- **Decide:** `software-engineer` (EP-001.01)
- **Revisa:** `docs/engineering/ENGINEERING_LOG.md` (EP-001.01)

## Contexto

`client/src/app/App.tsx` rinde únicamente `LoginExperience`; no existe enrutamiento. EP-001.01 debe entregar un shell autenticado con:

- rutas protegidas vs. públicas,
- decisión de enrutamiento post-login consultando `GET /api/auth/me` y `GET /api/properties` (0 → Zero Properties, 1 → Property Home, n → Portfolio),
- preservación del `returnUrl` a través del ciclo de login para entradas por deep-link (`/properties/:id`, `/invitations/:token/accept`),
- `code-splitting` futuro para Property Home / Portfolio / Wizard / Invitación sin reescribir el shell.

La aplicación es un SPA React 19 + Vite 7 + TypeScript 5.8. La pila actual ya depende de `i18next` y `motion`; cualquier librería adicional debe minimizar superficie y permanecer compatible con `StrictMode`, `Suspense` y pruebas por render.

Alternativas consideradas:

1. **React Router (v6/v7, Data Router)** — estándar de facto en React SPA, API declarativa + `createBrowserRouter`, loaders opcionales, integración documentada con `Suspense` y `ErrorBoundary`, navegación programática y `searchParams` para `returnUrl`, lazy routes, soporte activo, tipado.
2. **TanStack Router** — type-safe, potente, pero inmaduro para el equipo, documentación más extensa, migración futura más costosa para un aporte incremental; su modelo codegen-first excede el alcance de EP-001.01 (no hay backend nuevo ni queries complejas).
3. **Wouter / reach-router derivados** — ligeros pero con menor ecosistema, sin loaders/error handling integrado, menos probado en escenarios de preservación de estado de autenticación.

## Decisión

Se adopta **React Router** (paquete `react-router-dom`, rama v6-compat / v7 API estable `createBrowserRouter` + `RouterProvider`) como única librería de enrutamiento del cliente.

- Rutas definidas en `client/src/app/router.tsx` mediante `createBrowserRouter`.
- `AuthProvider` (`app/auth/AuthContext.tsx`) expone `{ status, principal, error, refresh, setPrincipal }` consultando `GET /api/auth/me` (`credentials: 'include'`), sin backend adicional. `status` es el enum `idle | loading | authenticated | unauthenticated | error` (en lugar de un booleano `loading` aislado); `logout` no se expone en este slice — ver Consecuencias — y el shell no presenta un control activo de cierre de sesión hasta que exista `POST /api/auth/logout`.
- `RequireAuth` (`app/auth/RequireAuth.tsx`) guarda rutas protegidas; si no autenticado, redirige a `/login?returnUrl=<encoded>`; si autenticado pero ruta protegida contiene lógica de conteo de propiedades, delega a `RootRedirect` que consulta `GET /api/properties`. `RootRedirect` propaga 401 hacia el cierre de sesión (vuelve a `unauthenticated` y redirige a login) en lugar de mapearlo a "zero properties".
- `RootRedirect` implementa la decisión 0/1/n en `app/routes/RootRedirect.tsx`.
- `AuthenticatedShell` (`app/shell/AuthenticatedShell.tsx`) renderiza navegación y `Outlet`; cumple WCAG AA (foco visible, navegación por teclado, `prefers-reduced-motion` respetado, targets ≥ 44px).
- `RouteErrorBoundary` y `AppErrorBoundary` cumplen **H-4**: cualquier error HTTP surfacing rinde una envolvente JSON-consistente (`{ error: string }` → UI) y preserva `Allow` cuando aplica.
- `returnUrl` se valida (solo paths internos, sin `//` ni esquema) antes de navegar, para evitar open-redirects.
- ADRs futuros (ADR-004…007) no se ven afectados; el router es ortogonal al modelo `Party`/`PropertyStakeholder`/Invitation/Payment.

## Consecuencias

### Ventajas

- Solución estándar, documentada y compatible con el ecosistema React/Vite existente.
- `returnUrl` como `searchParam` es trivial con `useSearchParams` + `navigate`.
- Lazy loading (`React.lazy` + `Suspense`) disponible sin costo adicional para nodos EP-001.04/.14.
- `ErrorBoundary` por ruta cumple H-4 sin librería adicional.

### Costos / riesgos

- Dependencia adicional (`react-router-dom` + peer). Mitigado por lockfile y bundle splitting.
- Regla de lint nueva: prohibido importar navegación imperativa fuera de helpers centralizados (evita desvíos).
- Sign-out queda deshabilitado en `AuthenticatedShell` en este slice (corrige AAR-001): el cookie de sesión es `httpOnly` y solo el servidor puede invalidarlo; presentar un control activo sin `POST /api/auth/logout` sería engañoso y contradiría `PublicOnly`. Cuando el backend exponga el endpoint, `AuthProvider` ganará `logout` y el control se reactivará.

### No objetivos

- No se introduce enrutamiento server-side ni SSR.
- No se modela `Property`/`Invitation`/`Payment`; solo placeholders tipados que EP-001.02/.03/.12/.13 reemplazarán.

## Alternativas descartadas

TanStack Router y Wouter se descartan por razones de madurez/costo y ecosistema, respectivamente (ver Contexto).

## Evolución

Reevaluar si Donarium introduce SSR, file-based routing, o un layout jerárquico con loaders concurrentes que requieran un enfoque codegen. Hasta entonces, esta decisión permanece vigente.
