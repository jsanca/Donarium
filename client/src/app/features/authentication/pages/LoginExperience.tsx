import { useTranslation } from 'react-i18next'
import { LoginArtwork } from '../components/LoginArtwork'
import { LoginForm } from '../components/LoginForm'

export function LoginExperience() {
  const { t } = useTranslation()

  return (
    <main className="grid min-h-[100dvh] bg-canvas lg:grid-cols-[minmax(0,1.15fr)_minmax(27rem,0.85fr)]">
      <LoginArtwork />
      <section className="relative flex min-h-[100dvh] flex-col items-center justify-center px-[var(--space-page)] py-10 sm:py-14 lg:px-10 xl:px-16">
        <LoginForm />
        <footer className="mt-8 text-center text-sm text-text-muted lg:absolute lg:bottom-7 lg:left-1/2 lg:-translate-x-1/2">
          {t('login.footer', { year: new Date().getFullYear() })}
        </footer>
      </section>
    </main>
  )
}
