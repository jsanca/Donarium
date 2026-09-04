import { useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { LoginExperience } from './features/authentication/pages/LoginExperience'

export function App() {
  const { t } = useTranslation()

  useEffect(() => {
    document.title = t('login.brandName')
  }, [t])

  return <LoginExperience />
}
