<script setup lang="ts">
import type { Coordinate } from '~/types/location'
import type { RestaurantPin } from '~/types/restaurant'

const props = withDefaults(defineProps<{
  center: Coordinate
  restaurants?: RestaurantPin[]
  selectedRestaurantId?: string | null
  zoom?: number
}>(), {
  restaurants: () => [],
  selectedRestaurantId: null,
  zoom: 16
})

const emit = defineEmits<{
  pickLocation: [value: Coordinate]
  selectRestaurant: [value: RestaurantPin]
  prefetchRestaurant: [value: RestaurantPin]
}>()

const mapElement = ref<HTMLElement | null>(null)

let leafletLib: typeof import('leaflet') | null = null
let map: import('leaflet').Map | null = null
let marker: import('leaflet').CircleMarker | null = null
let accuracyCircle: import('leaflet').Circle | null = null
let restaurantLayer: import('leaflet').LayerGroup | null = null
let mapResizeObserver: ResizeObserver | null = null
let canUseHoverPreview = false

function syncMapSize() {
  if (!map) {
    return
  }

  map.invalidateSize({
    pan: false,
    animate: false
  })
}

function syncMarker(nextCenter: Coordinate) {
  if (!leafletLib || !map) {
    return
  }

  const position: [number, number] = [nextCenter.lat, nextCenter.lng]

  if (!marker) {
    marker = leafletLib.circleMarker(position, {
      radius: 8,
      color: '#0b4a8f',
      fillColor: '#3aa6ff',
      fillOpacity: 0.92,
      weight: 2
    }).addTo(map)
  } else {
    marker.setLatLng(position)
  }

  if (nextCenter.accuracy && nextCenter.accuracy > 0) {
    if (!accuracyCircle) {
      accuracyCircle = leafletLib.circle(position, {
        radius: nextCenter.accuracy,
        color: '#1a6ec9',
        fillColor: '#1a6ec9',
        fillOpacity: 0.15,
        weight: 1
      }).addTo(map)
    } else {
      accuracyCircle.setLatLng(position)
      accuracyCircle.setRadius(nextCenter.accuracy)
    }
  } else if (accuracyCircle) {
    map.removeLayer(accuracyCircle)
    accuracyCircle = null
  }
}

function escapeHtml(input: string): string {
  return input
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function getPinPreviewHtml(restaurant: RestaurantPin, index: number): string {
  const safeTitle = escapeHtml(restaurant.title)
  const thumbnailUrl = restaurant.thumbnail ? escapeHtml(restaurant.thumbnail) : ''

  const imageBlock = thumbnailUrl
    ? `<img class="restaurant-pin-preview__thumb" src="${thumbnailUrl}" alt="${safeTitle}" />`
    : `<div class="restaurant-pin-preview__thumb restaurant-pin-preview__thumb--fallback">${index + 1}</div>`

  return [
    '<div class="restaurant-pin-preview">',
    imageBlock,
    `<p class="restaurant-pin-preview__title">${safeTitle}</p>`,
    '</div>'
  ].join('')
}

function syncRestaurants(restaurants: RestaurantPin[]) {
  if (!leafletLib || !map) {
    return
  }

  const leaflet = leafletLib

  if (!restaurantLayer) {
    restaurantLayer = leaflet.layerGroup().addTo(map)
  }

  const layer = restaurantLayer
  if (!layer) {
    return
  }

  layer.clearLayers()

  restaurants.forEach((restaurant, index) => {
    const isSelected = props.selectedRestaurantId === restaurant.id
    const markerColorClass = isSelected ? 'bg-cyan-500' : 'bg-red-500'
    const markerRestaurant = leaflet.marker([restaurant.lat, restaurant.lng], {
      icon: leaflet.divIcon({
        className: 'bg-transparent border-0 shadow-none',
        html: `<span class="inline-flex h-7 w-7 items-center justify-center rounded-full border-2 border-white text-xs font-bold text-white shadow-lg shadow-slate-900/30 ${markerColorClass}">${index + 1}</span>`,
        iconSize: [32, 32],
        iconAnchor: [16, 16]
      }),
      autoPanOnFocus: false,
      bubblingMouseEvents: false,
      keyboard: true
    })

    if (!markerRestaurant) {
      return
    }

    const ratingText = restaurant.rating !== null ? `&#9733; ${restaurant.rating.toFixed(1)}` : '&#9733; -'
    const addressText = restaurant.address ? escapeHtml(restaurant.address) : '-'
    const popupHtml = [
      '<div style="display:grid;gap:4px;max-width:260px;">',
      `<strong style="font-size:13px;line-height:1.25;">${escapeHtml(restaurant.title)}</strong>`,
      `<span style="font-size:12px;color:#334155;">${ratingText}</span>`,
      `<span style="font-size:11px;color:#64748b;line-height:1.3;">${addressText}</span>`,
      '</div>'
    ].join('')

    markerRestaurant.bindPopup(
      popupHtml,
      {
        autoPan: false
      }
    )

    if (canUseHoverPreview) {
      markerRestaurant.bindTooltip(getPinPreviewHtml(restaurant, index), {
        direction: 'top',
        offset: [0, -14],
        opacity: 1,
        className: 'restaurant-pin-tooltip',
        sticky: true
      })

      markerRestaurant.on('mouseover', () => {
        markerRestaurant.openTooltip()
        emit('prefetchRestaurant', restaurant)
      })

      markerRestaurant.on('mouseout', () => {
        markerRestaurant.closeTooltip()
      })
    }

    markerRestaurant.on('click', () => {
      emit('selectRestaurant', restaurant)
    })

    markerRestaurant.addTo(layer)
  })
}

onMounted(async () => {
  if (!mapElement.value) {
    return
  }

  leafletLib = await import('leaflet')
  canUseHoverPreview = typeof window !== 'undefined'
    && window.matchMedia('(hover: hover) and (pointer: fine)').matches

  map = leafletLib.map(mapElement.value, {
    zoomControl: false,
    attributionControl: false
  }).setView([props.center.lat, props.center.lng], props.zoom)

  leafletLib.control.zoom({ position: 'bottomright' }).addTo(map)

  leafletLib.control.attribution({
    position: 'bottomleft',
    prefix: false
  }).addTo(map).addAttribution('(c) OpenStreetMap contributors')

  leafletLib.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 20
  }).addTo(map)

  if (typeof ResizeObserver !== 'undefined') {
    mapResizeObserver = new ResizeObserver(() => {
      syncMapSize()
    })

    mapResizeObserver.observe(mapElement.value)
  }

  // Run once after initial layout settle so map fills the available area.
  requestAnimationFrame(() => {
    syncMapSize()
  })

  syncMarker(props.center)
  syncRestaurants(props.restaurants)

  map.on('click', (event) => {
    emit('pickLocation', {
      lat: event.latlng.lat,
      lng: event.latlng.lng
    })
  })
})

watch(() => props.center, (nextCenter) => {
  if (!map) {
    return
  }

  map.setView([nextCenter.lat, nextCenter.lng], map.getZoom(), {
    animate: true,
    duration: 0.5
  })

  syncMarker(nextCenter)
}, { deep: true })

watch(() => props.restaurants, (nextRestaurants) => {
  syncRestaurants(nextRestaurants)
}, { deep: true })

watch(() => props.selectedRestaurantId, (nextSelectedId) => {
  syncRestaurants(props.restaurants)

  if (!map || !nextSelectedId) {
    return
  }

  const selected = props.restaurants.find(item => item.id === nextSelectedId)
  if (!selected) {
    return
  }

  map.flyTo([selected.lat, selected.lng], Math.max(map.getZoom(), 15), {
    animate: true,
    duration: 0.4
  })
})

onUnmounted(() => {
  if (mapResizeObserver) {
    mapResizeObserver.disconnect()
    mapResizeObserver = null
  }

  if (map) {
    map.remove()
    map = null
  }
})
</script>

<template>
  <div
    ref="mapElement"
    class="h-full w-full"
  />
</template>

<style>
.restaurant-pin-tooltip {
  background: #ffffff;
  border: 1px solid #dbe2ea;
  border-radius: 12px;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.18);
  color: #0f172a;
  max-width: min(240px, calc(100vw - 28px));
  padding: 8px;
}

.restaurant-pin-preview {
  display: grid;
  gap: 8px;
  justify-items: center;
  width: min(156px, calc(100vw - 44px));
}

.restaurant-pin-preview__thumb {
  height: 86px;
  width: 140px;
  border-radius: 10px;
  object-fit: cover;
}

.restaurant-pin-preview__thumb--fallback {
  background: #e2e8f0;
  color: #0f172a;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
}

.restaurant-pin-preview__title {
  margin: 0;
  color: #0f172a;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.25;
  text-align: center;
  white-space: normal;
  word-break: break-word;
  display: -webkit-box;
  -line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.restaurant-pin-tooltip::before {
  border-top-color: #ffffff;
}

@media (max-width: 640px) {
  .restaurant-pin-tooltip {
    display: none;
  }
}
</style>
