export type CoordinateSource = 'browser' | 'ip' | 'default' | 'search'

export interface Coordinate {
  lat: number
  lng: number
  accuracy?: number
}

export interface LocationSearchResult {
  coordinate: Coordinate
  label: string
}
