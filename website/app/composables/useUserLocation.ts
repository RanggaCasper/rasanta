import type { Coordinate, CoordinateSource } from '~/types/location'
import {
  BROWSER_GEOLOCATION_OPTIONS,
  DEFAULT_COORDINATE,
  IP_LOCATION_ENDPOINT
} from '~/config/location'

interface IpWhoResponse {
  success: boolean
  latitude: number
  longitude: number
  city?: string
  region?: string
  country?: string
  message?: string
}

function browserCoordinate(browserUnsupportedMessage: string): Promise<Coordinate> {
  return new Promise((resolve, reject) => {
    if (!navigator.geolocation) {
      reject(new Error(browserUnsupportedMessage))
      return
    }

    navigator.geolocation.getCurrentPosition(
      (position) => {
        resolve({
          lat: position.coords.latitude,
          lng: position.coords.longitude,
          accuracy: position.coords.accuracy
        })
      },
      (error) => {
        reject(error)
      },
      BROWSER_GEOLOCATION_OPTIONS
    )
  })
}

async function ipCoordinate(messages: {
  requestFailed: string
  invalid: string
  fallbackLabel: string
}): Promise<{ coordinate: Coordinate, label: string }> {
  const response = await fetch(IP_LOCATION_ENDPOINT)

  if (!response.ok) {
    throw new Error(messages.requestFailed)
  }

  const result = await response.json() as IpWhoResponse

  if (!result.success || !Number.isFinite(result.latitude) || !Number.isFinite(result.longitude)) {
    throw new Error(result.message || messages.invalid)
  }

  const parts = [result.city, result.region, result.country].filter(Boolean)

  return {
    coordinate: {
      lat: result.latitude,
      lng: result.longitude
    },
    label: parts.length > 0 ? parts.join(', ') : messages.fallbackLabel
  }
}

export function useUserLocation() {
  const { t } = useAppI18n()

  const coordinate = ref<Coordinate>({ ...DEFAULT_COORDINATE })
  const source = ref<CoordinateSource>('default')
  const label = ref(t('location.defaultLabel'))
  const status = ref(t('location.waiting'))
  const loading = ref(false)
  const error = ref('')

  function updateLocation(next: Coordinate, nextSource: CoordinateSource, nextLabel: string) {
    coordinate.value = {
      lat: next.lat,
      lng: next.lng,
      accuracy: next.accuracy
    }
    source.value = nextSource
    label.value = nextLabel
  }

  async function resolveUserLocation() {
    if (import.meta.server) {
      return
    }

    loading.value = true
    error.value = ''
    status.value = t('location.requestingBrowser')

    try {
      try {
        const browser = await browserCoordinate(t('location.browserNotSupported'))
        updateLocation(browser, 'browser', t('location.currentLabel'))
        status.value = t('location.browserOk')
        return
      } catch {
        status.value = t('location.browserDenied')
      }

      try {
        const fromIp = await ipCoordinate({
          requestFailed: t('location.ipRequestFailed'),
          invalid: t('location.ipInvalid'),
          fallbackLabel: t('location.ipLabel')
        })
        updateLocation(fromIp.coordinate, 'ip', fromIp.label)
        status.value = t('location.ipOk')
        return
      } catch {
        updateLocation(DEFAULT_COORDINATE, 'default', t('location.defaultLabel'))
        status.value = t('location.fallback')
        error.value = t('location.readFailed')
      }
    } finally {
      loading.value = false
    }
  }

  function setManualLocation(next: Coordinate, nextSource: CoordinateSource = 'search', nextLabel?: string) {
    updateLocation(next, nextSource, nextLabel || t('source.search'))
    status.value = t('location.manualUpdated')
    error.value = ''
  }

  return {
    coordinate,
    source,
    label,
    status,
    loading,
    error,
    resolveUserLocation,
    setManualLocation
  }
}
