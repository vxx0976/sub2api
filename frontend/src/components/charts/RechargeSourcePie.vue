<template>
  <div class="card flex h-full flex-col p-4">
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
        {{ t('admin.dashboard.rechargeSourceTitle') }}
      </h3>
      <div class="text-xs text-gray-500 dark:text-gray-400">
        <span>{{ t('admin.dashboard.rangeTotal') }}:</span>
        <span class="ml-1 font-semibold text-gray-900 dark:text-white">
          ${{ formatCost(total) }}
        </span>
      </div>
    </div>

    <div v-if="loading" class="flex min-h-64 flex-1 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="total <= 0"
      class="flex min-h-64 flex-1 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
    <div
      v-else
      class="flex flex-1 flex-col items-center justify-center gap-4 lg:flex-row lg:gap-6"
    >
      <div class="h-56 w-56 flex-shrink-0">
        <Doughnut :data="chartData" :options="chartOptions" />
      </div>
      <div class="w-full flex-1 space-y-2">
        <div
          v-for="(item, idx) in legendItems"
          :key="item.key"
          class="flex items-center justify-between text-xs"
        >
          <div class="flex items-center gap-2">
            <span
              class="h-3 w-3 rounded-sm"
              :style="{ backgroundColor: colors[idx] }"
            ></span>
            <span class="text-gray-700 dark:text-gray-300">{{ item.label }}</span>
          </div>
          <div class="flex items-baseline gap-2">
            <span class="font-medium text-gray-900 dark:text-white"
              >${{ formatCost(item.value) }}</span
            >
            <span class="w-12 text-right text-gray-500 dark:text-gray-400"
              >{{ item.pct.toFixed(1) }}%</span
            >
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { RechargeBreakdown } from '@/api/admin/dashboard'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()

const props = defineProps<{
  breakdown: RechargeBreakdown | null
  loading?: boolean
}>()

// 固定 slice 的顺序与配色（支付宝蓝 / 微信绿 / USDT 青 / 兑换码橙 / 手工紫）
const colors = ['#3b82f6', '#10b981', '#14b8a6', '#f59e0b', '#a855f7']

const slices = computed(() => [
  { key: 'alipay', value: props.breakdown?.alipay ?? 0, label: t('admin.dashboard.srcAlipay') },
  { key: 'wxpay', value: props.breakdown?.wxpay ?? 0, label: t('admin.dashboard.srcWxpay') },
  { key: 'usdt', value: props.breakdown?.usdt ?? 0, label: t('admin.dashboard.srcUsdt') },
  { key: 'redeem_code', value: props.breakdown?.redeem_code ?? 0, label: t('admin.dashboard.srcRedeem') },
  { key: 'admin_manual', value: props.breakdown?.admin_manual ?? 0, label: t('admin.dashboard.srcAdminManual') }
])

const total = computed(() => slices.value.reduce((acc, s) => acc + s.value, 0))

const legendItems = computed(() =>
  slices.value.map((s) => ({
    ...s,
    pct: total.value > 0 ? (s.value / total.value) * 100 : 0
  }))
)

const chartData = computed(() => ({
  labels: slices.value.map((s) => s.label),
  datasets: [
    {
      data: slices.value.map((s) => s.value),
      backgroundColor: colors,
      borderColor: colors.map(() => 'transparent'),
      borderWidth: 0
    }
  ]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '55%',
  plugins: {
    legend: { display: false },
    tooltip: {
      callbacks: {
        label: (ctx: any) => {
          const v = Number(ctx.raw) || 0
          const pct = total.value > 0 ? (v / total.value) * 100 : 0
          return `${ctx.label}: $${formatCost(v)} (${pct.toFixed(1)}%)`
        }
      }
    }
  }
}))

const formatCost = (value: number): string => {
  const abs = Math.abs(value)
  if (abs >= 1000) return (value / 1000).toFixed(2) + 'K'
  if (abs >= 1) return value.toFixed(2)
  if (abs >= 0.01) return value.toFixed(3)
  return value.toFixed(4)
}
</script>
