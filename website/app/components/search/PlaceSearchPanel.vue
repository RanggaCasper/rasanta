<script setup lang="ts">
import type { SearchFilters } from '~/types/place'
import { DEFAULT_SEARCH_FILTERS } from '~/types/place'

interface SearchFormState {
  query: string
  lat: string
  lng: string
  limit: string
  saw: boolean
  ratingWeight: string
  reviewsWeight: string
  priceWeight: string
  hl: string
  gl: string
  authuser: string
}

const props = withDefaults(defineProps<{
  loading?: boolean
  initialValue?: Partial<SearchFilters>
}>(), {
  loading: false,
  initialValue: () => ({})
})

const emit = defineEmits<{
  search: [value: SearchFilters]
}>()

const advancedOpen = ref(false)
const form = reactive<SearchFormState>(createFormState({ ...DEFAULT_SEARCH_FILTERS, ...props.initialValue }))

function applyPreset(lat: number, lng: number, query: string) {
  form.lat = String(lat)
  form.lng = String(lng)
  form.query = query
}

function submitSearch() {
  const payload: SearchFilters = {
    query: form.query.trim() || DEFAULT_SEARCH_FILTERS.query,
    lat: toNumber(form.lat, DEFAULT_SEARCH_FILTERS.lat),
    lng: toNumber(form.lng, DEFAULT_SEARCH_FILTERS.lng),
    limit: Math.max(1, Math.round(toNumber(form.limit, DEFAULT_SEARCH_FILTERS.limit))),
    saw: form.saw,
    ratingWeight: toNumber(form.ratingWeight, DEFAULT_SEARCH_FILTERS.ratingWeight),
    reviewsWeight: toNumber(form.reviewsWeight, DEFAULT_SEARCH_FILTERS.reviewsWeight),
    priceWeight: toNumber(form.priceWeight, DEFAULT_SEARCH_FILTERS.priceWeight),
    hl: form.hl.trim() || DEFAULT_SEARCH_FILTERS.hl,
    gl: form.gl.trim() || DEFAULT_SEARCH_FILTERS.gl,
    authuser: form.authuser.trim() || DEFAULT_SEARCH_FILTERS.authuser
  }

  emit('search', payload)
}

function createFormState(filters: SearchFilters): SearchFormState {
  return {
    query: filters.query,
    lat: String(filters.lat),
    lng: String(filters.lng),
    limit: String(filters.limit),
    saw: filters.saw,
    ratingWeight: String(filters.ratingWeight),
    reviewsWeight: String(filters.reviewsWeight),
    priceWeight: String(filters.priceWeight),
    hl: filters.hl,
    gl: filters.gl,
    authuser: filters.authuser
  }
}

function toNumber(input: string, fallback: number): number {
  const parsed = Number(input.trim())
  return Number.isFinite(parsed) ? parsed : fallback
}
</script>

<template>
  <UCard class="search-panel">
    <template #header>
      <div class="search-panel__header">
        <h2 class="search-panel__title">
          Cari Tempat Makan
        </h2>
        <p class="search-panel__subtitle">
          Sesuaikan query, lokasi, dan bobot SAW untuk menemukan tempat terbaik.
        </p>
      </div>
    </template>

    <form
      class="search-panel__form"
      @submit.prevent="submitSearch"
    >
      <label
        class="field-label"
        for="query"
      >Kata kunci</label>
      <UInput
        id="query"
        v-model="form.query"
        icon="i-lucide-search"
        size="lg"
        placeholder="contoh: sate madura surabaya"
      />

      <div class="search-panel__grid">
        <div>
          <label
            class="field-label"
            for="lat"
          >Latitude</label>
          <UInput
            id="lat"
            v-model="form.lat"
            type="number"
            step="0.0001"
          />
        </div>
        <div>
          <label
            class="field-label"
            for="lng"
          >Longitude</label>
          <UInput
            id="lng"
            v-model="form.lng"
            type="number"
            step="0.0001"
          />
        </div>
        <div>
          <label
            class="field-label"
            for="limit"
          >Limit</label>
          <UInput
            id="limit"
            v-model="form.limit"
            type="number"
            min="1"
            max="100"
          />
        </div>
      </div>

      <div class="search-panel__preset">
        <UButton
          color="neutral"
          variant="soft"
          size="sm"
          @click="applyPreset(-7.2575, 112.7521, 'kuliner surabaya')"
        >
          Preset Surabaya
        </UButton>
        <UButton
          color="neutral"
          variant="soft"
          size="sm"
          @click="applyPreset(-8.4095, 115.1889, 'kuliner bali')"
        >
          Preset Bali
        </UButton>
      </div>

      <div class="search-panel__switch-row">
        <label class="search-panel__switch">
          <input
            v-model="form.saw"
            type="checkbox"
          >
          <span>Aktifkan ranking SAW</span>
        </label>

        <UButton
          type="button"
          color="neutral"
          variant="ghost"
          trailing-icon="i-lucide-sliders-horizontal"
          @click="advancedOpen = !advancedOpen"
        >
          {{ advancedOpen ? 'Sembunyikan' : 'Tampilkan' }} pengaturan lanjutan
        </UButton>
      </div>

      <div
        v-if="advancedOpen"
        class="search-panel__advanced"
      >
        <div>
          <label
            class="field-label"
            for="ratingWeight"
          >Rating Weight</label>
          <UInput
            id="ratingWeight"
            v-model="form.ratingWeight"
            type="number"
            step="0.1"
            min="0"
          />
        </div>
        <div>
          <label
            class="field-label"
            for="reviewsWeight"
          >Reviews Weight</label>
          <UInput
            id="reviewsWeight"
            v-model="form.reviewsWeight"
            type="number"
            step="0.1"
            min="0"
          />
        </div>
        <div>
          <label
            class="field-label"
            for="priceWeight"
          >Price Weight</label>
          <UInput
            id="priceWeight"
            v-model="form.priceWeight"
            type="number"
            step="0.1"
            min="0"
          />
        </div>
        <div>
          <label
            class="field-label"
            for="hl"
          >HL</label>
          <UInput
            id="hl"
            v-model="form.hl"
          />
        </div>
        <div>
          <label
            class="field-label"
            for="gl"
          >GL</label>
          <UInput
            id="gl"
            v-model="form.gl"
          />
        </div>
        <div>
          <label
            class="field-label"
            for="authuser"
          >Auth User</label>
          <UInput
            id="authuser"
            v-model="form.authuser"
          />
        </div>
      </div>

      <div class="search-panel__actions">
        <UButton
          type="submit"
          icon="i-lucide-sparkles"
          :loading="loading"
          size="lg"
        >
          Cari Tempat
        </UButton>
      </div>
    </form>
  </UCard>
</template>
