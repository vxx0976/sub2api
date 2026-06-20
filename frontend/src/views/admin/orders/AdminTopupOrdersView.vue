<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('topupOrders.adminTitle') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('topupOrders.adminDescription') }}</p>
        </div>
      </div>

      <!-- Filter Bar -->
      <div class="card p-4">
        <div class="flex flex-wrap gap-3">
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('topupOrders.filterChannel') }}</label>
            <select v-model="filters.channel" class="input w-36" @change="handleFilterChange">
              <option value="">{{ t('topupOrders.allChannels') }}</option>
              <option value="alimpay">{{ t('topupOrders.methodAlipay') }}</option>
              <option value="recharge">{{ t('topupOrders.methodWechat') }}</option>
              <option value="usdt">USDT</option>
            </select>
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('topupOrders.filterStatus') }}</label>
            <select v-model="filters.status" class="input w-36" @change="handleFilterChange">
              <option value="">{{ t('topupOrders.allStatus') }}</option>
              <option value="pending">{{ t('topupOrders.statusPending') }}</option>
              <option value="paid">{{ t('topupOrders.statusPaid') }}</option>
              <option value="expired">{{ t('topupOrders.statusExpired') }}</option>
              <option value="refunded">{{ t('topupOrders.statusRefunded') }}</option>
            </select>
          </div>
          <div class="flex items-center gap-2">
            <label class="text-sm text-gray-600 dark:text-gray-400">{{ t('topupOrders.userId') }}</label>
            <input
              v-model.number="filters.user_id"
              type="number"
              class="input w-32"
              :placeholder="t('topupOrders.userIdPlaceholder')"
              @change="handleFilterChange"
            />
          </div>
          <button class="btn btn-secondary" @click="resetFilters">{{ t('common.reset') }}</button>
        </div>
      </div>

      <!-- Loading / Table / Empty -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
      <TopupOrdersTable v-else-if="items.length > 0" :items="items" admin @refund="openRefund" />
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <p class="text-gray-500 dark:text-gray-400">{{ t('topupOrders.empty') }}</p>
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
      :title="t('topupOrders.refundTitle')"
      :message="t('topupOrders.refundConfirm', { orderNo: refundingOrder?.order_no, amount: refundingOrder?.credit_amount })"
      :confirm-text="refundSubmitting ? t('common.processing') : t('topupOrders.refund')"
      :danger="true"
      @confirm="confirmRefund"
      @cancel="showRefundDialog = false"
    >
      <textarea
        v-model="refundReason"
        rows="2"
        class="input w-full"
        :placeholder="t('topupOrders.refundReasonPlaceholder')"
      />
    </ConfirmDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { topupOrdersAPI, type MergedTopupOrder, type TopupChannel } from '@/api/topupOrders'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import TopupOrdersTable from '@/components/payment/TopupOrdersTable.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const items = ref<MergedTopupOrder[]>([])

const filters = reactive({
  channel: '' as TopupChannel | '',
  status: '',
  user_id: undefined as number | undefined
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0
})

function handleFilterChange() {
  pagination.page = 1
  loadOrders()
}

function resetFilters() {
  filters.channel = ''
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
    const result = await topupOrdersAPI.getAdminTopupOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      channel: filters.channel,
      status: filters.status,
      user_id: filters.user_id
    })
    items.value = result.items || []
    pagination.total = result.total
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

const showRefundDialog = ref(false)
const refundingOrder = ref<MergedTopupOrder | null>(null)
const refundReason = ref('')
const refundSubmitting = ref(false)

function openRefund(item: MergedTopupOrder) {
  refundingOrder.value = item
  refundReason.value = ''
  showRefundDialog.value = true
}

async function confirmRefund() {
  if (!refundingOrder.value || refundSubmitting.value) return
  refundSubmitting.value = true
  try {
    await topupOrdersAPI.refundTopupOrder(
      refundingOrder.value.channel,
      refundingOrder.value.order_no,
      refundReason.value
    )
    appStore.showSuccess(t('topupOrders.refundSuccess'))
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
