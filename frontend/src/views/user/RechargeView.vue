<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- Current Balance Card -->
      <div class="card overflow-hidden">
        <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
          <div
            class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm"
          >
            <Icon name="creditCard" size="xl" class="text-white" />
          </div>
          <p class="text-sm font-medium text-primary-100">{{ t('recharge.currentBalance') }}</p>
          <p class="mt-2 text-4xl font-bold text-white">
            ${{ user?.balance?.toFixed(2) || '0.00' }}
          </p>
        </div>
      </div>

      <!-- Loading Config -->
      <div v-if="loadingConfig" class="flex items-center justify-center py-12">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </div>

      <!-- Recharge Disabled -->
      <div v-else-if="!config?.enabled" class="card p-12 text-center">
        <Icon name="exclamationCircle" size="xl" class="mx-auto text-gray-400 dark:text-dark-500" />
        <h3 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('recharge.disabled') }}
        </h3>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('recharge.disabledDesc') }}
        </p>
      </div>

      <!-- Recharge Form -->
      <template v-else>
        <div class="card">
          <div class="p-6">
            <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('recharge.formTitle') }}
            </h2>

            <!-- Price Info -->
            <div v-if="config.selling_price > 0" class="mb-4 rounded-lg bg-blue-50 p-3 text-sm text-blue-700 dark:bg-blue-900/20 dark:text-blue-300">
              {{ t('recharge.priceInfo', { price: config.selling_price }) }}
            </div>

            <!-- Tier Info -->
            <div v-if="config.tiers && config.tiers.length > 0" class="mb-4 rounded-lg bg-amber-50 p-3 dark:bg-amber-900/20">
              <p class="mb-2 text-sm font-medium text-amber-800 dark:text-amber-300">{{ t('recharge.tierTitle') }}</p>
              <div class="space-y-1">
                <div v-for="(tier, idx) in config.tiers" :key="idx" class="flex items-center justify-between text-xs text-amber-700 dark:text-amber-400">
                  <span>
                    {{ t('recharge.tierRange', { min: tier.min, max: tier.max ? tier.max : '∞' }) }}
                  </span>
                  <span class="font-semibold">x{{ tier.multiplier }}</span>
                </div>
              </div>
            </div>

            <form @submit.prevent="handleCreateOrder" class="space-y-5">
              <!-- Amount Input -->
              <div>
                <label for="amount" class="input-label">{{ t('recharge.amountLabel') }}</label>
                <div class="relative mt-1">
                  <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
                    <span class="text-gray-400 dark:text-dark-500 font-medium">¥</span>
                  </div>
                  <input
                    id="amount"
                    v-model.number="amount"
                    type="number"
                    required
                    :min="config.min_amount"
                    :max="config.max_amount"
                    :step="0.01"
                    :placeholder="t('recharge.amountPlaceholder', { min: config.min_amount, max: config.max_amount })"
                    :disabled="submitting"
                    class="input py-3 pl-10 text-lg"
                  />
                </div>
                <p class="input-hint">
                  {{ t('recharge.amountHint', { min: config.min_amount, max: config.max_amount }) }}
                </p>
              </div>

              <!-- Credit Preview -->
              <div v-if="creditPreview > 0" class="rounded-lg bg-emerald-50 p-3 dark:bg-emerald-900/20">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-emerald-700 dark:text-emerald-300">{{ t('recharge.creditPreview') }}</span>
                  <span class="text-lg font-bold text-emerald-600 dark:text-emerald-400">${{ creditPreview.toFixed(2) }}</span>
                </div>
                <div v-if="currentMultiplier > 1" class="mt-1 text-xs text-emerald-600 dark:text-emerald-400">
                  {{ t('recharge.multiplierApplied', { multiplier: currentMultiplier }) }}
                </div>
              </div>

              <!-- Pay Type Selection -->
              <div>
                <label class="input-label">{{ t('recharge.payTypeLabel') }}</label>
                <div class="mt-2 grid grid-cols-2 gap-3">
                  <button
                    v-for="pt in config.pay_types"
                    :key="pt"
                    type="button"
                    :disabled="submitting"
                    :class="[
                      'flex items-center justify-center gap-2 rounded-xl border-2 px-4 py-3 text-sm font-medium transition-all',
                      payType === pt
                        ? 'border-primary-500 bg-primary-50 text-primary-700 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-300'
                        : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-dark-600 dark:text-dark-300 dark:hover:border-dark-500'
                    ]"
                    @click="payType = pt"
                  >
                    <span v-if="pt === 'alipay'" class="text-blue-500">
                      <svg class="h-5 w-5" viewBox="0 0 1024 1024" fill="currentColor"><path d="M789.024 647.104c-49.088-20.928-103.104-43.392-106.944-45.184 22.336-40.192 40.128-84.928 51.776-133.12h-150.72V416.512h181.12v-30.784h-181.12V301.44h-78.592s-5.568 0-5.568 8.128v76.16H320v30.784h178.976V468.8H352.96v30.784h268.352c-9.792 35.136-23.424 67.84-40.192 96.96a1603.84 1603.84 0 0 0-196.608-56.768C266.816 514.816 192 571.456 192 647.104c0 75.648 71.616 128.832 199.232 128.832 87.36 0 172.288-35.648 238.976-98.432 43.392 23.424 76.288 42.688 76.288 42.688l82.528-73.088zM371.968 736.576c-112.448 0-140.032-45.12-140.032-80.256 0-47.552 52.416-90.24 127.616-90.24 55.104 0 117.888 14.272 178.816 39.616-52.352 77.376-114.496 130.88-166.4 130.88z"/></svg>
                    </span>
                    <span v-else-if="pt === 'wxpay'" class="text-green-500">
                      <svg class="h-5 w-5" viewBox="0 0 1024 1024" fill="currentColor"><path d="M690.1 377.4c5.9 0 11.8.2 17.6.5-15.8-73.2-88.3-127.8-174.5-127.8-95.5 0-173.3 63.6-173.3 142.7 0 46.4 25.3 84.7 67.2 114.5l-16.8 50.4 58.4-29.2c21.1 4.2 37.9 8.4 58.4 8.4 5.8 0 11.5-.2 17.1-.7-3.6-12.1-5.6-24.8-5.6-37.9 0-67.3 64.2-120.9 151.5-120.9zM563.4 318.8c12.6 0 21.1 8.4 21.1 21.1 0 12.6-8.4 21.1-21.1 21.1-12.6 0-25.2-8.4-25.2-21.1 0-12.6 12.6-21.1 25.2-21.1zm-116.8 42.2c-12.6 0-25.2-8.4-25.2-21.1 0-12.6 12.6-21.1 25.2-21.1 12.6 0 21.1 8.4 21.1 21.1 0 12.6-8.5 21.1-21.1 21.1zm384 134.7c0-67.3-67.2-122-147.6-122-84.7 0-147.6 54.7-147.6 122s62.9 122 147.6 122c16.8 0 37.9-4.2 54.7-8.4l46.4 25.2-12.6-42.2c33.7-25.2 59.1-58.9 59.1-96.6zm-196.8-21.1c-8.4 0-16.8-8.4-16.8-16.8 0-8.4 8.4-16.8 16.8-16.8 12.6 0 21.1 8.4 21.1 16.8 0 8.4-8.5 16.8-21.1 16.8zm100.8 0c-8.4 0-16.8-8.4-16.8-16.8 0-8.4 8.4-16.8 16.8-16.8 12.6 0 21.1 8.4 21.1 16.8 0 8.4-8.5 16.8-21.1 16.8z"/></svg>
                    </span>
                    {{ t(`recharge.payType_${pt}`) }}
                  </button>
                </div>
              </div>

              <!-- Submit Button -->
              <button
                type="submit"
                :disabled="!canSubmit || submitting"
                class="btn btn-primary w-full py-3"
              >
                <svg v-if="submitting" class="-ml-1 mr-2 h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                {{ submitting ? t('recharge.submitting') : t('recharge.submitButton') }}
              </button>
            </form>
          </div>
        </div>

        <!-- Order History -->
        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('recharge.orderHistory') }}
            </h2>
          </div>
          <div class="p-6">
            <!-- Loading -->
            <div v-if="loadingOrders" class="flex items-center justify-center py-8">
              <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            </div>

            <!-- Order List -->
            <div v-else-if="orders.length > 0" class="space-y-3">
              <div
                v-for="order in orders"
                :key="order.order_no"
                class="flex items-center justify-between rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
              >
                <div class="flex items-center gap-4">
                  <div
                    :class="[
                      'flex h-10 w-10 items-center justify-center rounded-xl',
                      order.status === 'paid'
                        ? 'bg-emerald-100 dark:bg-emerald-900/30'
                        : order.status === 'pending'
                          ? 'bg-amber-100 dark:bg-amber-900/30'
                          : 'bg-gray-100 dark:bg-dark-700'
                    ]"
                  >
                    <Icon
                      :name="order.status === 'paid' ? 'checkCircle' : order.status === 'pending' ? 'clock' : 'exclamationCircle'"
                      size="md"
                      :class="[
                        order.status === 'paid'
                          ? 'text-emerald-600 dark:text-emerald-400'
                          : order.status === 'pending'
                            ? 'text-amber-600 dark:text-amber-400'
                            : 'text-gray-400 dark:text-dark-500'
                      ]"
                    />
                  </div>
                  <div>
                    <p class="text-sm font-medium text-gray-900 dark:text-white">
                      ¥{{ order.amount.toFixed(2) }}
                      <span class="text-xs text-gray-500 dark:text-dark-400"> → ${{ order.credit_amount.toFixed(2) }}</span>
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">
                      {{ formatDateTime(order.created_at) }}
                    </p>
                  </div>
                </div>
                <div class="text-right">
                  <span
                    :class="[
                      'inline-flex rounded-full px-2 py-0.5 text-xs font-medium',
                      order.status === 'paid'
                        ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                        : order.status === 'pending'
                          ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                          : 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-dark-400'
                    ]"
                  >
                    {{ t(`recharge.status_${order.status}`) }}
                  </span>
                  <p class="mt-1 text-xs text-gray-400 dark:text-dark-500">
                    {{ order.pay_type === 'alipay' ? t('recharge.payType_alipay') : t('recharge.payType_wxpay') }}
                  </p>
                </div>
              </div>
            </div>

            <!-- Empty State -->
            <div v-else class="empty-state py-8">
              <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800">
                <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
              </div>
              <p class="text-sm text-gray-500 dark:text-dark-400">
                {{ t('recharge.noOrders') }}
              </p>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { rechargeAPI, type RechargeConfig, type RechargeOrderItem } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)

const loadingConfig = ref(true)
const config = ref<RechargeConfig | null>(null)
const amount = ref<number | null>(null)
const payType = ref('alipay')
const submitting = ref(false)

const orders = ref<RechargeOrderItem[]>([])
const loadingOrders = ref(false)

// Polling for pending order
let pollTimer: ReturnType<typeof setInterval> | null = null
const pendingOrderNo = ref<string | null>(null)

const currentMultiplier = computed(() => {
  if (!config.value?.tiers || !amount.value) return 1
  for (const tier of config.value.tiers) {
    if (amount.value >= tier.min && (tier.max === null || amount.value <= tier.max)) {
      return tier.multiplier
    }
  }
  return 1
})

const creditPreview = computed(() => {
  if (!config.value || !amount.value || config.value.selling_price <= 0) return 0
  return Math.round(amount.value / config.value.selling_price * currentMultiplier.value * 100) / 100
})

const canSubmit = computed(() => {
  if (!config.value || !amount.value || !payType.value) return false
  return amount.value >= config.value.min_amount && amount.value <= config.value.max_amount
})

const fetchConfig = async () => {
  loadingConfig.value = true
  try {
    config.value = await rechargeAPI.getConfig()
    if (config.value.pay_types?.length > 0) {
      payType.value = config.value.pay_types[0]
    }
  } catch (err: any) {
    appStore.showError(t('recharge.loadConfigFailed'))
  } finally {
    loadingConfig.value = false
  }
}

const fetchOrders = async () => {
  loadingOrders.value = true
  try {
    const res = await rechargeAPI.listOrders(1, 20)
    orders.value = res.items || []
  } catch {
    // silently fail
  } finally {
    loadingOrders.value = false
  }
}

const handleCreateOrder = async () => {
  if (!canSubmit.value || !amount.value) return
  submitting.value = true
  try {
    const result = await rechargeAPI.createOrder(amount.value, payType.value)
    pendingOrderNo.value = result.order_no
    // Open payment page in new window
    window.open(result.pay_url, '_blank')
    appStore.showSuccess(t('recharge.orderCreated'))
    // Start polling for payment result
    startPolling(result.order_no)
    // Refresh order list
    await fetchOrders()
  } catch (err: any) {
    appStore.showError(err.response?.data?.detail || err.message || t('recharge.createOrderFailed'))
  } finally {
    submitting.value = false
  }
}

const startPolling = (orderNo: string) => {
  stopPolling()
  let count = 0
  pollTimer = setInterval(async () => {
    count++
    if (count > 60) { // Stop after 5 minutes (5s * 60)
      stopPolling()
      return
    }
    try {
      const status = await rechargeAPI.getOrderStatus(orderNo)
      if (status.status === 'paid') {
        stopPolling()
        appStore.showSuccess(t('recharge.paymentSuccess'))
        await authStore.refreshUser()
        await fetchOrders()
        pendingOrderNo.value = null
      } else if (status.status === 'expired') {
        stopPolling()
        pendingOrderNo.value = null
        await fetchOrders()
      }
    } catch {
      // ignore polling errors
    }
  }, 5000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  await fetchConfig()
  fetchOrders()

  // If returning from payment, check for pending order
  if (route.query.from === 'payment') {
    appStore.showInfo(t('recharge.checkingPayment'))
    // Refresh orders and user balance
    await fetchOrders()
    await authStore.refreshUser()
  }
})

onUnmounted(() => {
  stopPolling()
})
</script>
