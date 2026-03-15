<template>
  <div class="flex min-h-screen flex-col bg-[#FAF9F5] dark:bg-dark-950">
    <PublicHeader />

    <main class="mx-auto w-full max-w-3xl flex-1 px-4 py-8 sm:px-6">
      <!-- Overall Status Banner -->
      <div
        class="mb-6 rounded-xl border px-5 py-4"
        :class="overallBannerClass"
      >
        <div class="flex items-center gap-3">
          <span class="relative flex h-3 w-3">
            <span
              v-if="overallStatus === 'operational'"
              class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"
              :class="overallDotPingClass"
            ></span>
            <span class="relative inline-flex h-3 w-3 rounded-full" :class="overallDotClass"></span>
          </span>
          <span class="text-sm font-semibold" :class="overallTextClass">{{ overallMessage }}</span>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-16">
        <div class="text-center">
          <div class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-primary-500"></div>
          <p class="mt-3 text-sm text-gray-500 dark:text-dark-400">{{ t('status.checking') }}...</p>
        </div>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="flex items-center justify-center py-16">
        <div class="text-center">
          <p class="text-gray-600 dark:text-dark-300">{{ t('status.loadFailed') }}</p>
          <button
            @click="fetchStatus"
            class="mt-3 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white hover:bg-primary-600"
          >{{ t('status.retry') }}</button>
        </div>
      </div>

      <!-- Group Cards -->
      <div v-else class="space-y-3">
        <div
          v-for="group in groups"
          :key="group.id"
          class="rounded-xl border border-gray-200/80 bg-white px-5 py-4 dark:border-dark-700/60 dark:bg-dark-900"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <span class="relative flex h-2.5 w-2.5">
                <span
                  v-if="group.status === 'operational'"
                  class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
                ></span>
                <span class="relative inline-flex h-2.5 w-2.5 rounded-full" :class="statusDotClass(group.status)"></span>
              </span>
              <span class="text-sm font-medium text-gray-900 dark:text-white">{{ group.name }}</span>
            </div>
            <div class="flex items-center gap-4">
              <span v-if="group.total_requests > 0" class="text-xs text-gray-400 dark:text-dark-500">
                {{ group.total_requests }} {{ t('status.requests') }}
              </span>
              <span class="text-xs font-medium" :class="statusTextClass(group.status)">
                {{ statusLabel(group.status) }}
              </span>
              <span v-if="group.total_requests > 0" class="text-xs tabular-nums" :class="rateTextClass(group.status)">
                {{ group.success_rate.toFixed(1) }}%
              </span>
            </div>
          </div>
        </div>

        <!-- Empty state -->
        <div v-if="groups.length === 0" class="py-16 text-center text-sm text-gray-400 dark:text-dark-500">
          {{ t('status.noGroups') }}
        </div>
      </div>

      <!-- Footer -->
      <div v-if="updatedAt" class="mt-6 text-center text-xs text-gray-400 dark:text-dark-500">
        {{ t('status.lastUpdated') }}: {{ formattedTime }}
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import PublicHeader from '@/components/layout/PublicHeader.vue'
import { getGroupStatus, type GroupStatusItem } from '@/api/status'

const { t, locale } = useI18n()

const loading = ref(true)
const error = ref(false)
const groups = ref<GroupStatusItem[]>([])
const updatedAt = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

const overallStatus = computed(() => {
  if (groups.value.length === 0) return 'unknown'
  const hasDown = groups.value.some(g => g.status === 'down')
  if (hasDown) return 'down'
  const hasDegraded = groups.value.some(g => g.status === 'degraded')
  if (hasDegraded) return 'degraded'
  const allUnknown = groups.value.every(g => g.status === 'unknown')
  if (allUnknown) return 'unknown'
  return 'operational'
})

const overallMessage = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return t('status.allOperational')
    case 'degraded': return t('status.partialOutage')
    case 'down': return t('status.majorOutage')
    default: return t('status.checking')
  }
})

const overallBannerClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return 'border-emerald-200 bg-emerald-50 dark:border-emerald-900/40 dark:bg-emerald-950/30'
    case 'degraded': return 'border-amber-200 bg-amber-50 dark:border-amber-900/40 dark:bg-amber-950/30'
    case 'down': return 'border-red-200 bg-red-50 dark:border-red-900/40 dark:bg-red-950/30'
    default: return 'border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-900'
  }
})

const overallDotClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return 'bg-emerald-500'
    case 'degraded': return 'bg-amber-500'
    case 'down': return 'bg-red-500'
    default: return 'bg-gray-400'
  }
})

const overallDotPingClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return 'bg-emerald-400'
    default: return ''
  }
})

const overallTextClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return 'text-emerald-700 dark:text-emerald-400'
    case 'degraded': return 'text-amber-700 dark:text-amber-400'
    case 'down': return 'text-red-700 dark:text-red-400'
    default: return 'text-gray-600 dark:text-dark-400'
  }
})

function statusDotClass(status: string) {
  switch (status) {
    case 'operational': return 'bg-emerald-500'
    case 'degraded': return 'bg-amber-500'
    case 'down': return 'bg-red-500'
    default: return 'bg-gray-400'
  }
}

function statusTextClass(status: string) {
  switch (status) {
    case 'operational': return 'text-emerald-600 dark:text-emerald-400'
    case 'degraded': return 'text-amber-600 dark:text-amber-400'
    case 'down': return 'text-red-600 dark:text-red-400'
    default: return 'text-gray-400 dark:text-dark-500'
  }
}

function rateTextClass(status: string) {
  switch (status) {
    case 'operational': return 'text-emerald-600 dark:text-emerald-400'
    case 'degraded': return 'text-amber-600 dark:text-amber-400'
    case 'down': return 'text-red-600 dark:text-red-400'
    default: return 'text-gray-400 dark:text-dark-500'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'operational': return t('status.operational')
    case 'degraded': return t('status.degraded')
    case 'down': return t('status.down')
    default: return t('status.unknown')
  }
}

const formattedTime = computed(() => {
  if (!updatedAt.value) return ''
  const d = new Date(updatedAt.value)
  return d.toLocaleTimeString(locale.value === 'zh' ? 'zh-CN' : 'en-US')
})

async function fetchStatus() {
  try {
    if (!groups.value.length) {
      loading.value = true
    }
    error.value = false
    const data = await getGroupStatus()
    groups.value = data.groups || []
    updatedAt.value = data.updated_at
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStatus()
  pollTimer = setInterval(() => {
    // 只在非错误状态下轮询，避免持续失败时浪费请求
    if (!error.value) {
      fetchStatus()
    }
  }, 30000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>
