export interface RestaurantPin {
  id: string
  dataId: string | null
  placeId: string | null
  title: string
  lat: number
  lng: number
  rating: number | null
  reviews: number | null
  address: string | null
  openState: string | null
  type: string | null
  price: string | null
  thumbnail: string | null
  phone: string | null
  website: string | null
  serviceOptions: Record<string, boolean>
  extensionOfferings: string[]
}

export type RestaurantListFilter =
  | 'all'
  | 'price'
  | 'open_now'
  | 'delivery'
  | 'takeout'
  | 'halal'
  | 'alcohol'

export type RestaurantSortOption =
  | 'recommended'
  | 'rating_desc'
  | 'reviews_desc'
  | 'distance_asc'

export interface RestaurantDetail {
  summary: RestaurantPin
  raw: Record<string, unknown>
}

export interface RestaurantApiResponse {
  status_code: number
  count: number
  data: unknown[]
}

export interface RestaurantDetailResponse {
  status_code: number
  data: Record<string, unknown>
}
