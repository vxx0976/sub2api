<template>
  <div class="flex min-h-screen flex-col bg-[#FAF9F5] dark:bg-dark-950">
    <!-- Header -->
    <PublicHeader />

    <!-- Iframe Container -->
    <main class="relative flex-1">
      <!-- Loading Spinner -->
      <div
        v-if="loading"
        class="absolute inset-0 flex items-center justify-center bg-[#FAF9F5] dark:bg-dark-950"
      >
        <div class="text-center">
          <div class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-primary-500"></div>
          <p class="mt-3 text-sm text-gray-500 dark:text-dark-400">{{ t('status.checking') }}...</p>
        </div>
      </div>

      <!-- Load Failed -->
      <div
        v-if="loadFailed"
        class="absolute inset-0 flex items-center justify-center bg-[#FAF9F5] dark:bg-dark-950"
      >
        <div class="text-center">
          <p class="text-gray-600 dark:text-dark-300">{{ t('status.loadFailed') }}</p>
          <a
            href="https://status.claude.com"
            target="_blank"
            rel="noopener noreferrer"
            class="mt-3 inline-block rounded-lg bg-primary-500 px-5 py-2.5 text-sm font-medium text-white transition-colors hover:bg-primary-600"
          >
            {{ t('status.goToStatusPage') }}
          </a>
        </div>
      </div>

      <!-- Embedded Status Page -->
      <iframe
        v-show="!loading && !loadFailed"
        ref="iframeRef"
        src="https://status.claude.com"
        class="h-full w-full border-0"
        :style="{ minHeight: iframeHeight }"
        @load="onIframeLoad"
        @error="onIframeError"
        sandbox="allow-scripts allow-same-origin allow-popups"
        loading="lazy"
      ></iframe>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicHeader from '@/components/layout/PublicHeader.vue'

const { t } = useI18n()

const loading = ref(true)
const loadFailed = ref(false)
const iframeRef = ref<HTMLIFrameElement | null>(null)
const iframeHeight = ref('calc(100vh - 64px)')

let loadTimeout: ReturnType<typeof setTimeout> | null = null

function onIframeLoad() {
  loading.value = false
  if (loadTimeout) {
    clearTimeout(loadTimeout)
    loadTimeout = null
  }
}

function onIframeError() {
  loading.value = false
  loadFailed.value = true
}

onMounted(() => {
  // Timeout fallback: if iframe doesn't load in 10s, show error
  loadTimeout = setTimeout(() => {
    if (loading.value) {
      loading.value = false
      loadFailed.value = true
    }
  }, 10000)
})

onUnmounted(() => {
  if (loadTimeout) {
    clearTimeout(loadTimeout)
  }
})
</script>
