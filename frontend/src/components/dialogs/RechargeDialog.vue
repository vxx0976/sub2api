<template>
  <BaseDialog
    :show="show"
    :title="t('recharge.title')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <!-- 充值金额输入 -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('recharge.rechargeAmount') }}
        </label>
        <input
          v-model.number="amount"
          type="number"
          :min="config?.min_amount || 10"
          :max="config?.max_amount || 10000"
          step="0.01"
          class="input w-full"
          :placeholder="t('recharge.enterAmount')"
          @input="validateAmount"
        />
        <p v-if="amountError" class="mt-1 text-sm text-red-600 dark:text-red-400">
          {{ amountError }}
        </p>
      </div>

      <!-- 快捷金额 -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('recharge.quickAmounts') }}
        </label>
        <div class="grid grid-cols-4 gap-2">
          <button
            v-for="quickAmount in quickAmounts"
            :key="quickAmount"
            @click="amount = quickAmount"
            type="button"
            class="rounded-lg border border-gray-200 bg-gray-50 px-3 py-2 text-sm font-medium text-gray-700 transition-colors hover:border-gray-300 hover:bg-gray-100 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:border-dark-500 dark:hover:bg-dark-600"
          >
            ¥{{ quickAmount }}
          </button>
        </div>
      </div>

      <!-- 充值优惠信息 -->
      <div
        v-if="config && creditAmount > amount"
        class="rounded-xl border border-primary-200 bg-gradient-to-r from-primary-50 to-yellow-50 p-4 dark:border-primary-700 dark:from-primary-900/20 dark:to-yellow-900/20"
      >
        <div class="flex items-start">
          <Icon name="gift" size="md" class="mr-3 mt-1 text-primary-600 dark:text-primary-400" />
          <div class="flex-1">
            <div class="mb-1 text-sm font-medium text-gray-900 dark:text-white">
              🎁 {{ t('recharge.bonus') }}
            </div>
            <div class="mb-1 text-lg font-bold text-primary-600 dark:text-primary-400">
              {{ t('recharge.actualCredit') }}: ¥{{ creditAmount.toFixed(2) }}
            </div>
            <div class="text-xs text-gray-600 dark:text-gray-400">
              {{ t('recharge.multiplierInfo', { multiplier: currentMultiplier.toFixed(1) }) }}
            </div>
          </div>
        </div>
      </div>

      <!-- 阶梯倍率说明 -->
      <div v-if="config && config.tiers && config.tiers.length > 0" class="space-y-2">
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('recharge.tierInfo') }}
        </label>
        <div class="space-y-1">
          <div
            v-for="(tier, index) in config.tiers"
            :key="index"
            :class="[
              'flex items-center justify-between rounded px-3 py-2 text-sm',
              isCurrentTier(tier)
                ? 'bg-primary-100 text-primary-700 font-medium dark:bg-primary-900/30 dark:text-primary-300'
                : 'bg-gray-50 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
            ]"
          >
            <span>
              ¥{{ tier.min }}{{ tier.max ? ` - ¥${tier.max}` : '+' }}
            </span>
            <span class="flex items-center">
              {{ tier.multiplier }}×
              <span v-if="tier.multiplier > 1.0" class="ml-2 text-xs">
                (+{{ ((tier.multiplier - 1) * 100).toFixed(0) }}%)
              </span>
            </span>
          </div>
        </div>
      </div>

      <!-- 提示信息 -->
      <div class="text-xs text-gray-500 dark:text-dark-400">
        <p>{{ t('recharge.minAmount') }}: ¥{{ config?.min_amount || 10 }}</p>
        <p>{{ t('recharge.maxAmount') }}: ¥{{ config?.max_amount || 10000 }}</p>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end space-x-3">
        <button @click="handleClose" type="button" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button
          @click="handleConfirm"
          type="button"
          class="btn btn-primary"
          :disabled="!isValidAmount || submitting"
        >
          <span v-if="submitting">{{ t('recharge.processing') }}</span>
          <span v-else>
            {{ t('recharge.confirmRecharge') }}
            <span v-if="amount > 0"> ¥{{ amount.toFixed(2) }}</span>
          </span>
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
import type { RechargeConfig, RechargeTier } from '@/api/recharge'

interface Props {
  show: boolean
  config: RechargeConfig | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm', amount: number): void
}>()

const { t } = useI18n()

const amount = ref<number>(0)
const amountError = ref<string>('')
const submitting = ref(false)

const quickAmounts = computed(() => [10, 50, 100, 200, 500, 1000])

// 计算当前倍率
const currentMultiplier = computed(() => {
  if (!props.config || !props.config.tiers || amount.value <= 0) {
    return 1.0
  }

  for (const tier of props.config.tiers) {
    if (amount.value >= tier.min) {
      if (tier.max === null || amount.value <= tier.max) {
        return tier.multiplier
      }
    }
  }

  return 1.0
})

// 计算实际到账金额
const creditAmount = computed(() => {
  return amount.value * currentMultiplier.value
})

// 验证金额
const validateAmount = () => {
  amountError.value = ''

  if (!props.config) return

  if (amount.value < props.config.min_amount) {
    amountError.value = t('recharge.invalidAmount', {
      min: props.config.min_amount,
      max: props.config.max_amount
    })
  } else if (amount.value > props.config.max_amount) {
    amountError.value = t('recharge.invalidAmount', {
      min: props.config.min_amount,
      max: props.config.max_amount
    })
  }
}

const isValidAmount = computed(() => {
  if (!props.config) return false
  return (
    amount.value >= props.config.min_amount &&
    amount.value <= props.config.max_amount &&
    !amountError.value
  )
})

// 判断是否是当前阶梯
const isCurrentTier = (tier: RechargeTier): boolean => {
  if (amount.value <= 0) return false

  if (amount.value >= tier.min) {
    if (tier.max === null) {
      return true
    }
    return amount.value <= tier.max
  }

  return false
}

const handleClose = () => {
  amount.value = 0
  amountError.value = ''
  emit('close')
}

const handleConfirm = () => {
  if (!isValidAmount.value) return

  submitting.value = true
  emit('confirm', amount.value)
}

// 重置提交状态
watch(() => props.show, (newVal) => {
  if (!newVal) {
    submitting.value = false
  }
})
</script>
