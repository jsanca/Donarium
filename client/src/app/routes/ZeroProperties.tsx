import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { motion, useReducedMotion } from 'motion/react'
import { Building2, KeyRound, ArrowRight } from 'lucide-react'
import { calmTransition, revealUp } from '../../shared/motion/variants'

export function ZeroProperties() {
  const { t } = useTranslation()
  const reduceMotion = useReducedMotion()
  const [infoOpen, setInfoOpen] = useState(false)

  return (
    <div className="mx-auto max-w-[72rem]">
      <motion.div
        initial={reduceMotion ? false : 'hidden'}
        animate="visible"
        variants={revealUp}
        transition={calmTransition}
        className="mx-auto max-w-3xl text-center"
      >
        <p className="text-sm font-medium tracking-wide text-accent">{t('zero.eyebrow')}</p>
        <h1 className="mt-3 font-display text-[clamp(2rem,4vw,2.75rem)] leading-[1.05] text-ink">{t('zero.title')}</h1>
        <p className="mx-auto mt-4 max-w-2xl text-base leading-7 text-text-muted">{t('zero.subtitle')}</p>
      </motion.div>

      <div className="mx-auto mt-10 grid max-w-4xl gap-6 md:grid-cols-2">
        {/* Register */}
        <motion.section
          initial={reduceMotion ? false : 'hidden'}
          animate="visible"
          variants={revealUp}
          transition={{ ...calmTransition, delay: 0.08 }}
          aria-labelledby="zero-register-title"
          className="flex flex-col rounded-[var(--radius-panel)] border border-line bg-surface p-7 shadow-[var(--shadow-panel)] sm:p-8"
        >
          <span className="grid size-11 place-items-center rounded-xl bg-surface-subtle text-ink" aria-hidden="true">
            <Building2 className="size-5" />
          </span>
          <h2 id="zero-register-title" className="mt-5 font-display text-xl leading-tight text-ink">
            {t('zero.registerTitle')}
          </h2>
          <p className="mt-2 text-sm leading-6 text-text-muted">{t('zero.registerBody')}</p>
          <p className="mt-1 text-xs leading-5 text-text-muted">{t('zero.registerHint')}</p>

          <button
            type="button"
            aria-describedby="zero-register-hint"
            onClick={() => {
              // No backend yet: surface as informational note per handoff boundary
              const el = document.getElementById('zero-register-hint')
              el?.classList.remove('hidden')
              el?.focus()
            }}
            className="mt-6 inline-flex min-h-11 items-center justify-center gap-2 rounded-[var(--radius-control)] bg-ink px-5 text-sm font-semibold text-surface transition-colors hover:bg-ink-strong focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
          >
            {t('zero.registerAction')} <ArrowRight className="size-4" aria-hidden="true" />
          </button>
          <p id="zero-register-hint" tabIndex={-1} className="hidden mt-3 rounded-lg bg-surface-subtle px-3 py-2 text-xs leading-5 text-text-muted focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30">
            {t('zero.comingSoon')}
          </p>
        </motion.section>

        {/* Access existing */}
        <motion.section
          initial={reduceMotion ? false : 'hidden'}
          animate="visible"
          variants={revealUp}
          transition={{ ...calmTransition, delay: 0.14 }}
          aria-labelledby="zero-access-title"
          className="flex flex-col rounded-[var(--radius-panel)] border border-line bg-surface p-7 shadow-[var(--shadow-panel)] sm:p-8"
        >
          <span className="grid size-11 place-items-center rounded-xl bg-surface-subtle text-ink" aria-hidden="true">
            <KeyRound className="size-5" />
          </span>
          <h2 id="zero-access-title" className="mt-5 font-display text-xl leading-tight text-ink">
            {t('zero.accessTitle')}
          </h2>
          <p className="mt-2 text-sm leading-6 text-text-muted">{t('zero.accessBody')}</p>
          <p className="mt-1 text-xs leading-5 text-text-muted">{t('zero.accessHint')}</p>

          <button
            type="button"
            onClick={() => setInfoOpen(true)}
            className="mt-6 inline-flex min-h-11 items-center justify-center gap-2 rounded-[var(--radius-control)] border border-line bg-surface px-5 text-sm font-semibold text-ink transition-colors hover:bg-surface-subtle focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
          >
            {t('zero.accessAction')} <ArrowRight className="size-4" aria-hidden="true" />
          </button>
        </motion.section>
      </div>

      {/* Info dialog (accessible) */}
      {infoOpen && (
        <div
          role="dialog"
          aria-modal="true"
          aria-labelledby="zero-info-title"
          className="fixed inset-0 z-50 grid place-items-center bg-ink/40 p-4 backdrop-blur-sm"
          onClick={() => setInfoOpen(false)}
        >
          <div
            role="document"
            onClick={(e) => e.stopPropagation()}
            className="w-full max-w-lg rounded-[var(--radius-panel)] border border-line bg-surface p-7 shadow-[var(--shadow-panel)] sm:p-8"
          >
            <h3 id="zero-info-title" className="font-display text-lg text-ink">
              {t('zero.infoDialogTitle')}
            </h3>
            <p className="mt-3 text-sm leading-6 text-text-muted">{t('zero.infoDialogBody')}</p>
            <button
              type="button"
              autoFocus
              onClick={() => setInfoOpen(false)}
              className="mt-6 inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] bg-ink px-5 text-sm font-semibold text-surface transition-colors hover:bg-ink-strong focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
            >
              {t('zero.infoDialogClose')}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
