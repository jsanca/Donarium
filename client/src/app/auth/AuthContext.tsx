import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import type { AuthenticatedPrincipal } from '../lib/api'
import { fetchMe } from '../lib/api'

type AuthStatus = 'idle' | 'loading' | 'authenticated' | 'unauthenticated' | 'error'

type AuthState = {
  status: AuthStatus
  principal: AuthenticatedPrincipal | null
  error: string | null
  refresh: () => Promise<void>
  setPrincipal: (p: AuthenticatedPrincipal | null) => void
}

const AuthContext = createContext<AuthState | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<AuthStatus>('loading')
  const [principal, setPrincipal] = useState<AuthenticatedPrincipal | null>(null)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    const controller = new AbortController()
    setStatus('loading')
    setError(null)
    try {
      const p = await fetchMe(controller.signal)
      setPrincipal(p)
      setStatus('authenticated')
    } catch (err) {
      const apiErr = err as { status?: number; message?: string }
      if (apiErr?.status === 401) {
        setPrincipal(null)
        setStatus('unauthenticated')
      } else {
        setError(apiErr?.message ?? 'Unexpected error.')
        setStatus('error')
      }
    }
  }, [])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const value = useMemo<AuthState>(
    () => ({
      status,
      principal,
      error,
      refresh,
      setPrincipal: (p) => {
        setPrincipal(p)
        setStatus(p ? 'authenticated' : 'unauthenticated')
        setError(null)
      },
    }),
    [status, principal, error, refresh],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
