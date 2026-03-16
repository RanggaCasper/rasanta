<template>
  <header class="site-header">
    <UContainer class="site-header__inner">
      <NuxtLink
        to="/"
        class="brand-link"
        aria-label="Rasanta Home"
      >
        <span class="brand-dot" />
        <div>
          <p class="brand-title">Rasanta</p>
          <p class="brand-subtitle">Culinary Intelligence Explorer</p>
        </div>
      </NuxtLink>

      <div class="site-header__actions">
        <UBadge
          color="primary"
          variant="subtle"
          class="hidden sm:inline-flex"
        >
          Nuxt 4 + Nuxt UI
        </UBadge>
        <UColorModeButton />
      </div>
    </UContainer>
  </header>
</template>
