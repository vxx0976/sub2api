<template>
  <div
    class="group relative flex flex-col overflow-hidden rounded-3xl border transition-all duration-300"
    :class="[
      'border-yellow-200 bg-gradient-to-b from-yellow-50 to-white hover:border-yellow-300 hover:shadow-lg dark:from-yellow-900/10 dark:to-dark-800 dark:border-yellow-500/20 dark:hover:border-yellow-500/30'
    ]"
  >
    <!-- Card Content -->
    <div class="flex flex-1 flex-col p-6">
      <!-- Title -->
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          💰 {{ t('plans.paygo.title') }}
        </h3>
      </div>

      <!-- Description -->
      <p class="mb-4 text-sm text-gray-600 dark:text-gray-400">
        {{ t('plans.paygo.description') }}
      </p>

      <!-- Current Balance -->
      <div v-if="currentBalance !== null" class="mb-4 rounded-xl bg-primary-50 p-3 dark:bg-primary-900/20">
        <div class="mb-1 text-xs text-gray-600 dark:text-gray-400">
          {{ t('recharge.currentBalance') }}
        </div>
        <div class="text-2xl font-bold text-primary-600 dark:text-primary-400">
          ¥{{ currentBalance.toFixed(2) }}
        </div>
      </div>

      <!-- Bonus Tip -->
      <div v-if="config && config.enabled && hasBonusTier" class="mb-4 rounded-xl bg-yellow-50 p-3 dark:bg-yellow-900/20">
        <div class="flex items-start">
          <Icon name="gift" size="sm" class="mr-2 mt-0.5 text-yellow-600 dark:text-yellow-400" />
          <div class="text-xs text-yellow-700 dark:text-yellow-300">
            {{ t('recharge.bonusTip') }}
          </div>
        </div>
      </div>

      <!-- Features -->
      <ul class="mb-6 flex-1 space-y-3">
        <li v-for="(feature, index) in features" :key="index" class="flex items-start gap-3 text-sm">
          <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
            <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
          </div>
          <span class="text-gray-600 dark:text-gray-300">{{ feature }}</span>
        </li>
      </ul>

      <!-- Recharge Button -->
      <button
        @click="$emit('recharge')"
        :disabled="!config?.enabled"
        class="w-full rounded-xl px-4 py-3 text-sm font-semibold transition-all duration-200"
        :class="[
          config?.enabled
            ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/30 hover:bg-primary-600 hover:shadow-xl hover:shadow-primary-500/40'
            : 'bg-gray-300 text-gray-500 cursor-not-allowed dark:bg-dark-600 dark:text-dark-400'
        ]"
      >
        <span class="flex items-center justify-center gap-2">
          <Icon name="plus" size="sm" />
          {{ t('recharge.rechargeNow') }}
        </span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { RechargeConfig } from '@/api/recharge'

interface Props {
  config: RechargeConfig | null
  currentBalance: number | null
}

const props = defineProps<Props>()
defineEmits<{
  (e: 'recharge'): void
}>()

const { t } = useI18n()

const features = computed(() => [
  t('plans.paygo.features.anyAmount'),
  t('plans.paygo.features.payAsYouGo'),
  t('plans.paygo.features.neverExpires')
])

// 是否有充值优惠（倍率 > 1.0 的阶梯）
const hasBonusTier = computed(() => {
  if (!props.config || !props.config.tiers) return false
  return props.config.tiers.some(tier => tier.multiplier > 1.0)
})
</script>
