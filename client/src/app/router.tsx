import { lazy, Suspense } from 'react'
import { createBrowserRouter, Navigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { LoginExperience } from './features/authentication/pages/LoginExperience'
import { AuthenticatedShell } from './shell/AuthenticatedShell'
import { RequireAuth, PublicOnly } from './auth/RequireAuth'
import { RootRedirect } from './routes/RootRedirect'
import { RouteError } from '../shared/ui/RouteError'

const ZeroProperties = lazy(() => import('./routes/ZeroProperties').then((m) => ({ default: m.ZeroProperties })))
const PortfolioPlaceholder = lazy(() =>
  import('./routes/PortfolioPlaceholder').then((m) => ({ default: m.PortfolioPlaceholder })),
)
const PropertyHomePlaceholder = lazy(() =>
  import('./routes/PropertyHomePlaceholder').then((m) => ({ default: m.PropertyHomePlaceholder })),
)
const InvitationAcceptPlaceholder = lazy(() =>
  import('./routes/InvitationAcceptPlaceholder').then((m) => ({ default: m.InvitationAcceptPlaceholder })),
)
const NotFound = lazy(() => import('./routes/NotFound').then((m) => ({ default: m.NotFound })))

function LazyFallback() {
  const { t } = useTranslation()
  return (
    <div className="grid place-items-center py-16">
      <p className="text-sm text-text-muted" role="status" aria-live="polite">
        {t('common.loading')}
      </p>
    </div>
  )
}

function LazyWrap({ children }: { children: React.ReactNode }) {
  return <Suspense fallback={<LazyFallback />}>{children}</Suspense>
}

export const router = createBrowserRouter([
  {
    path: '/login',
    element: (
      <PublicOnly>
        <LoginExperience />
      </PublicOnly>
    ),
    errorElement: <RouteError />,
  },
  {
    path: '/',
    element: (
      <RequireAuth>
        <AuthenticatedShell />
      </RequireAuth>
    ),
    errorElement: <RouteError />,
    children: [
      { index: true, element: <RootRedirect /> },
      {
        path: 'welcome',
        element: (
          <LazyWrap>
            <ZeroProperties />
          </LazyWrap>
        ),
      },
      {
        path: 'portfolio',
        element: (
          <LazyWrap>
            <PortfolioPlaceholder />
          </LazyWrap>
        ),
      },
      // Keep /properties without :id as alias to /portfolio so placeholder nav works.
      { path: 'properties', element: <Navigate to="/portfolio" replace /> },
      {
        path: 'properties/:id',
        element: (
          <LazyWrap>
            <PropertyHomePlaceholder />
          </LazyWrap>
        ),
      },
      {
        path: 'invitations/:token/accept',
        element: (
          <LazyWrap>
            <InvitationAcceptPlaceholder />
          </LazyWrap>
        ),
      },
      {
        path: '*',
        element: (
          <LazyWrap>
            <NotFound />
          </LazyWrap>
        ),
      },
    ],
  },
  // Fallback for any other top-level unknown path while unauthenticated:
  {
    path: '*',
    element: (
      <LazyWrap>
        <NotFound />
      </LazyWrap>
    ),
    errorElement: <RouteError />,
  },
])
