<script setup lang="ts">
import LeafletMap from '~/components/maps/LeafletMap.vue'
import RestaurantExplorerPanel from '~/components/maps/RestaurantExplorerPanel.vue'
import type { Coordinate } from '~/types/location'
import type { RestaurantPin } from '~/types/restaurant'
import {
  DEFAULT_LOCALE,
  LOCALE_OPTIONS,
  type AppLocale
} from '~/config/i18n'

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
  count: restaurantCount,
  selectedRestaurant,
  detail,
  detailLoading,
  detailError,
  isDetailOpen,
  fetchRestaurantsByCoordinate,
  openRestaurantDetail,
  closeRestaurantDetail
} = useRestaurantMap()

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

onMounted(async () => {
  suppressWatchFetch = true
  await resolveUserLocation()
  suppressWatchFetch = false
  await fetchRestaurantsByCoordinate(coordinate.value)
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

    fetchRestaurantsByCoordinate(coordinate.value)
  }
)

watch(localeRequestKey, async () => {
  await fetchRestaurantsByCoordinate(coordinate.value)

  if (isDetailOpen.value && selectedRestaurant.value) {
    await openRestaurantDetail(selectedRestaurant.value)
  }
})

async function handleLocateMe() {
  suppressWatchFetch = true
  await resolveUserLocation()
  suppressWatchFetch = false
  await fetchRestaurantsByCoordinate(coordinate.value)
}

function handleMapPick(coordinateFromMap: Coordinate) {
  const formatted = `${coordinateFromMap.lat.toFixed(5)}, ${coordinateFromMap.lng.toFixed(5)}`
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

function toggleSidebar() {
  isSidebarOpen.value = !isSidebarOpen.value
}

function openSidebar() {
  if (!isSidebarOpen.value) {
    isSidebarOpen.value = true
  }
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
      :restaurants="restaurants"
      :restaurants-loading="restaurantsLoading"
      :restaurants-error="restaurantsError"
      :restaurant-count="restaurantCount"
      :selected-restaurant-id="selectedRestaurant?.id || null"
      :detail="detail"
      :detail-loading="detailLoading"
      :detail-error="detailError"
      :is-detail-open="isDetailOpen"
      @locate-me="handleLocateMe"
      @open-detail="handleOpenRestaurantDetail"
      @close-detail="handleCloseRestaurantDetail"
      @toggle-sidebar="toggleSidebar"
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
          :restaurants="restaurants"
          :selected-restaurant-id="selectedRestaurant?.id || null"
          :zoom="14"
          @pick-location="handleMapPick"
          @select-restaurant="handleOpenRestaurantDetail"
        />
      </ClientOnly>

      <p class="pointer-events-none absolute bottom-4 left-1/2 z-30 -translate-x-1/2 rounded-full border border-slate-400/50 bg-white/90 px-3 py-1.5 text-[0.7rem] tracking-[0.03em] text-slate-900 max-[640px]:bottom-3 max-[640px]:text-[0.68rem]">
        {{ t('map.credit') }}
      </p>
    </section>
  </main>
</template>
