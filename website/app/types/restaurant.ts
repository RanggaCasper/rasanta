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
}

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
