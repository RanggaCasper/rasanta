<script setup lang="ts">
import type { PlaceSummary } from '~/types/place'
import PlaceCard from '~/components/place/PlaceCard.vue'

defineProps<{
  places: PlaceSummary[]
  loading?: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  openDetail: [value: PlaceSummary]
}>()
</script>

<template>
  <section class="place-results">
    <div class="place-results__header">
      <h2 class="place-results__title">
        Hasil Pencarian
      </h2>
      <p class="place-results__subtitle">
        Bandingkan rating, ulasan, dan info layanan untuk menentukan pilihan.
      </p>
    </div>

    <UAlert
      v-if="error"
      color="error"
      variant="subtle"
      icon="i-lucide-triangle-alert"
      title="Pencarian gagal"
      :description="error"
    />

    <div
      v-else-if="loading"
      class="place-results__grid"
    >
      <UCard
        v-for="index in 6"
        :key="index"
        class="place-results__skeleton-card"
      >
        <USkeleton class="h-44 w-full rounded-lg" />
        <USkeleton class="mt-4 h-4 w-3/4" />
        <USkeleton class="mt-2 h-4 w-1/2" />
        <USkeleton class="mt-2 h-4 w-full" />
      </UCard>
    </div>

    <UCard
      v-else-if="places.length === 0"
      class="place-results__empty"
    >
      <h3>Belum ada hasil</h3>
      <p>Coba ganti kata kunci atau titik koordinat untuk melihat data tempat makan.</p>
    </UCard>

    <div
      v-else
      class="place-results__grid"
    >
      <PlaceCard
        v-for="place in places"
        :key="place.dataId"
        :place="place"
        @open-detail="emit('openDetail', $event)"
      />
    </div>
  </section>
</template>
