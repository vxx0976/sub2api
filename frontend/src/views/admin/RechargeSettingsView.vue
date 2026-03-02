<template>
  <AppLayout>
    <div class="mx-auto max-w-4xl space-y-6">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('admin.settings.recharge.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {{ t('admin.settings.recharge.description') }}
          </p>
        </div>
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <!-- Settings Form -->
      <form v-else @submit.prevent="saveSettings" class="space-y-6">
        <!-- Basic Settings -->
        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.settings.recharge.basicSettings') }}
            </h2>
          </div>
          <div class="space-y-6 p-6">
            <!-- Enable Recharge -->
            <div class="flex items-center justify-between">
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
                <div class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-primary-600 peer-checked:after:translate-x-full peer-checked:after:border-white peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary-300 dark:border-gray-600 dark:bg-gray-700 dark:peer-focus:ring-primary-800"></div>
              </label>
            </div>

            <!-- Min/Max Amount -->
            <div class="grid grid-cols-1 gap-6 md:grid-cols-3">
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
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.settings.recharge.minAmountHint') }}
                </p>
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
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.settings.recharge.maxAmountHint') }}
                </p>
              </div>
              <div>
                <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
                  {{ t('admin.settings.recharge.usdCnyRate') }}
                </label>
                <input
                  v-model.number="config.usd_cny_rate"
                  type="number"
                  min="0.01"
                  step="0.0001"
                  class="input"
                  placeholder="7.0"
                />
                <p class="mt-1.5 text-xs text-gray-500 dark:text-gray-400">
                  {{ t('admin.settings.recharge.usdCnyRateHint') }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Actions -->
        <div class="flex justify-end gap-3">
          <button
            type="button"
            @click="resetForm"
            class="btn btn-secondary"
            :disabled="saving"
          >
            {{ t('common.reset') }}
          </button>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="saving || !isValid"
          >
            <span v-if="saving">{{ t('common.saving') }}</span>
            <span v-else>{{ t('common.save') }}</span>
          </button>
        </div>
      </form>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { getRechargeConfigAdmin, updateRechargeConfig, type RechargeConfig } from '@/api/recharge'
import AppLayout from '@/components/layout/AppLayout.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const saving = ref(false)

const config = ref<RechargeConfig>({
  enabled: false,
  min_amount: 10,
  max_amount: 10000,
  usd_cny_rate: 7.0,
})

const originalConfig = ref<RechargeConfig | null>(null)

// Form validation
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
    originalConfig.value = JSON.parse(JSON.stringify(data))
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
    originalConfig.value = JSON.parse(JSON.stringify(config.value))
  } catch (error: any) {
    console.error('Failed to save recharge config:', error)
    appStore.showError(error.message || t('admin.settings.recharge.saveFailed'))
  } finally {
    saving.value = false
  }
}

function resetForm() {
  if (originalConfig.value) {
    config.value = JSON.parse(JSON.stringify(originalConfig.value))
  }
}

onMounted(() => {
  loadConfig()
})
</script>
