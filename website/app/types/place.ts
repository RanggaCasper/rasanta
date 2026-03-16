export interface PlaceCoordinates {
  latitude: number | null
  longitude: number | null
}

export interface PlaceSummary {
  position: number
  title: string
  dataId: string
  placeId: string
  rating: number | null
  reviews: number | null
  price: string | null
  type: string | null
  address: string | null
  openState: string | null
  phone: string | null
  website: string | null
  thumbnail: string | null
  coordinates: PlaceCoordinates
  serviceOptions: Record<string, boolean>
  reviewSnippets: string[]
}

export interface PlaceDetail {
  summary: PlaceSummary
  raw: Record<string, unknown>
}

export interface SearchFilters {
  query: string
  lat: number
  lng: number
  limit: number
  saw: boolean
  ratingWeight: number
  reviewsWeight: number
  priceWeight: number
  hl: string
  gl: string
  authuser: string
}

export const DEFAULT_SEARCH_FILTERS: SearchFilters = {
  query: 'kuliner surabaya',
  lat: -7.2575,
  lng: 112.7521,
  limit: 20,
  saw: true,
  ratingWeight: 0.6,
  reviewsWeight: 0.3,
  priceWeight: 0.1,
  hl: 'id',
  gl: 'id',
  authuser: '0'
}
