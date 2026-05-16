<template>
  <div class="card p-4">
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
        {{ t('admin.dashboard.financeTrendTitle') }}
      </h3>
      <div class="text-xs text-gray-500 dark:text-gray-400">
        <span>{{ t('admin.dashboard.rangeNet') }}:</span>
        <span
          class="ml-1 font-semibold"
          :class="net >= 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
        >
          {{ net >= 0 ? '+' : '' }}${{ formatCost(net) }}
        </span>
      </div>
    </div>
    <div v-if="loading" class="flex h-64 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="trendData.length > 0 && chartData" class="h-64">
      <Line :data="chartData" :options="lineOptions" />
    </div>
    <div
      v-else
      class="flex h-64 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
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
  consumption: '#ef4444'
}))

const net = computed(() =>
  props.trendData.reduce((acc, p) => acc + (p.recharge - p.consumption), 0)
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
          return `${t('admin.dashboard.net')}: ${sign}$${formatCost(delta)}`
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
