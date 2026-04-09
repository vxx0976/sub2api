import { useI18n } from 'vue-i18n'
import type { I18nMap } from '@/types'

/**
 * Resolve a localized value from an i18n map, falling back to the default value.
 * Priority: exact locale match → 'en' fallback → default value.
 */
export function useI18nField() {
  const { locale } = useI18n()

  function resolveI18n(i18nMap: I18nMap | null | undefined, fallback: string): string {
    if (!i18nMap) return fallback
    const lang = locale.value
    // zh is the default language stored in the name/description fields
    if (lang === 'zh') return fallback
    if (i18nMap[lang]) return i18nMap[lang]
    if (i18nMap['en']) return i18nMap['en']
    return fallback
  }

  return { resolveI18n }
}
