<template>
  <div class="space-y-5">
    <form @submit.prevent="handleCreateOrder" class="space-y-5">
      <div>
        <label for="alipay-amount" class="input-label">{{ t('recharge.amountLabel') }}</label>
        <div class="relative mt-1">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
            <span class="text-gray-400 dark:text-dark-500 font-medium">¥</span>
          </div>
          <input
            id="alipay-amount"
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
      </div>

      <button type="submit" :disabled="!canSubmit || submitting" class="btn btn-primary w-full py-3">
        <svg v-if="submitting" class="-ml-1 mr-2 h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
        </svg>
        {{ submitting ? t('recharge.submitting') : t('recharge.submitButton') }}
      </button>
    </form>

    <!-- QR Code Payment Modal -->
    <Teleport to="body">
      <!-- 不响应背景点击，避免扫码期间手滑关闭。用户必须点右上角 X 或扫码完成后自动关闭 -->
      <div
        v-if="showQRModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
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

          <!-- 金额警示卡：琥珀色柔和提示，金额用中性深色保留可读性 -->
          <div class="mb-4 rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-700/50 dark:bg-amber-900/10">
            <div class="flex items-start gap-2">
              <svg class="mt-0.5 h-4 w-4 flex-shrink-0 text-amber-500 dark:text-amber-400" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
              </svg>
              <p class="text-sm font-medium text-amber-800 dark:text-amber-200">
                {{ t('alimpay.payExactAmountNotice') }}
              </p>
            </div>
            <div class="mt-3 text-center">
              <p class="text-5xl font-bold tracking-tight text-amber-700 dark:text-amber-300">
                ¥{{ qrPaymentAmount?.toFixed(2) }}
              </p>
              <p class="mt-2 text-xs text-amber-700 dark:text-amber-400">
                {{ t('alimpay.amountMustMatch') }}
              </p>
            </div>
          </div>

          <div class="mb-4 flex justify-center">
            <div class="rounded-xl border-2 border-gray-100 bg-white p-3 dark:border-dark-600">
              <img :src="qrCodeDataURL" alt="Payment QR Code" class="h-52 w-52" />
            </div>
          </div>

          <p class="mb-3 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('recharge.scanQRHintAlipay') }}
          </p>

          <!-- 倒计时 -->
          <div v-if="remainingSeconds > 0" class="mb-2 flex items-center justify-center gap-2 text-sm font-medium text-orange-600 dark:text-orange-400">
            <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ t('alimpay.countdown', { time: formatCountdown(remainingSeconds) }) }}
          </div>
          <div v-else class="mb-2 text-center text-sm font-medium text-gray-500 dark:text-gray-400">
            {{ t('alimpay.expiredHint') }}
          </div>

          <div class="mt-2 flex items-center justify-center gap-2 text-xs text-gray-400 dark:text-dark-500">
            <svg class="h-3.5 w-3.5 animate-spin" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            {{ t('recharge.waitingPayment') }}
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
/**
 * 支付宝（AliMPay 个人免签）充值面板
 * 由统一充值页 RechargeView 在「支付宝」标签下渲染。config 由父组件传入，避免重复请求。
 */
import { ref, computed, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { alimpayAPI, type AliMPayConfig } from '@/api/alimpayOrder'

const props = defineProps<{ config: AliMPayConfig }>()
const emit = defineEmits<{ (e: 'paid'): void }>()

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const amount = ref<number | null>(null)
const submitting = ref(false)

let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null

const showQRModal = ref(false)
const qrCodeDataURL = ref('')
const qrPaymentAmount = ref<number | null>(null)
const remainingSeconds = ref(0)

const formatCountdown = (seconds: number) => {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

const creditPreview = computed(() => {
  if (!amount.value || props.config.selling_price <= 0) return 0
  return Math.round((amount.value / props.config.selling_price) * 100) / 100
})

const canSubmit = computed(() => {
  if (!amount.value) return false
  return amount.value >= props.config.min_amount && amount.value <= props.config.max_amount
})

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
      // 二维码生成失败时中止，避免误报「下单成功」并启动看不见的轮询
      appStore.showError(t('recharge.createOrderFailed'))
      return
    }
    appStore.showSuccess(t('recharge.orderCreated'))
    startCountdown(result.expires_in || 600)
    startPolling(result.order_no, result.expires_in || 600)
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
  // 轮询上限 = 订单有效期秒数 / 5 + 2 次 buffer，保证订单过期那一刻也能查到最终状态
  const maxCount = Math.ceil(expiresIn / 5) + 2
  let count = 0
  pollTimer = setInterval(async () => {
    count++
    if (count > maxCount) {
      stopPolling()
      return
    }
    try {
      const status = await alimpayAPI.getOrderStatus(orderNo)
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
