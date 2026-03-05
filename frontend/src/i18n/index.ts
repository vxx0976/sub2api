import { createI18n } from 'vue-i18n'

type LocaleCode = 'en' | 'zh' | 'ru'

type LocaleMessages = Record<string, any>

const LOCALE_KEY = 'sub2api_locale'
const DEFAULT_LOCALE: LocaleCode = 'en'
const SUPPORTED_LOCALES: LocaleCode[] = ['en', 'zh', 'ru']

/** Deep merge custom translations over base (custom keys override base) */
function deepMerge(base: LocaleMessages, custom: LocaleMessages): LocaleMessages {
  const result = { ...base }
  for (const key of Object.keys(custom)) {
    if (
      result[key] && typeof result[key] === 'object' && !Array.isArray(result[key]) &&
      typeof custom[key] === 'object' && !Array.isArray(custom[key])
    ) {
      result[key] = deepMerge(result[key], custom[key])
    } else {
      result[key] = custom[key]
    }
  }
  return result
}

/** Custom locale loaders (dev-only translations, separated to reduce merge conflicts) */
const customLocaleLoaders: Partial<Record<LocaleCode, () => Promise<{ default: LocaleMessages }>>> = {
  en: () => import('./locales/en.custom'),
  zh: () => import('./locales/zh.custom'),
}

const localeLoaders: Record<LocaleCode, () => Promise<{ default: LocaleMessages }>> = {
  en: () => import('./locales/en'),
  zh: () => import('./locales/zh'),
  ru: () => import('./locales/ru')
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
  if (browserLang.startsWith('ru')) {
    return 'ru'
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
  let messages = module.default

  // Merge custom translations if available
  const customLoader = customLocaleLoaders[locale]
  if (customLoader) {
    const customModule = await customLoader()
    messages = deepMerge(messages, customModule.default)
  }

  i18n.global.setLocaleMessage(locale, messages)
  loadedLocales.add(locale)
}

export async function initI18n(): Promise<void> {
  const current = getLocale()
  await loadLocaleMessages(current)
  const langMap: Record<string, string> = {
    zh: 'zh-CN',
    en: 'en',
    ru: 'ru'
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
    ru: 'ru'
  }
  document.documentElement.setAttribute('lang', langMap[locale] || locale)

  // 同步更新浏览器页签标题，使其跟随语言切换
  const { resolveDocumentTitle } = await import('@/router/title')
  const { default: router } = await import('@/router')
  const { useAppStore } = await import('@/stores/app')
  const route = router.currentRoute.value
  const appStore = useAppStore()
  document.title = resolveDocumentTitle(route.meta.title, appStore.siteName, route.meta.titleKey as string)
}

export function getLocale(): LocaleCode {
  const current = i18n.global.locale.value
  return isLocaleCode(current) ? current : DEFAULT_LOCALE
}

export const availableLocales = [
  { code: 'en', name: 'English' },
  { code: 'zh', name: '中文' },
  { code: 'ru', name: 'Русский' }
]

export default i18n
