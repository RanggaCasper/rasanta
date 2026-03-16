import {
  API_LOCALE_MAP,
  DEFAULT_LOCALE,
  I18N_MESSAGES,
  I18N_STORAGE_KEY,
  LOCALE_OPTIONS,
  type AppLocale,
  type TranslationParams
} from '~/config/i18n'

function resolveNestedKey(dict: Record<string, unknown>, key: string): string | undefined {
  const parts = key.split('.')
  let current: unknown = dict
  for (const part of parts) {
    if (typeof current !== 'object' || current === null) return undefined
    current = (current as Record<string, unknown>)[part]
  }
  return typeof current === 'string' ? current : undefined
}

function fillTemplate(template: string, params?: TranslationParams): string {
  if (!params) {
    return template
  }

  return template.replace(/\{(\w+)\}/g, (_, token: string) => {
    if (token in params) {
      return String(params[token])
    }

    return `{${token}}`
  })
}

function normalizeLocale(input: string | null | undefined): AppLocale {
  const matched = LOCALE_OPTIONS.find(option => option.value === input)
  if (matched) {
    return matched.value
  }

  return DEFAULT_LOCALE
}

export function useAppI18n() {
  const locale = useState<AppLocale>('app-locale', () => DEFAULT_LOCALE)
  const hydrationDone = useState<boolean>('app-locale-hydrated', () => false)
  const storageWatcherBound = useState<boolean>('app-locale-storage-watcher', () => false)

  if (import.meta.client && !hydrationDone.value) {
    const stored = window.localStorage.getItem(I18N_STORAGE_KEY)
    if (stored) {
      locale.value = normalizeLocale(stored)
    }

    hydrationDone.value = true
  }

  if (import.meta.client && !storageWatcherBound.value) {
    watch(locale, (nextLocale) => {
      window.localStorage.setItem(I18N_STORAGE_KEY, nextLocale)
    })
    storageWatcherBound.value = true
  }

  function setLocale(nextLocale: AppLocale) {
    locale.value = nextLocale
  }

  function t(key: string, params?: TranslationParams): string {
    const dict = I18N_MESSAGES[locale.value] || I18N_MESSAGES[DEFAULT_LOCALE]
    const template = resolveNestedKey(dict, key) ?? resolveNestedKey(I18N_MESSAGES[DEFAULT_LOCALE], key) ?? key
    return fillTemplate(template, params)
  }

  const restaurantApiLocale = computed(() => API_LOCALE_MAP[locale.value] || API_LOCALE_MAP[DEFAULT_LOCALE])

  const localeRequestKey = computed(() => `${restaurantApiLocale.value.hl}-${restaurantApiLocale.value.gl}`)

  return {
    locale,
    localeOptions: LOCALE_OPTIONS,
    restaurantApiLocale,
    localeRequestKey,
    setLocale,
    t
  }
}
