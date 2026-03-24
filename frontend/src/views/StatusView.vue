<template>
  <div class="flex min-h-screen flex-col bg-[#FAF9F5] dark:bg-dark-950">
    <PublicHeader />

    <main class="mx-auto w-full max-w-3xl flex-1 px-4 py-8 sm:px-6">
      <!-- Overall Status Banner (only show after data loaded successfully) -->
      <div
        v-if="!loading && !error"
        class="mb-8 rounded-xl border px-5 py-4"
        :class="overallBannerClass"
      >
        <div class="flex items-center gap-3">
          <span class="relative flex h-3 w-3">
            <span
              v-if="overallStatus === 'operational'"
              class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"
            ></span>
            <span class="relative inline-flex h-3 w-3 rounded-full" :class="overallDotClass"></span>
          </span>
          <span class="text-sm font-semibold" :class="overallTextClass">{{ overallMessage }}</span>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="text-center">
          <div class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-gray-300 border-t-primary-500"></div>
          <p class="mt-3 text-sm text-gray-500 dark:text-dark-400">{{ t('status.checking') }}...</p>
        </div>
      </div>

      <!-- Error -->
      <div v-else-if="error" class="flex items-center justify-center py-20">
        <div class="text-center">
          <p class="text-gray-600 dark:text-dark-300">{{ t('status.loadFailed') }}</p>
          <button
            @click="fetchStatus"
            class="mt-3 rounded-lg bg-primary-500 px-4 py-2 text-sm font-medium text-white hover:bg-primary-600"
          >{{ t('status.retry') }}</button>
        </div>
      </div>

      <!-- Group Cards -->
      <div v-else class="space-y-4">
        <div
          v-for="group in groups"
          :key="group.id"
          class="overflow-hidden rounded-xl border border-gray-200/80 bg-white shadow-sm dark:border-dark-700/60 dark:bg-dark-900"
        >
          <!-- Card Header: icon + name + status badge -->
          <div class="flex items-center justify-between px-5 pt-5 pb-3">
            <div class="flex items-center gap-3">
              <PlatformIcon :platform="group.platform" size="lg" class="shrink-0 text-gray-600 dark:text-gray-300" />
              <div class="min-w-0">
                <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ group.name }}</div>
                <div v-if="group.description" class="mt-0.5 text-xs text-gray-500 dark:text-dark-400 line-clamp-1">{{ group.description }}</div>
              </div>
            </div>
            <span
              class="shrink-0 rounded-full px-2.5 py-0.5 text-xs font-semibold"
              :class="statusBadgeClass(group.status)"
            >
              {{ statusLabel(group.status) }}
            </span>
          </div>

          <!-- Availability stats -->
          <div class="flex items-center justify-between border-t border-gray-100 px-5 py-3 dark:border-dark-800">
            <div class="text-xs text-gray-500 dark:text-dark-400">
              {{ t('status.availability7d') }}
              <span v-if="statsMap[group.id]?.total > 0" class="ml-1 text-gray-400 dark:text-dark-500">
                {{ statsMap[group.id].success }}/{{ statsMap[group.id].total }} {{ t('status.successCount') }}
              </span>
            </div>
            <span class="text-sm font-bold" :class="statusTextClass(group.status)">
              {{ statsMap[group.id]?.rate ?? '100.00' }}%
            </span>
          </div>

          <!-- 30-day History Bar -->
          <div class="border-t border-gray-100 px-5 py-3 dark:border-dark-800">
            <div class="mb-1.5 flex items-center justify-between text-[10px] text-gray-400 dark:text-dark-500">
              <span>{{ t('status.uptimeTitle') }}</span>
              <span>{{ nextUpdateText }}</span>
            </div>
            <div class="flex gap-[2px]">
              <template v-if="group.daily_history && group.daily_history.length > 0">
                <div
                  v-for="(day, idx) in group.daily_history"
                  :key="idx"
                  class="h-6 flex-1 rounded-[2px] transition-opacity hover:opacity-80 cursor-default"
                  :class="barClass(day.status)"
                  :title="barTooltip(day)"
                ></div>
              </template>
              <template v-else>
                <div v-for="i in 30" :key="'empty-'+i" class="h-6 flex-1 rounded-[2px] bg-emerald-500"></div>
              </template>
            </div>
            <div class="mt-1 flex items-center justify-between text-[10px] text-gray-400 dark:text-dark-500">
              <span>{{ t('status.daysAgo', { days: 30 }) }}</span>
              <span>{{ t('status.today') }}</span>
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
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import { getGroupStatus, type GroupStatusItem, type DailyStatus } from '@/api/status'

const { t, locale } = useI18n()

const loading = ref(true)
const error = ref(false)
const groups = ref<GroupStatusItem[]>([])
const updatedAt = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null
const countdown = ref(30)
let countdownTimer: ReturnType<typeof setInterval> | null = null

// === Overall status ===

const overallStatus = computed(() => {
  if (groups.value.length === 0) return 'operational'
  const hasDown = groups.value.some(g => g.status === 'down')
  if (hasDown) return 'down'
  const hasDegraded = groups.value.some(g => g.status === 'degraded')
  if (hasDegraded) return 'degraded'
  return 'operational'
})

const overallMessage = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return t('status.allOperational')
    case 'degraded': return t('status.partialOutage')
    case 'down': return t('status.majorOutage')
    default: return t('status.allOperational')
  }
})

const overallBannerClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return 'border-emerald-200 bg-emerald-50 dark:border-emerald-900/40 dark:bg-emerald-950/30'
    case 'degraded': return 'border-amber-200 bg-amber-50 dark:border-amber-900/40 dark:bg-amber-950/30'
    case 'down': return 'border-red-200 bg-red-50 dark:border-red-900/40 dark:bg-red-950/30'
    default: return 'border-emerald-200 bg-emerald-50 dark:border-emerald-900/40 dark:bg-emerald-950/30'
  }
})

const overallDotClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return 'bg-emerald-500'
    case 'degraded': return 'bg-amber-500'
    case 'down': return 'bg-red-500'
    default: return 'bg-emerald-500'
  }
})

const overallTextClass = computed(() => {
  switch (overallStatus.value) {
    case 'operational': return 'text-emerald-700 dark:text-emerald-400'
    case 'degraded': return 'text-amber-700 dark:text-amber-400'
    case 'down': return 'text-red-700 dark:text-red-400'
    default: return 'text-emerald-700 dark:text-emerald-400'
  }
})

// === Per-group helpers ===

function statusTextClass(status: string) {
  switch (status) {
    case 'operational': return 'text-emerald-600 dark:text-emerald-400'
    case 'degraded': return 'text-amber-600 dark:text-amber-400'
    case 'down': return 'text-red-600 dark:text-red-400'
    default: return 'text-emerald-600 dark:text-emerald-400'
  }
}

function statusBadgeClass(status: string) {
  switch (status) {
    case 'operational': return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'degraded': return 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    case 'down': return 'bg-red-50 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default: return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'operational': return t('status.operational')
    case 'degraded': return t('status.degraded')
    case 'down': return t('status.down')
    default: return t('status.operational')
  }
}

function barClass(status: string) {
  switch (status) {
    case 'operational': return 'bg-emerald-500'
    case 'degraded': return 'bg-amber-400'
    case 'down': return 'bg-red-500'
    default: return 'bg-emerald-500'
  }
}

function barTooltip(day: DailyStatus) {
  const date = formatDate(day.date)
  if (day.total === 0) return `${date}: ${t('status.noData')}`
  return `${date}: ${day.rate.toFixed(1)}% (${day.total} ${t('status.requests')})`
}

function formatDate(yyyymmdd: string) {
  if (!yyyymmdd || yyyymmdd.length !== 8) return yyyymmdd
  const y = yyyymmdd.slice(0, 4)
  const m = yyyymmdd.slice(4, 6)
  const d = yyyymmdd.slice(6, 8)
  return `${y}-${m}-${d}`
}

function calc7dStats(group: GroupStatusItem) {
  if (!group.daily_history || group.daily_history.length === 0) {
    return { success: 0, total: 0, rate: '100.00' }
  }
  const last7 = group.daily_history.slice(-7)
  let success = 0
  let total = 0
  for (const day of last7) {
    total += day.total
    success += Math.round(day.total * day.rate / 100)
  }
  const rate = total > 0 ? (success / total * 100).toFixed(2) : '100.00'
  return { success, total, rate }
}

const statsMap = computed(() => {
  const map: Record<number, { success: number; total: number; rate: string }> = {}
  for (const g of groups.value) {
    map[g.id] = calc7dStats(g)
  }
  return map
})

const nextUpdateText = computed(() => {
  return t('status.nextUpdate', { seconds: countdown.value })
})

const formattedTime = computed(() => {
  if (!updatedAt.value) return ''
  const d = new Date(updatedAt.value)
  return d.toLocaleTimeString(locale.value === 'zh' ? 'zh-CN' : 'en-US')
})

// === Data fetching ===

async function fetchStatus() {
  try {
    if (!groups.value.length) {
      loading.value = true
    }
    error.value = false
    const data = await getGroupStatus()
    groups.value = data.groups || []
    updatedAt.value = data.updated_at
    countdown.value = 30
  } catch {
    error.value = true
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchStatus()
  pollTimer = setInterval(() => {
    if (!error.value) fetchStatus()
  }, 30000)
  countdownTimer = setInterval(() => {
    if (countdown.value > 0) countdown.value--
  }, 1000)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (countdownTimer) clearInterval(countdownTimer)
})
</script>
