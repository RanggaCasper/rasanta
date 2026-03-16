<script setup lang="ts">
useHead({
  htmlAttrs: {
    lang: 'id'
  },
  meta: [
    { name: 'viewport', content: 'width=device-width, initial-scale=1' }
  ],
  link: [
    { rel: 'icon', href: '/favicon.ico' }
  ]
})

const title = 'Rasanta Maps'
const description = 'Peta interaktif berbasis Leaflet dengan deteksi lokasi otomatis dari browser dan fallback IP.'

useSeoMeta({
  title,
  description,
  ogTitle: title,
  ogDescription: description,
  twitterCard: 'summary_large_image'
})
</script>

<template>
  <UApp>
    <NuxtPage />
  </UApp>
</template>
