<template>
  <AppLayout>
    <div class="space-y-8">
      <!-- Plan Orders Section -->
      <div>
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('userOrders.planOrders') }}</h2>
          <router-link to="/plans" class="btn btn-primary btn-sm">
            <Icon name="plus" size="sm" />
            {{ t('userOrders.purchasePlan') }}
          </router-link>
        </div>
        <div v-if="planLoading" class="flex justify-center py-8">
          <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
        </div>
        <div v-else-if="planOrders.length === 0" class="card p-8 text-center">
          <div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
            <Icon name="document" size="lg" class="text-gray-400" />
          </div>
          <h3 class="mb-1 text-base font-semibold text-gray-900 dark:text-white">{{ t('userOrders.noOrders') }}</h3>
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('userOrders.noOrdersDesc') }}</p>
        </div>
        <div v-else class="card overflow-hidden">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('userOrders.orderNo') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('userOrders.planName') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('userOrders.amount') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('userOrders.status') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('userOrders.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('userOrders.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="order in planOrders" :key="order.id" class="hover:bg-gray-50 dark:hover:bg-dark-800/50">
                <td class="px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">{{ order.order_no }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">{{ order.group?.name || `Group #${order.group_id}` }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">¥{{ order.amount.toFixed(2) }}</td>
                <td class="px-4 py-3">
                  <span :class="['badge', getStatusClass(order.status)]">{{ t(`userOrders.statusLabels.${order.status}`) }}</span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-400">{{ formatDate(order.created_at) }}</td>
                <td class="px-4 py-3">
                  <button v-if="order.status === 'pending' && !isOrderExpired(order)" @click="handlePlanRepay(order)" :disabled="repaying === order.order_no" class="btn btn-sm btn-primary">
                    <span v-if="repaying === order.order_no">{{ t('userOrders.paying') }}</span>
                    <span v-else>{{ t('userOrders.continuePay') }}</span>
                  </button>
                  <span v-else-if="order.status === 'pending'" class="text-sm text-gray-400">{{ t('userOrders.expired') }}</span>
                  <span v-else class="text-sm text-gray-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Pagination v-if="planTotal > planPageSize" :page="planPage" :total="planTotal" :page-size="planPageSize" @update:page="handlePlanPageChange" class="mt-4" />
      </div>

      <!-- Recharge Orders Section -->
      <div>
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('userOrders.rechargeOrders') }}</h2>
          <router-link to="/plans" class="btn btn-primary btn-sm">
            <Icon name="plus" size="sm" />
            {{ t('recharge.rechargeNow') }}
          </router-link>
        </div>
        <div v-if="rechargeLoading" class="flex justify-center py-8">
          <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
        </div>
        <div v-else-if="rechargeOrders.length === 0" class="card p-8 text-center">
          <div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
            <Icon name="document" size="lg" class="text-gray-400" />
          </div>
          <h3 class="mb-1 text-base font-semibold text-gray-900 dark:text-white">{{ t('recharge.noOrders') }}</h3>
          <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('recharge.noOrdersDesc') }}</p>
        </div>
        <div v-else class="card overflow-hidden">
          <table class="w-full">
            <thead class="bg-gray-50 dark:bg-dark-800">
              <tr>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('recharge.orderNo') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('recharge.amount') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('recharge.creditAmount') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('recharge.multiplier') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('recharge.status') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('recharge.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-sm font-medium text-gray-500 dark:text-dark-400">{{ t('userOrders.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-for="order in rechargeOrders" :key="order.id" class="hover:bg-gray-50 dark:hover:bg-dark-800/50">
                <td class="px-4 py-3 text-sm font-medium text-gray-900 dark:text-white">{{ order.order_no }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">¥{{ order.amount.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm font-semibold text-primary-600 dark:text-primary-400">${{ order.credit_amount.toFixed(2) }}</td>
                <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-300">
                  {{ order.multiplier.toFixed(1) }}×
                  <span v-if="order.multiplier > 1.0" class="text-xs text-green-600 dark:text-green-400">(+{{ ((order.multiplier - 1) * 100).toFixed(0) }}%)</span>
                </td>
                <td class="px-4 py-3">
                  <span :class="['badge', getStatusClass(order.status)]">{{ t(`recharge.statusLabels.${order.status}`) }}</span>
                </td>
                <td class="px-4 py-3 text-sm text-gray-500 dark:text-dark-400">{{ formatDate(order.created_at) }}</td>
                <td class="px-4 py-3">
                  <button v-if="order.status === 'pending' && !isRechargeOrderExpired(order)" @click="handleRechargeRepay(order)" :disabled="repaying === order.order_no" class="btn btn-sm btn-primary">
                    <span v-if="repaying === order.order_no">{{ t('recharge.paying') }}</span>
                    <span v-else>{{ t('recharge.continuePay') }}</span>
                  </button>
                  <span v-else-if="order.status === 'pending'" class="text-sm text-gray-400">{{ t('recharge.expired') }}</span>
                  <span v-else class="text-sm text-gray-400">-</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Pagination v-if="rechargeTotal > rechargePageSize" :page="rechargePage" :total="rechargeTotal" :page-size="rechargePageSize" @update:page="handleRechargePageChange" class="mt-4" />
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onActivated } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { getOrders, repayOrder, type Order } from '@/api/plans'
import { getRechargeOrders, repayRechargeOrder, type RechargeOrder } from '@/api/recharge'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Plan orders state
const planLoading = ref(true)
const planOrders = ref<Order[]>([])
const planPage = ref(1)
const planPageSize = ref(10)
const planTotal = ref(0)

// Recharge orders state
const rechargeLoading = ref(true)
const rechargeOrders = ref<RechargeOrder[]>([])
const rechargePage = ref(1)
const rechargePageSize = ref(10)
const rechargeTotal = ref(0)

// Shared state
const repaying = ref<string | null>(null)

const loadPlanOrders = async () => {
  planLoading.value = true
  try {
    const response = await getOrders(planPage.value, planPageSize.value)
    planOrders.value = response.items || []
    planTotal.value = response.total
  } catch (error) {
    appStore.showError(t('userOrders.failedToLoad'))
    console.error('Error loading plan orders:', error)
  } finally {
    planLoading.value = false
  }
}

const loadRechargeOrders = async () => {
  rechargeLoading.value = true
  try {
    const response = await getRechargeOrders(rechargePage.value, rechargePageSize.value)
    rechargeOrders.value = response.items || []
    rechargeTotal.value = response.total
  } catch (error) {
    appStore.showError(t('recharge.failedToLoad'))
    console.error('Error loading recharge orders:', error)
  } finally {
    rechargeLoading.value = false
  }
}

const handlePlanPageChange = (newPage: number) => {
  planPage.value = newPage
  loadPlanOrders()
}

const handleRechargePageChange = (newPage: number) => {
  rechargePage.value = newPage
  loadRechargeOrders()
}

const getStatusClass = (status: string): string => {
  switch (status) {
    case 'paid': return 'badge-success'
    case 'pending': return 'badge-warning'
    case 'expired': return 'badge-gray'
    case 'refunded': return 'badge-info'
    default: return 'badge-gray'
  }
}

const formatDate = (dateStr: string): string => {
  return new Date(dateStr).toLocaleString()
}

const isOrderExpired = (order: Order): boolean => {
  if (!order.expired_at) return false
  return new Date() > new Date(order.expired_at)
}

const isRechargeOrderExpired = (order: RechargeOrder): boolean => {
  if (!order.expired_at) return false
  return new Date() > new Date(order.expired_at)
}

const handlePlanRepay = async (order: Order) => {
  repaying.value = order.order_no
  try {
    const result = await repayOrder(order.order_no)
    window.location.href = result.pay_url
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('userOrders.repayFailed'))
    loadPlanOrders()
  } finally {
    repaying.value = null
  }
}

const handleRechargeRepay = async (order: RechargeOrder) => {
  repaying.value = order.order_no
  try {
    const result = await repayRechargeOrder(order.order_no)
    window.location.href = result.pay_url
  } catch (error: any) {
    appStore.showError(error.response?.data?.message || t('recharge.repayFailed'))
    loadRechargeOrders()
  } finally {
    repaying.value = null
  }
}

onMounted(() => {
  loadPlanOrders()
  loadRechargeOrders()
})

onActivated(() => {
  loadPlanOrders()
  loadRechargeOrders()
})
</script>
