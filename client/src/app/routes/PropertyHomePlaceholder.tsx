import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export function PropertyHomePlaceholder() {
  const { id } = useParams()
  const { t } = useTranslation()

  return (
    <div className="mx-auto max-w-[80rem]">
      <header className="max-w-2xl">
        <p className="text-sm font-medium tracking-wide text-accent">{t('propertyHome.title')}</p>
        <h1 className="mt-1 font-display text-2xl leading-tight text-ink">
          {t('propertyHome.title')} <span className="font-sans text-sm font-normal text-text-muted">· {id}</span>
        </h1>
        <p className="mt-2 text-sm leading-6 text-text-muted">{t('propertyHome.subtitle')}</p>
      </header>

      <div
        role="status"
        aria-live="polite"
        className="mt-8 rounded-[var(--radius-panel)] border border-dashed border-line bg-surface p-8 shadow-[var(--shadow-panel)]"
      >
        <p className="text-sm font-medium text-ink">{t('propertyHome.comingSoon')}</p>
        <p className="mt-2 text-sm leading-6 text-text-muted">{t('propertyHome.deepLinkHint')}</p>
        <div className="mt-6 grid gap-4 md:grid-cols-3" aria-hidden="true">
          <div className="rounded-2xl border border-line bg-surface-subtle p-5">
            <div className="h-4 w-20 rounded bg-line/60" />
            <div className="mt-3 h-3 w-32 rounded bg-line/40" />
          </div>
          <div className="rounded-2xl border border-line bg-surface-subtle p-5">
            <div className="h-4 w-28 rounded bg-line/60" />
            <div className="mt-3 h-3 w-36 rounded bg-line/40" />
          </div>
          <div className="rounded-2xl border border-line bg-surface-subtle p-5">
            <div className="h-4 w-24 rounded bg-line/60" />
            <div className="mt-3 h-3 w-28 rounded bg-line/40" />
          </div>
        </div>
      </div>
    </div>
  )
}
