import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { House, LayoutGrid, Settings, CircleHelp, LogOut, Menu, X } from 'lucide-react'
import { useAuth } from '../auth/AuthContext'

export function AuthenticatedShell() {
  const { t } = useTranslation()
  const { principal } = useAuth()
  const [mobileOpen, setMobileOpen] = useState(false)

  const displayName = principal?.displayName ?? '—'
  const email = principal?.email ?? ''

  return (
    <div className="min-h-[100dvh] bg-canvas text-text">
      <a
        href="#main-content"
        className="sr-only left-4 top-4 z-50 rounded-md bg-ink px-4 py-2 text-sm font-medium text-surface focus:not-sr-only focus:absolute focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
      >
        {t('shell.skipToContent')}
      </a>

      {/* Top bar */}
      <header className="sticky top-0 z-40 border-b border-line bg-surface/90 backdrop-blur">
        <div className="mx-auto flex h-16 max-w-[80rem] items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
          <div className="flex items-center gap-3">
            <button
              type="button"
              aria-label={mobileOpen ? 'Close menu' : 'Open menu'}
              aria-expanded={mobileOpen}
              aria-controls="shell-mobile-nav"
              onClick={() => setMobileOpen((v) => !v)}
              className="grid size-11 place-items-center rounded-[var(--radius-control)] border border-line bg-surface text-ink transition-colors hover:bg-surface-subtle focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30 lg:hidden"
            >
              {mobileOpen ? <X className="size-5" aria-hidden="true" /> : <Menu className="size-5" aria-hidden="true" />}
            </button>

            <div className="flex items-center gap-2.5">
              <span className="grid size-9 place-items-center rounded-xl bg-ink text-surface" aria-hidden="true">
                <House className="size-5" />
              </span>
              <span className="font-display text-[1.05rem] font-semibold tracking-tight text-ink">Donarium</span>
            </div>
          </div>

          <nav aria-label="Primary" className="hidden items-center gap-1 lg:flex">
            <ShellNavLink to="/" label={t('shell.navPortfolio')} icon={LayoutGrid} />
            {/* Placeholders per plan: Settings / Help are nav placeholders only in EP-001 */}
            <ShellNavLink to="/properties" label={t('shell.navProperties')} icon={House} />
          </nav>

          <div className="flex items-center gap-2">
            <div className="hidden max-w-[14rem] flex-col items-end sm:flex">
              <span className="max-w-full truncate text-sm font-semibold leading-none text-ink">{displayName}</span>
              <span className="max-w-full truncate text-xs leading-none text-text-muted">{email}</span>
            </div>
            {/* AAR-001 remediation: sign-out is disabled until POST /api/auth/logout exists.
                Previously this control navigated to /login and was immediately bounced back
                by PublicOnly, leaving the httpOnly session cookie valid. Showing a disabled
                control is less misleading than a no-op that implies sign-out. */}
            <span
              aria-disabled="true"
              aria-label={t('shell.signOutSrHint')}
              title={t('shell.signOutSrHint')}
              className="inline-flex min-h-11 cursor-not-allowed items-center gap-2 rounded-[var(--radius-control)] border border-line bg-surface px-3 text-sm font-medium text-text-muted opacity-60"
            >
              <LogOut className="size-4 opacity-60" aria-hidden="true" />
              <span className="hidden sm:inline">{t('shell.signOut')}</span>
            </span>
          </div>
        </div>

        {/* Mobile nav */}
        {mobileOpen && (
          <nav
            id="shell-mobile-nav"
            aria-label="Primary mobile"
            className="border-t border-line bg-surface px-4 py-3 lg:hidden"
          >
            <div className="flex flex-col gap-1">
              <MobileNavLink to="/" label={t('shell.navPortfolio')} icon={LayoutGrid} onNavigate={() => setMobileOpen(false)} />
              <MobileNavLink to="/properties" label={t('shell.navProperties')} icon={House} onNavigate={() => setMobileOpen(false)} />
              <div className="my-2 border-t border-line" />
              <span className="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm text-text-muted">
                <Settings className="size-4" aria-hidden="true" /> {t('shell.navSettings')} <span className="ml-auto text-xs">—</span>
              </span>
              <span className="flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm text-text-muted">
                <CircleHelp className="size-4" aria-hidden="true" /> {t('shell.navHelp')} <span className="ml-auto text-xs">—</span>
              </span>
            </div>
          </nav>
        )}
      </header>

      <main id="main-content" className="mx-auto max-w-[80rem] px-4 py-8 sm:px-6 sm:py-10 lg:px-8">
        <Outlet />
      </main>
    </div>
  )
}

function ShellNavLink({
  to,
  label,
  icon: Icon,
}: {
  to: string
  label: string
  icon: React.ComponentType<{ className?: string }>
}) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        [
          'inline-flex min-h-11 items-center gap-2 rounded-[var(--radius-control)] px-3 text-sm font-medium transition-colors focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30',
          isActive ? 'bg-ink text-surface' : 'text-text-muted hover:bg-surface-subtle hover:text-ink',
        ].join(' ')
      }
    >
      <Icon className="size-4" aria-hidden="true" />
      {label}
    </NavLink>
  )
}

function MobileNavLink({
  to,
  label,
  icon: Icon,
  onNavigate,
}: {
  to: string
  label: string
  icon: React.ComponentType<{ className?: string }>
  onNavigate: () => void
}) {
  return (
    <NavLink
      to={to}
      onClick={onNavigate}
      className={({ isActive }) =>
        [
          'flex items-center gap-2 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30',
          isActive ? 'bg-ink text-surface' : 'text-ink hover:bg-surface-subtle',
        ].join(' ')
      }
    >
      <Icon className="size-4" aria-hidden="true" />
      {label}
    </NavLink>
  )
}
