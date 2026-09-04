import { useTranslation } from 'react-i18next'

export function PortfolioPlaceholder() {
  const { t } = useTranslation()
  return (
    <div className="mx-auto max-w-[80rem]">
      <header className="max-w-2xl">
        <h1 className="font-display text-2xl leading-tight text-ink">{t('portfolio.title')}</h1>
        <p className="mt-2 text-sm leading-6 text-text-muted">{t('portfolio.subtitle')}</p>
      </header>

      <div
        role="status"
        aria-live="polite"
        className="mt-8 rounded-[var(--radius-panel)] border border-dashed border-line bg-surface p-8 shadow-[var(--shadow-panel)]"
      >
        <p className="text-sm font-medium text-ink">{t('portfolio.empty')}</p>
        <p className="mt-2 text-sm leading-6 text-text-muted">{t('portfolio.placeholderComingSoon')}</p>
        {/* Skeleton cards to preserve layout rhythm without fabricating data */}
        <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3" aria-hidden="true">
          {[0, 1, 2].map((i) => (
            <div key={i} className="rounded-2xl border border-line bg-surface-subtle p-5">
              <div className="h-4 w-24 rounded bg-line/60" />
              <div className="mt-3 h-3 w-40 rounded bg-line/40" />
              <div className="mt-6 h-8 w-24 rounded-full bg-line/30" />
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
