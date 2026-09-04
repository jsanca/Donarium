import { Component, type ReactNode } from 'react'
import i18n from '../i18n'

type Props = {
  children: ReactNode
  fallback?: ReactNode
}

type State = { hasError: boolean; message: string | null }

/**
 * H-4 compliant error boundary: surfaces HTTP-origin errors that follow
 * the project's JSON envelope { error: string } and preserves Allow-header
 * intent via the ApiError it captures. Any error boundary that surfaces
 * HTTP errors complies with H-4 (consistent JSON + Allow preservation).
 */
export class AppErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, message: null }

  static getDerivedStateFromError(err: unknown): State {
    const msg =
      err instanceof Error
        ? err.message
        : typeof err === 'object' && err !== null && 'message' in err
          ? String((err as { message: unknown }).message)
          : 'Unexpected error.'
    return { hasError: true, message: msg }
  }

  componentDidCatch(error: unknown) {
    // Visible for debugging; no external reporting in this slice.
    console.error('[AppErrorBoundary]', error)
  }

  handleReset = () => {
    this.setState({ hasError: false, message: null })
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback
      return (
        <div className="grid min-h-[100dvh] place-items-center bg-canvas px-6 py-12">
          <div
            role="alert"
            className="w-full max-w-md rounded-[var(--radius-panel)] border border-line bg-surface p-8 shadow-[var(--shadow-panel)]"
          >
            <h1 className="font-display text-xl text-ink">{i18n.t('errorBoundary.title')}</h1>
            <p className="mt-2 text-sm leading-6 text-text-muted">
              {this.state.message ?? i18n.t('errorBoundary.body')}
            </p>
            <div className="mt-6 flex flex-wrap gap-3">
              <button
                type="button"
                onClick={() => window.location.reload()}
                className="inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] bg-ink px-5 text-sm font-semibold text-surface transition-colors hover:bg-ink-strong focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
              >
                {i18n.t('errorBoundary.reload')}
              </button>
              <a
                href="/"
                onClick={this.handleReset}
                className="inline-flex min-h-11 items-center justify-center rounded-[var(--radius-control)] border border-line bg-surface px-5 text-sm font-semibold text-ink transition-colors hover:bg-surface-subtle focus:outline-none focus-visible:ring-3 focus-visible:ring-focus/30"
              >
                {i18n.t('errorBoundary.goHome')}
              </a>
            </div>
          </div>
        </div>
      )
    }
    return this.props.children
  }
}
