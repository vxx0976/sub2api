<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.rechargeOrders.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.description') }}</p>
        </div>
      </div>

      <!-- Filter Bar -->
      <div class="card p-4">
        <div class="flex flex-wrap gap-3">
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('admin.rechargeOrders.filterStatus') }}</label>
            <select v-model="filters.status" class="input w-40" @change="handleFilterChange">
              <option value="">{{ t('admin.rechargeOrders.allStatus') }}</option>
              <option value="pending">{{ t('admin.rechargeOrders.statusPending') }}</option>
              <option value="paid">{{ t('admin.rechargeOrders.statusPaid') }}</option>
              <option value="expired">{{ t('admin.rechargeOrders.statusExpired') }}</option>
              <option value="refunded">{{ t('admin.rechargeOrders.statusRefunded') }}</option>
            </select>
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('admin.rechargeOrders.userId') }}</label>
            <input
              v-model.number="filters.user_id"
              type="number"
              class="input w-32"
              :placeholder="t('admin.rechargeOrders.userIdPlaceholder')"
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

      <!-- Table -->
      <div v-else-if="items.length > 0" class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.orderNo') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.userId') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.amount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.creditAmount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.status') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.payType') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.paidAt') }}</th>
                <th class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in items"
                :key="item.id"
                class="border-b border-gray-50 dark:border-dark-800"
              >
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.order_no }}</span>
                </td>
                <td class="px-4 py-3">
                  <div>
                    <span class="text-sm text-gray-900 dark:text-white">{{ item.user_email || '—' }}</span>
                    <span class="ml-1 text-xs text-gray-400 dark:text-gray-500">#{{ item.user_id }}</span>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">&yen;{{ item.amount }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">${{ item.credit_amount }}</span>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="statusClass(item.status)"
                  >
                    {{ statusLabel(item.status) }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ item.pay_type || '—' }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ item.paid_at ? formatDateTime(item.paid_at) : '—' }}</span>
                </td>
                <td class="px-4 py-3 text-right">
                  <button
                    v-if="item.status === 'paid'"
                    type="button"
                    class="text-sm font-medium text-red-600 hover:text-red-700 dark:text-red-400"
                    @click="openRefund(item)"
                  >
                    {{ t('admin.rechargeOrders.refund') }}
                  </button>
                  <span v-else class="text-sm text-gray-300 dark:text-gray-600">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <p class="text-gray-500 dark:text-gray-400">{{ t('admin.rechargeOrders.empty') }}</p>
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

    <ConfirmDialog
      :show="showRefundDialog"
      :title="t('admin.rechargeOrders.refundTitle')"
      :message="t('admin.rechargeOrders.refundConfirm', { orderNo: refundingOrder?.order_no, amount: refundingOrder?.credit_amount })"
      :confirm-text="refundSubmitting ? t('common.processing') : t('admin.rechargeOrders.refund')"
      :danger="true"
      @confirm="confirmRefund"
      @cancel="showRefundDialog = false"
    >
      <textarea
        v-model="refundReason"
        rows="2"
        class="input w-full"
        :placeholder="t('admin.rechargeOrders.refundReasonPlaceholder')"
      />
    </ConfirmDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminRechargeOrderItem } from '@/api/admin/rechargeOrders'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const items = ref<AdminRechargeOrderItem[]>([])

const filters = reactive({
  status: '',
  user_id: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

function statusClass(status: string) {
  switch (status) {
    case 'pending':
      return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
    case 'paid':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'expired':
      return 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
    case 'refunded':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'pending': return t('admin.rechargeOrders.statusPending')
    case 'paid': return t('admin.rechargeOrders.statusPaid')
    case 'expired': return t('admin.rechargeOrders.statusExpired')
    case 'refunded': return t('admin.rechargeOrders.statusRefunded')
    default: return status
  }
}

function handleFilterChange() {
  pagination.page = 1
  loadOrders()
}

function resetFilters() {
  filters.status = ''
  filters.user_id = undefined
  pagination.page = 1
  loadOrders()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadOrders()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadOrders()
}

async function loadOrders() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.status) params.status = filters.status
    if (filters.user_id) params.user_id = filters.user_id

    const result = await adminAPI.rechargeOrders.getAdminRechargeOrders(params)
    items.value = result.items || []
    pagination.total = result.total
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

const showRefundDialog = ref(false)
const refundingOrder = ref<AdminRechargeOrderItem | null>(null)
const refundReason = ref('')
const refundSubmitting = ref(false)

function openRefund(item: AdminRechargeOrderItem) {
  refundingOrder.value = item
  refundReason.value = ''
  showRefundDialog.value = true
}

async function confirmRefund() {
  if (!refundingOrder.value || refundSubmitting.value) return
  refundSubmitting.value = true
  try {
    await adminAPI.rechargeOrders.refundAdminRechargeOrder(refundingOrder.value.order_no, refundReason.value)
    appStore.showSuccess(t('admin.rechargeOrders.refundSuccess'))
    showRefundDialog.value = false
    loadOrders()
  } catch (error: any) {
    appStore.showError(error.message || t('common.error'))
  } finally {
    refundSubmitting.value = false
  }
}

onMounted(() => {
  loadOrders()
})
</script>
