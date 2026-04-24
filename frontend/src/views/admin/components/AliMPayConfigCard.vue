<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.settings.alimpay.title') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.alimpay.description') }}</p>
    </div>

    <div v-if="loading" class="p-6 text-sm text-gray-400">{{ t('common.loading') }}</div>

    <div v-else class="space-y-6 p-6">
      <!-- Enable Toggle -->
      <div class="flex items-center justify-between">
        <div>
          <label class="font-medium text-gray-900 dark:text-white">{{ t('admin.settings.alimpay.enabled') }}</label>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.alimpay.enabledHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <!-- Mode -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.settings.alimpay.mode') }}
        </label>
        <select v-model="form.mode" class="input">
          <option value="business_qr">{{ t('admin.settings.alimpay.modeBusinessQR') }}</option>
          <option value="transfer">{{ t('admin.settings.alimpay.modeTransfer') }}</option>
        </select>
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.alimpay.modeHint') }}</p>
      </div>

      <!-- App ID -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.settings.alimpay.appId') }}
        </label>
        <input v-model="form.app_id" type="text" class="input font-mono text-sm" :placeholder="t('admin.settings.alimpay.appIdPlaceholder')" />
      </div>

      <!-- Private Key -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.settings.alimpay.privateKey') }}
          <span v-if="hasPrivateKey" class="ml-2 text-xs text-emerald-600 dark:text-emerald-400">{{ t('admin.settings.alimpay.keyConfigured') }}</span>
        </label>
        <textarea
          v-model="form.private_key"
          rows="4"
          class="input font-mono text-xs"
          :placeholder="hasPrivateKey ? t('admin.settings.alimpay.keyKeepPlaceholder') : t('admin.settings.alimpay.privateKeyPlaceholder')"
        />
        <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.alimpay.keyHint') }}</p>
      </div>

      <!-- Alipay Public Key -->
      <div>
        <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('admin.settings.alimpay.alipayPublicKey') }}
          <span v-if="hasAlipayPublicKey" class="ml-2 text-xs text-emerald-600 dark:text-emerald-400">{{ t('admin.settings.alimpay.keyConfigured') }}</span>
        </label>
        <textarea
          v-model="form.alipay_public_key"
          rows="3"
          class="input font-mono text-xs"
          :placeholder="hasAlipayPublicKey ? t('admin.settings.alimpay.keyKeepPlaceholder') : t('admin.settings.alimpay.alipayPublicKeyPlaceholder')"
        />
      </div>

      <!-- Mode-specific fields -->
      <template v-if="form.mode === 'business_qr'">
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.settings.alimpay.businessQRURL') }}
          </label>
          <input v-model="form.business_qr_url" type="text" class="input font-mono text-sm" placeholder="https://qr.alipay.com/xxxxxx" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.alimpay.businessQRURLHint') }}</p>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.settings.alimpay.amountOffset') }}
          </label>
          <input v-model.number="form.amount_offset" type="number" step="0.01" min="0.01" class="input font-mono text-sm" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.alimpay.amountOffsetHint') }}</p>
        </div>
      </template>

      <template v-else-if="form.mode === 'transfer'">
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.settings.alimpay.transferUserId') }}
          </label>
          <input v-model="form.transfer_user_id" type="text" class="input font-mono text-sm" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.alimpay.transferUserIdHint') }}</p>
        </div>
      </template>

      <!-- Advanced -->
      <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
        <div class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.settings.alimpay.advanced') }}</div>
        <div class="grid grid-cols-2 gap-4">
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.alimpay.serverURL') }}</label>
            <input v-model="form.server_url" type="text" class="input font-mono text-sm" placeholder="https://openapi.alipay.com/gateway.do" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.alimpay.orderTimeoutSeconds') }}</label>
            <input v-model.number="form.order_timeout_seconds" type="number" min="60" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.alimpay.monitorIntervalSeconds') }}</label>
            <input v-model.number="form.monitor_interval_seconds" type="number" min="5" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.alimpay.matchToleranceSeconds') }}</label>
            <input v-model.number="form.match_tolerance_seconds" type="number" min="60" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.alimpay.queryMinutesBack') }}</label>
            <input v-model.number="form.query_minutes_back" type="number" min="1" class="input font-mono text-sm" />
          </div>
        </div>
      </div>

      <!-- Save button -->
      <div class="flex justify-end border-t border-gray-100 pt-4 dark:border-dark-700">
        <button class="btn btn-primary" :disabled="saving" @click="handleSave">
          {{ saving ? t('common.saving') : t('admin.settings.alimpay.saveButton') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { AdminAliMPayConfigUpdate } from '@/api/admin/alimpayConfig'
import { useAppStore } from '@/stores'
import Toggle from '@/components/common/Toggle.vue'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const saving = ref(false)
const hasPrivateKey = ref(false)
const hasAlipayPublicKey = ref(false)

const form = reactive<Required<AdminAliMPayConfigUpdate>>({
  enabled: false,
  mode: 'business_qr',
  app_id: '',
  private_key: '',
  alipay_public_key: '',
  server_url: '',
  transfer_user_id: '',
  business_qr_url: '',
  business_qr_path: '',
  amount_offset: 0.01,
  match_tolerance_seconds: 300,
  monitor_interval_seconds: 10,
  query_minutes_back: 30,
  order_timeout_seconds: 300
})

async function loadConfig() {
  loading.value = true
  try {
    const cfg = await adminAPI.alimpayConfig.getAliMPayConfig()
    form.enabled = cfg.enabled
    form.mode = cfg.mode || 'business_qr'
    form.app_id = cfg.app_id
    form.server_url = cfg.server_url
    form.transfer_user_id = cfg.transfer_user_id
    form.business_qr_url = cfg.business_qr_url
    form.business_qr_path = cfg.business_qr_path
    form.amount_offset = cfg.amount_offset > 0 ? cfg.amount_offset : 0.01
    form.match_tolerance_seconds = cfg.match_tolerance_seconds > 0 ? cfg.match_tolerance_seconds : 300
    form.monitor_interval_seconds = cfg.monitor_interval_seconds > 0 ? cfg.monitor_interval_seconds : 10
    form.query_minutes_back = cfg.query_minutes_back > 0 ? cfg.query_minutes_back : 30
    form.order_timeout_seconds = cfg.order_timeout_seconds > 0 ? cfg.order_timeout_seconds : 300
    hasPrivateKey.value = cfg.has_private_key
    hasAlipayPublicKey.value = cfg.has_alipay_public_key
    form.private_key = ''
    form.alipay_public_key = ''
  } catch (e) {
    appStore.showError(extractApiErrorMessage(e))
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const payload: AdminAliMPayConfigUpdate = {
      enabled: form.enabled,
      mode: form.mode,
      app_id: form.app_id,
      private_key: form.private_key,
      alipay_public_key: form.alipay_public_key,
      server_url: form.server_url,
      transfer_user_id: form.transfer_user_id,
      business_qr_url: form.business_qr_url,
      business_qr_path: form.business_qr_path,
      amount_offset: form.amount_offset,
      match_tolerance_seconds: form.match_tolerance_seconds,
      monitor_interval_seconds: form.monitor_interval_seconds,
      query_minutes_back: form.query_minutes_back,
      order_timeout_seconds: form.order_timeout_seconds
    }
    const updated = await adminAPI.alimpayConfig.updateAliMPayConfig(payload)
    hasPrivateKey.value = updated.has_private_key
    hasAlipayPublicKey.value = updated.has_alipay_public_key
    form.private_key = ''
    form.alipay_public_key = ''
    appStore.showSuccess(t('admin.settings.alimpay.saveSuccess'))
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
