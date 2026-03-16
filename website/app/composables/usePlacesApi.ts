import type { PlaceCoordinates, PlaceDetail, PlaceSummary, SearchFilters } from '~/types/place'
import { DEFAULT_SEARCH_FILTERS } from '~/types/place'

interface PlacesResponse {
  status_code: number
  count: number
  data: unknown[]
}

interface PlaceDetailResponse {
  status_code: number
  data: Record<string, unknown>
}

export function usePlacesApi() {
  const runtimeConfig = useRuntimeConfig()
  const apiBase = runtimeConfig.public.apiBase as string

  const filters = ref<SearchFilters>({ ...DEFAULT_SEARCH_FILTERS })
  const places = ref<PlaceSummary[]>([])
  const selectedPlace = ref<PlaceSummary | null>(null)
  const detail = ref<PlaceDetail | null>(null)

  const isLoading = ref(false)
  const isDetailLoading = ref(false)
  const isDetailOpen = ref(false)

  const error = ref<string | null>(null)
  const detailError = ref<string | null>(null)

  async function searchPlaces(nextFilters?: Partial<SearchFilters>) {
    filters.value = { ...filters.value, ...(nextFilters || {}) }

    isLoading.value = true
    error.value = null
    places.value = []
    selectedPlace.value = null
    detail.value = null
    detailError.value = null
    isDetailOpen.value = false

    try {
      const response = await $fetch<PlacesResponse>(`${apiBase}/api/v1/places`, {
        params: {
          query: filters.value.query,
          lat: filters.value.lat,
          lng: filters.value.lng,
          limit: filters.value.limit,
          saw: filters.value.saw,
          rating_weight: filters.value.ratingWeight,
          reviews_weight: filters.value.reviewsWeight,
          price_weight: filters.value.priceWeight,
          hl: filters.value.hl,
          gl: filters.value.gl,
          authuser: filters.value.authuser
        }
      })

      places.value = (response.data || []).map(item => normalizePlace(item)).filter(item => item.dataId !== '')
    } catch (err) {
      error.value = extractErrorMessage(err, 'Gagal mengambil data tempat. Pastikan backend berjalan.')
    } finally {
      isLoading.value = false
    }
  }

  async function openPlaceDetail(place: PlaceSummary) {
    selectedPlace.value = place
    isDetailOpen.value = true
    detailError.value = null
    detail.value = null

    isDetailLoading.value = true

    try {
      const lat = place.coordinates.latitude ?? filters.value.lat
      const lng = place.coordinates.longitude ?? filters.value.lng

      const response = await $fetch<PlaceDetailResponse>(`${apiBase}/api/v1/places/detail`, {
        params: {
          data_id: place.dataId,
          lat,
          long: lng,
          query: place.title,
          hl: filters.value.hl,
          gl: filters.value.gl,
          authuser: filters.value.authuser
        }
      })

      const summary = normalizePlace(response.data)
      detail.value = {
        summary: {
          ...summary,
          title: summary.title || place.title,
          dataId: summary.dataId || place.dataId,
          coordinates: {
            latitude: summary.coordinates.latitude ?? place.coordinates.latitude,
            longitude: summary.coordinates.longitude ?? place.coordinates.longitude
          }
        },
        raw: response.data
      }
    } catch (err) {
      detailError.value = extractErrorMessage(err, 'Gagal mengambil detail tempat.')
    } finally {
      isDetailLoading.value = false
    }
  }

  function closePlaceDetail() {
    isDetailOpen.value = false
  }

  return {
    filters,
    places,
    selectedPlace,
    detail,
    isLoading,
    isDetailLoading,
    isDetailOpen,
    error,
    detailError,
    searchPlaces,
    openPlaceDetail,
    closePlaceDetail
  }
}

function normalizePlace(input: unknown): PlaceSummary {
  const raw = asRecord(input)

  return {
    position: toNumber(raw.position) ?? 0,
    title: toText(raw.title) || 'Tanpa nama',
    dataId: toText(raw.data_id) || '',
    placeId: toText(raw.place_id) || '',
    rating: toNumber(raw.rating),
    reviews: toNumber(raw.reviews),
    price: toText(raw.price),
    type: toText(raw.type),
    address: toText(raw.address),
    openState: toText(raw.open_state),
    phone: toText(raw.phone),
    website: toText(raw.website),
    thumbnail: toText(raw.thumbnail),
    coordinates: toCoordinates(raw.gps_coordinates),
    serviceOptions: toServiceOptions(raw.service_options),
    reviewSnippets: toSnippetList(raw.review_snippets)
  }
}

function toCoordinates(input: unknown): PlaceCoordinates {
  const value = asRecord(input)
  return {
    latitude: toNumber(value.latitude),
    longitude: toNumber(value.longitude)
  }
}

function toServiceOptions(input: unknown): Record<string, boolean> {
  const value = asRecord(input)
  return {
    dine_in: Boolean(value.dine_in),
    takeout: Boolean(value.takeout),
    delivery: Boolean(value.delivery)
  }
}

function toSnippetList(input: unknown): string[] {
  if (!Array.isArray(input)) {
    return []
  }

  return input
    .map(item => toText(item))
    .filter((item): item is string => Boolean(item))
}

function asRecord(input: unknown): Record<string, unknown> {
  if (input && typeof input === 'object' && !Array.isArray(input)) {
    return input as Record<string, unknown>
  }
  return {}
}

function toText(input: unknown): string | null {
  if (typeof input === 'string') {
    const value = input.trim()
    return value.length > 0 ? value : null
  }

  if (typeof input === 'number' && Number.isFinite(input)) {
    return String(input)
  }

  return null
}

function toNumber(input: unknown): number | null {
  if (typeof input === 'number' && Number.isFinite(input)) {
    return input
  }

  if (typeof input === 'string') {
    const normalized = input.replace(',', '.').trim()
    if (normalized.length === 0) {
      return null
    }

    const parsed = Number(normalized)
    return Number.isFinite(parsed) ? parsed : null
  }

  return null
}

function extractErrorMessage(error: unknown, fallback: string): string {
  const input = error as {
    data?: { error?: string }
    message?: string
  }

  if (input?.data?.error) {
    return input.data.error
  }

  if (input?.message) {
    return input.message
  }

  return fallback
}
