import { isRouteErrorResponse, useRouteError, Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export function RouteError() {
  const { t } = useTranslation()
  const error = useRouteError()

  let title = t('routeError.title')
  let body = t('routeError.genericBody')
  let status: number | null = null

  if (isRouteErrorResponse(error)) {
    status = error.status
    if (status === 404) {
      title = t('routeError.notFoundTitle')
      body = t('routeError.notFoundBody')
    } else if (status === 401) {
      title = t('routeError.unauthorizedTitle')
      body = t('routeError.unauthorizedBody')
    } else if (status === 403) {
      title = t('routeError.forbiddenTitle')
      body = t('routeError.forbiddenBody')
    } else if (status === 410) {
      title = t('routeError.goneTitle')
      body = t('routeError.goneBody')
    } else {
      // Prefer JSON envelope { error: string } per H-4; fall back to generic
      const data = error.data as unknown
      if (data && typeof data === 'object' && 'error' in data && typeof (data as { error: unknown }).error === 'string') {
        body = (data as { error: string }).error
      } else if (typeof data === 'string' && data) {
        body = data
      }
    }
  } else if (error instanceof Error) {
    body = error.message
  }

  return (
    <div className="grid min-h-[100dvh] place-items-center bg-canvas px-6 py-12">
      <div
        role="alert"
        aria-live="assertive"
        className="w-full max-w-lg rounded-[var(--radius-panel)] border border-line bg-surface p-8 shadow-[var(--shadow-panel)]"
      >
        <p className="text-sm font-medium tracking-wide text-accent">{t('routeError.subtitle')}</p>
        <h1 className="mt-2 font-display text-2xl leading-tight text-ink">{title}</h1>
        <p className="mt-3 text-sm leading-6 text-text-muted">{body}</p>
        {status && <p className="mt-2 text-xs font-medium text-text-muted">HTTP {status}</p>}
        <div className="mt-6 flex flex-wrap gap-3">
          <Link
            to="/"
            className="inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] bg-ink px-5 text-sm font-semibold text-surface transition-colors hover:bg-ink-strong focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
          >
            {t('routeError.goHome')}
          </Link>
          <button
            type="button"
            onClick={() => window.location.reload()}
            className="inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] border border-line bg-surface px-5 text-sm font-semibold text-ink transition-colors hover:bg-surface-subtle focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
          >
            {t('routeError.retry')}
          </button>
        </div>
      </div>
    </div>
  )
}
