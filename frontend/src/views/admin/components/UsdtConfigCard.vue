<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.settings.usdt.title') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.description') }}</p>
    </div>

    <div v-if="loading" class="p-6 text-sm text-gray-400">{{ t('common.loading') }}</div>

    <div v-else class="space-y-6 p-6">
      <!-- Enable Toggle -->
      <div class="flex items-center justify-between">
        <div>
          <label class="font-medium text-gray-900 dark:text-white">{{ t('admin.settings.usdt.enabled') }}</label>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.enabledHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <!-- Receiving Address -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.settings.usdt.receivingAddress') }}</label>
        <input v-model="form.receiving_address" type="text" class="input font-mono text-sm" placeholder="T..." />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.receivingAddressHint') }}</p>
      </div>

      <!-- TronGrid API Key (sensitive) -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.settings.usdt.tronApiKey') }}
          <span v-if="hasTronApiKey" class="ml-2 text-xs text-emerald-600 dark:text-emerald-400">{{ t('admin.settings.usdt.keyConfigured') }}</span>
        </label>
        <input
          v-model="form.tron_api_key"
          type="text"
          class="input font-mono text-sm"
          :placeholder="hasTronApiKey ? t('admin.settings.usdt.keyKeepPlaceholder') : t('admin.settings.usdt.tronApiKeyPlaceholder')"
        />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.tronApiKeyHint') }}</p>
      </div>

      <!-- Rate config -->
      <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
        <div class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.settings.usdt.rateSection') }}</div>
        <div class="flex items-center justify-between">
          <div>
            <label class="text-sm text-gray-700 dark:text-gray-300">{{ t('admin.settings.usdt.rateAutoFetch') }}</label>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.rateAutoFetchHint') }}</p>
          </div>
          <Toggle v-model="form.rate_auto_fetch" />
        </div>
        <div class="mt-3 grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.manualRate') }}</label>
            <input v-model.number="form.manual_rate" type="number" step="0.0001" min="0" class="input font-mono text-sm" placeholder="7.2" />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.manualRateHint') }}</p>
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.rateMarkup') }}</label>
            <input v-model.number="form.rate_markup" type="number" step="0.001" min="0" max="0.5" class="input font-mono text-sm" placeholder="0" />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.rateMarkupHint') }}</p>
          </div>
        </div>
      </div>

      <!-- Advanced -->
      <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
        <div class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.settings.usdt.advanced') }}</div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.tronApiBaseUrl') }}</label>
            <input v-model="form.tron_api_base_url" type="text" class="input font-mono text-sm" placeholder="https://api.trongrid.io" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.amountOffset') }}</label>
            <input v-model.number="form.amount_offset" type="number" step="0.0001" min="0.0001" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.confirmSeconds') }}</label>
            <input v-model.number="form.confirm_seconds" type="number" min="0" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.orderTimeoutSeconds') }}</label>
            <input v-model.number="form.order_timeout_seconds" type="number" min="60" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.monitorIntervalSeconds') }}</label>
            <input v-model.number="form.monitor_interval_seconds" type="number" min="5" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.queryMinutesBack') }}</label>
            <input v-model.number="form.query_minutes_back" type="number" min="1" class="input font-mono text-sm" />
          </div>
        </div>
      </div>

      <!-- Save button -->
      <div class="flex justify-end border-t border-gray-100 pt-4 dark:border-dark-700">
        <button class="btn btn-primary" :disabled="saving" @click="handleSave">
          {{ saving ? t('common.saving') : t('admin.settings.usdt.saveButton') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminUsdtConfigUpdate } from '@/api/admin/usdtConfig'
import { useAppStore } from '@/stores'
import Toggle from '@/components/common/Toggle.vue'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const saving = ref(false)
const hasTronApiKey = ref(false)

const form = reactive<Required<AdminUsdtConfigUpdate>>({
  enabled: false,
  receiving_address: '',
  tron_api_base_url: '',
  tron_api_key: '',
  manual_rate: 7.2,
  rate_auto_fetch: false,
  rate_markup: 0,
  amount_offset: 0.0001,
  confirm_seconds: 60,
  monitor_interval_seconds: 15,
  query_minutes_back: 30,
  order_timeout_seconds: 1800
})

async function loadConfig() {
  loading.value = true
  try {
    const cfg = await adminAPI.usdtConfig.getUsdtConfig()
    form.enabled = cfg.enabled
    form.receiving_address = cfg.receiving_address
    form.tron_api_base_url = cfg.tron_api_base_url
    form.manual_rate = cfg.manual_rate > 0 ? cfg.manual_rate : 7.2
    form.rate_auto_fetch = cfg.rate_auto_fetch
    form.rate_markup = cfg.rate_markup >= 0 ? cfg.rate_markup : 0
    form.amount_offset = cfg.amount_offset > 0 ? cfg.amount_offset : 0.0001
    form.confirm_seconds = cfg.confirm_seconds >= 0 ? cfg.confirm_seconds : 60
    form.monitor_interval_seconds = cfg.monitor_interval_seconds > 0 ? cfg.monitor_interval_seconds : 15
    form.query_minutes_back = cfg.query_minutes_back > 0 ? cfg.query_minutes_back : 30
    form.order_timeout_seconds = cfg.order_timeout_seconds > 0 ? cfg.order_timeout_seconds : 1800
    hasTronApiKey.value = cfg.has_tron_api_key
    form.tron_api_key = ''
  } catch (e) {
    appStore.showError(extractApiErrorMessage(e))
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const payload: AdminUsdtConfigUpdate = {
      enabled: form.enabled,
      receiving_address: form.receiving_address,
      tron_api_base_url: form.tron_api_base_url,
      tron_api_key: form.tron_api_key,
      manual_rate: form.manual_rate,
      rate_auto_fetch: form.rate_auto_fetch,
      rate_markup: form.rate_markup,
      amount_offset: form.amount_offset,
      confirm_seconds: form.confirm_seconds,
      monitor_interval_seconds: form.monitor_interval_seconds,
      query_minutes_back: form.query_minutes_back,
      order_timeout_seconds: form.order_timeout_seconds
    }
    const updated = await adminAPI.usdtConfig.updateUsdtConfig(payload)
    hasTronApiKey.value = updated.has_tron_api_key
    form.tron_api_key = ''
    await appStore.fetchPublicSettings(true)
    appStore.showSuccess(t('admin.settings.usdt.saveSuccess'))
  } catch (e) {
    appStore.showError(extractApiErrorMessage(e))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>
