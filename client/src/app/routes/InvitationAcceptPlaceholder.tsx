import { useParams } from 'react-router-dom'
import { useTranslation } from 'react-i18next'

export function InvitationAcceptPlaceholder() {
  const { token } = useParams()
  const { t } = useTranslation()

  return (
    <div className="mx-auto max-w-2xl">
      <header>
        <p className="text-sm font-medium tracking-wide text-accent">{t('invitationAccept.title')}</p>
        <h1 className="mt-1 font-display text-2xl leading-tight text-ink">{t('invitationAccept.title')}</h1>
        <p className="mt-2 text-sm leading-6 text-text-muted">{t('invitationAccept.subtitle')}</p>
      </header>

      <div
        role="status"
        aria-live="polite"
        className="mt-8 rounded-[var(--radius-panel)] border border-dashed border-line bg-surface p-8 shadow-[var(--shadow-panel)]"
      >
        <p className="text-sm leading-6 text-text-muted">{t('invitationAccept.placeholder')}</p>
        <p className="mt-3 text-xs leading-5 text-text-muted">
          Token: <code className="rounded bg-surface-subtle px-1.5 py-0.5 font-mono text-xs text-ink">{token ?? '—'}</code>
        </p>
        <p className="mt-6 text-xs leading-5 text-text-muted">
          Four-outcome authorization (401 / 404 / 410 / 403 / 201) lands with EP-001.13; deep-link
          preservation through login is already active in this shell.
        </p>
      </div>
    </div>
  )
}
