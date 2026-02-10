import { createI18n } from 'vue-i18n'
import en from './locales/en'
import zh from './locales/zh'
import ja from './locales/ja'
import ko from './locales/ko'
import zhTW from './locales/zhTW'
import ru from './locales/ru'
import fr from './locales/fr'
import de from './locales/de'
import es from './locales/es'
import pt from './locales/pt'

const LOCALE_KEY = 'sub2api_locale'
const SUPPORTED_LOCALES = ['en', 'zh', 'ja', 'ko', 'zh-TW', 'ru', 'fr', 'de', 'es', 'pt']

function getDefaultLocale(): string {
  // Check localStorage first (user manual selection)
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved && SUPPORTED_LOCALES.includes(saved)) {
    return saved
  }

  // Check server-injected default locale
  const appConfig = (window as any).__APP_CONFIG__
  if (appConfig?.default_locale && SUPPORTED_LOCALES.includes(appConfig.default_locale)) {
    return appConfig.default_locale
  }

  // Check browser language
  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('en')) {
    return 'en'
  }
  if (browserLang.startsWith('ja')) {
    return 'ja'
  }
  if (browserLang.startsWith('ko')) {
    return 'ko'
  }
  if (browserLang === 'zh-tw' || browserLang === 'zh-hant') {
    return 'zh-TW'
  }
  if (browserLang.startsWith('ru')) {
    return 'ru'
  }
  if (browserLang.startsWith('fr')) {
    return 'fr'
  }
  if (browserLang.startsWith('de')) {
    return 'de'
  }
  if (browserLang.startsWith('es')) {
    return 'es'
  }
  if (browserLang.startsWith('pt')) {
    return 'pt'
  }

  // 默认中文
  return 'zh'
}

export const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: 'zh',
  messages: {
    en,
    zh,
    ja,
    ko,
    'zh-TW': zhTW,
    ru,
    fr,
    de,
    es,
    pt
  },
  // 禁用 HTML 消息警告 - 引导步骤使用富文本内容（driver.js 支持 HTML）
  // 这些内容是内部定义的，不存在 XSS 风险
  warnHtmlMessage: false
})

export function setLocale(locale: string) {
  if (SUPPORTED_LOCALES.includes(locale)) {
    i18n.global.locale.value = locale as 'en' | 'zh' | 'ja' | 'ko' | 'zh-TW' | 'ru' | 'fr' | 'de' | 'es' | 'pt'
    localStorage.setItem(LOCALE_KEY, locale)
    const langMap: Record<string, string> = {
      zh: 'zh-CN',
      en: 'en',
      ja: 'ja',
      ko: 'ko',
      'zh-TW': 'zh-TW',
      ru: 'ru',
      fr: 'fr',
      de: 'de',
      es: 'es',
      pt: 'pt-BR'
    }
    document.documentElement.setAttribute('lang', langMap[locale] || locale)
  }
}

export function getLocale(): string {
  return i18n.global.locale.value
}

export const availableLocales = [
  { code: 'en', name: 'English' },
  { code: 'zh', name: '中文' },
  { code: 'zh-TW', name: '繁體中文' },
  { code: 'ja', name: '日本語' },
  { code: 'ko', name: '한국어' },
  { code: 'ru', name: 'Русский' },
  { code: 'fr', name: 'Français' },
  { code: 'de', name: 'Deutsch' },
  { code: 'es', name: 'Español' },
  { code: 'pt', name: 'Português' }
]

export default i18n
