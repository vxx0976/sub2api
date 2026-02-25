import { createI18n } from 'vue-i18n'

type LocaleCode = 'en' | 'zh' | 'ja' | 'ko' | 'zh-TW' | 'ru' | 'fr' | 'de' | 'es' | 'pt'

type LocaleMessages = Record<string, any>

const LOCALE_KEY = 'sub2api_locale'
const DEFAULT_LOCALE: LocaleCode = 'en'
const SUPPORTED_LOCALES: LocaleCode[] = ['en', 'zh', 'ja', 'ko', 'zh-TW', 'ru', 'fr', 'de', 'es', 'pt']

const localeLoaders: Record<LocaleCode, () => Promise<{ default: LocaleMessages }>> = {
  en: () => import('./locales/en'),
  zh: () => import('./locales/zh'),
  ja: () => import('./locales/ja'),
  ko: () => import('./locales/ko'),
  'zh-TW': () => import('./locales/zhTW'),
  ru: () => import('./locales/ru'),
  fr: () => import('./locales/fr'),
  de: () => import('./locales/de'),
  es: () => import('./locales/es'),
  pt: () => import('./locales/pt')
}

function isLocaleCode(value: string): value is LocaleCode {
  return SUPPORTED_LOCALES.includes(value as LocaleCode)
}

function getDefaultLocale(): LocaleCode {
  // Check localStorage first (user manual selection)
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved && isLocaleCode(saved)) {
    return saved
  }

  // Check server-injected default locale
  const appConfig = (window as any).__APP_CONFIG__
  if (appConfig?.default_locale && isLocaleCode(appConfig.default_locale)) {
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
  fallbackLocale: DEFAULT_LOCALE,
  messages: {},
  // 禁用 HTML 消息警告 - 引导步骤使用富文本内容（driver.js 支持 HTML）
  // 这些内容是内部定义的，不存在 XSS 风险
  warnHtmlMessage: false
})

const loadedLocales = new Set<LocaleCode>()

export async function loadLocaleMessages(locale: LocaleCode): Promise<void> {
  if (loadedLocales.has(locale)) {
    return
  }

  const loader = localeLoaders[locale]
  const module = await loader()
  i18n.global.setLocaleMessage(locale, module.default)
  loadedLocales.add(locale)
}

export async function initI18n(): Promise<void> {
  const current = getLocale()
  await loadLocaleMessages(current)
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
  document.documentElement.setAttribute('lang', langMap[current] || current)
}

export async function setLocale(locale: string): Promise<void> {
  if (!isLocaleCode(locale)) {
    return
  }

  await loadLocaleMessages(locale)
  i18n.global.locale.value = locale
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

export function getLocale(): LocaleCode {
  const current = i18n.global.locale.value
  return isLocaleCode(current) ? current : DEFAULT_LOCALE
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
