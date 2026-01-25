<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <!-- Header -->
    <PublicHeader />

    <!-- Main Content -->
    <main class="mx-auto max-w-7xl px-6 py-8">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
import PublicHeader from './PublicHeader.vue'
</script>
