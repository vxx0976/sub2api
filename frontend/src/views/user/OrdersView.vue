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
          {{ t('userOrders.noOrders') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('userOrders.noOrdersDesc') }}
        </p>
        <router-link to="/plans" class="btn btn-primary mt-4">
          {{ t('userOrders.purchasePlan') }}
        </router-link>
      </div>

      <!-- Orders List -->
      <div v-else class="space-y-6">
        <!-- Header -->
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('userOrders.myOrders') }}
          </h2>
          <router-link to="/plans" class="btn btn-primary">
            <Icon name="plus" size="sm" />
            {{ t('userOrders.purchasePlan') }}
          </router-link>
        </div>

        <!-- Orders Table -->
        <div class="card overflow-hidden">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('userOrders.orderNo') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('userOrders.planName') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('userOrders.amount') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('userOrders.status') }}
                </th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">
                  {{ t('userOrders.createdAt') }}
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
                  {{ order.group?.name || `Group #${order.group_id}` }}
                </td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                  ¥{{ order.amount.toFixed(2) }}
                </td>
                <td class="px-4 py-3">
                  <span :class="['badge', getStatusClass(order.status)]">
                    {{ t(`userOrders.statusLabels.${order.status}`) }}
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
                    <span v-if="repaying === order.order_no">{{ t('userOrders.paying') }}</span>
                    <span v-else>{{ t('userOrders.continuePay') }}</span>
                  </button>
                  <span v-else-if="order.status === 'pending'" class="text-sm text-gray-400">
                    {{ t('userOrders.expired') }}
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
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { getOrders, repayOrder, type Order } from '@/api/plans'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const orders = ref<Order[]>([])
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const repaying = ref<string | null>(null)

const loadOrders = async () => {
  loading.value = true
  try {
    const response = await getOrders(page.value, pageSize.value)
    orders.value = response.items || []
    total.value = response.total
  } catch (error) {
    appStore.showError(t('userOrders.failedToLoad'))
    console.error('Error loading orders:', error)
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
  const date = new Date(dateStr)
  return date.toLocaleString()
}

const isOrderExpired = (order: Order): boolean => {
  if (!order.expired_at) return false
  return new Date() > new Date(order.expired_at)
}

const handleRepay = async (order: Order) => {
  repaying.value = order.order_no
  try {
    const result = await repayOrder(order.order_no)
    // Redirect to payment page
    window.location.href = result.pay_url
  } catch (error: any) {
    const message = error.response?.data?.message || t('userOrders.repayFailed')
    appStore.showError(message)
    // Refresh orders to update status
    loadOrders()
  } finally {
    repaying.value = null
  }
}

onMounted(() => {
  loadOrders()
})
</script>
