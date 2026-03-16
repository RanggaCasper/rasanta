import { defineStore } from 'pinia'
import type { Coordinate } from '~/types/location'
import type {
  RestaurantApiResponse,
  RestaurantDetail,
  RestaurantDetailResponse,
  RestaurantPin
} from '~/types/restaurant'
import { RESTAURANT_SEARCH_CONFIG, RESTAURANT_SEARCH_GL } from '~/config/places'

const THUMBNAIL_SCALE_FACTOR = 5
const THUMBNAIL_MAX_DIMENSION = 8192

interface FetchRestaurantsOptions {
  saw?: boolean
  query?: string
}

export const useRestaurantMapStore = defineStore('restaurant-map', () => {
  const { restaurantApiLocale, t } = useAppI18n()

  const runtimeConfig = useRuntimeConfig()
  const apiBase = runtimeConfig.public.apiBase as string

  const restaurants = ref<RestaurantPin[]>([])
  const loading = ref(false)
  const error = ref('')
  const count = ref(0)
  const selectedRestaurant = ref<RestaurantPin | null>(null)
  const detail = ref<RestaurantDetail | null>(null)
  const detailLoading = ref(false)
  const detailError = ref('')
  const isDetailOpen = ref(false)

  const detailCache = ref<Record<string, RestaurantDetail>>({})

  const pendingDetailRequests = new Map<string, Promise<RestaurantDetailResponse>>()
  let requestId = 0

  async function fetchRestaurantsByCoordinate(coordinate: Coordinate, options: FetchRestaurantsOptions = {}) {
    if (import.meta.server) {
      return
    }

    requestId += 1
    const currentRequest = requestId

    loading.value = true
    error.value = ''

    try {
      const response = await $fetch<RestaurantApiResponse>(`${apiBase}/api/v1/places`, {
        params: {
          query: options.query ?? RESTAURANT_SEARCH_CONFIG.query,
          lat: coordinate.lat,
          lng: coordinate.lng,
          limit: RESTAURANT_SEARCH_CONFIG.limit,
          saw: options.saw ?? RESTAURANT_SEARCH_CONFIG.saw,
          hl: restaurantApiLocale.value.hl,
          gl: RESTAURANT_SEARCH_GL,
          authuser: RESTAURANT_SEARCH_CONFIG.authuser
        }
      })

      if (currentRequest !== requestId) {
        return
      }

      const mapped = (response.data || [])
        .map(item => normalizeRestaurant(item))
        .filter((item): item is RestaurantPin => item !== null)

      restaurants.value = mapped
      count.value = mapped.length

      if (selectedRestaurant.value) {
        const matched = mapped.find(item => item.id === selectedRestaurant.value?.id)
        if (matched) {
          selectedRestaurant.value = matched
        } else {
          closeRestaurantDetail()
        }
      }
    } catch (err) {
      if (currentRequest !== requestId) {
        return
      }

      restaurants.value = []
      count.value = 0
      error.value = extractErrorMessage(err, t('api.fetchRestaurantsFailed'))
      closeRestaurantDetail()
    } finally {
      if (currentRequest === requestId) {
        loading.value = false
      }
    }
  }

  async function openRestaurantDetail(restaurant: RestaurantPin) {
    selectedRestaurant.value = restaurant
    detail.value = null
    detailError.value = ''
    detailLoading.value = true
    isDetailOpen.value = true

    if (!restaurant.dataId) {
      detail.value = {
        summary: restaurant,
        raw: {}
      }
      detailError.value = t('api.dataIdMissing')
      detailLoading.value = false
      return
    }

    const cacheKey = makeDetailCacheKey(restaurant)
    const cached = detailCache.value[cacheKey]
    if (cached) {
      detail.value = cached
      detailLoading.value = false
      return
    }

    try {
      const response = await requestRestaurantDetail(restaurant)

      const normalized = normalizeRestaurant(response.data, {
        lat: restaurant.lat,
        lng: restaurant.lng
      })

      const nextDetail: RestaurantDetail = {
        summary: normalized || restaurant,
        raw: asRecord(response.data)
      }

      detail.value = nextDetail
      detailCache.value[cacheKey] = nextDetail
    } catch (err) {
      detailError.value = extractErrorMessage(err, t('api.fetchDetailFailed'))
      detail.value = {
        summary: restaurant,
        raw: {}
      }
    } finally {
      detailLoading.value = false
    }
  }

  async function requestRestaurantDetail(restaurant: RestaurantPin): Promise<RestaurantDetailResponse> {
    if (!restaurant.dataId) {
      throw new Error(t('api.dataIdMissing'))
    }

    const requestKey = makeDetailCacheKey(restaurant)
    const existing = pendingDetailRequests.get(requestKey)
    if (existing) {
      return existing
    }

    const requestPromise = $fetch<RestaurantDetailResponse>(`${apiBase}/api/v1/places/detail`, {
      params: {
        data_id: restaurant.dataId,
        lat: restaurant.lat,
        long: restaurant.lng,
        query: RESTAURANT_SEARCH_CONFIG.query,
        hl: restaurantApiLocale.value.hl,
        gl: restaurantApiLocale.value.gl,
        authuser: RESTAURANT_SEARCH_CONFIG.authuser
      }
    }).finally(() => {
      pendingDetailRequests.delete(requestKey)
    })

    pendingDetailRequests.set(requestKey, requestPromise)
    return requestPromise
  }

  function makeDetailCacheKey(restaurant: RestaurantPin): string {
    if (!restaurant.dataId) {
      return `${restaurant.id}:${restaurantApiLocale.value.hl}:${restaurantApiLocale.value.gl}`
    }

    return `${restaurant.dataId}:${restaurantApiLocale.value.hl}:${restaurantApiLocale.value.gl}`
  }

  function applyThumbnail(restaurantId: string, thumbnail: string, activeRequestId: number) {
    if (activeRequestId !== requestId) {
      return
    }

    const index = restaurants.value.findIndex(item => item.id === restaurantId)
    if (index === -1) {
      return
    }

    const current = restaurants.value[index]
    if (!current || current.thumbnail) {
      return
    }

    restaurants.value[index] = {
      ...current,
      thumbnail
    }

    if (selectedRestaurant.value?.id === restaurantId && !selectedRestaurant.value.thumbnail) {
      selectedRestaurant.value = {
        ...selectedRestaurant.value,
        thumbnail
      }
    }

    if (detail.value?.summary?.id === restaurantId && !detail.value.summary.thumbnail) {
      detail.value = {
        ...detail.value,
        summary: {
          ...detail.value.summary,
          thumbnail
        }
      }
    }
  }

  async function prefetchRestaurantDetail(restaurant: RestaurantPin) {
    if (!restaurant.dataId) {
      return
    }

    const cacheKey = makeDetailCacheKey(restaurant)
    if (detailCache.value[cacheKey]) {
      return
    }

    try {
      const response = await requestRestaurantDetail(restaurant)

      const normalized = normalizeRestaurant(response.data, {
        lat: restaurant.lat,
        lng: restaurant.lng
      })

      const nextDetail: RestaurantDetail = {
        summary: normalized || restaurant,
        raw: asRecord(response.data)
      }

      detailCache.value[cacheKey] = nextDetail

      if (normalized?.thumbnail) {
        applyThumbnail(restaurant.id, normalized.thumbnail, requestId)
      }
    } catch {
      // Ignore prefetch errors silently.
    }
  }

  function closeRestaurantDetail() {
    isDetailOpen.value = false
    selectedRestaurant.value = null
    detail.value = null
    detailError.value = ''
    detailLoading.value = false
  }

  function clearDetailCache() {
    detailCache.value = {}
  }

  return {
    restaurants,
    loading,
    error,
    count,
    selectedRestaurant,
    detail,
    detailLoading,
    detailError,
    isDetailOpen,
    detailCache,
    fetchRestaurantsByCoordinate,
    openRestaurantDetail,
    prefetchRestaurantDetail,
    closeRestaurantDetail,
    clearDetailCache
  }
})

function normalizeRestaurant(input: unknown, fallbackCoordinate?: { lat: number, lng: number }): RestaurantPin | null {
  const raw = asRecord(input)
  const gps = asRecord(raw.gps_coordinates)

  const lat = toNumber(gps.latitude) ?? fallbackCoordinate?.lat ?? null
  const lng = toNumber(gps.longitude) ?? fallbackCoordinate?.lng ?? null
  const title = toText(raw.title)

  if (lat === null || lng === null || title === null) {
    return null
  }

  const id = toText(raw.data_id) || toText(raw.place_id) || `${lat},${lng}-${title}`

  return {
    id,
    dataId: toText(raw.data_id),
    placeId: toText(raw.place_id),
    title,
    lat,
    lng,
    rating: toNumber(raw.rating),
    reviews: toNumber(raw.reviews),
    address: toText(raw.address),
    openState: toText(raw.open_state),
    type: toText(raw.type),
    price: toText(raw.price),
    thumbnail: upscaleThumbnailUrl(toText(raw.thumbnail)),
    phone: toText(raw.phone),
    website: toText(raw.website),
    serviceOptions: toServiceOptions(raw.service_options),
    extensionOfferings: toExtensionOfferings(
      raw.extension ?? raw.extensions ?? raw.feature_extensions
    )
  }
}

function toServiceOptions(input: unknown): Record<string, boolean> {
  const value = asRecord(input)

  return Object.entries(value).reduce<Record<string, boolean>>((acc, [key, rawValue]) => {
    acc[key] = Boolean(rawValue)
    return acc
  }, {})
}

function toExtensionOfferings(input: unknown): string[] {
  const offerings: string[] = []
  const seen = new Set<string>()

  const pushOffering = (value: unknown) => {
    const text = toText(value)
    if (!text) {
      return
    }

    const normalized = text.toLowerCase()
    if (seen.has(normalized)) {
      return
    }

    seen.add(normalized)
    offerings.push(text)
  }

  const walk = (value: unknown) => {
    if (Array.isArray(value)) {
      value.forEach(item => {
        walk(item)
      })
      return
    }

    if (!value || typeof value !== 'object') {
      return
    }

    const record = asRecord(value)

    Object.entries(record).forEach(([key, entry]) => {
      if (key.trim().toLowerCase() === 'offerings') {
        if (Array.isArray(entry)) {
          entry.forEach(pushOffering)
          return
        }

        pushOffering(entry)
        return
      }

      walk(entry)
    })
  }

  walk(input)
  return offerings
}

function upscaleThumbnailUrl(url: string | null): string | null {
  if (!url) {
    return null
  }

  const scaleValue = (value: string): string => {
    const parsed = Number(value)
    if (!Number.isFinite(parsed) || parsed <= 0) {
      return value
    }

    const scaled = Math.min(Math.round(parsed * THUMBNAIL_SCALE_FACTOR), THUMBNAIL_MAX_DIMENSION)
    return String(scaled)
  }

  let result = url

  // Pattern: ?w=80&h=92
  result = result.replace(/([?&](?:w|h|s)=)(\d{1,5})/gi, (_, prefix: string, value: string) => {
    return `${prefix}${scaleValue(value)}`
  })

  // Pattern: =w80-h92-k-no or -s40-k
  result = result.replace(/([=\-])(w|h|s)(\d{1,5})(?=[\-.,_=&/?]|$)/gi, (
    _,
    separator: string,
    token: string,
    value: string
  ) => {
    return `${separator}${token}${scaleValue(value)}`
  })

  return result
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
    const value = input.trim().replace(',', '.')
    if (value.length === 0) {
      return null
    }

    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : null
  }

  return null
}

function extractErrorMessage(err: unknown, fallback: string): string {
  const input = err as {
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
