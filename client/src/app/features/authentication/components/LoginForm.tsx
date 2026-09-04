import { zodResolver } from '@hookform/resolvers/zod'
import { Eye, EyeOff, LockKeyhole, Mail } from 'lucide-react'
import { AnimatePresence, motion, useReducedMotion } from 'motion/react'
import { useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { useTranslation } from 'react-i18next'
import { z } from 'zod'
import { BrandSignature } from './BrandSignature'
import { calmTransition, revealUp } from '../../../../shared/motion/variants'

type LoginFields = {
  email: string
  password: string
}

type Feedback = 'idle' | 'forgot' | 'submitted'

export function LoginForm() {
  const { t } = useTranslation()
  const reduceMotion = useReducedMotion()
  const [isPasswordVisible, setPasswordVisible] = useState(false)
  const [feedback, setFeedback] = useState<Feedback>('idle')

  const schema = useMemo(
    () =>
      z.object({
        email: z
          .string()
          .min(1, t('login.validation.emailRequired'))
          .email(t('login.validation.emailInvalid')),
        password: z
          .string()
          .min(1, t('login.validation.passwordRequired'))
          .min(8, t('login.validation.passwordShort')),
      }),
    [t],
  )

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFields>({
    resolver: zodResolver(schema),
    mode: 'onBlur',
  })

  const submit = async () => {
    setFeedback('submitted')
  }

  const emailErrorId = 'login-email-error'
  const passwordErrorId = 'login-password-error'

  return (
    <motion.section
      className="w-full max-w-[27.5rem] rounded-[var(--radius-panel)] border border-line bg-surface px-6 py-10 shadow-[var(--shadow-panel)] sm:px-10 sm:py-12 lg:px-11"
      initial={reduceMotion ? false : 'hidden'}
      animate="visible"
      variants={revealUp}
      transition={calmTransition}
      aria-labelledby="login-title"
    >
      <div className="mb-10 text-center">
        <BrandSignature />
        <p className="mt-7 text-sm font-medium text-accent">{t('login.eyebrow')}</p>
        <h1 id="login-title" className="mt-2 font-display text-[clamp(2.25rem,5vw,3.2rem)] leading-[0.98] text-ink">
          {t('login.title')}
        </h1>
        <p className="mx-auto mt-4 max-w-xs text-base leading-7 text-text-muted">{t('login.subtitle')}</p>
      </div>

      <form className="space-y-5" noValidate onSubmit={handleSubmit(submit)}>
        <div>
          <label className="mb-2 block text-sm font-semibold text-text" htmlFor="login-email">
            {t('login.emailLabel')}
          </label>
          <div className="relative">
            <Mail className="pointer-events-none absolute left-4 top-1/2 size-5 -translate-y-1/2 text-text-muted" aria-hidden="true" />
            <input
              id="login-email"
              className="h-15 w-full rounded-[var(--radius-control)] border border-line bg-surface px-12 text-base text-text shadow-[var(--shadow-control)] transition-[border-color,box-shadow] duration-200 placeholder:text-text-muted hover:border-text-muted focus:border-focus focus:outline-none focus:ring-3 focus:ring-focus/20"
              type="email"
              autoComplete="email"
              placeholder={t('login.emailPlaceholder')}
              aria-invalid={Boolean(errors.email)}
              aria-describedby={errors.email ? emailErrorId : undefined}
              {...register('email')}
            />
          </div>
          <AnimatePresence initial={false}>
            {errors.email && (
              <motion.p
                id={emailErrorId}
                className="mt-2 text-sm text-danger"
                role="alert"
                initial={reduceMotion ? false : { opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                exit={reduceMotion ? undefined : { opacity: 0, y: -4 }}
                transition={calmTransition}
              >
                {errors.email.message}
              </motion.p>
            )}
          </AnimatePresence>
        </div>

        <div>
          <label className="mb-2 block text-sm font-semibold text-text" htmlFor="login-password">
            {t('login.passwordLabel')}
          </label>
          <div className="relative">
            <LockKeyhole className="pointer-events-none absolute left-4 top-1/2 size-5 -translate-y-1/2 text-text-muted" aria-hidden="true" />
            <input
              id="login-password"
              className="h-15 w-full rounded-[var(--radius-control)] border border-line bg-surface px-12 pr-13 text-base text-text shadow-[var(--shadow-control)] transition-[border-color,box-shadow] duration-200 placeholder:text-text-muted hover:border-text-muted focus:border-focus focus:outline-none focus:ring-3 focus:ring-focus/20"
              type={isPasswordVisible ? 'text' : 'password'}
              autoComplete="current-password"
              placeholder={t('login.passwordPlaceholder')}
              aria-invalid={Boolean(errors.password)}
              aria-describedby={errors.password ? passwordErrorId : undefined}
              {...register('password')}
            />
            <button
              className="absolute right-2 top-1/2 grid size-10 -translate-y-1/2 place-items-center rounded-md text-text-muted transition-colors duration-200 hover:bg-surface-subtle hover:text-ink"
              type="button"
              aria-label={t(isPasswordVisible ? 'login.hidePassword' : 'login.showPassword')}
              onClick={() => setPasswordVisible((visible) => !visible)}
            >
              {isPasswordVisible ? <EyeOff className="size-5" aria-hidden="true" /> : <Eye className="size-5" aria-hidden="true" />}
            </button>
          </div>
          <AnimatePresence initial={false}>
            {errors.password && (
              <motion.p
                id={passwordErrorId}
                className="mt-2 text-sm text-danger"
                role="alert"
                initial={reduceMotion ? false : { opacity: 0, y: -4 }}
                animate={{ opacity: 1, y: 0 }}
                exit={reduceMotion ? undefined : { opacity: 0, y: -4 }}
                transition={calmTransition}
              >
                {errors.password.message}
              </motion.p>
            )}
          </AnimatePresence>
        </div>

        <button
          className="-mt-1 rounded-md text-sm font-medium text-accent transition-colors duration-200 hover:text-accent-hover"
          type="button"
          onClick={() => setFeedback('forgot')}
        >
          {t('login.forgotPassword')}
        </button>

        <button
          className="flex h-15 w-full items-center justify-center rounded-[var(--radius-control)] bg-ink px-5 text-base font-semibold text-surface transition-[background-color,box-shadow] duration-200 hover:bg-ink-strong focus:outline-none focus:ring-3 focus:ring-focus/30 disabled:cursor-wait disabled:opacity-70"
          type="submit"
          disabled={isSubmitting}
        >
          {isSubmitting ? t('login.submitting') : t('login.submit')}
        </button>

        <AnimatePresence initial={false}>
          {feedback !== 'idle' && (
            <motion.p
              className={feedback === 'submitted' ? 'text-sm text-success' : 'text-sm text-text-muted'}
              role="status"
              initial={reduceMotion ? false : { opacity: 0, y: -4 }}
              animate={{ opacity: 1, y: 0 }}
              exit={reduceMotion ? undefined : { opacity: 0, y: -4 }}
              transition={calmTransition}
            >
              {t(feedback === 'submitted' ? 'login.submitted' : 'login.forgotPasswordFeedback')}
            </motion.p>
          )}
        </AnimatePresence>
      </form>

      <p className="mt-9 text-center text-sm leading-6 text-text-muted">
        {t('login.contactPrompt')}{' '}
        <a className="font-medium text-accent transition-colors duration-200 hover:text-accent-hover" href="mailto:administrator@donarium.invalid">
          {t('login.contactAction')}
        </a>
      </p>
    </motion.section>
  )
}
