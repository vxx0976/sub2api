<template>
  <div class="space-y-5">
    <form @submit.prevent="handleCreateOrder" class="space-y-5">
      <div>
        <label for="wechat-amount" class="input-label">{{ t('recharge.amountLabel') }}</label>
        <div class="relative mt-1">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-4">
            <span class="text-gray-400 dark:text-dark-500 font-medium">¥</span>
          </div>
          <input
            id="wechat-amount"
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
      <div
        v-if="showQRModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
        @click.self="closeQRModal"
      >
        <div class="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-800">
          <div class="mb-4 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('recharge.scanToPay') }}
            </h3>
            <button
              class="rounded-lg p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-dark-300"
              @click="closeQRModal"
            >
              <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Amount Info -->
          <div class="mb-4 rounded-lg bg-gray-50 p-3 text-center dark:bg-dark-700">
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('recharge.payAmount') }}</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">¥{{ qrOrderAmount?.toFixed(2) }}</p>
          </div>

          <!-- QR Code -->
          <div class="mb-4 flex justify-center">
            <div class="rounded-xl border-2 border-gray-100 bg-white p-3 dark:border-dark-600">
              <img :src="qrCodeURL" alt="Payment QR Code" class="h-52 w-52" />
            </div>
          </div>

          <p class="mb-3 text-center text-sm text-gray-500 dark:text-dark-400">
            {{ t('recharge.scanQRHintWechat') }}
          </p>

          <!-- Fallback link -->
          <a
            v-if="qrPayURL"
            :href="qrPayURL"
            target="_blank"
            class="block text-center text-xs text-primary-500 hover:text-primary-600 dark:text-primary-400"
          >
            {{ t('recharge.openPayPage') }}
          </a>

          <!-- Waiting indicator -->
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
  </div>
</template>

<script setup lang="ts">
/**
 * 微信充值面板（走易支付 EPAY 网关，pay_type 固定 wxpay）
 * 由统一充值页 RechargeView 在「微信」标签下渲染。config 由父组件传入。
 */
import { ref, computed, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import QRCode from 'qrcode'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { rechargeAPI, type RechargeConfig } from '@/api/recharge'

const props = defineProps<{ config: RechargeConfig }>()
const emit = defineEmits<{ (e: 'paid'): void }>()

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const PAY_TYPE = 'wxpay'

const amount = ref<number | null>(null)
const submitting = ref(false)

let pollTimer: ReturnType<typeof setInterval> | null = null

const showQRModal = ref(false)
const qrCodeURL = ref('')
const qrPayURL = ref('')
const qrOrderAmount = ref<number | null>(null)

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
    const result = await rechargeAPI.createOrder(amount.value, PAY_TYPE)
    qrPayURL.value = result.pay_url || ''
    qrOrderAmount.value = result.amount
    const qrContent = result.qrcode || result.pay_url || ''
    if (qrContent) {
      try {
        qrCodeURL.value = await QRCode.toDataURL(qrContent, { width: 280, margin: 2 })
        showQRModal.value = true
      } catch {
        if (qrPayURL.value) window.open(qrPayURL.value, '_blank')
      }
    } else if (qrPayURL.value) {
      window.open(qrPayURL.value, '_blank')
    }
    appStore.showSuccess(t('recharge.orderCreated'))
    startPolling(result.order_no)
  } catch (err: any) {
    appStore.showError(err.response?.data?.detail || err.message || t('recharge.createOrderFailed'))
  } finally {
    submitting.value = false
  }
}

const closeQRModal = () => {
  showQRModal.value = false
  stopPolling()
}

const startPolling = (orderNo: string) => {
  stopPolling()
  let count = 0
  pollTimer = setInterval(async () => {
    count++
    if (count > 60) {
      // Stop after 5 minutes (5s * 60)
      stopPolling()
      return
    }
    try {
      const status = await rechargeAPI.getOrderStatus(orderNo)
      if (status.status === 'paid') {
        stopPolling()
        showQRModal.value = false
        amount.value = null
        appStore.showSuccess(t('recharge.paymentSuccess'))
        await authStore.refreshUser()
        emit('paid')
      } else if (status.status === 'expired') {
        stopPolling()
        showQRModal.value = false
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

onUnmounted(() => {
  stopPolling()
})
</script>
