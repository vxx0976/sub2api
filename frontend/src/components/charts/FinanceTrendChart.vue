<template>
  <div class="card flex h-full flex-col p-4">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
        {{ t('admin.dashboard.financeTrendTitle') }}
      </h3>
      <div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
        <div>
          <span>{{ t('admin.dashboard.rangeNet') }}:</span>
          <span
            class="ml-1 font-semibold"
            :class="net >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
          >
            {{ net >= 0 ? '+' : '' }}${{ formatCost(net) }}
          </span>
        </div>
        <div>
          <span>{{ t('admin.dashboard.rangeMargin') }}:</span>
          <span
            class="ml-1 font-semibold"
            :class="rangeMargin >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
          >
            {{ rangeMargin >= 0 ? '+' : '' }}${{ formatCost(rangeMargin) }}
          </span>
        </div>
      </div>
    </div>
    <div v-if="loading" class="flex min-h-64 flex-1 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="trendData.length > 0 && chartData" class="min-h-64 flex-1">
      <Line :data="chartData" :options="lineOptions" />
    </div>
    <div
      v-else
      class="flex min-h-64 flex-1 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { FinanceTrendPoint } from '@/api/admin/dashboard'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()

const props = defineProps<{
  trendData: FinanceTrendPoint[]
  loading?: boolean
}>()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb',
  recharge: '#10b981',
  consumption: '#ef4444',
  accountCost: '#f59e0b'
}))

const net = computed(() =>
  props.trendData.reduce((acc, p) => acc + (p.recharge - p.consumption), 0)
)

// 毛利 = 用户消耗 − 账号成本
const rangeMargin = computed(() =>
  props.trendData.reduce((acc, p) => acc + (p.consumption - p.account_cost), 0)
)

const chartData = computed(() => {
  if (!props.trendData?.length) return null
  return {
    labels: props.trendData.map((d) => d.date),
    datasets: [
      {
        label: t('admin.dashboard.recharge'),
        data: props.trendData.map((d) => d.recharge),
        borderColor: chartColors.value.recharge,
        backgroundColor: `${chartColors.value.recharge}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: t('admin.dashboard.consumption'),
        data: props.trendData.map((d) => d.consumption),
        borderColor: chartColors.value.consumption,
        backgroundColor: `${chartColors.value.consumption}20`,
        fill: true,
        tension: 0.3
      },
      {
        label: t('admin.dashboard.accountCostLine'),
        data: props.trendData.map((d) => d.account_cost),
        borderColor: chartColors.value.accountCost,
        backgroundColor: `${chartColors.value.accountCost}20`,
        borderDash: [4, 4],
        fill: false,
        tension: 0.3
      }
    ]
  }
})

const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: { size: 11 }
      }
    },
    tooltip: {
      callbacks: {
        label: (context: any) => `${context.dataset.label}: $${formatCost(context.raw)}`,
        footer: (items: any) => {
          const idx = items[0]?.dataIndex
          if (idx === undefined || !props.trendData[idx]) return ''
          const p = props.trendData[idx]
          const delta = p.recharge - p.consumption
          const sign = delta >= 0 ? '+' : ''
          const margin = p.consumption - p.account_cost
          const marginSign = margin >= 0 ? '+' : ''
          return [
            `${t('admin.dashboard.net')}: ${sign}$${formatCost(delta)}`,
            `${t('admin.dashboard.margin')}: ${marginSign}$${formatCost(margin)}`
          ]
        }
      }
    }
  },
  scales: {
    x: {
      grid: { color: chartColors.value.grid },
      ticks: { color: chartColors.value.text, font: { size: 10 } }
    },
    y: {
      grid: { color: chartColors.value.grid },
      ticks: {
        color: chartColors.value.text,
        font: { size: 10 },
        callback: (value: string | number) => `$${formatCost(Number(value))}`
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
