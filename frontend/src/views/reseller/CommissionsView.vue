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
      <div v-if="summary" class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.priceMultiplier') }}</p>
          <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">{{ summary.price_multiplier }}x</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCost') }}</p>
          <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">${{ summary.total_cost }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCommission') }}</p>
          <p class="mt-1 text-xl font-bold text-emerald-600 dark:text-emerald-400">${{ summary.total_commission }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.withdrawn') }}</p>
          <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">${{ summary.withdrawn }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.pending') }}</p>
          <p class="mt-1 text-xl font-bold text-yellow-600 dark:text-yellow-400">${{ summary.pending }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.available') }}</p>
          <p class="mt-1 text-xl font-bold text-blue-600 dark:text-blue-400">${{ summary.available }}</p>
        </div>
      </div>

      <!-- Filter Bar -->
      <div class="card p-4">
        <div class="flex flex-wrap gap-3">
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('reseller.commissions.startDate') }}</label>
            <input
              v-model="filters.start_date"
              type="date"
              class="input w-40"
              @change="handleFilterChange"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('reseller.commissions.endDate') }}</label>
            <input
              v-model="filters.end_date"
              type="date"
              class="input w-40"
              @change="handleFilterChange"
            />
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('reseller.commissions.userId') }}</label>
            <input
              v-model.number="filters.user_id"
              type="number"
              class="input w-32"
              :placeholder="t('reseller.commissions.userIdPlaceholder')"
              @change="handleFilterChange"
            />
          </div>
          <button @click="resetFilters" class="btn btn-secondary">
            {{ t('common.reset') }}
          </button>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <!-- Detail Table -->
      <div v-else-if="items.length > 0" class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.userId') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.model') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCost') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.merchantRateSnapshot') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.commission') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.createdAt') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in items"
                :key="item.id"
                class="border-b border-gray-50 dark:border-dark-800"
              >
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.user_id }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ item.model }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">${{ item.total_cost }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ item.merchant_rate_snapshot }}x</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm font-medium text-emerald-600 dark:text-emerald-400">${{ item.commission }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.empty') }}</p>
      </div>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { CommissionDetailItem, CommissionSummary } from '@/api/reseller/commissions'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const summary = ref<CommissionSummary | null>(null)
const items = ref<CommissionDetailItem[]>([])

const filters = reactive({
  start_date: '',
  end_date: '',
  user_id: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

async function loadSummary() {
  try {
    summary.value = await resellerAPI.commissions.getCommissionSummary()
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  }
}

async function loadDetail() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.start_date) params.start_date = filters.start_date
    if (filters.end_date) params.end_date = filters.end_date
    if (filters.user_id) params.user_id = filters.user_id

    const result = await resellerAPI.commissions.getCommissionDetail(params)
    items.value = result.items || []
    pagination.total = result.total
    pagination.pages = result.pages
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

function handleFilterChange() {
  pagination.page = 1
  loadDetail()
}

function resetFilters() {
  filters.start_date = ''
  filters.end_date = ''
  filters.user_id = undefined
  pagination.page = 1
  loadDetail()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadDetail()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadDetail()
}

onMounted(() => {
  loadSummary()
  loadDetail()
})
</script>
