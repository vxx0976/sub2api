<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.withdrawals.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.description') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <div v-if="summary" class="rounded-lg bg-blue-50 px-4 py-2 dark:bg-blue-900/20">
            <p class="text-xs text-blue-600 dark:text-blue-400">{{ t('reseller.withdrawals.available') }}</p>
            <p class="text-lg font-bold text-blue-700 dark:text-blue-300">¥{{ Number(summary.available).toFixed(2) }}</p>
          </div>
          <button @click="openApplyDialog" class="btn">
            {{ t('reseller.withdrawals.apply') }}
          </button>
        </div>
      </div>

      <!-- Filter Bar -->
      <div class="card p-4">
        <div class="flex flex-wrap items-center gap-3">
          <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('reseller.withdrawals.filterStatus') }}</label>
          <select v-model="statusFilter" class="input w-40" @change="handleFilterChange">
            <option value="">{{ t('reseller.withdrawals.allStatus') }}</option>
            <option value="pending">{{ t('reseller.withdrawals.statusPending') }}</option>
            <option value="paid">{{ t('reseller.withdrawals.statusPaid') }}</option>
            <option value="rejected">{{ t('reseller.withdrawals.statusRejected') }}</option>
          </select>
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
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.amount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.status') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.paymentMethod') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.paymentAccount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.paymentName') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.actions') }}</th>
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
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</span>
                </td>
                <td class="px-4 py-3">
                  <button
                    v-if="item.status === 'pending'"
                    @click="handleCancel(item)"
                    class="text-sm text-red-600 hover:text-red-700 dark:text-red-400 dark:hover:text-red-300"
                  >
                    {{ t('reseller.withdrawals.cancel') }}
                  </button>
                  <span v-else class="text-sm text-gray-400">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.withdrawals.empty') }}</p>
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

    <!-- Apply Withdrawal Dialog -->
    <BaseDialog
      :show="showApplyDialog"
      :title="t('reseller.withdrawals.applyTitle')"
      width="narrow"
      @close="closeApplyDialog"
    >
      <form id="withdrawal-form" @submit.prevent="handleApplySubmit" class="space-y-4">
        <div>
          <label class="input-label">{{ t('reseller.withdrawals.amount') }}</label>
          <div class="relative">
            <div class="absolute left-3 top-1/2 -translate-y-1/2 font-medium text-gray-500">$</div>
            <input
              v-model.number="applyForm.amount"
              type="number"
              step="0.01"
              min="0.01"
              required
              class="input pl-8"
            />
          </div>
          <p v-if="summary" class="mt-1 text-xs text-gray-500 dark:text-gray-400">
            {{ t('reseller.withdrawals.availableBalance') }}: ¥{{ Number(summary.available).toFixed(2) }}
          </p>
        </div>
        <div>
          <label class="input-label">{{ t('reseller.withdrawals.paymentMethod') }}</label>
          <select v-model="applyForm.payment_method" required class="input">
            <option value="">{{ t('reseller.withdrawals.selectPaymentMethod') }}</option>
            <option value="alipay">{{ t('reseller.withdrawals.alipay') }}</option>
            <option value="wechat">{{ t('reseller.withdrawals.wechat') }}</option>
            <option value="bank">{{ t('reseller.withdrawals.bank') }}</option>
          </select>
        </div>
        <div>
          <label class="input-label">{{ t('reseller.withdrawals.paymentAccount') }}</label>
          <input
            v-model="applyForm.payment_account"
            type="text"
            required
            class="input"
            :placeholder="t('reseller.withdrawals.paymentAccountPlaceholder')"
          />
        </div>
        <div>
          <label class="input-label">{{ t('reseller.withdrawals.paymentName') }}</label>
          <input
            v-model="applyForm.payment_name"
            type="text"
            required
            class="input"
            :placeholder="t('reseller.withdrawals.paymentNamePlaceholder')"
          />
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeApplyDialog" class="btn btn-secondary">{{ t('common.cancel') }}</button>
          <button
            type="submit"
            form="withdrawal-form"
            :disabled="applySubmitting"
            class="btn"
          >
            {{ applySubmitting ? t('common.saving') : t('reseller.withdrawals.submit') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { WithdrawalItem } from '@/api/reseller/withdrawals'
import type { CommissionSummary } from '@/api/reseller/commissions'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const summary = ref<CommissionSummary | null>(null)
const items = ref<WithdrawalItem[]>([])
const statusFilter = ref('')

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

// Apply Dialog
const showApplyDialog = ref(false)
const applySubmitting = ref(false)
const applyForm = reactive({
  amount: 0,
  payment_method: '',
  payment_account: '',
  payment_name: ''
})

function openApplyDialog() {
  applyForm.amount = 0
  applyForm.payment_method = ''
  applyForm.payment_account = ''
  applyForm.payment_name = ''
  showApplyDialog.value = true
}

function closeApplyDialog() {
  showApplyDialog.value = false
}

async function handleApplySubmit() {
  if (!applyForm.amount || applyForm.amount <= 0) {
    appStore.showError(t('reseller.withdrawals.amountRequired'))
    return
  }
  if (!applyForm.payment_method) {
    appStore.showError(t('reseller.withdrawals.paymentMethodRequired'))
    return
  }
  applySubmitting.value = true
  try {
    await resellerAPI.withdrawals.createWithdrawal({
      amount: applyForm.amount,
      payment_method: applyForm.payment_method,
      payment_account: applyForm.payment_account,
      payment_name: applyForm.payment_name
    })
    appStore.showSuccess(t('reseller.withdrawals.applySuccess'))
    closeApplyDialog()
    loadSummary()
    loadWithdrawals()
  } catch (error: any) {
    appStore.showError(error.message || t('common.error'))
  } finally {
    applySubmitting.value = false
  }
}

async function handleCancel(item: WithdrawalItem) {
  if (!confirm(t('reseller.withdrawals.cancelConfirm'))) return
  try {
    await resellerAPI.withdrawals.cancelWithdrawal(item.id)
    appStore.showSuccess(t('reseller.withdrawals.cancelSuccess'))
    loadSummary()
    loadWithdrawals()
  } catch (error: any) {
    appStore.showError(error.message || t('common.error'))
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
    case 'pending': return t('reseller.withdrawals.statusPending')
    case 'paid': return t('reseller.withdrawals.statusPaid')
    case 'rejected': return t('reseller.withdrawals.statusRejected')
    default: return status
  }
}

function paymentMethodLabel(method: string) {
  switch (method) {
    case 'alipay': return t('reseller.withdrawals.alipay')
    case 'wechat': return t('reseller.withdrawals.wechat')
    case 'bank': return t('reseller.withdrawals.bank')
    default: return method
  }
}

function handleFilterChange() {
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

async function loadSummary() {
  try {
    summary.value = await resellerAPI.commissions.getCommissionSummary()
  } catch {
    // silently ignore
  }
}

async function loadWithdrawals() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (statusFilter.value) params.status = statusFilter.value

    const result = await resellerAPI.withdrawals.getWithdrawals(params)
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
  loadSummary()
  loadWithdrawals()
})
</script>
