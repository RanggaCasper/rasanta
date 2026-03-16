export const RESTAURANT_SEARCH_CONFIG = {
  query: 'restaurant',
  limit: 20,
  saw: true,
  authuser: '0'
} as const

export const RESTAURANT_SEARCH_GL = 'US'

export interface QueryPreset {
  value: string
  label: string
}

export const QUERY_PRESETS: QueryPreset[] = [
  { value: 'restaurant', label: 'Restaurant' },
  { value: 'cafe', label: 'Café' },
  { value: 'bakso', label: 'Bakso' },
  { value: 'mie', label: 'Mie' },
  { value: 'soto', label: 'Soto' },
  { value: 'ayam', label: 'Ayam' },
  { value: 'nasi', label: 'Nasi' },
  { value: 'seafood', label: 'Seafood' },
  { value: 'pizza', label: 'Pizza' },
  { value: 'sushi', label: 'Sushi' }
]
