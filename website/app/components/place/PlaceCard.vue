<script setup lang="ts">
import type { PlaceSummary } from '~/types/place'

const props = defineProps<{
  place: PlaceSummary
}>()

const emit = defineEmits<{
  openDetail: [value: PlaceSummary]
}>()

const serviceBadges = computed(() => {
  const rows: string[] = []
  if (props.place.serviceOptions.dine_in) rows.push('Dine In')
  if (props.place.serviceOptions.takeout) rows.push('Takeout')
  if (props.place.serviceOptions.delivery) rows.push('Delivery')
  return rows
})

const ratingText = computed(() => {
  if (props.place.rating == null) {
    return 'N/A'
  }
  return props.place.rating.toFixed(1)
})

const reviewText = computed(() => {
  if (props.place.reviews == null) {
    return '0 review'
  }

  return `${new Intl.NumberFormat('id-ID').format(props.place.reviews)} review`
})
</script>

<template>
  <UCard class="place-card">
    <div class="place-card__visual">
      <img
        v-if="place.thumbnail"
        class="place-card__image"
        :src="place.thumbnail"
        :alt="`Foto ${place.title}`"
        loading="lazy"
      >
      <div
        v-else
        class="place-card__placeholder"
      >
        <span>{{ place.title.slice(0, 1).toUpperCase() }}</span>
      </div>
    </div>

    <div class="place-card__content">
      <div class="place-card__title-row">
        <h3 class="place-card__title">
          {{ place.title }}
        </h3>
        <UBadge
          color="primary"
          variant="subtle"
        >
          #{{ place.position }}
        </UBadge>
      </div>

      <p class="place-card__meta">
        <span class="place-card__rating">{{ ratingText }}</span>
        <span>•</span>
        <span>{{ reviewText }}</span>
        <span v-if="place.type">•</span>
        <span v-if="place.type">{{ place.type }}</span>
      </p>

      <p class="place-card__address">
        {{ place.address || 'Alamat tidak tersedia' }}
      </p>

      <div class="place-card__chips">
        <UBadge
          v-for="service in serviceBadges"
          :key="service"
          color="neutral"
          variant="outline"
          size="sm"
        >
          {{ service }}
        </UBadge>
        <UBadge
          v-if="place.openState"
          color="success"
          variant="soft"
          size="sm"
        >
          {{ place.openState }}
        </UBadge>
      </div>

      <div class="place-card__actions">
        <UButton
          color="primary"
          variant="solid"
          icon="i-lucide-map-pin"
          @click="emit('openDetail', place)"
        >
          Lihat Detail
        </UButton>
        <UButton
          v-if="place.website"
          color="neutral"
          variant="ghost"
          icon="i-lucide-globe"
          :to="place.website"
          target="_blank"
        >
          Website
        </UButton>
      </div>
    </div>
  </UCard>
</template>
