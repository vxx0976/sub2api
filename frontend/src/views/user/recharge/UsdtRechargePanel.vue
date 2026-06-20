<template>
  <div class="space-y-5">
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
        <label for="usdt-amount" class="input-label">{{ t('usdt.amountLabel') }}</label>
        <div class="relative mt-1">
          <input
            id="usdt-amount"
            v-model.number="amount"
            type="number"
            required
            :min="usdtMin"
            :max="usdtMax > 0 ? usdtMax : undefined"
            :step="0.01"
            :placeholder="t('usdt.amountPlaceholder')"
            :disabled="submitting"
            class="input py-3 pr-16 text-lg"
          />
          <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-4">
            <span class="text-gray-400 dark:text-dark-500 font-medium">USDT</span>
          </div>
        </div>
        <p class="input-hint">
          {{ usdtMax > 0
            ? t('usdt.amountHint', { min: usdtMin, max: usdtMax })
            : t('usdt.amountHintNoMax', { min: usdtMin }) }}
        </p>
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
              <p class="text-3xl font-bold tracking-tight text-amber-700 dark:text-amber-300">{{ trimUsdt(qrUsdtAmount) }} USDT</p>
              <button
                class="rounded-lg p-1.5 text-amber-600 hover:bg-amber-100 dark:text-amber-400 dark:hover:bg-amber-900/30"
                :title="t('usdt.copy')"
                @click="copy(trimUsdt(qrUsdtAmount))"
              >
                <Icon name="clipboard" size="sm" />
              </button>
            </div>
            <p class="mt-2 text-center text-xs text-amber-700 dark:text-amber-400">{{ t('usdt.amountMustMatch') }}</p>
          </div>

          <!-- QR of address -->
          <div class="mb-3 flex justify-center">
            <div class="rounded-xl border-2 border-gray-100 bg-white p-3 dark:border-dark-600">
              <img :src="qrCodeDataURL" alt="USDT Address QR Code" class="h-64 w-64" style="image-rendering: pixelated" />
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

          <!-- 耐心等待 + 客服提示 -->
          <div class="mb-2 rounded-lg bg-gray-50 p-2.5 text-center text-xs leading-relaxed text-gray-500 dark:bg-dark-700/50 dark:text-dark-400">
            {{ t('usdt.waitNotice') }}
          </div>

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
  </div>
</template>

<script setup lang="ts">
/**
 * USDT（自建多链）充值面板
 * 由统一充值页 RechargeView 在「USDT」标签下渲染。config 由父组件传入。
 */
import { ref, computed, watch, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { usdtAPI, type UsdtConfig } from '@/api/usdtOrder'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ config: UsdtConfig }>()
const emit = defineEmits<{ (e: 'paid'): void }>()

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const amount = ref<number | null>(null)
const selectedChain = ref('')
const submitting = ref(false)

const CHAIN_LABELS: Record<string, string> = { trc20: 'TRC20 (TRON)', bep20: 'BEP20 (BSC)', ton: 'TON' }
const chainLabel = (c: string) => CHAIN_LABELS[c] || c.toUpperCase()

// 去掉金额尾部多余的 0（1.500000 → 1.5；1.000000 → 1；1.55 → 1.55）
const trimUsdt = (s: string) => {
  if (!s || s.indexOf('.') < 0) return s || ''
  return s.replace(/0+$/, '').replace(/\.$/, '')
}

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
  if (!amount.value || props.config.rate <= 0) return 0
  return Math.round(amount.value * props.config.rate * 100) / 100
})

// 限额直接是 USDT 口径（min 默认 0.1，max=0 表示无上限）
const usdtMin = computed(() => (props.config.min_amount > 0 ? props.config.min_amount : 0.1))
const usdtMax = computed(() => props.config.max_amount)

const canSubmit = computed(() => {
  if (!amount.value || !selectedChain.value || props.config.rate <= 0) return false
  if (!props.config.chains?.includes(selectedChain.value)) return false
  if (amount.value < usdtMin.value) return false
  if (usdtMax.value > 0 && amount.value > usdtMax.value) return false
  return true
})

const ensureChain = () => {
  if (props.config.chains?.length && !props.config.chains.includes(selectedChain.value)) {
    selectedChain.value = props.config.chains[0]
  }
}
watch(() => props.config, ensureChain, { immediate: true })

const copy = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    appStore.showSuccess(t('usdt.copied'))
  } catch {
    appStore.showError(t('usdt.copyFailed'))
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
      // 高纠错 + 足够留白 + 高分辨率，避免缩放发虚导致钱包扫不出
      qrCodeDataURL.value = await QRCode.toDataURL(result.address, { errorCorrectionLevel: 'H', margin: 4, width: 512 })
      showQRModal.value = true
    } catch {
      // 二维码生成失败时中止，避免误报「下单成功」并启动看不见的轮询
      appStore.showError(t('recharge.createOrderFailed'))
      return
    }
    appStore.showSuccess(t('recharge.orderCreated'))
    startCountdown(result.expires_in || 1800)
    startPolling(result.order_no, result.expires_in || 1800)
  } catch (err: any) {
    appStore.showError(err.response?.data?.detail || err.message || t('recharge.createOrderFailed'))
  } finally {
    submitting.value = false
  }
}

const closeQRModal = () => {
  showQRModal.value = false
  stopPolling()
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
        amount.value = null
        appStore.showSuccess(t('recharge.paymentSuccess'))
        await authStore.refreshUser()
        emit('paid')
      } else if (status.status === 'expired') {
        stopPolling()
        stopCountdown()
        showQRModal.value = false
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

onUnmounted(() => {
  stopPolling()
  stopCountdown()
})
</script>
