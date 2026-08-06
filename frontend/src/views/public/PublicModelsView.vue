<template>
  <div class="min-h-screen bg-gray-50 text-gray-900 dark:bg-dark-950 dark:text-white">
    <PublicHeader />

    <main class="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:py-10">
      <!-- Hero banner with live stats -->
      <section class="relative mb-6 overflow-hidden rounded-2xl bg-gradient-to-br from-primary-600 via-primary-600 to-primary-800 px-6 py-7 text-white shadow-sm sm:mb-8 sm:px-9 sm:py-9">
        <div aria-hidden="true" class="pointer-events-none absolute -right-16 -top-24 h-64 w-64 rounded-full bg-white/10 blur-2xl"></div>
        <div aria-hidden="true" class="pointer-events-none absolute -bottom-24 left-1/3 h-56 w-56 rounded-full bg-primary-300/20 blur-3xl"></div>
        <div class="relative flex flex-col gap-7 sm:flex-row sm:items-center sm:justify-between">
          <div class="min-w-0">
            <span class="inline-flex items-center gap-1.5 rounded-full bg-white/15 px-3 py-1 text-xs font-medium ring-1 ring-white/20 backdrop-blur">
              <Icon name="grid" size="xs" />
              {{ t('publicModels.badge') }}
            </span>
            <h1 class="mt-3 break-words text-2xl font-bold tracking-tight sm:text-4xl">
              {{ t('publicModels.title') }}
            </h1>
            <p class="mt-2.5 max-w-xl text-sm text-white/80">
              {{ t('publicModels.subtitle') }}
            </p>
          </div>
          <div class="flex shrink-0 items-center gap-5 sm:gap-9">
            <div class="text-center">
              <div class="text-3xl font-bold tabular-nums sm:text-4xl">{{ groups.length }}</div>
              <p class="mt-1 text-xs text-white/70">{{ t('publicModels.statGroups') }}</p>
            </div>
            <div class="h-10 w-px bg-white/20"></div>
            <div class="text-center">
              <div class="text-3xl font-bold tabular-nums sm:text-4xl">{{ totalModelCount }}</div>
              <p class="mt-1 text-xs text-white/70">{{ t('publicModels.statModels') }}</p>
            </div>
            <div class="h-10 w-px bg-white/20"></div>
            <div class="text-center">
              <div class="text-3xl font-bold tabular-nums sm:text-4xl">{{ platformOptions.length }}</div>
              <p class="mt-1 text-xs text-white/70">{{ t('publicModels.statPlatforms') }}</p>
            </div>
          </div>
        </div>
      </section>

      <!-- API base URL -->
      <div
        v-if="apiBaseUrl"
        class="mb-6 flex flex-col gap-3 rounded-xl border border-gray-200 bg-white p-4 sm:flex-row sm:items-center sm:justify-between sm:gap-4 dark:border-dark-700 dark:bg-dark-800/40"
      >
        <div class="flex items-start gap-3">
          <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary-700 dark:bg-primary-500/10 dark:text-primary-300">
            <Icon name="link" size="md" />
          </span>
          <div class="min-w-0">
            <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('publicModels.apiBaseTitle') }}</p>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-dark-400">{{ t('publicModels.apiBaseHint') }}</p>
          </div>
        </div>
        <button
          type="button"
          class="inline-flex min-w-0 items-center gap-3 self-start rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 transition hover:border-primary-300 sm:self-auto dark:border-dark-700 dark:bg-dark-900/60 dark:hover:border-primary-500/40"
          :title="t('publicModels.copyApiBase')"
          @click="copyApiBase"
        >
          <code class="truncate font-mono text-sm text-gray-800 dark:text-dark-100">{{ apiBaseUrl }}/v1</code>
          <span class="flex-shrink-0 text-xs font-medium text-primary-600 dark:text-primary-300">{{ t('publicModels.copyApiBase') }}</span>
        </button>
      </div>

      <!-- Controls: search + price mode + refresh -->
      <div class="mb-4 flex flex-col gap-3 sm:flex-row sm:items-center">
        <div class="relative min-w-0 flex-1">
          <Icon name="search" size="sm" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-dark-400" />
          <input
            v-model="searchQuery"
            type="search"
            :placeholder="t('publicModels.searchPlaceholder')"
            class="w-full rounded-lg border border-gray-200 bg-white py-2.5 pl-9 pr-3 text-sm text-gray-900 placeholder:text-gray-400 transition focus:border-primary-400 focus:outline-none focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-800/60 dark:text-white dark:placeholder:text-dark-400"
          />
        </div>
        <div class="flex items-center gap-2">
          <!-- Official / Site price toggle -->
          <div class="inline-flex rounded-lg border border-gray-200 bg-white p-0.5 text-xs dark:border-dark-700 dark:bg-dark-800">
            <button
              type="button"
              class="rounded-md px-3 py-1.5 font-medium transition"
              :class="priceMode === 'site' ? 'bg-primary-500 text-white shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
              @click="priceMode = 'site'"
            >{{ t('publicModels.sitePrice') }}</button>
            <button
              type="button"
              class="rounded-md px-3 py-1.5 font-medium transition"
              :class="priceMode === 'official' ? 'bg-primary-500 text-white shadow-sm' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
              @click="priceMode = 'official'"
            >{{ t('publicModels.officialPrice') }}</button>
          </div>
          <button
            type="button"
            class="inline-flex flex-shrink-0 items-center gap-1.5 rounded-lg border border-gray-200 bg-white px-3 py-2.5 text-sm font-medium text-gray-700 transition hover:border-primary-300 hover:text-primary-700 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-700 dark:bg-dark-800/40 dark:text-dark-200 dark:hover:border-primary-500/40"
            :disabled="loading"
            @click="reload"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            <span class="hidden sm:inline">{{ t('publicModels.refresh') }}</span>
          </button>
        </div>
      </div>

      <!-- Platform filter + fx note -->
      <div class="mb-6 flex flex-wrap items-center gap-2">
        <button
          type="button"
          class="platform-chip"
          :class="platformFilter === '' ? 'platform-chip-active' : ''"
          @click="platformFilter = ''"
        >
          <Icon name="grid" size="xs" class="mr-1" />
          {{ t('publicModels.filterAll') }}
          <span class="ml-1 text-[10px] opacity-70">{{ groups.length }}</span>
        </button>
        <button
          v-for="p in platformOptions"
          :key="p.name"
          type="button"
          class="platform-chip"
          :class="platformFilter === p.name ? 'platform-chip-active' : ''"
          @click="platformFilter = p.name"
        >
          <PlatformIcon :platform="(p.name as GroupPlatform)" size="xs" class="mr-1" />
          {{ p.name }}
          <span class="ml-1 text-[10px] opacity-70">{{ p.count }}</span>
        </button>
        <span v-if="showFxNote" class="ml-auto text-xs text-gray-400 dark:text-dark-400">
          {{ t('publicModels.fxNote', { rate: fxRate.toFixed(2) }) }}
        </span>
      </div>

      <!-- Loading skeleton -->
      <div v-if="loading && !groups.length" class="space-y-4">
        <div v-for="i in 4" :key="i" class="h-40 animate-pulse rounded-xl bg-white dark:bg-dark-800/40"></div>
      </div>

      <!-- Error -->
      <div
        v-else-if="loadError"
        class="rounded-lg border border-red-200 bg-red-50 p-6 text-red-700 dark:border-red-500/30 dark:bg-red-500/10 dark:text-red-200"
      >
        <h2 class="text-base font-semibold">{{ t('publicModels.loadErrorTitle') }}</h2>
        <p class="mt-2 text-sm">{{ t('publicModels.loadErrorDescription') }}</p>
      </div>

      <!-- Empty -->
      <div
        v-else-if="!filteredGroups.length"
        class="rounded-lg border border-dashed border-gray-300 bg-white px-6 py-14 text-center text-sm text-gray-500 dark:border-dark-700 dark:bg-dark-900 dark:text-dark-400"
      >
        {{ searchQuery.trim() ? t('publicModels.searchEmpty') : t('publicModels.empty') }}
      </div>

      <!-- Group list with per-model price comparison -->
      <div v-else class="space-y-4">
        <article v-for="group in filteredGroups" :key="group.id" class="group-card">
          <header class="flex flex-wrap items-center gap-3 border-b border-gray-100 px-4 py-3.5 dark:border-dark-700/50">
            <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-700 dark:bg-dark-800 dark:text-dark-200">
              <PlatformIcon :platform="(group.platform as GroupPlatform)" size="md" />
            </span>
            <div class="min-w-0">
              <h3 class="truncate text-base font-semibold text-gray-950 dark:text-white">{{ group.name }}</h3>
              <p class="text-xs uppercase tracking-wide text-gray-500 dark:text-dark-400">{{ group.platform }}</p>
            </div>
            <span class="rate-badge">
              <Icon name="bolt" size="xs" class="mr-0.5" />
              {{ formatRate(group.rate_multiplier) }}
            </span>
            <span class="ml-auto text-xs text-gray-400 dark:text-dark-400">
              {{ t('publicModels.modelCount', { count: group.models.length }) }}
            </span>
          </header>

          <div v-if="!group.models.length" class="px-4 py-8 text-center text-sm text-gray-400 dark:text-dark-500">
            {{ t('publicModels.noModels') }}
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="text-left text-xs font-medium uppercase tracking-wide text-gray-400 dark:text-dark-500">
                  <th class="px-4 py-2.5">
                    <span class="inline-flex items-center gap-1"><Icon name="cube" size="xs" />{{ t('publicModels.columns.model') }}</span>
                  </th>
                  <th class="px-3 py-2.5 text-right">
                    <span class="inline-flex items-center justify-end gap-1"><Icon name="arrowDown" size="xs" />{{ t('publicModels.columns.input') }}</span>
                  </th>
                  <th class="px-3 py-2.5 text-right">
                    <span class="inline-flex items-center justify-end gap-1"><Icon name="arrowUp" size="xs" />{{ t('publicModels.columns.output') }}</span>
                  </th>
                  <th class="px-3 py-2.5 text-right">
                    <span class="inline-flex items-center justify-end gap-1"><Icon name="database" size="xs" />{{ t('publicModels.columns.cacheWrite') }}</span>
                  </th>
                  <th class="px-3 py-2.5 text-right">
                    <span class="inline-flex items-center justify-end gap-1"><Icon name="eye" size="xs" />{{ t('publicModels.columns.cacheRead') }}</span>
                  </th>
                </tr>
              </thead>
              <tbody>
                <template v-for="model in displayedModels(group)" :key="model.name">
                  <tr class="border-t border-gray-50 transition hover:bg-primary-50/30 dark:border-dark-800 dark:hover:bg-primary-900/10">
                    <td class="px-4 py-2.5">
                      <div class="flex flex-wrap items-center gap-2">
                        <button
                          type="button"
                          class="model-name"
                          :title="t('publicModels.copyModelHint')"
                          @click="copyModel(model.name)"
                        >{{ model.name }}</button>
                        <span
                          v-if="modeBadge(model)"
                          class="inline-flex items-center rounded-md bg-amber-100 px-1.5 py-0.5 text-[10px] font-medium text-amber-700 dark:bg-amber-900/30 dark:text-amber-300"
                        >{{ modeBadge(model) }}</span>
                      </div>
                    </td>
                    <template v-if="isPerRequestMode(model)">
                      <td colspan="4" class="px-3 py-2.5 text-right font-mono text-xs text-gray-400 dark:text-dark-500">—</td>
                    </template>
                    <template v-else>
                      <td class="px-3 py-2.5 text-right font-mono text-sm" :class="priceCellTone">{{ formatPrice(basePrice(model, 'input'), group.rate_multiplier, model.price_currency) }}</td>
                      <td class="px-3 py-2.5 text-right font-mono text-sm" :class="priceCellTone">{{ formatPrice(basePrice(model, 'output'), group.rate_multiplier, model.price_currency) }}</td>
                      <td class="px-3 py-2.5 text-right font-mono text-sm" :class="priceCellTone">{{ formatPrice(basePrice(model, 'cache_write'), group.rate_multiplier, model.price_currency) }}</td>
                      <td class="px-3 py-2.5 text-right font-mono text-sm" :class="priceCellTone">{{ formatPrice(basePrice(model, 'cache_read'), group.rate_multiplier, model.price_currency) }}</td>
                    </template>
                  </tr>
                  <tr v-if="hasTierBlock(model)" class="border-t border-gray-50 bg-gray-50/30 dark:border-dark-800 dark:bg-dark-900/40">
                    <td colspan="5" class="px-4 py-3">
                      <table
                        v-if="tierMatrix(model)"
                        class="min-w-[240px] border-separate border-spacing-0 overflow-hidden rounded-md border border-gray-200 text-xs dark:border-dark-700"
                      >
                        <thead>
                          <tr class="bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                            <th class="px-3 py-1.5 text-left font-medium"></th>
                            <th v-for="col in tierMatrix(model)!.cols" :key="col" class="px-3 py-1.5 text-right font-medium">{{ col }}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr
                            v-for="row in tierMatrix(model)!.rows"
                            :key="row"
                            class="even:bg-white odd:bg-gray-50 dark:even:bg-dark-900/40 dark:odd:bg-dark-800/40"
                          >
                            <td class="px-3 py-1.5 text-left font-medium text-gray-700 dark:text-gray-200">{{ row }}</td>
                            <td
                              v-for="col in tierMatrix(model)!.cols"
                              :key="col"
                              class="px-3 py-1.5 text-right font-mono"
                              :class="priceCellTone"
                            >{{ formatPerItem(tierMatrix(model)!.cells[row]?.[col], group.rate_multiplier, model.price_currency) }}</td>
                          </tr>
                        </tbody>
                      </table>
                      <table v-else class="min-w-[200px] border-separate border-spacing-0 overflow-hidden rounded-md border border-gray-200 text-xs dark:border-dark-700">
                        <thead>
                          <tr class="bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                            <th class="px-3 py-1.5 text-left font-medium">{{ t('publicModels.tier.label') }}</th>
                            <th class="px-3 py-1.5 text-right font-medium">{{ t('publicModels.tier.price') }}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr
                            v-for="iv in (model.intervals ?? [])"
                            :key="iv.tier_label ?? `${iv.min_tokens}-${iv.max_tokens ?? ''}`"
                            class="even:bg-white odd:bg-gray-50 dark:even:bg-dark-900/40 dark:odd:bg-dark-800/40"
                          >
                            <td class="px-3 py-1.5 text-left text-gray-700 dark:text-gray-200">{{ iv.tier_label || '-' }}</td>
                            <td class="px-3 py-1.5 text-right font-mono" :class="priceCellTone">{{ formatPerItem(iv.per_request_price, group.rate_multiplier, model.price_currency) }}</td>
                          </tr>
                        </tbody>
                      </table>
                    </td>
                  </tr>
                </template>
              </tbody>
            </table>
            <button
              v-if="group.models.length > MAX_ROWS && !isExpanded(group.id)"
              type="button"
              class="w-full border-t border-gray-50 px-4 py-2.5 text-center text-xs font-medium text-primary-600 transition hover:bg-primary-50/40 dark:border-dark-800 dark:text-primary-300 dark:hover:bg-primary-900/10"
              @click="expand(group.id)"
            >
              {{ t('publicModels.showMore', { count: group.models.length - MAX_ROWS }) }}
            </button>
          </div>
        </article>
      </div>

      <p class="mt-8 text-center text-xs text-gray-500 dark:text-dark-400">
        {{ t('publicModels.footnote') }}
      </p>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { trimTrailingZeros } from '@/utils/formatters'
import { useI18n } from 'vue-i18n'
import PublicHeader from '@/components/layout/PublicHeader.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformIcon from '@/components/common/PlatformIcon.vue'
import userChannelsAPI, { type UserPricingGroup, type UserPricingModel } from '@/api/channels'
import { useAppStore } from '@/stores'
import { useClipboard } from '@/composables/useClipboard'
import { DEFAULT_CNY_PER_USD } from '@/utils/pricing'
import { costSymbol } from '@/utils/usagePricing'
import type { GroupPlatform } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const MAX_ROWS = 30

const groups = ref<UserPricingGroup[]>([])
const loading = ref(false)
const loadError = ref(false)
const platformFilter = ref<string>('')
const searchQuery = ref('')
const priceMode = ref<'official' | 'site'>('site')
const fxRate = ref<number>(DEFAULT_CNY_PER_USD)
const expandedGroups = ref<number[]>([])

const platformOptions = computed(() => {
  const counts = new Map<string, number>()
  for (const g of groups.value) {
    counts.set(g.platform, (counts.get(g.platform) ?? 0) + 1)
  }
  return Array.from(counts.entries())
    .map(([name, count]) => ({ name, count }))
    .sort((a, b) => b.count - a.count || a.name.localeCompare(b.name))
})

const totalModelCount = computed(() => {
  const names = new Set<string>()
  for (const g of groups.value) {
    for (const m of g.models) names.add(m.name)
  }
  return names.size
})

const filteredGroups = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  let result = platformFilter.value
    ? groups.value.filter((g) => g.platform === platformFilter.value)
    : groups.value
  if (q) {
    result = result.filter(
      (g) =>
        g.name.toLowerCase().includes(q) ||
        g.platform.toLowerCase().includes(q) ||
        g.models.some((m) => m.name.toLowerCase().includes(q)),
    )
  }
  return result
})

function isExpanded(groupId: number): boolean {
  return expandedGroups.value.includes(groupId)
}
function expand(groupId: number) {
  if (!expandedGroups.value.includes(groupId)) expandedGroups.value = [...expandedGroups.value, groupId]
}

function displayedModels(group: UserPricingGroup): UserPricingModel[] {
  const q = searchQuery.value.trim().toLowerCase()
  let list = group.models
  if (q) {
    // Surface matching models first so the hit is visible.
    const matched = group.models.filter((m) => m.name.toLowerCase().includes(q))
    const rest = group.models.filter((m) => !m.name.toLowerCase().includes(q))
    list = [...matched, ...rest]
  }
  if (list.length > MAX_ROWS && !isExpanded(group.id)) return list.slice(0, MAX_ROWS)
  return list
}

function copyModel(name: string) {
  copyToClipboard(name)
}

// API 接入域名：优先后台配置的 api_base_url，否则回退当前站点 origin。
const apiBaseUrl = computed(() => {
  const configured = (appStore.cachedPublicSettings?.api_base_url || '').trim()
  if (configured) return configured.replace(/\/+$/, '')
  if (typeof window !== 'undefined' && window.location?.origin) return window.location.origin
  return ''
})

function copyApiBase() {
  if (apiBaseUrl.value) copyToClipboard(`${apiBaseUrl.value}/v1`)
}

function formatRate(rate: number): string {
  const r = Number(rate || 1)
  if (Math.abs(r - 1) < 1e-6) return '1x'
  if (r >= 10) return `${r.toFixed(0)}x`
  return `${parseFloat(r.toFixed(3))}x`
}

const priceCellTone = computed(() =>
  priceMode.value === 'site'
    ? 'text-primary-600 dark:text-primary-400'
    : 'text-gray-700 dark:text-gray-300',
)

// 仅在配置了真实汇率（≠ 1:1）时才提示汇率；默认 1¥=1$ 模型下汇率说明无意义。
const showFxNote = computed(() => priceMode.value === 'site' && Math.abs(fxRate.value - 1) > 1e-6)

/**
 * basePrice 选取展示用的基础单价（per-token USD）。
 *   - official 模式：始终用 LiteLLM 官方价（official_*）
 *   - site     模式：优先渠道显式配置的单价，未配置回退到 official
 */
function basePrice(
  model: UserPricingModel,
  field: 'input' | 'output' | 'cache_write' | 'cache_read',
): number | null | undefined {
  const officialKey = `official_${field}_price` as const
  if (priceMode.value === 'official') {
    return model[officialKey]
  }
  const siteKey = `${field}_price` as const
  return model[siteKey] ?? model[officialKey]
}

/**
 * 价格格式化（per-token 单价 → 每 M token）：
 *   - official 模式：直接 × 1M
 *   - site     模式：(group.rate / cny_per_usd) × 1M
 *
 * 货币符号按 price_currency 显示：国产人民币计价模型 '¥'，其余 '$'
 * （与用量页 costSymbol 同口径，依赖部署 cny_to_usd_rate=1.0）。
 */
function formatPrice(perToken: number | null | undefined, groupRate: number, currency?: string): string {
  if (perToken == null) return '-'
  const sym = costSymbol(currency)
  const officialPerM = perToken * 1_000_000
  if (priceMode.value === 'official') {
    return `${sym}${trimNum(officialPerM)}/M`
  }
  const sitePerM = (groupRate / fxRate.value) * officialPerM
  return `${sym}${trimNum(sitePerM)}/M`
}

const TIER_SEP = '-'

function isPerRequestMode(model: UserPricingModel): boolean {
  const mode = model.billing_mode
  return mode === 'per_request' || mode === 'image'
}

function hasTierBlock(model: UserPricingModel): boolean {
  return isPerRequestMode(model) && (model.intervals?.length ?? 0) > 0
}

function modeBadge(model: UserPricingModel): string {
  if (model.billing_mode === 'image') return t('publicModels.modeBadge.image')
  if (model.billing_mode === 'per_request') return t('publicModels.modeBadge.perRequest')
  return ''
}

type TierMatrixData = {
  rows: string[]
  cols: string[]
  cells: Record<string, Record<string, number | null | undefined>>
}

function tierMatrix(model: UserPricingModel): TierMatrixData | null {
  const ivs = model.intervals ?? []
  if (ivs.length === 0) return null
  const rows: string[] = []
  const cols: string[] = []
  const cells: Record<string, Record<string, number | null | undefined>> = {}
  for (const iv of ivs) {
    const label = (iv.tier_label ?? '').trim()
    const sepIdx = label.indexOf(TIER_SEP)
    if (sepIdx <= 0 || sepIdx >= label.length - 1) return null
    const row = label.slice(0, sepIdx).trim()
    const col = label.slice(sepIdx + 1).trim()
    if (!row || !col) return null
    if (!rows.includes(row)) rows.push(row)
    if (!cols.includes(col)) cols.push(col)
    if (!cells[row]) cells[row] = {}
    cells[row][col] = iv.per_request_price
  }
  return { rows, cols, cells }
}

function formatPerItem(perItem: number | null | undefined, groupRate: number, currency?: string): string {
  if (perItem == null) return '-'
  const sym = costSymbol(currency)
  if (priceMode.value === 'official') {
    return sym + trimNum(perItem)
  }
  return sym + trimNum((groupRate / fxRate.value) * perItem)
}

function trimNum(n: number): string {
  if (n === 0) return '0'
  const digits = n >= 100 ? 0 : n >= 10 ? 2 : 4
  const fixed = n.toFixed(digits)
  return trimTrailingZeros(fixed) || '0'
}

async function reload() {
  loading.value = true
  loadError.value = false
  try {
    const [list, fx] = await Promise.all([
      userChannelsAPI.getPublicPricingGroups(),
      userChannelsAPI.getPublicFXRate().catch(() => null),
    ])
    groups.value = list
    if (fx && fx.cny_per_usd > 0) fxRate.value = fx.cny_per_usd
  } catch {
    loadError.value = true
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  reload()
})
</script>

<style scoped>
.platform-chip {
  @apply inline-flex items-center rounded-full border border-gray-200 bg-white px-3 py-1.5 text-xs font-medium text-gray-700
         transition hover:border-primary-300 hover:text-primary-700
         dark:border-dark-700 dark:bg-dark-800/40 dark:text-dark-200 dark:hover:border-primary-500/40;
}
.platform-chip-active {
  @apply border-primary-500 bg-primary-50 text-primary-700
         dark:border-primary-500 dark:bg-primary-500/15 dark:text-primary-200;
}
.group-card {
  @apply overflow-hidden rounded-xl border border-gray-200 bg-white transition-all duration-200
         hover:border-primary-200 hover:shadow-md
         dark:border-dark-700 dark:bg-dark-800/40 dark:hover:border-primary-500/40;
}
.rate-badge {
  @apply inline-flex flex-shrink-0 items-center rounded-full bg-emerald-100 px-2 py-0.5 text-xs font-semibold text-emerald-700
         dark:bg-emerald-900/30 dark:text-emerald-300;
}
.model-name {
  @apply cursor-pointer font-mono text-sm font-medium text-gray-900 transition-colors hover:text-primary-600
         dark:text-gray-100 dark:hover:text-primary-300;
}
</style>
