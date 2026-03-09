<template>
  <div v-if="hasAnyCostLimit" class="flex flex-col gap-1.5">
    <!-- API Key 日/周/总配额（quota_daily_limit / quota_weekly_limit / quota_limit） -->
    <div v-if="showDailyQuota" class="flex items-center gap-1">
      <QuotaBadge :used="account.quota_daily_used ?? 0" :limit="account.quota_daily_limit!" label="D" />
    </div>
    <div v-if="showWeeklyQuota" class="flex items-center gap-1">
      <QuotaBadge :used="account.quota_weekly_used ?? 0" :limit="account.quota_weekly_limit!" label="W" />
    </div>
    <div v-if="showTotalQuota" class="flex items-center gap-1">
      <QuotaBadge :used="account.quota_used ?? 0" :limit="account.quota_limit!" />
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
import QuotaBadge from './QuotaBadge.vue'

const props = defineProps<{
  account: Account
}>()

const { t } = useI18n()

const isAnthropicOAuthOrSetupToken = computed(() =>
  props.account.platform === 'anthropic' &&
  (props.account.type === 'oauth' || props.account.type === 'setup-token')
)

// ========== API Key 配额（日/周/总） ==========

const showDailyQuota = computed(() =>
  props.account.type === 'apikey' && (props.account.quota_daily_limit ?? 0) > 0
)

const showWeeklyQuota = computed(() =>
  props.account.type === 'apikey' && (props.account.quota_weekly_limit ?? 0) > 0
)

const showTotalQuota = computed(() =>
  props.account.type === 'apikey' && (props.account.quota_limit ?? 0) > 0
)

// ========== 5h窗口费用 ==========

const showWindowCost = computed(() =>
  isAnthropicOAuthOrSetupToken.value &&
  (props.account.window_cost_limit ?? 0) > 0
)

const currentWindowCost = computed(() => props.account.current_window_cost ?? 0)

const windowCostClass = computed(() => {
  if (!showWindowCost.value) return ''
  const current = currentWindowCost.value
  const limit = props.account.window_cost_limit || 0
  const reserve = props.account.window_cost_sticky_reserve || 10
  if (current >= limit + reserve) return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
  if (current >= limit) return 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400'
  if (current >= limit * 0.8) return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
  return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
})

const windowCostTooltip = computed(() => {
  if (!showWindowCost.value) return ''
  const current = currentWindowCost.value
  const limit = props.account.window_cost_limit || 0
  const reserve = props.account.window_cost_sticky_reserve || 10
  if (current >= limit + reserve) return t('admin.accounts.capacity.windowCost.blocked')
  if (current >= limit) return t('admin.accounts.capacity.windowCost.stickyOnly')
  return t('admin.accounts.capacity.windowCost.normal')
})

// ========== 通用 ==========

const hasAnyCostLimit = computed(() =>
  showDailyQuota.value || showWeeklyQuota.value || showTotalQuota.value || showWindowCost.value
)

const formatCost = (value: number | null | undefined) => {
  if (value === null || value === undefined) return '0'
  return value.toFixed(2)
}
</script>
