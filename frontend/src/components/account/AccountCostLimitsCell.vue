<template>
  <div v-if="hasAnyCostLimit" class="flex flex-col gap-1.5">
    <!-- 每日费用限额 -->
    <div v-if="showDailyCost" class="flex items-center gap-1">
      <span
        :class="[
          'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium',
          dailyCostClass
        ]"
        :title="dailyCostTooltip"
      >
        <svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
        </svg>
        <span class="font-mono">${{ formatCost(currentDailyCost) }}</span>
        <span class="text-gray-400 dark:text-gray-500">/</span>
        <span class="font-mono">${{ formatCost(account.daily_cost_limit) }}</span>
        <span class="text-[9px] opacity-60">/d</span>
      </span>
    </div>

    <!-- 每周费用限额 -->
    <div v-if="showWeeklyCost" class="flex items-center gap-1">
      <span
        :class="[
          'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium',
          weeklyCostClass
        ]"
        :title="weeklyCostTooltip"
      >
        <svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" />
        </svg>
        <span class="font-mono">${{ formatCost(currentWeeklyCost) }}</span>
        <span class="text-gray-400 dark:text-gray-500">/</span>
        <span class="font-mono">${{ formatCost(account.weekly_cost_limit) }}</span>
        <span class="text-[9px] opacity-60">/w</span>
      </span>
    </div>

    <!-- 5h窗口费用限制 -->
    <div v-if="showWindowCost" class="flex items-center gap-1">
      <span
        :class="[
          'inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[10px] font-medium',
          windowCostClass
        ]"
        :title="windowCostTooltip"
      >
        <svg class="h-2.5 w-2.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="font-mono">${{ formatCost(currentWindowCost) }}</span>
        <span class="text-gray-400 dark:text-gray-500">/</span>
        <span class="font-mono">${{ formatCost(account.window_cost_limit) }}</span>
        <span class="text-[9px] opacity-60">/5h</span>
      </span>
    </div>
  </div>
  <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Account } from '@/types'

const props = defineProps<{
  account: Account
}>()

const { t } = useI18n()

const isAnthropic = computed(() => props.account.platform === 'anthropic')

const isAnthropicOAuthOrSetupToken = computed(() => {
  return (
    isAnthropic.value &&
    (props.account.type === 'oauth' || props.account.type === 'setup-token')
  )
})

// ========== 每日费用 ==========

const showDailyCost = computed(() => {
  return (
    isAnthropic.value &&
    props.account.daily_cost_limit !== undefined &&
    props.account.daily_cost_limit !== null &&
    props.account.daily_cost_limit > 0
  )
})

const currentDailyCost = computed(() => props.account.current_daily_cost ?? 0)

const dailyCostClass = computed(() => {
  if (!showDailyCost.value) return ''
  const current = currentDailyCost.value
  const limit = props.account.daily_cost_limit || 0
  if (current >= limit) {
    return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
  }
  if (current >= limit * 0.8) {
    return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
  }
  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
})

const dailyCostTooltip = computed(() => {
  if (!showDailyCost.value) return ''
  const current = currentDailyCost.value
  const limit = props.account.daily_cost_limit || 0
  if (current >= limit) {
    return t('admin.accounts.capacity.dailyCost.blocked')
  }
  return t('admin.accounts.capacity.dailyCost.normal')
})

// ========== 每周费用 ==========

const showWeeklyCost = computed(() => {
  return (
    isAnthropic.value &&
    props.account.weekly_cost_limit !== undefined &&
    props.account.weekly_cost_limit !== null &&
    props.account.weekly_cost_limit > 0
  )
})

const currentWeeklyCost = computed(() => props.account.current_weekly_cost ?? 0)

const weeklyCostClass = computed(() => {
  if (!showWeeklyCost.value) return ''
  const current = currentWeeklyCost.value
  const limit = props.account.weekly_cost_limit || 0
  if (current >= limit) {
    return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
  }
  if (current >= limit * 0.8) {
    return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
  }
  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
})

const weeklyCostTooltip = computed(() => {
  if (!showWeeklyCost.value) return ''
  const current = currentWeeklyCost.value
  const limit = props.account.weekly_cost_limit || 0
  if (current >= limit) {
    return t('admin.accounts.capacity.weeklyCost.blocked')
  }
  return t('admin.accounts.capacity.weeklyCost.normal')
})

// ========== 5h窗口费用 ==========

const showWindowCost = computed(() => {
  return (
    isAnthropicOAuthOrSetupToken.value &&
    props.account.window_cost_limit !== undefined &&
    props.account.window_cost_limit !== null &&
    props.account.window_cost_limit > 0
  )
})

const currentWindowCost = computed(() => props.account.current_window_cost ?? 0)

const windowCostClass = computed(() => {
  if (!showWindowCost.value) return ''
  const current = currentWindowCost.value
  const limit = props.account.window_cost_limit || 0
  const reserve = props.account.window_cost_sticky_reserve || 10
  if (current >= limit + reserve) {
    return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
  }
  if (current >= limit) {
    return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
  }
  if (current >= limit * 0.8) {
    return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
  }
  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
})

const windowCostTooltip = computed(() => {
  if (!showWindowCost.value) return ''
  const current = currentWindowCost.value
  const limit = props.account.window_cost_limit || 0
  const reserve = props.account.window_cost_sticky_reserve || 10
  if (current >= limit + reserve) {
    return t('admin.accounts.capacity.windowCost.blocked')
  }
  if (current >= limit) {
    return t('admin.accounts.capacity.windowCost.stickyOnly')
  }
  return t('admin.accounts.capacity.windowCost.normal')
})

// ========== 通用 ==========

const hasAnyCostLimit = computed(() => showDailyCost.value || showWeeklyCost.value || showWindowCost.value)

const formatCost = (value: number | null | undefined) => {
  if (value === null || value === undefined) return '0'
  return value.toFixed(2)
}
</script>
