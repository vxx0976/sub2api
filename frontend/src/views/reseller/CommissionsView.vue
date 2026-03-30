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
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalRecharge') }}</p>
          <p class="mt-1 text-xl font-bold text-purple-600 dark:text-purple-400">¥{{ f2(summary.total_recharge) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCommission') }}</p>
          <p class="mt-1 text-xl font-bold text-emerald-600 dark:text-emerald-400">¥{{ f2(summary.total_commission) }}</p>
          <p class="mt-0.5 text-xs text-gray-400 dark:text-gray-500">{{ t('reseller.commissions.commissionRate') }}: {{ f2p(summary.commission_rate) }}%</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.withdrawn') }}</p>
          <p class="mt-1 text-xl font-bold text-gray-900 dark:text-white">¥{{ f2(summary.withdrawn) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.pending') }}</p>
          <p class="mt-1 text-xl font-bold text-yellow-600 dark:text-yellow-400">¥{{ f2(summary.pending) }}</p>
        </div>
        <div class="card p-4">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.available') }}</p>
          <p class="mt-1 text-xl font-bold text-blue-600 dark:text-blue-400">¥{{ f2(summary.available) }}</p>
        </div>
      </div>

      <!-- Recharge List -->
      <div v-if="rechargeLoading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <div v-else-if="rechargeItems.length > 0" class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.userId') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.userEmail') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.orderNo') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.creditAmount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.paidAt') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(item, idx) in rechargeItems" :key="idx" class="border-b border-gray-50 dark:border-dark-800">
                <td class="px-4 py-3"><span class="text-sm text-gray-900 dark:text-white">{{ item.user_id }}</span></td>
                <td class="px-4 py-3"><span class="text-sm text-gray-600 dark:text-gray-300">{{ item.user_email || '-' }}</span></td>
                <td class="px-4 py-3"><span class="text-sm font-mono text-gray-600 dark:text-gray-300">{{ item.order_no }}</span></td>
                <td class="px-4 py-3"><span class="text-sm font-medium text-blue-600 dark:text-blue-400">¥{{ f2(item.credit_amount) }}</span></td>
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
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { CommissionSummary, RechargeDetailItem } from '@/api/reseller/commissions'
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
function f2p(v: number) {
  return (v * 100).toFixed(0)
}

const summary = ref<CommissionSummary | null>(null)

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
  loadRecharge()
})
</script>
