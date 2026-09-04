import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { RouterProvider } from 'react-router-dom'
import { AuthProvider } from './auth/AuthContext'
import { router } from './router'
import { AppErrorBoundary } from '../shared/ui/ErrorBoundary'

export function App() {
  const { t } = useTranslation()

  useEffect(() => {
    document.title = t('login.brandName')
  }, [t])

  return (
    <AppErrorBoundary>
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </AppErrorBoundary>
  )
}
