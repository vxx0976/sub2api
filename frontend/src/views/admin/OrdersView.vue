<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-start justify-between gap-4">
          <!-- Left: Tabs + Search + filters -->
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <!-- Tabs -->
            <div class="flex rounded-lg bg-gray-100 p-1 dark:bg-dark-700">
              <button
                @click="activeTab = 'subscription'"
                :class="[
                  'px-4 py-1.5 text-sm font-medium rounded-md transition-colors',
                  activeTab === 'subscription'
                    ? 'bg-white text-primary-600 shadow dark:bg-dark-600 dark:text-primary-400'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
                ]"
              >
                {{ t('admin.orders.tabs.subscription') }}
              </button>
              <button
                @click="activeTab = 'recharge'"
                :class="[
                  'px-4 py-1.5 text-sm font-medium rounded-md transition-colors',
                  activeTab === 'recharge'
                    ? 'bg-white text-primary-600 shadow dark:bg-dark-600 dark:text-primary-400'
                    : 'text-gray-600 hover:text-gray-900 dark:text-gray-400 dark:hover:text-white'
                ]"
              >
                {{ t('admin.orders.tabs.recharge') }}
              </button>
            </div>

            <!-- Order Search -->
            <div class="relative w-full sm:w-64">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400"
              />
              <input
                v-model="searchKeyword"
                type="text"
                :placeholder="activeTab === 'subscription' ? t('admin.orders.searchPlaceholder') : t('admin.rechargeOrders.searchPlaceholder')"
                class="input pl-10"
                @input="debounceSearch"
              />
            </div>

            <!-- Status Filter -->
            <div class="w-full sm:w-40">
              <Select
                v-model="filters.status"
                :options="statusOptions"
                :placeholder="t('admin.orders.allStatus')"
                @change="applyFilters"
              />
            </div>
          </div>

          <!-- Right: Actions -->
          <div class="ml-auto flex flex-wrap items-center justify-end gap-3">
            <!-- Recharge Settings Button (only show on recharge tab) -->
            <button
              v-if="activeTab === 'recharge'"
              @click="showRechargeSettings = true"
              class="btn btn-secondary"
              :title="t('admin.settings.recharge.title')"
            >
              <Icon name="cog" size="md" class="mr-2" />
              {{ t('admin.orders.rechargeSettings') }}
            </button>
            <button
              @click="loadData"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <!-- Subscription Orders Table -->
      <template #table v-if="activeTab === 'subscription'">
        <DataTable :columns="subscriptionColumns" :data="subscriptionOrders" :loading="loading">
          <template #cell-order_no="{ value }">
            <span class="font-mono text-sm font-medium text-gray-900 dark:text-white">
              {{ value }}
            </span>
          </template>

          <template #cell-user="{ row }">
            <div class="flex items-center gap-2">
              <div
                class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30"
              >
                <span class="text-sm font-medium text-primary-700 dark:text-primary-300">
                  {{ row.user?.email?.charAt(0).toUpperCase() || '?' }}
                </span>
              </div>
              <span class="font-medium text-gray-900 dark:text-white">
                {{ row.user?.email || t('admin.orders.userPrefix', { id: row.user_id }) }}
              </span>
            </div>
          </template>

          <template #cell-group="{ row }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ row.group?.name || `Group #${row.group_id}` }}
            </span>
          </template>

          <template #cell-amount="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">
              ¥{{ value.toFixed(2) }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', getStatusClass(value)]">
              {{ t(`admin.orders.statusLabels.${value}`) }}
            </span>
          </template>

          <template #cell-pay_type="{ value }">
            <div v-if="value" class="flex items-center gap-1.5">
              <svg v-if="value === 'alipay'" class="h-5 w-5" viewBox="0 0 24 24" fill="none">
                <rect width="24" height="24" rx="4" fill="#1677FF"/>
                <path d="M18.5 14.2c-1.1-.4-2.3-.9-3.6-1.4.4-.9.7-1.9.9-2.9h-3.3v-1h3.8V8h-3.8V6h-1.6v2H7.2v.9h3.7v1H7.5v.9h5.8c-.2.7-.4 1.4-.7 2-1.4-.4-2.6-.7-3.6-.7-2.2 0-3.5 1.1-3.5 2.5 0 1.6 1.5 2.6 3.8 2.6 1.8 0 3.4-.7 4.8-2 1.4.7 2.9 1.4 4.4 2v-1.2z M8.8 16.2c-1.5 0-2.3-.5-2.3-1.3 0-.7.6-1.2 1.8-1.2.9 0 1.9.2 3.1.6-1 1.2-2 1.9-2.6 1.9z" fill="white"/>
              </svg>
              <svg v-else-if="value === 'wxpay'" class="h-5 w-5" viewBox="0 0 24 24" fill="none">
                <rect width="24" height="24" rx="4" fill="#07C160"/>
                <path d="M15.5 8.5c-2.5 0-4.5 1.7-4.5 3.8 0 2.1 2 3.8 4.5 3.8.5 0 1-.1 1.4-.2l1.4.8-.4-1.2c.9-.7 1.6-1.8 1.6-3.2 0-2.1-2-3.8-4.5-3.8zm-1.3 2.5c.3 0 .5.2.5.5s-.2.5-.5.5-.5-.2-.5-.5.2-.5.5-.5zm2.6 0c.3 0 .5.2.5.5s-.2.5-.5.5-.5-.2-.5-.5.2-.5.5-.5z" fill="white"/>
                <path d="M9.5 5C6.5 5 4 7 4 9.5c0 1.5.8 2.8 2 3.7l-.5 1.5 1.8-1c.5.2 1.1.3 1.7.3.2 0 .4 0 .5 0-.1-.3-.1-.7-.1-1 0-2.5 2.3-4.5 5.1-4.5.2 0 .3 0 .5 0C14.5 6.2 12.2 5 9.5 5zM7.3 8c.4 0 .7.3.7.7s-.3.7-.7.7-.7-.3-.7-.7.3-.7.7-.7zm4.4 0c.4 0 .7.3.7.7s-.3.7-.7.7-.7-.3-.7-.7.3-.7.7-.7z" fill="white"/>
              </svg>
              <span v-else class="text-sm text-gray-700 dark:text-gray-300">{{ value }}</span>
            </div>
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDate(value) }}
            </span>
          </template>

          <template #cell-paid_at="{ value }">
            <span v-if="value" class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDate(value) }}
            </span>
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.orders.noOrdersYet')"
              :description="t('admin.orders.noOrdersDesc')"
            />
          </template>
        </DataTable>
      </template>

      <!-- Recharge Orders Table -->
      <template #table v-else>
        <DataTable :columns="rechargeColumns" :data="rechargeOrders" :loading="loading">
          <template #cell-order_no="{ row }">
            <div>
              <span class="font-mono text-sm font-medium text-gray-900 dark:text-white">
                {{ row.order_no }}
              </span>
              <div v-if="row.trade_no" class="text-xs text-gray-500 dark:text-gray-400">
                {{ t('admin.rechargeOrders.tradeNo') }}: {{ row.trade_no }}
              </div>
            </div>
          </template>

          <template #cell-user="{ row }">
            <div class="flex items-center gap-2">
              <div
                class="flex h-8 w-8 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30"
              >
                <span class="text-sm font-medium text-primary-700 dark:text-primary-300">
                  {{ row.user?.email?.charAt(0).toUpperCase() || row.user?.username?.charAt(0).toUpperCase() || '?' }}
                </span>
              </div>
              <div>
                <div class="font-medium text-gray-900 dark:text-white">
                  {{ row.user?.username || `User #${row.user_id}` }}
                </div>
                <div v-if="row.user?.email" class="text-xs text-gray-500 dark:text-gray-400">
                  {{ row.user.email }}
                </div>
              </div>
            </div>
          </template>

          <template #cell-amount="{ value }">
            <span class="font-medium text-gray-900 dark:text-white">
              ¥{{ value.toFixed(2) }}
            </span>
          </template>

          <template #cell-credit_amount="{ value }">
            <span class="font-semibold text-primary-600 dark:text-primary-400">
              ¥{{ value.toFixed(2) }}
            </span>
          </template>

          <template #cell-multiplier="{ row }">
            <span class="text-sm text-gray-700 dark:text-gray-300">
              {{ row.multiplier.toFixed(1) }}×
            </span>
            <span
              v-if="row.multiplier > 1.0"
              class="ml-1 text-xs text-green-600 dark:text-green-400"
            >
              (+{{ ((row.multiplier - 1) * 100).toFixed(0) }}%)
            </span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', getStatusClass(value)]">
              {{ t(`recharge.statusLabels.${value}`) }}
            </span>
          </template>

          <template #cell-pay_type="{ value }">
            <div v-if="value" class="flex items-center gap-1.5">
              <svg v-if="value === 'alipay'" class="h-5 w-5" viewBox="0 0 24 24" fill="none">
                <rect width="24" height="24" rx="4" fill="#1677FF"/>
                <path d="M18.5 14.2c-1.1-.4-2.3-.9-3.6-1.4.4-.9.7-1.9.9-2.9h-3.3v-1h3.8V8h-3.8V6h-1.6v2H7.2v.9h3.7v1H7.5v.9h5.8c-.2.7-.4 1.4-.7 2-1.4-.4-2.6-.7-3.6-.7-2.2 0-3.5 1.1-3.5 2.5 0 1.6 1.5 2.6 3.8 2.6 1.8 0 3.4-.7 4.8-2 1.4.7 2.9 1.4 4.4 2v-1.2z M8.8 16.2c-1.5 0-2.3-.5-2.3-1.3 0-.7.6-1.2 1.8-1.2.9 0 1.9.2 3.1.6-1 1.2-2 1.9-2.6 1.9z" fill="white"/>
              </svg>
              <svg v-else-if="value === 'wxpay'" class="h-5 w-5" viewBox="0 0 24 24" fill="none">
                <rect width="24" height="24" rx="4" fill="#07C160"/>
                <path d="M15.5 8.5c-2.5 0-4.5 1.7-4.5 3.8 0 2.1 2 3.8 4.5 3.8.5 0 1-.1 1.4-.2l1.4.8-.4-1.2c.9-.7 1.6-1.8 1.6-3.2 0-2.1-2-3.8-4.5-3.8zm-1.3 2.5c.3 0 .5.2.5.5s-.2.5-.5.5-.5-.2-.5-.5.2-.5.5-.5zm2.6 0c.3 0 .5.2.5.5s-.2.5-.5.5-.5-.2-.5-.5.2-.5.5-.5z" fill="white"/>
                <path d="M9.5 5C6.5 5 4 7 4 9.5c0 1.5.8 2.8 2 3.7l-.5 1.5 1.8-1c.5.2 1.1.3 1.7.3.2 0 .4 0 .5 0-.1-.3-.1-.7-.1-1 0-2.5 2.3-4.5 5.1-4.5.2 0 .3 0 .5 0C14.5 6.2 12.2 5 9.5 5zM7.3 8c.4 0 .7.3.7.7s-.3.7-.7.7-.7-.3-.7-.7.3-.7.7-.7zm4.4 0c.4 0 .7.3.7.7s-.3.7-.7.7-.7-.3-.7-.7.3-.7.7-.7z" fill="white"/>
              </svg>
              <span v-else class="text-sm text-gray-700 dark:text-gray-300">{{ value }}</span>
            </div>
            <span v-else class="text-sm text-gray-400 dark:text-dark-500">-</span>
          </template>

          <template #cell-created_at="{ row }">
            <div class="text-sm text-gray-500 dark:text-dark-400">
              {{ formatDate(row.created_at) }}
            </div>
            <div v-if="row.paid_at" class="text-xs text-gray-400 dark:text-gray-500">
              {{ t('admin.rechargeOrders.paidAt') }}: {{ formatDate(row.paid_at) }}
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.rechargeOrders.noOrders')"
              :description="t('admin.rechargeOrders.noOrdersDesc')"
            />
          </template>
        </DataTable>
      </template>

      <!-- Pagination -->
      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Recharge Settings Modal -->
    <RechargeSettingsModal
      :show="showRechargeSettings"
      @close="showRechargeSettings = false"
      @saved="loadData"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { getAllRechargeOrders, type RechargeOrder } from '@/api/recharge'
import type { AdminOrder } from '@/api/admin/orders'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import RechargeSettingsModal from '@/components/admin/RechargeSettingsModal.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Active tab state
const activeTab = ref<'subscription' | 'recharge'>('subscription')
const showRechargeSettings = ref(false)

// Subscription order columns
const subscriptionColumns = computed<Column[]>(() => [
  { key: 'order_no', label: t('admin.orders.columns.orderNo'), sortable: true },
  { key: 'user', label: t('admin.orders.columns.user'), sortable: true },
  { key: 'group', label: t('admin.orders.columns.group'), sortable: true },
  { key: 'amount', label: t('admin.orders.columns.amount'), sortable: true },
  { key: 'status', label: t('admin.orders.columns.status'), sortable: true },
  { key: 'pay_type', label: t('admin.orders.columns.payType'), sortable: false },
  { key: 'created_at', label: t('admin.orders.columns.createdAt'), sortable: true },
  { key: 'paid_at', label: t('admin.orders.columns.paidAt'), sortable: true }
])

// Recharge order columns
const rechargeColumns = computed<Column[]>(() => [
  { key: 'order_no', label: t('recharge.orderNo'), sortable: true },
  { key: 'user', label: t('admin.rechargeOrders.user'), sortable: true },
  { key: 'amount', label: t('recharge.amount'), sortable: true },
  { key: 'credit_amount', label: t('recharge.creditAmount'), sortable: true },
  { key: 'multiplier', label: t('recharge.multiplier'), sortable: true },
  { key: 'status', label: t('recharge.status'), sortable: true },
  { key: 'pay_type', label: t('admin.rechargeOrders.payMethod'), sortable: false },
  { key: 'created_at', label: t('recharge.createdAt'), sortable: true }
])

const statusOptions = computed(() => [
  { value: '', label: t('admin.orders.allStatus') },
  { value: 'pending', label: t('admin.orders.statusLabels.pending') },
  { value: 'paid', label: t('admin.orders.statusLabels.paid') },
  { value: 'expired', label: t('admin.orders.statusLabels.expired') },
  { value: 'refunded', label: t('admin.orders.statusLabels.refunded') }
])

const subscriptionOrders = ref<AdminOrder[]>([])
const rechargeOrders = ref<RechargeOrder[]>([])
const loading = ref(false)
const searchKeyword = ref('')
let abortController: AbortController | null = null
let searchTimeout: ReturnType<typeof setTimeout> | null = null

const filters = reactive({
  status: '',
  search: ''
})

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const debounceSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  searchTimeout = setTimeout(() => {
    filters.search = searchKeyword.value
    applyFilters()
  }, 300)
}

const applyFilters = () => {
  pagination.page = 1
  loadData()
}

const loadData = async () => {
  if (activeTab.value === 'subscription') {
    await loadSubscriptionOrders()
  } else {
    await loadRechargeOrders()
  }
}

const loadSubscriptionOrders = async () => {
  if (abortController) {
    abortController.abort()
  }
  const requestController = new AbortController()
  abortController = requestController
  const { signal } = requestController

  loading.value = true
  try {
    const response = await adminAPI.orders.list(
      pagination.page,
      pagination.page_size,
      {
        status: filters.status || undefined,
        search: filters.search || undefined
      },
      { signal }
    )
    if (signal.aborted || abortController !== requestController) return
    subscriptionOrders.value = response.items
    pagination.total = response.total
    pagination.pages = response.pages
  } catch (error: any) {
    if (signal.aborted || error?.name === 'AbortError' || error?.code === 'ERR_CANCELED') {
      return
    }
    appStore.showError(t('admin.orders.failedToLoad'))
    console.error('Error loading orders:', error)
  } finally {
    if (abortController === requestController) {
      loading.value = false
      abortController = null
    }
  }
}

const loadRechargeOrders = async () => {
  loading.value = true
  try {
    const response = await getAllRechargeOrders(pagination.page, pagination.page_size, {
      keyword: filters.search || undefined,
      status: filters.status || undefined
    })
    rechargeOrders.value = response.items || []
    pagination.total = response.total
  } catch (error) {
    appStore.showError(t('admin.rechargeOrders.failedToLoad'))
    console.error('Error loading recharge orders:', error)
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page: number) => {
  pagination.page = page
  loadData()
}

const handlePageSizeChange = (pageSize: number) => {
  pagination.page_size = pageSize
  pagination.page = 1
  loadData()
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

// Watch tab changes to reload data
watch(activeTab, () => {
  pagination.page = 1
  searchKeyword.value = ''
  filters.status = ''
  filters.search = ''
  loadData()
})

onMounted(() => {
  loadData()
})

onUnmounted(() => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
})
</script>
