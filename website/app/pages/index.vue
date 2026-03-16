<script setup lang="ts">
import LeafletMap from '~/components/maps/LeafletMap.vue'
import RestaurantExplorerPanel from '~/components/maps/RestaurantExplorerPanel.vue'
import type { Coordinate } from '~/types/location'
import type {
  RestaurantListFilter,
  RestaurantPin,
  RestaurantSortOption
} from '~/types/restaurant'
import {
  DEFAULT_LOCALE,
  LOCALE_OPTIONS,
  type AppLocale
} from '~/config/i18n'
import { QUERY_PRESETS, RESTAURANT_SEARCH_CONFIG } from '~/config/places'

const {
  coordinate,
  source,
  label,
  status,
  loading,
  resolveUserLocation,
  setManualLocation
} = useUserLocation()

const {
  restaurants,
  loading: restaurantsLoading,
  error: restaurantsError,
  selectedRestaurant,
  detail,
  detailLoading,
  detailError,
  isDetailOpen,
  fetchRestaurantsByCoordinate,
  openRestaurantDetail,
  prefetchRestaurantDetail,
  closeRestaurantDetail
} = useRestaurantMap()

const sawEnabled = ref(true)
const activeFilter = ref<RestaurantListFilter>('all')
const sortBy = ref<RestaurantSortOption>('recommended')
const activeQuery = ref<string>(RESTAURANT_SEARCH_CONFIG.query)

const visibleRestaurants = computed(() => {
  const filtered = filterRestaurants(restaurants.value, activeFilter.value)
  return sortRestaurants(filtered, sortBy.value, coordinate.value)
})

const visibleRestaurantCount = computed(() => {
  return visibleRestaurants.value.length
})

const {
  locale,
  localeOptions,
  localeRequestKey,
  setLocale,
  t
} = useAppI18n()

const isSidebarOpen = ref(true)

const localeModel = computed({
  get: () => locale.value,
  set: (nextLocale: string) => {
    const matched = LOCALE_OPTIONS.find(option => option.value === nextLocale)
    if (matched) {
      setLocale(matched.value as AppLocale)
      return
    }

    setLocale(DEFAULT_LOCALE)
  }
})

const mainLayoutClass = computed(() => {
  return isSidebarOpen.value
    ? 'grid min-h-screen grid-cols-1 lg:h-screen lg:overflow-hidden lg:grid-cols-[minmax(420px,56%)_minmax(360px,44%)]'
    : 'grid min-h-screen grid-cols-1 lg:h-screen lg:overflow-hidden lg:grid-cols-1'
})

const mapSectionClass = computed(() => {
  return isSidebarOpen.value
    ? 'relative hidden lg:block lg:h-screen lg:min-h-0'
    : 'relative min-h-screen lg:h-screen lg:min-h-0'
})

let suppressWatchFetch = false

async function fetchRestaurantsWithCurrentOptions() {
  await fetchRestaurantsByCoordinate(coordinate.value, {
    saw: sawEnabled.value,
    query: activeQuery.value
  })
}

onMounted(async () => {
  suppressWatchFetch = true
  await resolveUserLocation()
  suppressWatchFetch = false
  await fetchRestaurantsWithCurrentOptions()
})

watch(
  () => [coordinate.value.lat, coordinate.value.lng] as const,
  ([nextLat, nextLng], [prevLat, prevLng]) => {
    if (suppressWatchFetch) {
      return
    }

    if (nextLat === prevLat && nextLng === prevLng) {
      return
    }

    fetchRestaurantsWithCurrentOptions()
  }
)

watch(localeRequestKey, async () => {
  await fetchRestaurantsWithCurrentOptions()

  if (isDetailOpen.value && selectedRestaurant.value) {
    await openRestaurantDetail(selectedRestaurant.value)
  }
})

watch(sawEnabled, async () => {
  await fetchRestaurantsWithCurrentOptions()
})

async function handleLocateMe() {
  suppressWatchFetch = true
  await resolveUserLocation()
  suppressWatchFetch = false
  await fetchRestaurantsWithCurrentOptions()
}

function handleMapPick(coordinateFromMap: Coordinate) {
  setManualLocation(
    coordinateFromMap,
    'search',
    t('location.pinLabel', {
      lat: coordinateFromMap.lat.toFixed(5),
      lng: coordinateFromMap.lng.toFixed(5)
    })
  )
}

function handleOpenRestaurantDetail(restaurant: RestaurantPin) {
  openRestaurantDetail(restaurant)
}

function handleCloseRestaurantDetail() {
  closeRestaurantDetail()
}

function handleToggleSaw() {
  sawEnabled.value = !sawEnabled.value
}

function handleChangeFilter(nextFilter: RestaurantListFilter) {
  activeFilter.value = nextFilter
}

function handleChangeSort(nextSort: RestaurantSortOption) {
  sortBy.value = nextSort
}

async function handleChangeQuery(nextQuery: string) {
  activeQuery.value = nextQuery
  await fetchRestaurantsWithCurrentOptions()
}

function handlePrefetchDetail(restaurant: RestaurantPin) {
  prefetchRestaurantDetail(restaurant)
}

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

function openSidebar() {
  if (!isSidebarOpen.value) {
    isSidebarOpen.value = true
  }
}

function filterRestaurants(input: RestaurantPin[], filter: RestaurantListFilter): RestaurantPin[] {
  if (filter === 'all') {
    return input
  }

  if (filter === 'price') {
    return input.filter(item => Boolean(item.price))
  }

  if (filter === 'open_now') {
    return input.filter(item => isOpenNow(item.openState))
  }

  if (filter === 'delivery') {
    return input.filter(item => hasServiceOption(item, key => key.includes('delivery') || key.includes('antar')))
  }

  if (filter === 'takeout') {
    return input.filter(item => hasServiceOption(item, key => key.includes('takeout') || key.includes('bawa')))
  }

  if (filter === 'halal') {
    return input.filter(item => hasOfferingTag(item, [/\bhalal\b/i, /\bsyariah\b/i]))
  }

  if (filter === 'alcohol') {
    return input.filter(item => hasOfferingTag(item, [
      /\balcohol\b/i,
      /\balkohol\b/i,
      /\bbeer\b/i,
      /\bwine\b/i,
      /\bcocktail\b/i,
      /\bpub\b/i,
      /\bbar\b/i,
      /\bliquor\b/i,
      /\bbrewery\b/i,
      /\bvodka\b/i,
      /\bwhisk(?:y|e)y\b/i,
      /\bminuman keras\b/i
    ]))
  }

  return input
}

function sortRestaurants(
  input: RestaurantPin[],
  sort: RestaurantSortOption,
  center: Coordinate
): RestaurantPin[] {
  if (sort === 'recommended') {
    return [...input]
  }

  const list = [...input]

  if (sort === 'rating_desc') {
    return list.sort((a, b) => {
      const ratingA = a.rating ?? Number.NEGATIVE_INFINITY
      const ratingB = b.rating ?? Number.NEGATIVE_INFINITY

      if (ratingA !== ratingB) {
        return ratingB - ratingA
      }

      const reviewsA = a.reviews ?? Number.NEGATIVE_INFINITY
      const reviewsB = b.reviews ?? Number.NEGATIVE_INFINITY
      return reviewsB - reviewsA
    })
  }

  if (sort === 'reviews_desc') {
    return list.sort((a, b) => {
      const reviewsA = a.reviews ?? Number.NEGATIVE_INFINITY
      const reviewsB = b.reviews ?? Number.NEGATIVE_INFINITY
      return reviewsB - reviewsA
    })
  }

  return list.sort((a, b) => {
    return distanceScore(a, center) - distanceScore(b, center)
  })
}

function hasServiceOption(
  restaurant: RestaurantPin,
  match: (key: string) => boolean
): boolean {
  return Object.entries(restaurant.serviceOptions).some(([key, value]) => {
    return Boolean(value) && match(key.toLowerCase())
  })
}

function isOpenNow(openState: string | null): boolean {
  if (!openState) {
    return false
  }

  const normalized = openState.toLowerCase().replace(/\s+/g, ' ').trim()

  if (normalized.length === 0) {
    return false
  }

  if (/(^|[\s\u00b7])(?:buka|open)\s*24\s*jam\b/.test(normalized) || /open\s*24\s*hours\b/.test(normalized)) {
    return true
  }

  if (/^(?:segera\s+tutup|closing\s+soon)\b/.test(normalized)) {
    return true
  }

  if (/^(?:buka|open)\b/.test(normalized)) {
    return true
  }

  if (/^(?:tutup|closed)\b/.test(normalized)) {
    return false
  }

  if (/(^|[\s\u00b7])(?:segera\s+tutup|closing\s+soon)\b/.test(normalized)) {
    return true
  }

  if (/(^|[\s\u00b7])(?:buka|open)\b/.test(normalized) && !/(^|[\s\u00b7])(?:tutup|closed)\b/.test(normalized)) {
    return true
  }

  return false
}

function hasOfferingTag(restaurant: RestaurantPin, patterns: RegExp[]): boolean {
  const haystack = restaurant.extensionOfferings
    .map(item => item.trim())
    .filter(item => item.length > 0)
    .join(' ')

  if (haystack.length === 0) {
    return false
  }

  return patterns.some(pattern => pattern.test(haystack))
}

function distanceScore(restaurant: RestaurantPin, center: Coordinate): number {
  const latDiff = restaurant.lat - center.lat
  const lngDiff = restaurant.lng - center.lng
  return (latDiff * latDiff) + (lngDiff * lngDiff)
}
</script>

<template>
  <main :class="mainLayoutClass">
    <RestaurantExplorerPanel
      v-if="isSidebarOpen"
      class="max-[1023px]:fixed max-[1023px]:inset-0 max-[1023px]:z-1300 max-[1023px]:h-screen max-[1023px]:min-h-screen lg:h-screen lg:min-h-0"
      :coordinate="coordinate"
      :source="source"
      :label="label"
      :status="status"
      :loading="loading"
      :restaurants="visibleRestaurants"
      :restaurants-loading="restaurantsLoading"
      :restaurants-error="restaurantsError"
      :restaurant-count="visibleRestaurantCount"
      :saw-enabled="sawEnabled"
      :active-filter="activeFilter"
      :sort-by="sortBy"
      :active-query="activeQuery"
      :query-presets="QUERY_PRESETS"
      :selected-restaurant-id="selectedRestaurant?.id || null"
      :detail="detail"
      :detail-loading="detailLoading"
      :detail-error="detailError"
      :is-detail-open="isDetailOpen"
      @locate-me="handleLocateMe"
      @open-detail="handleOpenRestaurantDetail"
      @close-detail="handleCloseRestaurantDetail"
      @toggle-sidebar="toggleSidebar"
      @toggle-saw="handleToggleSaw"
      @change-filter="handleChangeFilter"
      @change-sort="handleChangeSort"
      @change-query="handleChangeQuery"
      @prefetch-detail="handlePrefetchDetail"
    />

    <section :class="mapSectionClass">
      <div class="absolute right-4 top-4 z-1200 flex items-center gap-2 rounded-xl bg-white/95 px-3 py-2 shadow-lg ring-1 ring-slate-200 backdrop-blur-sm">
        <label
          for="language-select"
          class="text-xs font-semibold uppercase tracking-wide text-slate-600"
        >
          {{ t('language.label') }}
        </label>
        <select
          id="language-select"
          v-model="localeModel"
          class="rounded-lg border border-slate-300 bg-white px-2.5 py-1.5 text-sm font-medium text-slate-800 outline-none ring-sky-500 transition focus:ring-2"
        >
          <option
            v-for="option in localeOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </option>
        </select>
      </div>

      <UButton
        v-if="isSidebarOpen"
        color="neutral"
        variant="solid"
        icon="i-lucide-panel-left-open"
        class="absolute left-4 top-4 z-1200 font-bold shadow-2xl shadow-slate-900/35 ring-2 ring-white"
        @click="toggleSidebar()"
      >
        {{ t('action.hideSidebar') }}
      </UButton>

      <UButton
        v-else
        color="neutral"
        variant="solid"
        icon="i-lucide-panel-left-open"
        class="absolute left-4 top-4 z-1200 font-bold shadow-2xl shadow-slate-900/35 ring-2 ring-white"
        @click="openSidebar()"
      >
        {{ t('action.openSidebar') }}
      </UButton>

      <ClientOnly>
        <LeafletMap
          :center="coordinate"
          :restaurants="visibleRestaurants"
          :selected-restaurant-id="selectedRestaurant?.id || null"
          :zoom="14"
          @pick-location="handleMapPick"
          @select-restaurant="handleOpenRestaurantDetail"
          @prefetch-restaurant="handlePrefetchDetail"
        />
      </ClientOnly>

      <p class="pointer-events-none absolute bottom-4 left-1/2 z-30 -translate-x-1/2 rounded-full border border-slate-400/50 bg-white/90 px-3 py-1.5 text-[0.7rem] tracking-[0.03em] text-slate-900 max-[640px]:bottom-3 max-[640px]:text-[0.68rem]">
        {{ t('map.credit') }}
      </p>
    </section>
  </main>
</template>
