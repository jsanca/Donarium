import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export function NotFound() {
  const { t } = useTranslation()
  return (
    <div className="grid place-items-center py-16">
      <div
        role="alert"
        className="w-full max-w-lg rounded-[var(--radius-panel)] border border-line bg-surface p-8 text-center shadow-[var(--shadow-panel)]"
      >
        <h1 className="font-display text-xl text-ink">{t('routeError.notFoundTitle')}</h1>
        <p className="mt-2 text-sm leading-6 text-text-muted">{t('routeError.notFoundBody')}</p>
        <Link
          to="/"
          className="mt-6 inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] bg-ink px-5 text-sm font-semibold text-surface transition-colors hover:bg-ink-strong focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
        >
          {t('routeError.goHome')}
        </Link>
      </div>
    </div>
  )
}
