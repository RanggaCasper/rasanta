<script setup lang="ts">
import type { Coordinate, CoordinateSource } from '~/types/location'
import type { RestaurantPin } from '~/types/restaurant'

const props = defineProps<{
  coordinate: Coordinate
  source: CoordinateSource
  label: string
  status: string
  loading: boolean
  error: string
  restaurantsLoading: boolean
  restaurantsError: string
  restaurantCount: number
  restaurants: RestaurantPin[]
}>()

const emit = defineEmits<{
  locateMe: []
}>()

const sourceLabel = computed(() => {
  if (props.source === 'browser') {
    return 'GPS Browser'
  } else if (props.source === 'ip') {
    return 'Perkiraan IP'
  } else if (props.source === 'search') {
    return 'Pencarian / klik peta'
  }

  return 'Default'
})

const coordinateText = computed(() => {
  return `${props.coordinate.lat.toFixed(6)}, ${props.coordinate.lng.toFixed(6)}`
})

const restaurantPreview = computed(() => {
  return props.restaurants.slice(0, 8)
})
</script>

<template>
  <aside class="map-panel">
    <p class="map-panel__eyebrow">
      Rasanta Maps
    </p>

    <h1 class="map-panel__title">
      Peta interaktif lokasi Anda
    </h1>

    <p class="map-panel__status">
      {{ status }}
    </p>

    <p
      v-if="error"
      class="map-panel__error"
    >
      {{ error }}
    </p>

    <p
      v-if="restaurantsError"
      class="map-panel__error"
    >
      {{ restaurantsError }}
    </p>

    <div class="map-panel__actions">
      <UButton
        color="neutral"
        icon="i-lucide-locate-fixed"
        :loading="loading"
        @click="emit('locateMe')"
      >
        Gunakan lokasi saya
      </UButton>
    </div>

    <p class="map-panel__status">
      Query backend: <strong>restaurant</strong>
    </p>

    <p class="map-panel__status">
      {{ restaurantsLoading ? 'Mengambil data restaurant...' : `Restaurant ditemukan: ${restaurantCount}` }}
    </p>

    <dl class="map-panel__meta">
      <div class="map-panel__meta-row">
        <dt class="map-panel__meta-key">
          Sumber koordinat
        </dt>
        <dd class="map-panel__meta-value">
          {{ sourceLabel }}
        </dd>
      </div>

      <div class="map-panel__meta-row">
        <dt class="map-panel__meta-key">
          Koordinat aktif
        </dt>
        <dd class="map-panel__meta-value">
          {{ coordinateText }}
        </dd>
      </div>

      <div class="map-panel__meta-row">
        <dt class="map-panel__meta-key">
          Label titik
        </dt>
        <dd class="map-panel__meta-value">
          {{ label }}
        </dd>
      </div>
    </dl>

    <p class="map-panel__hint">
      Klik area peta untuk memindahkan pin. Data restaurant otomatis di-refresh berdasarkan lat/lng pin.
    </p>

    <ul
      v-if="restaurantPreview.length > 0"
      class="map-panel__restaurant-list"
    >
      <li
        v-for="restaurant in restaurantPreview"
        :key="restaurant.id"
        class="map-panel__restaurant-item"
      >
        <p class="map-panel__restaurant-title">
          {{ restaurant.title }}
        </p>
        <p class="map-panel__restaurant-meta">
          {{ restaurant.rating ?? '-' }} • {{ restaurant.lat.toFixed(5) }}, {{ restaurant.lng.toFixed(5) }}
        </p>
      </li>
    </ul>
  </aside>
</template>
