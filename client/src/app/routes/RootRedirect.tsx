import { useEffect, useState } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useAuth } from '../auth/AuthContext'
import { fetchProperties, type ApiError } from '../lib/api'
import { buildLoginUrlWithReturnUrl } from '../lib/returnUrl'

type State =
  | { status: 'loading' }
  | { status: 'zero' }
  | { status: 'single'; id: string }
  | { status: 'many' }
  | { status: 'unauthenticated'; returnUrl: string }
  | { status: 'error'; error: ApiError }

export function RootRedirect() {
  const { t } = useTranslation()
  const { setPrincipal } = useAuth()
  const location = useLocation()
  const [state, setState] = useState<State>({ status: 'loading' })

  useEffect(() => {
    const ctrl = new AbortController()
    let alive = true
    const returnUrlSnapshot = location.pathname + location.search + location.hash

    async function run() {
      try {
        const items = await fetchProperties(ctrl.signal)
        if (!alive) return
        if (items.length === 0) setState({ status: 'zero' })
        else if (items.length === 1 && items[0]) setState({ status: 'single', id: items[0].id })
        else setState({ status: 'many' })
      } catch (err) {
        if (!alive) return
        const apiErr = err as ApiError
        // AAR-002: 401 must not be conflated with "zero properties". An expired/
        // invalid session is an authentication failure, not an empty portfolio.
        // Clear the principal so RequireAuth can redirect to login with returnUrl.
        if (apiErr?.status === 401) {
          setPrincipal(null)
          setState({ status: 'unauthenticated', returnUrl: returnUrlSnapshot })
          return
        }
        setState({ status: 'error', error: apiErr })
      }
    }

    void run()
    return () => {
      alive = false
      ctrl.abort()
    }
  }, [location.pathname, location.search, location.hash, setPrincipal])

  if (state.status === 'loading') {
    return (
      <div className="grid place-items-center py-16">
        <p className="text-sm text-text-muted" role="status" aria-live="polite">
          {t('common.loading')}
        </p>
      </div>
    )
  }

  if (state.status === 'unauthenticated') {
    return <Navigate to={buildLoginUrlWithReturnUrl(state.returnUrl)} replace />
  }

  if (state.status === 'zero') return <Navigate to="/welcome" replace />
  if (state.status === 'single') return <Navigate to={`/properties/${state.id}`} replace />
  if (state.status === 'many') return <Navigate to="/portfolio" replace />

  return (
    <div role="alert" className="rounded-[var(--radius-panel)] border border-line bg-surface p-8 shadow-[var(--shadow-panel)]">
      <h1 className="font-display text-xl text-ink">{t('routeError.title')}</h1>
      <p className="mt-2 text-sm leading-6 text-text-muted">{state.error.message}</p>
      <button
        type="button"
        onClick={() => window.location.reload()}
        className="mt-6 inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] bg-ink px-5 text-sm font-semibold text-surface transition-colors hover:bg-ink-strong focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
      >
        {t('routeError.retry')}
      </button>
    </div>
  )
}
