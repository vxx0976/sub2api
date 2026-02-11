<template>
  <Teleport to="body">
    <Transition name="modal">
      <div
        v-if="show"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        @click.self="handleClose"
      >
        <div class="relative mx-4 w-full max-w-sm rounded-2xl bg-white shadow-xl dark:bg-dark-800" @click.stop>
          <!-- Header -->
          <div class="flex items-center justify-between border-b border-gray-200 px-6 py-4 dark:border-dark-600">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ mode === 'business_qr' ? t('payment.scanToPay') : t('payment.transferToPay') }}
            </h3>
            <button
              @click="handleClose"
              class="-mr-2 rounded-xl p-2 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:text-dark-500 dark:hover:bg-dark-700 dark:hover:text-dark-300"
              :aria-label="t('common.close')"
            >
              <Icon name="x" size="md" />
            </button>
          </div>

          <!-- Body -->
          <div class="space-y-5 px-6 py-5">
            <!-- QR Code -->
            <div class="flex justify-center">
              <div class="rounded-xl border border-gray-200 bg-white p-3 dark:border-dark-600 dark:bg-white">
                <img
                  :src="`data:image/png;base64,${qrCode}`"
                  :alt="t('payment.qrCodeAlt')"
                  class="h-52 w-52"
                />
              </div>
            </div>

            <!-- Payment amount -->
            <div class="text-center">
              <div class="text-sm text-gray-500 dark:text-gray-400">
                {{ t('payment.amountToPay') }}
              </div>
              <div class="mt-1 text-3xl font-bold text-primary-500">
                ¥{{ paymentAmount.toFixed(2) }}
              </div>
              <div v-if="amount !== paymentAmount" class="mt-0.5 text-xs text-gray-400 line-through dark:text-dark-500">
                ¥{{ amount.toFixed(2) }}
              </div>
            </div>

            <!-- Mode-specific instructions -->
            <div class="rounded-lg bg-gray-50 px-4 py-3 dark:bg-dark-700">
              <!-- Business QR mode -->
              <template v-if="mode === 'business_qr'">
                <p class="text-center text-sm text-gray-600 dark:text-gray-400">
                  {{ t('payment.businessQrInstruction') }}
                </p>
              </template>

              <!-- Transfer mode -->
              <template v-else>
                <p class="text-sm text-gray-600 dark:text-gray-400">
                  {{ t('payment.transferInstruction') }}
                </p>
                <div class="mt-2 flex items-center justify-between rounded-md bg-white px-3 py-2 dark:bg-dark-800">
                  <div>
                    <div class="text-xs text-gray-400 dark:text-dark-500">{{ t('payment.transferMemo') }}</div>
                    <div class="font-mono text-sm font-medium text-gray-900 dark:text-white">{{ orderNo }}</div>
                  </div>
                </div>
                <a
                  v-if="qrCodeUrl"
                  :href="qrCodeUrl"
                  class="mt-3 flex w-full items-center justify-center gap-2 rounded-lg bg-[#1677FF] px-4 py-2.5 text-sm font-medium text-white transition-colors hover:bg-[#1266DB]"
                >
                  <Icon name="externalLink" size="sm" />
                  {{ t('payment.openAlipay') }}
                </a>
              </template>
            </div>

            <!-- Polling status -->
            <div class="flex items-center justify-center gap-2 text-sm text-gray-500 dark:text-gray-400">
              <LoadingSpinner size="sm" color="primary" />
              <span>{{ t('payment.waitingForPayment') }}</span>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import { apiClient } from '@/api/client'

const { t } = useI18n()

interface Props {
  show: boolean
  orderNo: string
  amount: number
  paymentAmount: number
  qrCode: string
  qrCodeUrl: string
  mode: string
}

interface Emits {
  (e: 'close'): void
  (e: 'paid'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

let pollTimer: ReturnType<typeof setInterval> | null = null
const isPaid = ref(false)

const pollPaymentStatus = async () => {
  if (!props.orderNo || isPaid.value) return

  try {
    const response = await apiClient.get<{ status: string; order_no: string }>(
      `/orders/${props.orderNo}/payment-status`
    )
    const status = response.data.status
    if (status === 'paid') {
      isPaid.value = true
      stopPolling()
      emit('paid')
    }
  } catch {
    // Silently ignore polling errors; will retry on next interval
  }
}

const startPolling = () => {
  stopPolling()
  isPaid.value = false
  // Poll immediately, then every 3 seconds
  pollPaymentStatus()
  pollTimer = setInterval(pollPaymentStatus, 3000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

const handleClose = () => {
  stopPolling()
  emit('close')
}

// Start/stop polling based on modal visibility
watch(
  () => props.show,
  (isVisible) => {
    if (isVisible) {
      startPolling()
    } else {
      stopPolling()
    }
  },
  { immediate: true }
)

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .relative,
.modal-leave-active .relative {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .relative {
  transform: scale(0.95);
}
</style>
