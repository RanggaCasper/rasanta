import idJson from '../../i18n/locales/id.json'
import enJson from '../../i18n/locales/en.json'

export type AppLocale = 'id' | 'en'

export type TranslationParams = Record<string, string | number>

export interface LocaleOption {
  value: AppLocale
  label: string
}

export interface RestaurantApiLocale {
  hl: string
  gl: string
}

export const I18N_STORAGE_KEY = 'rasanta.locale'

export const DEFAULT_LOCALE: AppLocale = 'id'

export const LOCALE_OPTIONS: LocaleOption[] = [
  { value: 'id', label: 'Bahasa Indonesia' },
  { value: 'en', label: 'English' }
]

export const API_LOCALE_MAP: Record<AppLocale, RestaurantApiLocale> = {
  id: {
    hl: 'id',
    gl: 'id'
  },
  en: {
    hl: 'en',
    gl: 'us'
  }
}

export const I18N_MESSAGES: Record<AppLocale, Record<string, unknown>> = {
  id: idJson as Record<string, unknown>,
  en: enJson as Record<string, unknown>
}
