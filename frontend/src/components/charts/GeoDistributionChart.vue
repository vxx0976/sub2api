<template>
  <div class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
      {{ t('admin.dashboard.geoDistribution') }}
    </h3>
    <div v-if="loading" class="flex h-72 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="distribution.length > 0" class="flex flex-col gap-4 lg:flex-row">
      <div ref="chartRef" class="h-72 w-full lg:w-2/3"></div>
      <div class="max-h-72 w-full overflow-y-auto lg:w-1/3">
        <table class="w-full text-xs">
          <thead>
            <tr class="text-gray-500 dark:text-gray-400">
              <th class="pb-2 text-left">#</th>
              <th class="pb-2 text-left">{{ t('admin.dashboard.country') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.requests') }}</th>
              <th class="pb-2 text-right">%</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(item, index) in topCountries"
              :key="item.country_code"
              class="border-t border-gray-100 dark:border-gray-700"
            >
              <td class="py-1.5 text-gray-400">{{ index + 1 }}</td>
              <td class="py-1.5 font-medium text-gray-900 dark:text-white">
                {{ getCountryFlag(item.country_code) }} {{ getCountryName(item.country_code) }}
              </td>
              <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                {{ item.count.toLocaleString() }}
              </td>
              <td class="py-1.5 text-right text-gray-400 dark:text-gray-500">
                {{ total > 0 ? ((item.count / total) * 100).toFixed(1) : 0 }}%
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else
      class="flex h-72 flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-gray-400"
    >
      <span>{{ t('admin.dashboard.noDataAvailable') }}</span>
      <button
        v-if="!backfilling"
        class="rounded-md bg-blue-600 px-3 py-1.5 text-xs text-white hover:bg-blue-700"
        @click="handleBackfill"
      >
        {{ t('admin.dashboard.geoBackfill') }}
      </button>
      <span v-else class="text-xs text-blue-500">{{ t('admin.dashboard.geoBackfilling') }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import * as echarts from 'echarts/core'
import { MapChart } from 'echarts/charts'
import { VisualMapComponent, TooltipComponent, GeoComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { GeoDistributionItem } from '@/types'
import worldJson from '@/assets/map/world.json'
import { backfillGeoData } from '@/api/admin/dashboard'
import { useAppStore } from '@/stores/app'

echarts.use([MapChart, VisualMapComponent, TooltipComponent, GeoComponent, CanvasRenderer])
echarts.registerMap('world', worldJson as any)

const { t } = useI18n()
const appStore = useAppStore()

const props = defineProps<{
  distribution: GeoDistributionItem[]
  total: number
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'backfillDone'): void
}>()

const chartRef = ref<HTMLElement>()
let chart: echarts.ECharts | null = null
const backfilling = ref(false)

async function handleBackfill() {
  backfilling.value = true
  try {
    const result = await backfillGeoData()
    appStore.showSuccess(
      `${t('admin.dashboard.geoBackfillDone')}: ${result.ips_processed} IPs, ${result.rows_updated} rows`
    )
    emit('backfillDone')
  } catch {
    appStore.showError(t('admin.dashboard.geoBackfillFailed'))
  } finally {
    backfilling.value = false
  }
}

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))

const topCountries = computed(() => props.distribution.slice(0, 15))

// ISO 3166-1 alpha-2 → ECharts map name (common mappings)
const codeToName: Record<string, string> = {
  AF: 'Afghanistan', AL: 'Albania', DZ: 'Algeria', AD: 'Andorra', AO: 'Angola',
  AR: 'Argentina', AM: 'Armenia', AU: 'Australia', AT: 'Austria', AZ: 'Azerbaijan',
  BS: 'Bahamas', BH: 'Bahrain', BD: 'Bangladesh', BB: 'Barbados', BY: 'Belarus',
  BE: 'Belgium', BZ: 'Belize', BJ: 'Benin', BT: 'Bhutan', BO: 'Bolivia',
  BA: 'Bosnia and Herz.', BW: 'Botswana', BR: 'Brazil', BN: 'Brunei', BG: 'Bulgaria',
  BF: 'Burkina Faso', BI: 'Burundi', KH: 'Cambodia', CM: 'Cameroon', CA: 'Canada',
  CF: 'Central African Rep.', TD: 'Chad', CL: 'Chile', CN: 'China', CO: 'Colombia',
  CG: 'Congo', CD: 'Dem. Rep. Congo', CR: 'Costa Rica', CI: "Côte d'Ivoire", HR: 'Croatia',
  CU: 'Cuba', CY: 'Cyprus', CZ: 'Czech Rep.', DK: 'Denmark', DJ: 'Djibouti',
  DO: 'Dominican Rep.', EC: 'Ecuador', EG: 'Egypt', SV: 'El Salvador', GQ: 'Eq. Guinea',
  ER: 'Eritrea', EE: 'Estonia', ET: 'Ethiopia', FJ: 'Fiji', FI: 'Finland',
  FR: 'France', GA: 'Gabon', GM: 'Gambia', GE: 'Georgia', DE: 'Germany',
  GH: 'Ghana', GR: 'Greece', GT: 'Guatemala', GN: 'Guinea', GW: 'Guinea-Bissau',
  GY: 'Guyana', HT: 'Haiti', HN: 'Honduras', HU: 'Hungary', IS: 'Iceland',
  IN: 'India', ID: 'Indonesia', IR: 'Iran', IQ: 'Iraq', IE: 'Ireland',
  IL: 'Israel', IT: 'Italy', JM: 'Jamaica', JP: 'Japan', JO: 'Jordan',
  KZ: 'Kazakhstan', KE: 'Kenya', KP: 'Dem. Rep. Korea', KR: 'Korea', KW: 'Kuwait',
  KG: 'Kyrgyzstan', LA: 'Lao PDR', LV: 'Latvia', LB: 'Lebanon', LS: 'Lesotho',
  LR: 'Liberia', LY: 'Libya', LT: 'Lithuania', LU: 'Luxembourg', MK: 'Macedonia',
  MG: 'Madagascar', MW: 'Malawi', MY: 'Malaysia', ML: 'Mali', MR: 'Mauritania',
  MX: 'Mexico', MD: 'Moldova', MN: 'Mongolia', ME: 'Montenegro', MA: 'Morocco',
  MZ: 'Mozambique', MM: 'Myanmar', NA: 'Namibia', NP: 'Nepal', NL: 'Netherlands',
  NZ: 'New Zealand', NI: 'Nicaragua', NE: 'Niger', NG: 'Nigeria', NO: 'Norway',
  OM: 'Oman', PK: 'Pakistan', PA: 'Panama', PG: 'Papua New Guinea', PY: 'Paraguay',
  PE: 'Peru', PH: 'Philippines', PL: 'Poland', PT: 'Portugal', QA: 'Qatar',
  RO: 'Romania', RU: 'Russia', RW: 'Rwanda', SA: 'Saudi Arabia', SN: 'Senegal',
  RS: 'Serbia', SL: 'Sierra Leone', SG: 'Singapore', SK: 'Slovakia', SI: 'Slovenia',
  SB: 'Solomon Is.', SO: 'Somalia', ZA: 'South Africa', SS: 'S. Sudan', ES: 'Spain',
  LK: 'Sri Lanka', SD: 'Sudan', SR: 'Suriname', SZ: 'Swaziland', SE: 'Sweden',
  CH: 'Switzerland', SY: 'Syria', TW: 'Taiwan', TJ: 'Tajikistan', TZ: 'Tanzania',
  TH: 'Thailand', TL: 'Timor-Leste', TG: 'Togo', TT: 'Trinidad and Tobago', TN: 'Tunisia',
  TR: 'Turkey', TM: 'Turkmenistan', UG: 'Uganda', UA: 'Ukraine', AE: 'United Arab Emirates',
  GB: 'United Kingdom', US: 'United States', UY: 'Uruguay', UZ: 'Uzbekistan',
  VE: 'Venezuela', VN: 'Vietnam', YE: 'Yemen', ZM: 'Zambia', ZW: 'Zimbabwe',
  HK: 'China', MO: 'China', PS: 'Palestine',
}

// Display names for the table (more readable)
const codeToDisplayName: Record<string, string> = {
  US: 'United States', CN: 'China', JP: 'Japan', DE: 'Germany', GB: 'United Kingdom',
  FR: 'France', KR: 'South Korea', IN: 'India', CA: 'Canada', AU: 'Australia',
  BR: 'Brazil', RU: 'Russia', SG: 'Singapore', HK: 'Hong Kong', TW: 'Taiwan',
  NL: 'Netherlands', SE: 'Sweden', CH: 'Switzerland', IE: 'Ireland', IL: 'Israel',
  AE: 'UAE', MO: 'Macao', PH: 'Philippines', TH: 'Thailand', VN: 'Vietnam',
  MY: 'Malaysia', ID: 'Indonesia', MX: 'Mexico', AR: 'Argentina', PL: 'Poland',
  UA: 'Ukraine', CZ: 'Czech Republic', RO: 'Romania', FI: 'Finland', NO: 'Norway',
  DK: 'Denmark', AT: 'Austria', BE: 'Belgium', PT: 'Portugal', ES: 'Spain',
  IT: 'Italy', TR: 'Turkey', ZA: 'South Africa', EG: 'Egypt', NG: 'Nigeria',
  KE: 'Kenya', PK: 'Pakistan', BD: 'Bangladesh', NZ: 'New Zealand', CL: 'Chile',
  CO: 'Colombia', PE: 'Peru',
}

function getCountryName(code: string): string {
  return codeToDisplayName[code] || codeToName[code] || code
}

function getCountryFlag(code: string): string {
  if (code.length !== 2) return ''
  const offset = 0x1F1E6
  return String.fromCodePoint(
    code.charCodeAt(0) - 65 + offset,
    code.charCodeAt(1) - 65 + offset
  )
}

function getMapData() {
  return props.distribution.map((item) => ({
    name: codeToName[item.country_code] || item.country_code,
    value: item.count
  }))
}

function getMaxValue() {
  if (!props.distribution.length) return 100
  return props.distribution[0].count
}

function renderChart() {
  if (!chartRef.value) return
  if (!chart) {
    chart = echarts.init(chartRef.value)
  }

  const dark = isDarkMode.value
  chart.setOption({
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        if (params.value) {
          return `${params.name}: ${params.value.toLocaleString()}`
        }
        return `${params.name}: 0`
      }
    },
    visualMap: {
      min: 0,
      max: getMaxValue(),
      text: ['High', 'Low'],
      realtime: false,
      calculable: true,
      inRange: {
        color: dark
          ? ['#1e3a5f', '#2563eb', '#60a5fa']
          : ['#e0f2fe', '#3b82f6', '#1d4ed8']
      },
      textStyle: { color: dark ? '#e5e7eb' : '#374151', fontSize: 11 },
      left: 'left',
      bottom: 10
    },
    series: [
      {
        type: 'map',
        map: 'world',
        roam: true,
        scaleLimit: { min: 1, max: 8 },
        emphasis: {
          label: { show: true, color: dark ? '#fff' : '#000' },
          itemStyle: { areaColor: '#fbbf24' }
        },
        itemStyle: {
          areaColor: dark ? '#1f2937' : '#f3f4f6',
          borderColor: dark ? '#374151' : '#d1d5db'
        },
        data: getMapData()
      }
    ]
  })
}

function handleResize() {
  chart?.resize()
}

watch(
  () => [props.distribution, props.loading],
  () => {
    if (!props.loading && props.distribution.length > 0) {
      setTimeout(renderChart, 50)
    }
  },
  { deep: true }
)

onMounted(() => {
  window.addEventListener('resize', handleResize)
  if (props.distribution.length > 0) {
    setTimeout(renderChart, 50)
  }
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chart?.dispose()
  chart = null
})
</script>
