import { createI18n } from 'vue-i18n'
import ca from './locales/ca.json'
import es from './locales/es.json'
import en from './locales/en.json'

export type SupportedLanguage = 'ca' | 'es' | 'en'

const getInitialLanguage = (): SupportedLanguage => {
  try {
    if (typeof window !== 'undefined' && window.localStorage) {
      const saved = window.localStorage.getItem('encertia_lang') as SupportedLanguage | null
      if (saved && ['ca', 'es', 'en'].includes(saved)) {
        return saved
      }
    }
  } catch {
    // ignore
  }
  return 'ca'
}

export const i18n = createI18n({
  legacy: false,
  locale: getInitialLanguage(),
  fallbackLocale: 'ca',
  messages: {
    ca,
    es,
    en
  }
})

export function setAppLanguage(lang: SupportedLanguage) {
  i18n.global.locale.value = lang
  try {
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.setItem('encertia_lang', lang)
    }
  } catch {
    // ignore storage restrictions
  }
}
