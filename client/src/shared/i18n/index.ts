import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import { en } from './en'
import { es } from './es'

const STORAGE_KEY = 'donarium:lng'

function detectLanguage(): string {
  const stored = (() => {
    try {
      return localStorage.getItem(STORAGE_KEY)
    } catch {
      return null
    }
  })()
  if (stored === 'en' || stored === 'es') return stored

  const nav = navigator.language?.toLowerCase() ?? ''
  if (nav.startsWith('en')) return 'en'
  if (nav.startsWith('es')) return 'es'
  return 'es'
}

const initialLng = detectLanguage()

void i18n.use(initReactI18next).init({
  resources: { es, en },
  lng: initialLng,
  fallbackLng: 'es',
  interpolation: { escapeValue: false },
})

i18n.on('languageChanged', (lng) => {
  try {
    localStorage.setItem(STORAGE_KEY, lng)
  } catch {
    // ignore
  }
})

export default i18n
