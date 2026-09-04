import { Navigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from './AuthContext'
import { buildLoginUrlWithReturnUrl } from '../lib/returnUrl'
import { sanitizeReturnUrl } from '../lib/returnUrl'

export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { status } = useAuth()
  const location = useLocation()
  const { t } = useTranslation()

  if (status === 'loading' || status === 'idle') {
    return (
      <div className="grid min-h-[100dvh] place-items-center bg-canvas px-6">
        <p className="text-sm text-text-muted" role="status" aria-live="polite">
          {t('common.loading')}
        </p>
      </div>
    )
  }

  if (status === 'unauthenticated') {
    const returnUrl = location.pathname + location.search + location.hash
    // Do not store login itself as returnUrl
    const safe = returnUrl !== '/login' ? returnUrl : '/'
    return <Navigate to={buildLoginUrlWithReturnUrl(safe)} replace />
  }

  if (status === 'error') {
    // Allow the shell to surface the error while still offering retry
    return (
      <div className="grid min-h-[100dvh] place-items-center bg-canvas px-6 py-12">
        <div className="w-full max-w-md rounded-[var(--radius-panel)] border border-line bg-surface p-8 shadow-[var(--shadow-panel)]">
          <h1 className="font-display text-xl text-ink">{t('errorBoundary.title')}</h1>
          <p className="mt-2 text-sm leading-6 text-text-muted">{t('routeError.unauthorizedBody')}</p>
          <a
            href={buildLoginUrlWithReturnUrl(location.pathname + location.search)}
            className="mt-6 inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] bg-ink px-5 text-sm font-semibold text-surface transition-colors hover:bg-ink-strong focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
          >
            {t('login.submit')}
          </a>
        </div>
      </div>
    )
  }

  return <>{children}</>
}

export function PublicOnly({ children }: { children: React.ReactNode }) {
  const { status } = useAuth()
  const location = useLocation()
  const { t } = useTranslation()
  const search = new URLSearchParams(location.search)
  const raw = search.get('returnUrl')

  // AAR-006: reuse shared sanitizer (includes CRLF rejection) instead of inline duplicate.
  const returnUrl = sanitizeReturnUrl(raw) ?? '/'

  if (status === 'authenticated') {
    return <Navigate to={returnUrl} replace />
  }

  if (status === 'loading' || status === 'idle') {
    return (
      <div className="grid min-h-[100dvh] place-items-center bg-canvas px-6">
        <p className="text-sm text-text-muted" role="status" aria-live="polite">
          {t('common.loading')}
        </p>
      </div>
    )
  }

  return <>{children}</>
}
