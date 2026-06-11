<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.merchantWithdrawals.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.description') }}</p>
        </div>
      </div>

      <!-- Filter Bar -->
      <div class="card p-4">
        <div class="flex flex-wrap gap-3">
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('admin.merchantWithdrawals.filterStatus') }}</label>
            <select v-model="filters.status" class="input w-40" @change="handleFilterChange">
              <option value="">{{ t('admin.merchantWithdrawals.allStatus') }}</option>
              <option value="pending">{{ t('admin.merchantWithdrawals.statusPending') }}</option>
              <option value="paid">{{ t('admin.merchantWithdrawals.statusPaid') }}</option>
              <option value="rejected">{{ t('admin.merchantWithdrawals.statusRejected') }}</option>
            </select>
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('admin.merchantWithdrawals.resellerId') }}</label>
            <input
              v-model.number="filters.reseller_id"
              type="number"
              class="input w-32"
              :placeholder="t('admin.merchantWithdrawals.resellerIdPlaceholder')"
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
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">ID</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.resellerId') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.amount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.status') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.paymentMethod') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.paymentAccount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.paymentName') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.adminNotes') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in items"
                :key="item.id"
                class="border-b border-gray-50 dark:border-dark-800"
              >
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.id }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.reseller_id }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">${{ item.amount }}</span>
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
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ paymentMethodLabel(item.payment_method) }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ item.payment_account }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ item.payment_name }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="max-w-32 truncate text-sm text-gray-500 dark:text-gray-400" :title="item.admin_notes">
                    {{ item.admin_notes || '—' }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</span>
                </td>
                <td class="px-4 py-3">
                  <div v-if="item.status === 'pending'" class="flex items-center gap-2">
                    <button
                      @click="openActionDialog(item, 'pay')"
                      class="rounded px-2 py-1 text-xs font-medium text-emerald-600 transition-colors hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-900/20"
                    >
                      {{ t('admin.merchantWithdrawals.pay') }}
                    </button>
                    <button
                      @click="openActionDialog(item, 'reject')"
                      class="rounded px-2 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                    >
                      {{ t('admin.merchantWithdrawals.reject') }}
                    </button>
                  </div>
                  <span v-else class="text-sm text-gray-400">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <p class="text-gray-500 dark:text-gray-400">{{ t('admin.merchantWithdrawals.empty') }}</p>
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

    <!-- Action Dialog (Pay / Reject) -->
    <BaseDialog
      :show="showActionDialog"
      :title="actionType === 'pay' ? t('admin.merchantWithdrawals.payTitle') : t('admin.merchantWithdrawals.rejectTitle')"
      width="narrow"
      @close="closeActionDialog"
    >
      <form v-if="actionItem" id="withdrawal-action-form" @submit.prevent="handleActionSubmit" class="space-y-4">
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
          <div class="flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('admin.merchantWithdrawals.amount') }}</span>
            <span class="font-medium text-gray-900 dark:text-white">${{ actionItem.amount }}</span>
          </div>
          <div class="mt-1 flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('admin.merchantWithdrawals.paymentMethod') }}</span>
            <span class="text-gray-900 dark:text-white">{{ paymentMethodLabel(actionItem.payment_method) }}</span>
          </div>
          <div class="mt-1 flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('admin.merchantWithdrawals.paymentAccount') }}</span>
            <span class="text-gray-900 dark:text-white">{{ actionItem.payment_account }}</span>
          </div>
          <div class="mt-1 flex justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('admin.merchantWithdrawals.paymentName') }}</span>
            <span class="text-gray-900 dark:text-white">{{ actionItem.payment_name }}</span>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('admin.merchantWithdrawals.adminNotes') }}</label>
          <textarea
            v-model="actionNotes"
            rows="3"
            class="input"
            :placeholder="t('admin.merchantWithdrawals.adminNotesPlaceholder')"
          ></textarea>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeActionDialog" class="btn btn-secondary">{{ t('common.cancel') }}</button>
          <button
            type="submit"
            form="withdrawal-action-form"
            :disabled="actionSubmitting"
            class="btn"
            :class="actionType === 'pay' ? 'bg-emerald-600 text-white hover:bg-emerald-700' : 'btn-danger'"
          >
            {{ actionSubmitting ? t('common.saving') : (actionType === 'pay' ? t('admin.merchantWithdrawals.confirmPay') : t('admin.merchantWithdrawals.confirmReject')) }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminWithdrawalItem } from '@/api/admin/merchantWithdrawals'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const items = ref<AdminWithdrawalItem[]>([])

const filters = reactive({
  status: '',
  reseller_id: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

// Action Dialog
const showActionDialog = ref(false)
const actionSubmitting = ref(false)
const actionItem = ref<AdminWithdrawalItem | null>(null)
const actionType = ref<'pay' | 'reject'>('pay')
const actionNotes = ref('')

function openActionDialog(item: AdminWithdrawalItem, type: 'pay' | 'reject') {
  actionItem.value = item
  actionType.value = type
  actionNotes.value = ''
  showActionDialog.value = true
}

function closeActionDialog() {
  showActionDialog.value = false
  actionItem.value = null
}

async function handleActionSubmit() {
  if (!actionItem.value) return
  actionSubmitting.value = true
  try {
    const payload = { admin_notes: actionNotes.value }
    if (actionType.value === 'pay') {
      await adminAPI.merchantWithdrawals.payWithdrawal(actionItem.value.id, payload)
      appStore.showSuccess(t('admin.merchantWithdrawals.paySuccess'))
    } else {
      await adminAPI.merchantWithdrawals.rejectWithdrawal(actionItem.value.id, payload)
      appStore.showSuccess(t('admin.merchantWithdrawals.rejectSuccess'))
    }
    closeActionDialog()
    loadWithdrawals()
  } catch (error: any) {
    appStore.showError(error.message || t('common.error'))
  } finally {
    actionSubmitting.value = false
  }
}

function statusClass(status: string) {
  switch (status) {
    case 'pending':
      return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
    case 'paid':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'rejected':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
  }
}

function statusLabel(status: string) {
  switch (status) {
    case 'pending': return t('admin.merchantWithdrawals.statusPending')
    case 'paid': return t('admin.merchantWithdrawals.statusPaid')
    case 'rejected': return t('admin.merchantWithdrawals.statusRejected')
    default: return status
  }
}

function paymentMethodLabel(method: string) {
  switch (method) {
    case 'alipay': return t('admin.merchantWithdrawals.alipay')
    case 'wechat': return t('admin.merchantWithdrawals.wechat')
    case 'bank': return t('admin.merchantWithdrawals.bank')
    default: return method
  }
}

function handleFilterChange() {
  pagination.page = 1
  loadWithdrawals()
}

function resetFilters() {
  filters.status = ''
  filters.reseller_id = undefined
  pagination.page = 1
  loadWithdrawals()
}

function handlePageChange(page: number) {
  pagination.page = page
  loadWithdrawals()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadWithdrawals()
}

async function loadWithdrawals() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.status) params.status = filters.status
    if (filters.reseller_id) params.reseller_id = filters.reseller_id

    const result = await adminAPI.merchantWithdrawals.getAdminWithdrawals(params)
    items.value = result.items || []
    pagination.total = result.total
    pagination.pages = result.pages
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadWithdrawals()
})
</script>
