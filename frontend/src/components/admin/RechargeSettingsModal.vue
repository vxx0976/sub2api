<template>
  <BaseDialog
    :show="show"
    :title="t('admin.settings.recharge.title')"
    width="wide"
    @close="$emit('close')"
  >
    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
    </div>

    <!-- Settings Form -->
    <form v-else @submit.prevent="saveSettings" class="space-y-6">
      <!-- Basic Settings -->
      <div class="space-y-4">
        <h3 class="text-base font-medium text-gray-900 dark:text-white">
          {{ t('admin.settings.recharge.basicSettings') }}
        </h3>

        <!-- Enable Recharge -->
        <div class="flex items-center justify-between rounded-lg bg-gray-50 p-4 dark:bg-dark-700">
          <div>
            <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.recharge.enabled') }}
            </label>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.recharge.enabledHint') }}
            </p>
          </div>
          <label class="relative inline-flex cursor-pointer items-center">
            <input
              type="checkbox"
              v-model="config.enabled"
              class="peer sr-only"
            />
            <div class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-primary-600 peer-checked:after:translate-x-full peer-checked:after:border-white peer-focus:outline-none dark:border-gray-600 dark:bg-gray-700"></div>
          </label>
        </div>

        <!-- Min/Max Amount -->
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.recharge.minAmount') }}
            </label>
            <input
              v-model.number="config.min_amount"
              type="number"
              min="1"
              step="0.01"
              class="input"
              :placeholder="t('admin.settings.recharge.minAmountPlaceholder')"
            />
          </div>
          <div>
            <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('admin.settings.recharge.maxAmount') }}
            </label>
            <input
              v-model.number="config.max_amount"
              type="number"
              min="1"
              step="0.01"
              class="input"
              :placeholder="t('admin.settings.recharge.maxAmountPlaceholder')"
            />
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          type="button"
          @click="$emit('close')"
          class="btn btn-secondary"
          :disabled="saving"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          @click="saveSettings"
          class="btn btn-primary"
          :disabled="saving || !isValid"
        >
          <span v-if="saving">{{ t('common.saving') }}</span>
          <span v-else>{{ t('common.save') }}</span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { getRechargeConfigAdmin, updateRechargeConfig, type RechargeConfig } from '@/api/recharge'
import BaseDialog from '@/components/common/BaseDialog.vue'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)

const config = ref<RechargeConfig>({
  enabled: false,
  min_amount: 10,
  max_amount: 10000,
  usd_cny_rate: 7.0,
})

const isValid = computed(() => {
  if (config.value.min_amount >= config.value.max_amount) return false
  if (config.value.min_amount < 1) return false
  return true
})

async function loadConfig() {
  try {
    loading.value = true
    const data = await getRechargeConfigAdmin()
    config.value = { ...data }
  } catch (error: any) {
    console.error('Failed to load recharge config:', error)
    appStore.showError(error.message || t('admin.settings.recharge.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  try {
    saving.value = true
    await updateRechargeConfig(config.value)
    appStore.showSuccess(t('admin.settings.recharge.saveSuccess'))
    emit('saved')
    emit('close')
  } catch (error: any) {
    console.error('Failed to save recharge config:', error)
    appStore.showError(error.message || t('admin.settings.recharge.saveFailed'))
  } finally {
    saving.value = false
  }
}

// Load config when modal opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadConfig()
  }
}, { immediate: true })
</script>
