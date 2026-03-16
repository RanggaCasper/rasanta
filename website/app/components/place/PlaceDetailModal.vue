<script setup lang="ts">
import type { PlaceDetail, PlaceSummary } from '~/types/place'

const props = defineProps<{
  open: boolean
  place: PlaceSummary | null
  detail: PlaceDetail | null
  loading?: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
}>()

const modelOpen = computed({
  get: () => props.open,
  set: (value: boolean) => emit('update:open', value)
})

const summary = computed(() => props.detail?.summary || props.place)

const serviceRows = computed(() => {
  const options = summary.value?.serviceOptions
  if (!options) {
    return []
  }

  return [
    { label: 'Dine In', value: options.dine_in },
    { label: 'Takeout', value: options.takeout },
    { label: 'Delivery', value: options.delivery }
  ]
})

const rawJson = computed(() => {
  if (!props.detail?.raw) {
    return ''
  }

  return JSON.stringify(props.detail.raw, null, 2)
})
</script>

<template>
  <UModal
    v-model:open="modelOpen"
    :ui="{ content: 'sm:max-w-3xl' }"
  >
    <template #content>
      <div class="detail-modal">
        <div class="detail-modal__header">
          <p class="detail-modal__eyebrow">
            Place Detail
          </p>
          <h2 class="detail-modal__title">
            {{ summary?.title || 'Detail Tempat' }}
          </h2>
          <p class="detail-modal__subtitle">
            {{ summary?.address || 'Alamat tidak tersedia' }}
          </p>
        </div>

        <UAlert
          v-if="error"
          color="error"
          variant="subtle"
          title="Gagal memuat detail"
          :description="error"
          icon="i-lucide-triangle-alert"
        />

        <div
          v-else-if="loading"
          class="detail-modal__loading"
        >
          <USkeleton class="h-5 w-1/3" />
          <USkeleton class="h-4 w-full" />
          <USkeleton class="h-4 w-full" />
          <USkeleton class="h-56 w-full" />
        </div>

        <div
          v-else-if="summary"
          class="detail-modal__body"
        >
          <div class="detail-modal__stats">
            <UCard>
              <p class="detail-modal__stat-label">
                Rating
              </p>
              <p class="detail-modal__stat-value">
                {{ summary.rating ?? 'N/A' }}
              </p>
            </UCard>
            <UCard>
              <p class="detail-modal__stat-label">
                Reviews
              </p>
              <p class="detail-modal__stat-value">
                {{ summary.reviews ?? 0 }}
              </p>
            </UCard>
            <UCard>
              <p class="detail-modal__stat-label">
                Koordinat
              </p>
              <p class="detail-modal__stat-value detail-modal__stat-value--small">
                {{ summary.coordinates.latitude ?? '-' }}, {{ summary.coordinates.longitude ?? '-' }}
              </p>
            </UCard>
          </div>

          <UCard>
            <h3 class="detail-modal__section-title">
              Service Options
            </h3>
            <div class="detail-modal__badges">
              <UBadge
                v-for="item in serviceRows"
                :key="item.label"
                :color="item.value ? 'success' : 'neutral'"
                :variant="item.value ? 'soft' : 'outline'"
              >
                {{ item.label }}: {{ item.value ? 'Yes' : 'No' }}
              </UBadge>
            </div>
          </UCard>

          <UCard>
            <h3 class="detail-modal__section-title">
              Snippet Ulasan
            </h3>
            <ul
              v-if="summary.reviewSnippets.length"
              class="detail-modal__snippet-list"
            >
              <li
                v-for="snippet in summary.reviewSnippets"
                :key="snippet"
              >
                {{ snippet }}
              </li>
            </ul>
            <p v-else>
              Snippet ulasan belum tersedia.
            </p>
          </UCard>

          <UCard>
            <h3 class="detail-modal__section-title">
              Raw Payload
            </h3>
            <pre class="detail-modal__code">{{ rawJson }}</pre>
          </UCard>
        </div>
      </div>
    </template>
  </UModal>
</template>
