<template>
  <BaseDialog
    :show="show"
    :title="t('recharge.title')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <!-- 支付金额 -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('recharge.youPay') }} ({{ currency }})
        </label>
        <div class="relative">
          <span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm font-medium text-gray-500 dark:text-gray-400">
            {{ currencySymbol }}
          </span>
          <input
            :value="amount || ''"
            type="number"
            :min="minAmount"
            :max="maxAmount || undefined"
            step="0.01"
            class="input w-full pl-7"
            :placeholder="t('recharge.enterAmount')"
            @input="onAmountInput"
          />
        </div>
        <p v-if="amountError" class="mt-1 text-sm text-red-600 dark:text-red-400">
          {{ amountError }}
        </p>
      </div>

      <!-- 快捷金额 -->
      <div class="flex flex-wrap gap-2">
        <button
          v-for="q in quickAmounts"
          :key="q"
          @click="amount = q"
          type="button"
          :class="[
            'rounded-lg border px-3 py-1.5 text-sm font-medium transition-all',
            amount === q
              ? 'border-primary-400 bg-primary-50 text-primary-700 dark:border-primary-600 dark:bg-primary-900/30 dark:text-primary-300'
              : 'border-gray-200 bg-gray-50 text-gray-600 hover:bg-gray-100 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600'
          ]"
        >
          {{ currencySymbol }}{{ q }}
        </button>
      </div>

      <!-- 到账金额 -->
      <div v-if="amount > 0" class="rounded-xl border border-primary-200 bg-gradient-to-r from-primary-50 to-blue-50 p-4 dark:border-primary-800 dark:from-primary-900/20 dark:to-blue-900/20">
        <div class="flex items-center justify-between">
          <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('recharge.youReceive') }}</div>
          <div class="text-xs text-gray-400 dark:text-gray-500">1 {{ currency }} = ${{ unitRate.toFixed(2) }}</div>
        </div>
        <div class="mt-1 text-2xl font-bold text-primary-600 dark:text-primary-400">
          ${{ platformBalance.toFixed(2) }}
          <span class="ml-1 text-sm font-normal text-gray-500 dark:text-gray-400">{{ t('recharge.platformBalance') }}</span>
        </div>
      </div>

      <!-- 非中文：虚拟币支付 + 联系方式 -->
      <div v-if="!isChinese" class="space-y-2">
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
        </div>
      </div>

      <!-- 汇率说明 -->
      <div class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-dark-600 dark:bg-dark-700">
        <div class="flex items-start gap-2">
          <Icon name="infoCircle" size="sm" class="mt-0.5 text-gray-400 dark:text-gray-500" />
          <div class="flex-1 text-xs text-gray-500 dark:text-gray-400">
            <p>{{ rateDescription }}</p>
            <p class="mt-1">{{ t('recharge.balanceNeverExpires') }}</p>
          </div>
        </div>
      </div>

      <div v-if="isChinese" class="text-xs text-gray-400 dark:text-gray-500">
        {{ t('recharge.amountRange', { min: minAmount, max: maxAmount, symbol: currencySymbol }) }}
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button @click="handleClose" type="button" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button v-if="isChinese" @click="handleConfirm" type="button" class="btn btn-primary" :disabled="!isValidAmount || submitting">
          <span v-if="submitting">{{ t('recharge.processing') }}</span>
          <span v-else>{{ t('recharge.confirmRecharge') }}<span v-if="amount > 0"> {{ currencySymbol }}{{ amount.toFixed(2) }}</span></span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'

interface Props {
  show: boolean
}

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'close'): void; (e: 'confirm', amount: number): void }>()
const { t, locale } = useI18n()
const appStore = useAppStore()

const amount = ref<number>(0)
const amountError = ref<string>('')
const submitting = ref(false)

// 固定倍率：1 CNY = $2 平台余额
const MULTIPLIER = 2

const isChinese = computed(() => locale.value.startsWith('zh'))
const currency = computed(() => isChinese.value ? 'CNY' : 'USDT')
const currencySymbol = computed(() => isChinese.value ? '¥' : '$')
const contactTelegram = computed(() => appStore.cachedPublicSettings?.contact_telegram || '')
const contactWechat = computed(() => appStore.cachedPublicSettings?.contact_wechat || '')

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

// 中文：1 CNY = $2；非中文：1 USDT ≈ 1 CNY * MULTIPLIER（按汇率换算）
const cnyAmount = computed(() => isChinese.value ? amount.value : amount.value * 7)
const unitRate = computed(() => isChinese.value ? MULTIPLIER : 7 * MULTIPLIER)
const platformBalance = computed(() => cnyAmount.value * MULTIPLIER)

const rateDescription = computed(() => {
  if (isChinese.value) return t('recharge.rateDescCny', { multiplier: MULTIPLIER.toFixed(1) })
  return t('recharge.rateDescUsd', { rate: '7.0', multiplier: MULTIPLIER.toFixed(1), total: (7 * MULTIPLIER).toFixed(2) })
})

// 中文：10-500；非中文：无限制
const quickAmounts = computed(() => isChinese.value ? [10, 50, 200, 500] : [10, 50, 100, 500])
const minAmount = computed(() => isChinese.value ? 10 : 1)
const maxAmount = computed(() => isChinese.value ? 500 : 0) // 0 = 无上限

const onAmountInput = (event: Event) => {
  const val = parseFloat((event.target as HTMLInputElement).value)
  amount.value = isNaN(val) ? 0 : val
  validateAmount()
}

const validateAmount = () => {
  amountError.value = ''
  if (amount.value > 0 && amount.value < minAmount.value) {
    amountError.value = t('recharge.invalidAmountRange', { min: minAmount.value, max: maxAmount.value || '∞', symbol: currencySymbol.value })
  }
  if (isChinese.value && amount.value > maxAmount.value) {
    amountError.value = t('recharge.invalidAmountRange', { min: minAmount.value, max: maxAmount.value, symbol: currencySymbol.value })
  }
}

const isValidAmount = computed(() => {
  if (amount.value < minAmount.value) return false
  if (isChinese.value && amount.value > maxAmount.value) return false
  return !amountError.value
})

const handleClose = () => { amount.value = 0; amountError.value = ''; emit('close') }
const handleConfirm = () => {
  if (!isValidAmount.value) return
  submitting.value = true
  emit('confirm', Math.round(cnyAmount.value * 100) / 100)
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

watch(() => props.show, (newVal) => {
  if (!newVal) { submitting.value = false }
})
</script>
