<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="orders.length === 0" class="card p-12 text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
        >
          <Icon name="document" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('recharge.noOrders') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('recharge.noOrdersDesc') }}
        </p>
        <router-link to="/plans" class="btn btn-primary mt-4">
          {{ t('recharge.rechargeNow') }}
        </router-link>
      </div>

      <!-- Orders List -->
      <div v-else class="space-y-6">
        <!-- Header -->
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('recharge.myOrders') }}
          </h2>
          <router-link to="/plans" class="btn btn-primary">
            <Icon name="plus" size="sm" />
            {{ t('recharge.rechargeNow') }}
          </router-link>
        </div>

        <!-- Orders Table -->
        <div class="card overflow-hidden">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.orderNo') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.amount') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.creditAmount') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.multiplier') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.status') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.createdAt') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('userOrders.actions') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="order in orders" :key="order.id" class="hover:bg-gray-50 dark:hover:bg-dark-800/50">
                <td class="px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">
                  {{ order.order_no }}
                </td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                  ¥{{ order.amount.toFixed(2) }}
                </td>
                <td class="px-4 py-3 text-sm font-semibold text-primary-600 dark:text-primary-400">
                  ¥{{ order.credit_amount.toFixed(2) }}
                </td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ order.multiplier.toFixed(1) }}×
                  <span v-if="order.multiplier > 1.0" class="text-xs text-green-600 dark:text-green-400">
                    (+{{ ((order.multiplier - 1) * 100).toFixed(0) }}%)
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span :class="['badge', getStatusClass(order.status)]">
                    {{ t(`recharge.statusLabels.${order.status}`) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-400">
                  {{ formatDate(order.created_at) }}
                </td>
                <td class="px-4 py-3">
                  <button
                    v-if="order.status === 'pending' && !isOrderExpired(order)"
                    @click="handleRepay(order)"
                    :disabled="repaying === order.order_no"
                    class="btn btn-sm btn-primary"
                  >
                    <span v-if="repaying === order.order_no">{{ t('recharge.paying') }}</span>
                    <span v-else>{{ t('recharge.continuePay') }}</span>
                  </button>
                  <span v-else-if="order.status === 'pending'" class="text-sm text-gray-400">
                    {{ t('recharge.expired') }}
                  </span>
                  <span v-else class="text-sm text-gray-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <Pagination
          v-if="total > pageSize"
          :page="page"
          :total="total"
          :page-size="pageSize"
          @update:page="handlePageChange"
        />
      </div>
    </div>
    <!-- Payment QR Code Modal -->
    <PaymentQRCodeModal
      :show="showPaymentModal"
      :order-no="paymentInfo.orderNo"
      :amount="paymentInfo.amount"
      :payment-amount="paymentInfo.paymentAmount"
      :qr-code="paymentInfo.qrCode"
      :qr-code-url="paymentInfo.qrCodeUrl"
      :mode="paymentInfo.mode"
      @close="showPaymentModal = false"
      @paid="handlePaymentSuccess"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onActivated } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { getRechargeOrders, repayRechargeOrder, type RechargeOrder } from '@/api/recharge'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import PaymentQRCodeModal from '@/components/common/PaymentQRCodeModal.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const orders = ref<RechargeOrder[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const repaying = ref<string | null>(null)

// Payment modal state
const showPaymentModal = ref(false)
const paymentInfo = reactive({
  orderNo: '',
  amount: 0,
  paymentAmount: 0,
  qrCode: '',
  qrCodeUrl: '',
  mode: ''
})

const loadOrders = async () => {
  loading.value = true
  try {
    const response = await getRechargeOrders(page.value, pageSize.value)
    orders.value = response.items || []
    total.value = response.total
  } catch (error) {
    appStore.showError(t('recharge.failedToLoad'))
    console.error('Error loading recharge orders:', error)
  } finally {
    loading.value = false
  }
}

const handlePageChange = (newPage: number) => {
  page.value = newPage
  loadOrders()
}

const getStatusClass = (status: string): string => {
  switch (status) {
    case 'paid':
      return 'badge-success'
    case 'pending':
      return 'badge-warning'
    case 'expired':
      return 'badge-gray'
    case 'refunded':
      return 'badge-info'
    default:
      return 'badge-gray'
  }
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  // 后端返回的时间可能是本地时间但被标记为 UTC
  // 如果时间字符串以 Z 结尾，移除它让 JavaScript 将其解析为本地时间
  let adjustedDateStr = dateStr
  if (dateStr.endsWith('Z')) {
    adjustedDateStr = dateStr.slice(0, -1)
  }
  const date = new Date(adjustedDateStr)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleString()
}

const isOrderExpired = (order: RechargeOrder): boolean => {
  if (!order.expired_at) return false
  return new Date() > new Date(order.expired_at)
}

const handleRepay = async (order: RechargeOrder) => {
  repaying.value = order.order_no
  try {
    const result = await repayRechargeOrder(order.order_no)
    paymentInfo.orderNo = result.order_no
    paymentInfo.amount = result.amount
    paymentInfo.paymentAmount = result.payment_amount
    paymentInfo.qrCode = result.qr_code
    paymentInfo.qrCodeUrl = result.qr_code_url
    paymentInfo.mode = result.mode
    showPaymentModal.value = true
  } catch (error: any) {
    const message = error.response?.data?.message || t('recharge.repayFailed')
    appStore.showError(message)
    loadOrders()
  } finally {
    repaying.value = null
  }
}

const handlePaymentSuccess = () => {
  showPaymentModal.value = false
  appStore.showSuccess(t('payment.paymentSuccess'))
  loadOrders()
}

onMounted(() => {
  loadOrders()
})

// Reload orders when component is activated (e.g., navigated back from payment)
onActivated(() => {
  loadOrders()
})
</script>
