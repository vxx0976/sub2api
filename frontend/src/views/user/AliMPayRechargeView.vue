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

      <div v-if="loadingConfig" class="flex items-center justify-center py-12">
        <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
      </div>

      <div v-else-if="!config?.enabled" class="card p-12 text-center">
        <Icon name="exclamationCircle" size="xl" class="mx-auto text-gray-400 dark:text-dark-500" />
        <h3 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('recharge.disabled') }}
        </h3>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t('recharge.disabledDesc') }}
        </p>
      </div>

      <template v-else>
        <div class="card">
          <div class="p-6">
            <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('alimpay.formTitle') }}
            </h2>

            <div v-if="hasEffectiveTiers" class="mb-4 rounded-lg bg-amber-50 p-3 dark:bg-amber-900/20">
              <p class="mb-2 text-sm font-medium text-amber-800 dark:text-amber-300">{{ t('recharge.tierTitle') }}</p>
              <div class="space-y-1">
                <div v-for="(tier, idx) in config.tiers" :key="idx" class="flex items-center justify-between text-xs text-amber-700 dark:text-amber-400">
                  <span>{{ t('recharge.tierRange', { min: tier.min, max: tier.max ? tier.max : '∞' }) }}</span>
                  <span class="font-semibold">x{{ tier.multiplier }}</span>
                </div>
              </div>
            </div>

            <form @submit.prevent="handleCreateOrder" class="space-y-5">
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
                <p class="input-hint">{{ t('recharge.amountHint', { min: config.min_amount, max: config.max_amount }) }}</p>
              </div>

              <div v-if="creditPreview > 0" class="rounded-lg bg-emerald-50 p-3 dark:bg-emerald-900/20">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-emerald-700 dark:text-emerald-300">{{ t('recharge.creditPreview') }}</span>
                  <span class="text-lg font-bold text-emerald-600 dark:text-emerald-400">${{ creditPreview.toFixed(2) }}</span>
                </div>
                <div v-if="currentMultiplier > 1" class="mt-1 text-xs text-emerald-600 dark:text-emerald-400">
                  {{ t('recharge.multiplierApplied', { multiplier: currentMultiplier }) }}
                </div>
              </div>

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
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('recharge.orderHistory') }}</h2>
          </div>
          <div class="p-6">
            <div v-if="loadingOrders" class="flex items-center justify-center py-8">
              <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            </div>
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
                      ¥{{ order.payment_amount ? order.payment_amount.toFixed(2) : order.amount.toFixed(2) }}
                      <span class="text-xs text-gray-500 dark:text-dark-400"> → ${{ order.credit_amount.toFixed(2) }}</span>
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">{{ formatDateTime(order.created_at) }}</p>
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
                </div>
              </div>
            </div>
            <div v-else class="empty-state py-8">
              <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800">
                <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
              </div>
              <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('recharge.noOrders') }}</p>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- QR Code Payment Modal -->
    <Teleport to="body">
      <div
        v-if="showQRModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
        @click.self="closeQRModal"
      >
        <div class="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-800">
          <div class="mb-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('recharge.scanToPay') }}</h3>
            <button
              class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-dark-300"
              @click="closeQRModal"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- 强调：必须支付精确金额 -->
          <div class="mb-4 rounded-lg bg-amber-50 p-3 dark:bg-amber-900/20">
            <p class="text-xs font-medium text-amber-800 dark:text-amber-300">
              {{ t('alimpay.payExactAmountNotice') }}
            </p>
            <p class="mt-1 text-center text-2xl font-bold text-amber-700 dark:text-amber-300">
              ¥{{ qrPaymentAmount?.toFixed(2) }}
            </p>
          </div>

          <div class="mb-4 flex justify-center">
            <div class="rounded-xl border-2 border-gray-100 bg-white p-3 dark:border-dark-600">
              <img :src="qrCodeDataURL" alt="Payment QR Code" class="h-52 w-52" />
            </div>
          </div>

          <p class="mb-3 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('recharge.scanQRHint') }}
          </p>

          <div class="mt-4 flex items-center justify-center gap-2 text-sm text-gray-400 dark:text-dark-500">
            <svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            {{ t('recharge.waitingPayment') }}
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { alimpayAPI, type AliMPayConfig, type AliMPayOrderItem } from '@/api/alimpayOrder'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)

const loadingConfig = ref(true)
const config = ref<AliMPayConfig | null>(null)
const amount = ref<number | null>(null)
const submitting = ref(false)

const orders = ref<AliMPayOrderItem[]>([])
const loadingOrders = ref(false)

let pollTimer: ReturnType<typeof setInterval> | null = null

const showQRModal = ref(false)
const qrCodeDataURL = ref('')
const qrPaymentAmount = ref<number | null>(null)

const hasEffectiveTiers = computed(() => {
  if (!config.value?.tiers || config.value.tiers.length === 0) return false
  return config.value.tiers.some(t => t.multiplier !== 1)
})

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
  return Math.round((amount.value / config.value.selling_price) * currentMultiplier.value * 100) / 100
})

const canSubmit = computed(() => {
  if (!config.value || !amount.value) return false
  return amount.value >= config.value.min_amount && amount.value <= config.value.max_amount
})

const fetchConfig = async () => {
  loadingConfig.value = true
  try {
    config.value = await alimpayAPI.getConfig()
  } catch {
    appStore.showError(t('recharge.loadConfigFailed'))
  } finally {
    loadingConfig.value = false
  }
}

const fetchOrders = async () => {
  loadingOrders.value = true
  try {
    const res = await alimpayAPI.listOrders(1, 20)
    orders.value = res.items || []
  } catch {
    // silent
  } finally {
    loadingOrders.value = false
  }
}

const handleCreateOrder = async () => {
  if (!canSubmit.value || !amount.value) return
  submitting.value = true
  try {
    const result = await alimpayAPI.createOrder(amount.value)
    qrPaymentAmount.value = result.payment_amount
    try {
      qrCodeDataURL.value = await QRCode.toDataURL(result.qrcode_url, { width: 280, margin: 2 })
      showQRModal.value = true
    } catch {
      appStore.showError(t('recharge.createOrderFailed'))
    }
    appStore.showSuccess(t('recharge.orderCreated'))
    startPolling(result.order_no)
    await fetchOrders()
  } catch (err: any) {
    appStore.showError(err.response?.data?.detail || err.message || t('recharge.createOrderFailed'))
  } finally {
    submitting.value = false
  }
}

const closeQRModal = () => {
  showQRModal.value = false
}

const startPolling = (orderNo: string) => {
  stopPolling()
  let count = 0
  pollTimer = setInterval(async () => {
    count++
    if (count > 90) {
      stopPolling()
      return
    }
    try {
      const status = await alimpayAPI.getOrderStatus(orderNo)
      if (status.status === 'paid') {
        stopPolling()
        showQRModal.value = false
        appStore.showSuccess(t('recharge.paymentSuccess'))
        await authStore.refreshUser()
        await fetchOrders()
      } else if (status.status === 'expired') {
        stopPolling()
        showQRModal.value = false
        await fetchOrders()
      }
    } catch {
      // ignore
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
})

onUnmounted(() => {
  stopPolling()
})
</script>
