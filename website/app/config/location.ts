import type { Coordinate } from '~/types/location'

export const DEFAULT_COORDINATE: Coordinate = {
  lat: -6.175392,
  lng: 106.827153
}

export const BROWSER_GEOLOCATION_OPTIONS: PositionOptions = {
  enableHighAccuracy: true,
  timeout: 8000,
  maximumAge: 0
}

export const IP_LOCATION_ENDPOINT = 'https://ipwho.is/'
