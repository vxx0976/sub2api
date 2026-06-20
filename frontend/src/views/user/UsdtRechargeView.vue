<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- Current Balance Card -->
      <div class="card overflow-hidden">
        <div class="bg-gradient-to-br from-primary-500 to-primary-600 px-6 py-8 text-center">
          <div class="mb-4 inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-white/20 backdrop-blur-sm">
            <Icon name="creditCard" size="xl" class="text-white" />
          </div>
          <p class="text-sm font-medium text-primary-100">{{ t('recharge.currentBalance') }}</p>
          <p class="mt-2 text-4xl font-bold text-white">${{ user?.balance?.toFixed(2) || '0.00' }}</p>
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
        <h3 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('recharge.disabled') }}</h3>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">{{ t('recharge.disabledDesc') }}</p>
      </div>

      <template v-else>
        <div class="card">
          <div class="p-6">
            <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('usdt.formTitle') }}</h2>

            <form @submit.prevent="handleCreateOrder" class="space-y-5">
              <div v-if="config.chains && config.chains.length">
                <label class="input-label">{{ t('usdt.selectChain') }}</label>
                <div class="mt-1 grid grid-cols-3 gap-2">
                  <button
                    v-for="c in config.chains"
                    :key="c"
                    type="button"
                    :class="[
                      'rounded-lg border px-3 py-2 text-sm font-medium transition-colors',
                      selectedChain === c
                        ? 'border-primary-500 bg-primary-50 text-primary-600 dark:border-primary-400 dark:bg-primary-900/20 dark:text-primary-300'
                        : 'border-gray-200 text-gray-600 hover:border-gray-300 dark:border-dark-600 dark:text-dark-300'
                    ]"
                    @click="selectedChain = c"
                  >
                    {{ chainLabel(c) }}
                  </button>
                </div>
              </div>

              <div>
                <label for="amount" class="input-label">{{ t('usdt.amountLabel') }}</label>
                <div class="relative mt-1">
                  <input
                    id="amount"
                    v-model.number="amount"
                    type="number"
                    required
                    :min="usdtMin"
                    :max="usdtMax"
                    :step="0.01"
                    :placeholder="t('usdt.amountPlaceholder')"
                    :disabled="submitting"
                    class="input py-3 pr-16 text-lg"
                  />
                  <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-4">
                    <span class="text-gray-400 dark:text-dark-500 font-medium">USDT</span>
                  </div>
                </div>
                <p class="input-hint">{{ t('usdt.amountHint', { min: usdtMin.toFixed(2), max: usdtMax.toFixed(2) }) }}</p>
              </div>

              <div v-if="amount && amount > 0 && creditPreview > 0" class="rounded-lg bg-emerald-50 p-3 dark:bg-emerald-900/20">
                <div class="flex items-center justify-between">
                  <span class="text-sm text-emerald-700 dark:text-emerald-300">{{ t('recharge.creditPreview') }}</span>
                  <span class="text-lg font-bold text-emerald-600 dark:text-emerald-400">${{ creditPreview.toFixed(2) }}</span>
                </div>
              </div>

              <button type="submit" :disabled="!canSubmit || submitting" class="btn btn-primary w-full py-3">
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
                      {{ order.usdt_amount_str }} USDT
                      <span class="text-xs text-gray-500 dark:text-dark-400"> → ${{ order.credit_amount.toFixed(2) }}</span>
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">{{ chainLabel(order.chain) }} · {{ formatDateTime(order.created_at) }}</p>
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

    <!-- USDT Payment Modal -->
    <Teleport to="body">
      <div v-if="showQRModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
        <div class="mx-4 max-h-[92vh] w-full max-w-sm overflow-y-auto rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-800">
          <div class="mb-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('usdt.scanToPay') }}</h3>
            <button
              class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-dark-300"
              @click="closeQRModal"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Network warning -->
          <div class="mb-4 rounded-xl border border-red-200 bg-red-50 p-3 dark:border-red-700/50 dark:bg-red-900/10">
            <div class="flex items-start gap-2">
              <svg class="mt-0.5 h-4 w-4 flex-shrink-0 text-red-500 dark:text-red-400" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              <p class="text-sm font-medium text-red-700 dark:text-red-300">{{ t('usdt.networkWarning', { chain: chainLabel(qrChain) }) }}</p>
            </div>
          </div>

          <!-- Exact amount -->
          <div class="mb-4 rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-700/50 dark:bg-amber-900/10">
            <p class="text-center text-sm font-medium text-amber-800 dark:text-amber-200">{{ t('usdt.payExactAmountNotice') }}</p>
            <div class="mt-3 flex items-center justify-center gap-2">
              <p class="text-3xl font-bold tracking-tight text-amber-700 dark:text-amber-300">{{ qrUsdtAmount }} USDT</p>
              <button
                class="rounded-lg p-1.5 text-amber-600 hover:bg-amber-100 dark:text-amber-400 dark:hover:bg-amber-900/30"
                :title="t('usdt.copy')"
                @click="copy(qrUsdtAmount)"
              >
                <Icon name="clipboard" size="sm" />
              </button>
            </div>
            <p class="mt-2 text-center text-xs text-amber-700 dark:text-amber-400">{{ t('usdt.amountMustMatch') }}</p>
          </div>

          <!-- QR of address -->
          <div class="mb-3 flex justify-center">
            <div class="rounded-xl border-2 border-gray-100 bg-white p-3 dark:border-dark-600">
              <img :src="qrCodeDataURL" alt="USDT Address QR Code" class="h-48 w-48" />
            </div>
          </div>

          <!-- Address with copy -->
          <div class="mb-3">
            <label class="mb-1 block text-xs text-gray-500 dark:text-dark-400">{{ t('usdt.addressLabel', { chain: chainLabel(qrChain) }) }}</label>
            <div class="flex items-center gap-2 rounded-lg bg-gray-50 p-2 dark:bg-dark-700">
              <span class="flex-1 break-all font-mono text-xs text-gray-800 dark:text-gray-200">{{ qrAddress }}</span>
              <button
                class="flex-shrink-0 rounded-lg p-1.5 text-gray-500 hover:bg-gray-200 dark:text-gray-400 dark:hover:bg-dark-600"
                :title="t('usdt.copy')"
                @click="copy(qrAddress)"
              >
                <Icon name="clipboard" size="sm" />
              </button>
            </div>
          </div>

          <!-- Countdown -->
          <div v-if="remainingSeconds > 0" class="mb-2 flex items-center justify-center gap-2 text-sm font-medium text-orange-600 dark:text-orange-400">
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ t('usdt.countdown', { time: formatCountdown(remainingSeconds) }) }}
          </div>
          <div v-else class="mb-2 text-center text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('usdt.expiredHint') }}</div>

          <div class="mt-2 flex items-center justify-center gap-2 text-xs text-gray-400 dark:text-dark-500">
            <svg class="h-3.5 w-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            {{ t('usdt.waitingPayment') }}
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
import { usdtAPI, type UsdtConfig, type UsdtOrderItem } from '@/api/usdtOrder'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)

const loadingConfig = ref(true)
const config = ref<UsdtConfig | null>(null)
const amount = ref<number | null>(null)
const selectedChain = ref('')
const submitting = ref(false)

const CHAIN_LABELS: Record<string, string> = { trc20: 'TRC20 (TRON)', bep20: 'BEP20 (BSC)', ton: 'TON' }
const chainLabel = (c: string) => CHAIN_LABELS[c] || c.toUpperCase()

const orders = ref<UsdtOrderItem[]>([])
const loadingOrders = ref(false)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null

const showQRModal = ref(false)
const qrCodeDataURL = ref('')
const qrAddress = ref('')
const qrUsdtAmount = ref('')
const qrChain = ref('')
const remainingSeconds = ref(0)

const formatCountdown = (seconds: number) => {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

// 用户输入 USDT 数量；到账余额 = USDT × 汇率
const creditPreview = computed(() => {
  if (!config.value || !amount.value || config.value.rate <= 0) return 0
  return Math.round(amount.value * config.value.rate * 100) / 100
})

// 充值限额是余额(CNY)口径，换算成 USDT 输入范围
const usdtMin = computed(() => (config.value && config.value.rate > 0 ? config.value.min_amount / config.value.rate : 0))
const usdtMax = computed(() => (config.value && config.value.rate > 0 ? config.value.max_amount / config.value.rate : 0))

const canSubmit = computed(() => {
  if (!config.value || !amount.value || !selectedChain.value || config.value.rate <= 0) return false
  if (!config.value.chains?.includes(selectedChain.value)) return false
  return creditPreview.value >= config.value.min_amount && creditPreview.value <= config.value.max_amount
})

const copy = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(t('usdt.copied'))
  } catch {
    appStore.showError(t('usdt.copyFailed'))
  }
}

const fetchConfig = async () => {
  loadingConfig.value = true
  try {
    config.value = await usdtAPI.getConfig()
    if (config.value.chains?.length && !config.value.chains.includes(selectedChain.value)) {
      selectedChain.value = config.value.chains[0]
    }
  } catch {
    appStore.showError(t('recharge.loadConfigFailed'))
  } finally {
    loadingConfig.value = false
  }
}

const fetchOrders = async () => {
  loadingOrders.value = true
  try {
    const res = await usdtAPI.listOrders(1, 20)
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
    const result = await usdtAPI.createOrder(amount.value, selectedChain.value)
    qrAddress.value = result.address
    qrUsdtAmount.value = result.usdt_amount_str
    qrChain.value = result.chain
    try {
      qrCodeDataURL.value = await QRCode.toDataURL(result.address, { width: 280, margin: 2 })
      showQRModal.value = true
    } catch {
      appStore.showError(t('recharge.createOrderFailed'))
    }
    appStore.showSuccess(t('recharge.orderCreated'))
    startCountdown(result.expires_in || 1800)
    startPolling(result.order_no, result.expires_in || 1800)
    await fetchOrders()
  } catch (err: any) {
    appStore.showError(err.response?.data?.detail || err.message || t('recharge.createOrderFailed'))
  } finally {
    submitting.value = false
  }
}

const closeQRModal = () => {
  showQRModal.value = false
  stopCountdown()
}

const startCountdown = (seconds: number) => {
  stopCountdown()
  remainingSeconds.value = seconds
  countdownTimer = setInterval(() => {
    if (remainingSeconds.value <= 0) {
      stopCountdown()
      return
    }
    remainingSeconds.value--
  }, 1000)
}

const stopCountdown = () => {
  if (countdownTimer) {
    clearInterval(countdownTimer)
    countdownTimer = null
  }
}

const startPolling = (orderNo: string, expiresIn: number) => {
  stopPolling()
  const maxCount = Math.ceil(expiresIn / 5) + 2
  let count = 0
  pollTimer = setInterval(async () => {
    count++
    if (count > maxCount) {
      stopPolling()
      return
    }
    try {
      const status = await usdtAPI.getOrderStatus(orderNo)
      if (status.status === 'paid') {
        stopPolling()
        stopCountdown()
        showQRModal.value = false
        appStore.showSuccess(t('recharge.paymentSuccess'))
        await authStore.refreshUser()
        await fetchOrders()
      } else if (status.status === 'expired') {
        stopPolling()
        stopCountdown()
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
  stopCountdown()
})
</script>
