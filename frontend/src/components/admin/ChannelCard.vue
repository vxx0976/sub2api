<template>
  <div
    class="group relative flex flex-col overflow-hidden rounded-3xl border transition-all duration-300"
    :class="cardClass"
  >
    <!-- Recommended-like Badge for Active -->
    <div v-if="channel.status === 'active'" class="absolute right-4 top-4">
      <span class="inline-flex items-center gap-1 rounded-full bg-emerald-500 px-2.5 py-1 text-xs font-semibold text-white shadow-lg shadow-emerald-500/30">
        <span class="h-1.5 w-1.5 rounded-full bg-white animate-pulse" />
        {{ t('admin.channels.status.active') }}
      </span>
    </div>
    <div v-else-if="channel.status === 'error'" class="absolute right-4 top-4">
      <span class="inline-flex items-center gap-1 rounded-full bg-red-500 px-2.5 py-1 text-xs font-semibold text-white shadow-lg shadow-red-500/30">
        <Icon name="exclamationCircle" size="xs" />
        {{ t('admin.channels.status.error') }}
      </span>
    </div>
    <div v-else class="absolute right-4 top-4">
      <span class="inline-flex items-center gap-1 rounded-full bg-gray-400 px-2.5 py-1 text-xs font-semibold text-white">
        {{ t('admin.channels.status.inactive') }}
      </span>
    </div>

    <!-- Card Content -->
    <div class="flex flex-1 flex-col p-6">
      <!-- Header: Icon + Name -->
      <div class="mb-4 flex items-start gap-4">
        <!-- Platform Icon -->
        <div
          class="flex h-14 w-14 flex-shrink-0 items-center justify-center rounded-2xl shadow-lg"
          :class="iconBgClass"
        >
          <img
            v-if="channel.icon_url && !imageError"
            :src="channel.icon_url"
            :alt="channel.name"
            class="h-9 w-9 rounded-lg object-contain"
            @error="handleImageError"
          />
          <span v-else class="text-2xl font-bold text-white drop-shadow">
            {{ channel.name.charAt(0).toUpperCase() }}
          </span>
        </div>

        <!-- Name & Description -->
        <div class="min-w-0 flex-1 pr-16">
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ channel.name }}
          </h3>
          <p
            v-if="channel.description"
            class="mt-1 text-sm text-gray-500 dark:text-gray-400 line-clamp-2"
          >
            {{ channel.description }}
          </p>
          <div v-if="channel.platform" class="mt-2">
            <span
              class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium"
              :class="platformBadgeClass"
            >
              {{ channel.platform }}
            </span>
          </div>
        </div>
      </div>

      <!-- Balance Display - Main Feature -->
      <div class="mb-6">
        <div class="flex items-end gap-1">
          <span class="text-lg font-medium text-gray-500 dark:text-gray-400">
            {{ channel.balance_unit }}
          </span>
          <span class="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">
            {{ formatBalanceInteger(channel.cached_balance) }}
          </span>
          <span class="mb-1 text-xl font-semibold text-gray-500 dark:text-gray-400">
            .{{ formatBalanceDecimal(channel.cached_balance) }}
          </span>
        </div>
        <div class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          <span v-if="channel.last_check_at">
            {{ t('admin.channels.lastUpdated') }}: {{ formatTime(channel.last_check_at) }}
          </span>
          <span v-else>{{ t('admin.channels.balance') }}</span>
        </div>
      </div>

      <!-- Error Message -->
      <div
        v-if="channel.last_error"
        class="mb-4 flex items-start gap-2 rounded-xl bg-red-50 p-3 dark:bg-red-900/20"
      >
        <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30">
          <Icon name="exclamationCircle" size="xs" class="text-red-600 dark:text-red-400" />
        </div>
        <p class="line-clamp-2 text-xs text-red-600 dark:text-red-400">
          {{ channel.last_error }}
        </p>
      </div>

      <!-- Spacer -->
      <div class="flex-1" />

      <!-- Actions -->
      <div class="flex items-center justify-between gap-2 border-t border-gray-100 pt-4 dark:border-dark-700">
        <!-- Left: Website link -->
        <div>
          <a
            v-if="channel.website_url"
            :href="channel.website_url"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-1.5 rounded-xl bg-primary-50 px-3 py-2 text-sm font-medium text-primary-600 transition-colors hover:bg-primary-100 dark:bg-primary-900/20 dark:text-primary-400 dark:hover:bg-primary-900/30"
          >
            <Icon name="externalLink" size="sm" />
            {{ t('admin.channels.recharge') }}
          </a>
        </div>
        <!-- Right: Actions -->
        <div class="flex items-center gap-1">
          <button
            @click="$emit('refresh', channel)"
            :disabled="refreshing"
            class="inline-flex items-center justify-center rounded-xl p-2.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 disabled:opacity-50 dark:text-gray-400 dark:hover:bg-dark-600 dark:hover:text-gray-200"
            :title="t('admin.channels.refreshBalance')"
          >
            <Icon name="refresh" size="sm" :class="{ 'animate-spin': refreshing }" />
          </button>
          <button
            @click="$emit('edit', channel)"
            class="inline-flex items-center justify-center rounded-xl p-2.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-700 dark:text-gray-400 dark:hover:bg-dark-600 dark:hover:text-gray-200"
            :title="t('common.edit')"
          >
            <Icon name="edit" size="sm" />
          </button>
          <button
            @click="$emit('delete', channel)"
            class="inline-flex items-center justify-center rounded-xl p-2.5 text-red-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:text-red-400 dark:hover:bg-red-900/20 dark:hover:text-red-300"
            :title="t('common.delete')"
          >
            <Icon name="trash" size="sm" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Channel } from '@/types'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const props = defineProps<{
  channel: Channel
  refreshing?: boolean
}>()

defineEmits<{
  (e: 'refresh', channel: Channel): void
  (e: 'edit', channel: Channel): void
  (e: 'delete', channel: Channel): void
}>()

const imageError = ref(false)

const handleImageError = () => {
  imageError.value = true
}

const cardClass = computed(() => {
  if (props.channel.status === 'error') {
    return 'border-red-300/50 bg-gradient-to-b from-red-50 to-white shadow-lg shadow-red-500/10 dark:border-red-500/30 dark:from-red-900/20 dark:to-dark-800'
  }
  if (props.channel.status === 'inactive') {
    return 'border-gray-200 bg-gradient-to-b from-gray-50 to-white opacity-60 dark:border-dark-600 dark:from-dark-700/50 dark:to-dark-800'
  }
  // Active status - primary color scheme like recommended plan
  return 'border-emerald-300/50 bg-gradient-to-b from-emerald-50 to-white shadow-lg shadow-emerald-500/10 hover:shadow-xl dark:border-emerald-500/30 dark:from-emerald-900/20 dark:to-dark-800'
})

const iconBgClass = computed(() => {
  const platform = props.channel.platform
  switch (platform) {
    case 'anthropic':
      return 'bg-gradient-to-br from-orange-400 to-orange-600 shadow-orange-500/30'
    case 'openai':
      return 'bg-gradient-to-br from-emerald-400 to-emerald-600 shadow-emerald-500/30'
    case 'gemini':
      return 'bg-gradient-to-br from-blue-400 to-blue-600 shadow-blue-500/30'
    default:
      return 'bg-gradient-to-br from-gray-500 to-gray-700 shadow-gray-500/30'
  }
})

const platformBadgeClass = computed(() => {
  const platform = props.channel.platform
  switch (platform) {
    case 'anthropic':
      return 'bg-gradient-to-r from-orange-100 to-amber-100 text-orange-700 dark:from-orange-900/30 dark:to-amber-900/30 dark:text-orange-400'
    case 'openai':
      return 'bg-gradient-to-r from-emerald-100 to-green-100 text-emerald-700 dark:from-emerald-900/30 dark:to-green-900/30 dark:text-emerald-400'
    case 'gemini':
      return 'bg-gradient-to-r from-blue-100 to-indigo-100 text-blue-700 dark:from-blue-900/30 dark:to-indigo-900/30 dark:text-blue-400'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300'
  }
})

// Round to 2 decimal places first, then format
function getRoundedBalance(balance: number | null | undefined): number | null {
  if (balance === null || balance === undefined) {
    return null
  }
  return Math.round(balance * 100) / 100
}

function formatBalanceInteger(balance: number | null | undefined): string {
  const rounded = getRoundedBalance(balance)
  if (rounded === null) {
    return '--'
  }
  return Math.floor(rounded).toLocaleString('en-US')
}

function formatBalanceDecimal(balance: number | null | undefined): string {
  const rounded = getRoundedBalance(balance)
  if (rounded === null) {
    return '00'
  }
  const decimal = rounded % 1
  return Math.round(decimal * 100).toString().padStart(2, '0')
}

function formatTime(dateStr: string | null | undefined): string {
  if (!dateStr) return '--'
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)

  if (diffMins < 1) return t('common.time.justNow')
  if (diffMins < 60) return t('common.time.minutesAgo', { n: diffMins })
  const diffHours = Math.floor(diffMins / 60)
  if (diffHours < 24) return t('common.time.hoursAgo', { n: diffHours })
  const diffDays = Math.floor(diffHours / 24)
  return t('common.time.daysAgo', { n: diffDays })
}
</script>
