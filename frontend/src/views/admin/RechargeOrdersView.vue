<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header with Search -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.rechargeOrders.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.rechargeOrders.description') }}
          </p>
        </div>
      </div>

      <!-- Filters -->
      <div class="card">
        <div class="p-4">
          <div class="grid grid-cols-1 gap-4 md:grid-cols-4">
            <!-- Search -->
            <div class="md:col-span-2">
              <input
                v-model="filters.keyword"
                type="text"
                class="input"
                :placeholder="t('admin.rechargeOrders.searchPlaceholder')"
                @input="debouncedSearch"
              />
            </div>

            <!-- Status Filter -->
            <div>
              <select v-model="filters.status" class="input" @change="loadOrders">
                <option value="">{{ t('admin.rechargeOrders.allStatus') }}</option>
                <option value="pending">{{ t('recharge.statusLabels.pending') }}</option>
                <option value="paid">{{ t('recharge.statusLabels.paid') }}</option>
                <option value="expired">{{ t('recharge.statusLabels.expired') }}</option>
                <option value="refunded">{{ t('recharge.statusLabels.refunded') }}</option>
              </select>
            </div>

            <!-- Reset Button -->
            <div>
              <button @click="resetFilters" class="btn btn-secondary w-full">
                {{ t('common.reset') }}
              </button>
            </div>
          </div>
        </div>
      </div>

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
          {{ t('admin.rechargeOrders.noOrders') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('admin.rechargeOrders.noOrdersDesc') }}
        </p>
      </div>

      <!-- Orders Table -->
      <div v-else class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.orderNo') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('admin.rechargeOrders.user') }}
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
                  {{ t('admin.rechargeOrders.payMethod') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('recharge.createdAt') }}
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr
                v-for="order in orders"
                :key="order.id"
                class="hover:bg-gray-50 dark:hover:bg-dark-800/50"
              >
                <td class="px-4 py-3">
                  <div class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ order.order_no }}
                  </div>
                  <div v-if="order.trade_no" class="text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.rechargeOrders.tradeNo') }}: {{ order.trade_no }}
                  </div>
                </td>
                <td class="px-4 py-3">
                  <div class="text-sm text-gray-900 dark:text-white">
                    {{ order.user?.username || `User #${order.user_id}` }}
                  </div>
                  <div v-if="order.user?.email" class="text-xs text-gray-500 dark:text-gray-400">
                    {{ order.user.email }}
                  </div>
                </td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                  ¥{{ order.amount.toFixed(2) }}
                </td>
                <td class="px-4 py-3 text-sm font-semibold text-primary-600 dark:text-primary-400">
                  ¥{{ order.credit_amount.toFixed(2) }}
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-700 dark:text-gray-300">
                    {{ order.multiplier.toFixed(1) }}×
                  </span>
                  <span
                    v-if="order.multiplier > 1.0"
                    class="ml-1 text-xs text-green-600 dark:text-green-400"
                  >
                    (+{{ ((order.multiplier - 1) * 100).toFixed(0) }}%)
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span :class="['badge', getStatusClass(order.status)]">
                    {{ t(`recharge.statusLabels.${order.status}`) }}
                  </span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                  <span v-if="order.pay_type">
                    {{ order.pay_type === 'alipay' ? '支付宝' : order.pay_type === 'wxpay' ? '微信' : order.pay_type }}
                  </span>
                  <span v-else class="text-gray-400">-</span>
                </td>
                <td class="px-4 py-3">
                  <div class="text-sm text-gray-500 dark:text-dark-400">
                    {{ formatDate(order.created_at) }}
                  </div>
                  <div v-if="order.paid_at" class="text-xs text-gray-400 dark:text-gray-500">
                    {{ t('admin.rechargeOrders.paidAt') }}: {{ formatDate(order.paid_at) }}
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div v-if="total > 0" class="border-t border-gray-100 px-4 py-3 dark:border-dark-700">
          <Pagination
            :page="page"
            :total="total"
            :page-size="pageSize"
            @update:page="handlePageChange"
          />
        </div>
      </div>

      <!-- Statistics -->
      <div v-if="!loading && orders.length > 0" class="grid grid-cols-1 gap-6 md:grid-cols-4">
        <div class="card p-6">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.rechargeOrders.totalOrders') }}
          </div>
          <div class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">
            {{ total }}
          </div>
        </div>
        <div class="card p-6">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.rechargeOrders.totalAmount') }}
          </div>
          <div class="mt-2 text-2xl font-bold text-primary-600 dark:text-primary-400">
            ¥{{ calculateTotalAmount().toFixed(2) }}
          </div>
        </div>
        <div class="card p-6">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.rechargeOrders.totalCredit') }}
          </div>
          <div class="mt-2 text-2xl font-bold text-green-600 dark:text-green-400">
            ¥{{ calculateTotalCredit().toFixed(2) }}
          </div>
        </div>
        <div class="card p-6">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.rechargeOrders.averageMultiplier') }}
          </div>
          <div class="mt-2 text-2xl font-bold text-purple-600 dark:text-purple-400">
            {{ calculateAvgMultiplier().toFixed(2) }}×
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { getAllRechargeOrders, type RechargeOrder } from '@/api/recharge'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const orders = ref<RechargeOrder[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const filters = ref({
  keyword: '',
  status: ''
})

let searchTimeout: ReturnType<typeof setTimeout> | null = null

const debouncedSearch = () => {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    page.value = 1
    loadOrders()
  }, 500)
}

async function loadOrders() {
  loading.value = true
  try {
    const response = await getAllRechargeOrders(page.value, pageSize.value, {
      keyword: filters.value.keyword || undefined,
      status: filters.value.status || undefined
    })
    orders.value = response.items || []
    total.value = response.total
  } catch (error) {
    appStore.showError(t('admin.rechargeOrders.failedToLoad'))
    console.error('Error loading recharge orders:', error)
  } finally {
    loading.value = false
  }
}

function handlePageChange(newPage: number) {
  page.value = newPage
  loadOrders()
}

function resetFilters() {
  filters.value = {
    keyword: '',
    status: ''
  }
  page.value = 1
  loadOrders()
}

function getStatusClass(status: string): string {
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

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString()
}

function calculateTotalAmount(): number {
  return orders.value
    .filter(o => o.status === 'paid')
    .reduce((sum, o) => sum + o.amount, 0)
}

function calculateTotalCredit(): number {
  return orders.value
    .filter(o => o.status === 'paid')
    .reduce((sum, o) => sum + o.credit_amount, 0)
}

function calculateAvgMultiplier(): number {
  const paidOrders = orders.value.filter(o => o.status === 'paid')
  if (paidOrders.length === 0) return 0
  const sum = paidOrders.reduce((sum, o) => sum + o.multiplier, 0)
  return sum / paidOrders.length
}

onMounted(() => {
  loadOrders()
})
</script>
