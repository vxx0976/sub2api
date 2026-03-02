<template>
  <AppLayout>
    <div class="mx-auto max-w-2xl space-y-6">
      <!-- Header -->
      <div class="text-center">
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
          {{ t('recharge.pageTitle') }}
        </h1>
        <p class="mt-2 text-gray-500 dark:text-dark-400">
          {{ t('recharge.pageSubtitle') }}
        </p>
      </div>

      <!-- First Recharge Special Offer Card -->
      <div
        v-if="firstRechargeEligible && isChinese"
        class="relative overflow-hidden rounded-2xl bg-gradient-to-r from-amber-500 via-orange-500 to-red-500 p-6 text-white shadow-lg shadow-orange-500/20"
      >
        <div class="absolute -right-4 -top-4 h-24 w-24 rounded-full bg-white/10"></div>
        <div class="absolute -bottom-6 -left-6 h-32 w-32 rounded-full bg-white/5"></div>
        <div class="relative">
          <div class="flex items-center gap-2">
            <span class="text-3xl font-bold">¥1</span>
            <span class="text-lg font-semibold">{{ t('recharge.firstRecharge.title') }}</span>
          </div>
          <p class="mt-1 text-sm text-white/80">
            {{ t('recharge.firstRecharge.description') }}
          </p>
          <div class="mt-4 flex items-center gap-3">
            <button
              @click="handleFirstRecharge"
              :disabled="submitting"
              class="rounded-xl bg-white px-6 py-2.5 text-sm font-semibold text-orange-600 shadow-lg transition-all hover:bg-orange-50 disabled:opacity-50"
            >
              {{ t('recharge.firstRecharge.button') }}
            </button>
            <span class="text-xs text-white/60">{{ t('recharge.firstRecharge.limitOnce') }}</span>
          </div>
        </div>
      </div>

      <!-- Recharge Amount Card -->
      <div class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-600 dark:bg-dark-800">
        <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('recharge.selectAmount') }} ({{ currency }})
        </h2>

        <!-- Quick Amount Buttons -->
        <div class="grid grid-cols-4 gap-3">
          <button
            v-for="q in quickAmounts"
            :key="q"
            @click="selectAmount(q)"
            type="button"
            :class="[
              'rounded-xl border-2 px-3 py-3 text-center text-sm font-semibold transition-all',
              selectedQuickAmount === q
                ? 'border-primary-500 bg-primary-50 text-primary-700 shadow-md shadow-primary-500/10 dark:border-primary-400 dark:bg-primary-900/30 dark:text-primary-300'
                : 'border-gray-200 bg-gray-50 text-gray-700 hover:border-gray-300 hover:bg-gray-100 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:border-dark-500'
            ]"
          >
            {{ currencySymbol }}{{ q }}
          </button>
        </div>

        <!-- Custom Amount -->
        <div class="mt-5">
          <label class="mb-2 block text-sm font-medium text-gray-600 dark:text-gray-400">
            {{ t('recharge.customAmount') }}
          </label>
          <div class="relative">
            <span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm font-medium text-gray-500 dark:text-gray-400">
              {{ currencySymbol }}
            </span>
            <input
              v-model.number="customAmountInput"
              type="number"
              :min="minAmount"
              :max="maxAmount || undefined"
              step="1"
              class="input w-full pl-7"
              :placeholder="t('recharge.enterAmount')"
              @input="onCustomInput"
            />
          </div>
          <p v-if="isChinese" class="mt-1.5 text-xs text-gray-400 dark:text-gray-500">
            {{ t('recharge.amountRange', { min: minAmount, max: maxAmount, symbol: currencySymbol }) }}
          </p>
        </div>

        <!-- Create Order Button (Chinese only - Alipay) -->
        <button
          v-if="isChinese"
          @click="handleCreateOrder"
          :disabled="!isValidAmount || submitting"
          class="mt-5 w-full rounded-xl bg-gradient-to-r from-primary-500 to-primary-600 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-primary-500/20 transition-all hover:from-primary-600 hover:to-primary-700 hover:shadow-xl disabled:opacity-50 disabled:shadow-none"
        >
          <span v-if="submitting" class="flex items-center justify-center gap-2">
            <div class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
            {{ t('recharge.processing') }}
          </span>
          <span v-else>{{ t('recharge.createOrder') }}</span>
        </button>
      </div>

      <!-- Amount Preview -->
      <div
        v-if="inputAmount > 0"
        class="rounded-2xl border border-primary-200 bg-gradient-to-r from-primary-50 to-blue-50 p-5 dark:border-primary-800 dark:from-primary-900/20 dark:to-blue-900/20"
      >
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('recharge.youWillReceive') }}</div>
          <div class="text-xs text-gray-400 dark:text-gray-500">
            <template v-if="isChinese">$1 = ¥{{ cnyPerDollar.toFixed(2) }}</template>
            <template v-else>$1 USDT = ${{ USD_RATE }}</template>
          </div>
        </div>
        <div class="mt-1 text-2xl font-bold text-primary-600 dark:text-primary-400">
          ${{ platformBalance.toFixed(2) }}
          <span class="ml-1 text-sm font-normal text-gray-500 dark:text-gray-400">{{ t('recharge.platformBalance') }}</span>
        </div>
      </div>

      <!-- Non-Chinese: Crypto payment + contact info -->
      <div v-if="!isChinese" class="rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-600 dark:bg-dark-800 space-y-4">
        <div v-if="cryptoAddresses.length > 0">
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('recharge.cryptoPayment') }}</label>
          <div v-for="(addr, index) in cryptoAddresses" :key="index" class="mt-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2.5 dark:border-dark-600 dark:bg-dark-700">
            <div class="flex items-center justify-between gap-2">
              <span class="inline-flex items-center rounded-md bg-gray-200 px-1.5 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-600 dark:text-dark-300">{{ addr.chain }}</span>
              <button @click="copyAddress(addr.address)" type="button" class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-200 hover:text-gray-600 dark:text-dark-500 dark:hover:bg-dark-600 dark:hover:text-dark-300">
                <Icon name="clipboard" size="xs" />
              </button>
            </div>
            <p class="mt-1 break-all font-mono text-xs leading-relaxed text-gray-500 dark:text-dark-400">{{ addr.address }}</p>
          </div>
        </div>
        <p class="text-xs text-gray-400 dark:text-gray-500">
          {{ t('recharge.cryptoHint') }}
        </p>
        <div class="flex flex-wrap items-center gap-3 text-xs">
          <a v-if="contactTelegram" :href="`https://t.me/${contactTelegram}`" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 font-medium text-blue-500 hover:text-blue-600 dark:text-blue-400">
            Telegram @{{ contactTelegram }}
            <Icon name="externalLink" size="xs" />
          </a>
          <span v-if="contactWechat" class="inline-flex items-center gap-1 text-gray-500 dark:text-gray-400">
            WeChat: {{ contactWechat }}
          </span>
          <a href="mailto:vanxuehan@gmail.com" class="inline-flex items-center gap-1 font-medium text-red-500 hover:text-red-600 dark:text-red-400">
            vanxuehan@gmail.com
          </a>
        </div>
      </div>

      <!-- Recharge Notes (Chinese only) -->
      <div v-if="isChinese" class="rounded-2xl border border-amber-200 bg-amber-50 p-5 dark:border-amber-800/50 dark:bg-amber-900/10">
        <h3 class="mb-3 flex items-center gap-2 text-sm font-semibold text-amber-800 dark:text-amber-300">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
          </svg>
          {{ t('recharge.notesTitle') }}
        </h3>
        <ul class="space-y-2 text-xs text-amber-700 dark:text-amber-400">
          <li class="flex items-start gap-2">
            <span class="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-amber-400"></span>
            {{ t('recharge.noteAlipay') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-amber-400"></span>
            {{ t('recharge.noteAmountMatch') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-amber-400"></span>
            {{ t('recharge.noteAutoCredit') }}
          </li>
          <li class="flex items-start gap-2">
            <span class="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-amber-400"></span>
            {{ t('recharge.noteContact') }}
          </li>
        </ul>
      </div>

      <!-- Rate info -->
      <div class="rounded-2xl border border-gray-200 bg-gray-50 p-4 dark:border-dark-600 dark:bg-dark-700">
        <div class="flex items-start gap-2">
          <Icon name="infoCircle" size="sm" class="mt-0.5 text-gray-400 dark:text-gray-500" />
          <div class="flex-1 text-xs text-gray-500 dark:text-gray-400">
            <p v-if="isChinese">{{ t('recharge.rateDescCny', { cnyPerDollar: cnyPerDollar.toFixed(2), rate: usdCnyRate.toFixed(2) }) }}</p>
            <p class="mt-1">{{ t('recharge.balanceNeverExpires') }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Payment QR Code Modal -->
    <PaymentQRCodeModal
      :show="showPaymentModal"
      :order-no="paymentInfo.orderNo"
      :amount="paymentInfo.amount"
      :payment-amount="paymentInfo.paymentAmount"
      :qr-code="paymentInfo.qrCode"
      :qr-code-url="paymentInfo.qrCodeUrl"
      :mode="paymentInfo.mode"
      @close="showPaymentModal = false"
      @paid="handlePaymentSuccess"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { createRechargeOrder, getRechargeConfig } from '@/api/recharge'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PaymentQRCodeModal from '@/components/common/PaymentQRCodeModal.vue'

const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

// Locale
const isChinese = computed(() => locale.value.startsWith('zh'))

// Constants
const USD_RATE = 10 // 1 USDT = $10 platform balance
const usdCnyRate = ref(7.0) // fetched from backend

// Locale-aware amounts
const currency = computed(() => isChinese.value ? 'CNY' : 'USDT')
const currencySymbol = computed(() => isChinese.value ? '¥' : '$')
const quickAmounts = computed(() => isChinese.value ? [10, 50, 200, 500] : [10, 50, 100, 500])
const minAmount = computed(() => isChinese.value ? 10 : 1)
const maxAmount = computed(() => isChinese.value ? 500 : 0) // 0 = no limit
// CNY rate derived from USD_RATE: ¥1 = $USD_RATE / usdCnyRate
const cnyMultiplier = computed(() => USD_RATE / usdCnyRate.value)
const unitRate = computed(() => isChinese.value ? cnyMultiplier.value : USD_RATE)
// ¥X = $1 platform balance (how many CNY per $1 platform balance)
const cnyPerDollar = computed(() => usdCnyRate.value / USD_RATE)

// Contact info
const contactTelegram = computed(() => appStore.cachedPublicSettings?.contact_telegram || '')
const contactWechat = computed(() => appStore.cachedPublicSettings?.contact_wechat || '')

// Crypto addresses
interface CryptoAddress { chain: string; address: string }
const cryptoAddresses = computed<CryptoAddress[]>(() => {
  const raw = appStore.cachedPublicSettings?.crypto_addresses
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed
  } catch { /* ignore */ }
  return []
})

// State
const selectedQuickAmount = ref<number | null>(null)
const customAmountInput = ref<number | null>(null)
const submitting = ref(false)
const firstRechargeEligible = ref(false)

// Payment modal state
const showPaymentModal = ref(false)
const paymentInfo = reactive({
  orderNo: '',
  amount: 0,
  paymentAmount: 0,
  qrCode: '',
  qrCodeUrl: '',
  mode: ''
})

// Computed: input amount in local currency
const inputAmount = computed(() => {
  if (customAmountInput.value && customAmountInput.value > 0) {
    return customAmountInput.value
  }
  return selectedQuickAmount.value ?? 0
})

// Computed: CNY equivalent (for order creation)
const cnyAmount = computed(() => isChinese.value ? inputAmount.value : inputAmount.value * usdCnyRate.value)

// Computed: platform balance received
const platformBalance = computed(() => inputAmount.value * unitRate.value)

// Validation
const isValidAmount = computed(() => {
  const amt = inputAmount.value
  if (amt < minAmount.value) return false
  if (isChinese.value && maxAmount.value > 0 && amt > maxAmount.value) return false
  return true
})

function selectAmount(amount: number) {
  selectedQuickAmount.value = amount
  customAmountInput.value = null
}

function onCustomInput() {
  if (customAmountInput.value && customAmountInput.value > 0) {
    selectedQuickAmount.value = null
  }
}

async function handleCreateOrder() {
  if (!isValidAmount.value || submitting.value) return
  await submitOrder(Math.round(cnyAmount.value * 100) / 100)
}

async function copyAddress(address: string) {
  try {
    await navigator.clipboard.writeText(address)
    appStore.showSuccess(t('recharge.copySuccess'))
  } catch {
    const ta = document.createElement('textarea')
    ta.value = address
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    appStore.showSuccess(t('recharge.copySuccess'))
  }
}

async function handleFirstRecharge() {
  if (submitting.value) return
  await submitOrder(1) // ¥1
}

async function submitOrder(amount: number) {
  try {
    submitting.value = true
    const result = await createRechargeOrder(amount)
    paymentInfo.orderNo = result.order_no
    paymentInfo.amount = result.amount
    paymentInfo.paymentAmount = result.payment_amount
    paymentInfo.qrCode = result.qr_code
    paymentInfo.qrCodeUrl = result.qr_code_url
    paymentInfo.mode = result.mode
    showPaymentModal.value = true
  } catch (err: any) {
    console.error('Failed to create recharge order:', err)
    appStore.showError(err.message || t('recharge.rechargeFailed'))
  } finally {
    submitting.value = false
  }
}

function handlePaymentSuccess() {
  showPaymentModal.value = false
  appStore.showSuccess(t('payment.paymentSuccess'))
  authStore.refreshUser()
}

// Check first recharge eligibility
async function checkFirstRecharge() {
  try {
    const { apiClient } = await import('@/api/client')
    const response = await apiClient.get('/recharge/first-recharge-status')
    firstRechargeEligible.value = response.data?.eligible ?? false
  } catch {
    // API may not exist yet, silently ignore
    firstRechargeEligible.value = false
  }
}

async function loadRechargeConfig() {
  try {
    const config = await getRechargeConfig()
    if (config.usd_cny_rate > 0) {
      usdCnyRate.value = config.usd_cny_rate
    }
  } catch {
    // Use default rate
  }
}

onMounted(() => {
  checkFirstRecharge()
  loadRechargeConfig()
})
</script>
