<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.commissions.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.description') }}</p>
        </div>
      </div>

      <!-- Summary Cards -->
      <div v-if="summary" class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCost') }}</p>
          <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">${{ f2(summary.total_cost) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCommission') }}</p>
          <p class="mt-1 text-xl font-bold text-emerald-600 dark:text-emerald-400">${{ f2(summary.total_commission) }}</p>
          <p class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">{{ t('reseller.commissions.commissionRate') }}: {{ f2p(summary.commission_rate) }}%</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.withdrawn') }}</p>
          <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">${{ f2(summary.withdrawn) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.pending') }}</p>
          <p class="mt-1 text-xl font-bold text-yellow-600 dark:text-yellow-400">${{ f2(summary.pending) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.available') }}</p>
          <p class="mt-1 text-xl font-bold text-blue-600 dark:text-blue-400">${{ f2(summary.available) }}</p>
        </div>
      </div>

      <!-- Tabs -->
      <div class="flex gap-1 rounded-lg bg-gray-100 p-1 dark:bg-dark-700">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          :class="activeTab === tab.key
            ? 'bg-white text-gray-900 shadow dark:bg-dark-600 dark:text-white'
            : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'"
          class="flex-1 rounded-md px-4 py-2 text-sm font-medium transition-all"
        >
          {{ tab.label }}
        </button>
      </div>

      <!-- Consumption Tab -->
      <template v-if="activeTab === 'consumption'">
        <!-- Filter Bar -->
        <div class="card p-4">
          <div class="flex flex-wrap gap-3">
            <div class="flex items-center gap-2">
              <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('reseller.commissions.startDate') }}</label>
              <input v-model="consumptionFilters.start_date" type="date" class="input w-40" @change="handleConsumptionFilterChange" />
            </div>
            <div class="flex items-center gap-2">
              <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('reseller.commissions.endDate') }}</label>
              <input v-model="consumptionFilters.end_date" type="date" class="input w-40" @change="handleConsumptionFilterChange" />
            </div>
            <div class="flex items-center gap-2">
              <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('reseller.commissions.userId') }}</label>
              <input
                v-model.number="consumptionFilters.user_id"
                type="number"
                class="input w-32"
                :placeholder="t('reseller.commissions.userIdPlaceholder')"
                @change="handleConsumptionFilterChange"
              />
            </div>
            <button @click="resetConsumptionFilters" class="btn btn-secondary">{{ t('common.reset') }}</button>
          </div>
        </div>

        <div v-if="consumptionLoading" class="flex items-center justify-center py-12">
          <LoadingSpinner />
        </div>

        <div v-else-if="consumptionItems.length > 0" class="card overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-gray-100 dark:border-dark-700">
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.userId') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.model') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCost') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.commission') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.createdAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in consumptionItems" :key="item.id" class="border-b border-gray-50 dark:border-dark-800">
                  <td class="px-4 py-3"><span class="text-sm text-gray-900 dark:text-white">{{ item.user_id }}</span></td>
                  <td class="px-4 py-3"><span class="text-sm text-gray-600 dark:text-gray-300">{{ item.model }}</span></td>
                  <td class="px-4 py-3"><span class="text-sm text-gray-900 dark:text-white">${{ f2(item.total_cost) }}</span></td>
                  <td class="px-4 py-3"><span class="text-sm font-medium text-emerald-600 dark:text-emerald-400">${{ f2(item.commission) }}</span></td>
                  <td class="px-4 py-3"><span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
          <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.empty') }}</p>
        </div>

        <Pagination
          v-if="consumptionPagination.total > 0"
          :page="consumptionPagination.page"
          :total="consumptionPagination.total"
          :page-size="consumptionPagination.page_size"
          @update:page="handleConsumptionPageChange"
          @update:pageSize="handleConsumptionPageSizeChange"
        />
      </template>

      <!-- Recharge Tab -->
      <template v-if="activeTab === 'recharge'">
        <div v-if="rechargeLoading" class="flex items-center justify-center py-12">
          <LoadingSpinner />
        </div>

        <div v-else-if="rechargeItems.length > 0" class="card overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full">
              <thead>
                <tr class="border-b border-gray-100 dark:border-dark-700">
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.userId') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.orderNo') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.creditAmount') }}</th>
                  <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.paidAt') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(item, idx) in rechargeItems" :key="idx" class="border-b border-gray-50 dark:border-dark-800">
                  <td class="px-4 py-3"><span class="text-sm text-gray-900 dark:text-white">{{ item.user_id }}</span></td>
                  <td class="px-4 py-3"><span class="text-sm font-mono text-gray-600 dark:text-gray-300">{{ item.order_no }}</span></td>
                  <td class="px-4 py-3"><span class="text-sm font-medium text-blue-600 dark:text-blue-400">${{ f4(item.credit_amount) }}</span></td>
                  <td class="px-4 py-3"><span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(item.paid_at) }}</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
          <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.empty') }}</p>
        </div>

        <Pagination
          v-if="rechargePagination.total > 0"
          :page="rechargePagination.page"
          :total="rechargePagination.total"
          :page-size="rechargePagination.page_size"
          @update:page="handleRechargePageChange"
          @update:pageSize="handleRechargePageSizeChange"
        />
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { CommissionDetailItem, CommissionSummary, RechargeDetailItem } from '@/api/reseller/commissions'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'

const { t } = useI18n()
const appStore = useAppStore()

function f2(v: string | number) {
  return Number(v).toFixed(2)
}
function f4(v: string | number) {
  return Number(v).toFixed(4)
}
function f2p(v: number) {
  return (v * 100).toFixed(0)
}

const activeTab = ref<'consumption' | 'recharge'>('consumption')
const tabs = computed(() => [
  { key: 'consumption' as const, label: t('reseller.commissions.tabConsumption') },
  { key: 'recharge' as const, label: t('reseller.commissions.tabRecharge') },
])

const summary = ref<CommissionSummary | null>(null)

// ── Consumption tab ──
const consumptionLoading = ref(true)
const consumptionItems = ref<CommissionDetailItem[]>([])
const consumptionFilters = reactive({ start_date: '', end_date: '', user_id: undefined as number | undefined })
const consumptionPagination = reactive({ page: 1, page_size: 20, total: 0, pages: 0 })

// ── Recharge tab ──
const rechargeLoading = ref(true)
const rechargeItems = ref<RechargeDetailItem[]>([])
const rechargePagination = reactive({ page: 1, page_size: 20, total: 0, pages: 0 })

async function loadSummary() {
  try {
    summary.value = await resellerAPI.commissions.getCommissionSummary()
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  }
}

async function loadConsumption() {
  consumptionLoading.value = true
  try {
    const params: Record<string, any> = {
      page: consumptionPagination.page,
      page_size: consumptionPagination.page_size
    }
    if (consumptionFilters.start_date) params.start_date = consumptionFilters.start_date
    if (consumptionFilters.end_date) params.end_date = consumptionFilters.end_date
    if (consumptionFilters.user_id) params.user_id = consumptionFilters.user_id
    const result = await resellerAPI.commissions.getCommissionDetail(params)
    consumptionItems.value = result.items || []
    consumptionPagination.total = result.total
    consumptionPagination.pages = result.pages
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    consumptionLoading.value = false
  }
}

async function loadRecharge() {
  rechargeLoading.value = true
  try {
    const result = await resellerAPI.commissions.getRechargeDetail({
      page: rechargePagination.page,
      page_size: rechargePagination.page_size
    })
    rechargeItems.value = result.items || []
    rechargePagination.total = result.total
    rechargePagination.pages = result.pages
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    rechargeLoading.value = false
  }
}

function handleConsumptionFilterChange() {
  consumptionPagination.page = 1
  loadConsumption()
}
function resetConsumptionFilters() {
  consumptionFilters.start_date = ''
  consumptionFilters.end_date = ''
  consumptionFilters.user_id = undefined
  consumptionPagination.page = 1
  loadConsumption()
}
function handleConsumptionPageChange(page: number) {
  consumptionPagination.page = page
  loadConsumption()
}
function handleConsumptionPageSizeChange(size: number) {
  consumptionPagination.page_size = size
  consumptionPagination.page = 1
  loadConsumption()
}
function handleRechargePageChange(page: number) {
  rechargePagination.page = page
  loadRecharge()
}
function handleRechargePageSizeChange(size: number) {
  rechargePagination.page_size = size
  rechargePagination.page = 1
  loadRecharge()
}

onMounted(() => {
  loadSummary()
  loadConsumption()
  loadRecharge()
})
</script>
