<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('admin.settings.usdt.title') }}</h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.description') }}</p>
    </div>

    <div v-if="loading" class="p-6 text-sm text-gray-400">{{ t('common.loading') }}</div>

    <div v-else class="space-y-6 p-6">
      <!-- Master enable -->
      <div class="flex items-center justify-between">
        <div>
          <label class="font-medium text-gray-900 dark:text-white">{{ t('admin.settings.usdt.enabled') }}</label>
          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.enabledHint') }}</p>
        </div>
        <Toggle v-model="form.enabled" />
      </div>

      <!-- Limits (USDT) -->
      <div class="grid grid-cols-2 gap-4">
        <div>
          <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.minAmount') }}</label>
          <input v-model.number="form.min_amount" type="number" step="0.1" min="0" class="input font-mono text-sm" />
        </div>
        <div>
          <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.maxAmount') }}</label>
          <input v-model.number="form.max_amount" type="number" step="1" min="0" class="input font-mono text-sm" />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.maxAmountHint') }}</p>
        </div>
      </div>

      <!-- Per-chain config -->
      <div class="border-t border-gray-100 pt-4 dark:border-dark-700">
        <div class="mb-3 text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.settings.usdt.chainsSection') }}</div>
        <div class="space-y-4">
          <div
            v-for="chain in CHAINS"
            :key="chain.key"
            class="rounded-xl border border-gray-200 p-4 dark:border-dark-700"
          >
            <div class="mb-3 flex items-center justify-between">
              <span class="font-medium text-gray-900 dark:text-white">{{ chain.label }}</span>
              <Toggle v-model="form.chains[chain.key].enabled" />
            </div>
            <div class="space-y-3">
              <div>
                <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.chainAddress') }}</label>
                <input v-model="form.chains[chain.key].address" type="text" class="input font-mono text-sm" :placeholder="chain.addrPlaceholder" />
              </div>
              <div>
                <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">
                  {{ chain.apiKeyLabel }}
                  <span v-if="hasApiKey[chain.key]" class="ml-2 text-emerald-600 dark:text-emerald-400">{{ t('admin.settings.usdt.keyConfigured') }}</span>
                </label>
                <input
                  v-model="form.chains[chain.key].api_key"
                  type="text"
                  class="input font-mono text-sm"
                  :placeholder="hasApiKey[chain.key] ? t('admin.settings.usdt.keyKeepPlaceholder') : chain.apiKeyPlaceholder"
                />
                <p v-if="chain.note" class="mt-1 text-xs text-amber-600 dark:text-amber-400">{{ t(chain.note) }}</p>
              </div>
              <div>
                <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.chainApiBaseUrl') }}</label>
                <input v-model="form.chains[chain.key].api_base_url" type="text" class="input font-mono text-sm" :placeholder="chain.baseUrlPlaceholder" />
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Rate -->
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
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.amountOffset') }}</label>
            <input v-model.number="form.amount_offset" type="number" step="0.01" min="0.01" class="input font-mono text-sm" />
          </div>
          <div>
            <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">{{ t('admin.settings.usdt.amountTolerance') }}</label>
            <input v-model.number="form.amount_tolerance" type="number" step="0.001" min="0" class="input font-mono text-sm" />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.settings.usdt.amountToleranceHint') }}</p>
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
import type { AdminUsdtConfigUpdate, AdminUsdtChainConfigUpdate } from '@/api/admin/usdtConfig'
import { useAppStore } from '@/stores'
import Toggle from '@/components/common/Toggle.vue'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()

interface ChainMeta {
  key: string
  label: string
  addrPlaceholder: string
  apiKeyLabel: string
  apiKeyPlaceholder: string
  baseUrlPlaceholder: string
  note?: string
}

const CHAINS: ChainMeta[] = [
  { key: 'trc20', label: 'TRC20 (TRON)', addrPlaceholder: 'T...', apiKeyLabel: 'TronGrid API Key (可选)', apiKeyPlaceholder: '可选，留空走匿名', baseUrlPlaceholder: 'https://api.trongrid.io', note: 'admin.settings.usdt.trc20KeyNote' },
  { key: 'bep20', label: 'BEP20 (BSC)', addrPlaceholder: '0x...', apiKeyLabel: 'RPC API Key (可选)', apiKeyPlaceholder: '公共节点无需填', baseUrlPlaceholder: 'https://bsc-rpc.publicnode.com', note: 'admin.settings.usdt.bep20PaidNote' },
  { key: 'ton', label: 'TON', addrPlaceholder: 'EQ... / UQ...', apiKeyLabel: 'TonCenter API Key', apiKeyPlaceholder: 'TonCenter key', baseUrlPlaceholder: 'https://toncenter.com/api/v3' }
]

const loading = ref(true)
const saving = ref(false)
const hasApiKey = reactive<Record<string, boolean>>({ trc20: false, bep20: false, ton: false })

interface ChainForm {
  enabled: boolean
  address: string
  api_key: string
  api_base_url: string
}

const form = reactive({
  enabled: false,
  min_amount: 0.1,
  max_amount: 0,
  manual_rate: 7.2,
  rate_auto_fetch: false,
  rate_markup: 0,
  amount_offset: 0.05,
  amount_tolerance: 0.01,
  confirm_seconds: 60,
  monitor_interval_seconds: 15,
  query_minutes_back: 30,
  order_timeout_seconds: 1800,
  chains: {
    trc20: { enabled: false, address: '', api_key: '', api_base_url: '' } as ChainForm,
    bep20: { enabled: false, address: '', api_key: '', api_base_url: '' } as ChainForm,
    ton: { enabled: false, address: '', api_key: '', api_base_url: '' } as ChainForm
  } as Record<string, ChainForm>
})

async function loadConfig() {
  loading.value = true
  try {
    const cfg = await adminAPI.usdtConfig.getUsdtConfig()
    form.enabled = cfg.enabled
    form.min_amount = cfg.min_amount > 0 ? cfg.min_amount : 0.1
    form.max_amount = cfg.max_amount >= 0 ? cfg.max_amount : 0
    form.manual_rate = cfg.manual_rate > 0 ? cfg.manual_rate : 7.2
    form.rate_auto_fetch = cfg.rate_auto_fetch
    form.rate_markup = cfg.rate_markup >= 0 ? cfg.rate_markup : 0
    form.amount_offset = cfg.amount_offset > 0 ? cfg.amount_offset : 0.05
    form.amount_tolerance = cfg.amount_tolerance >= 0 ? cfg.amount_tolerance : 0.01
    form.confirm_seconds = cfg.confirm_seconds >= 0 ? cfg.confirm_seconds : 60
    form.monitor_interval_seconds = cfg.monitor_interval_seconds > 0 ? cfg.monitor_interval_seconds : 15
    form.query_minutes_back = cfg.query_minutes_back > 0 ? cfg.query_minutes_back : 30
    form.order_timeout_seconds = cfg.order_timeout_seconds > 0 ? cfg.order_timeout_seconds : 1800
    for (const c of CHAINS) {
      const cc = cfg.chains?.[c.key]
      hasApiKey[c.key] = !!cc?.has_api_key
      form.chains[c.key] = {
        enabled: !!cc?.enabled,
        address: cc?.address || '',
        api_key: '',
        api_base_url: cc?.api_base_url || ''
      }
    }
  } catch (e) {
    appStore.showError(extractApiErrorMessage(e))
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    const chains: Record<string, AdminUsdtChainConfigUpdate> = {}
    for (const c of CHAINS) {
      const f = form.chains[c.key]
      chains[c.key] = {
        enabled: f.enabled,
        address: f.address,
        api_key: f.api_key,
        api_base_url: f.api_base_url
      }
    }
    const payload: AdminUsdtConfigUpdate = {
      enabled: form.enabled,
      min_amount: form.min_amount,
      max_amount: form.max_amount,
      manual_rate: form.manual_rate,
      rate_auto_fetch: form.rate_auto_fetch,
      rate_markup: form.rate_markup,
      amount_offset: form.amount_offset,
      amount_tolerance: form.amount_tolerance,
      confirm_seconds: form.confirm_seconds,
      monitor_interval_seconds: form.monitor_interval_seconds,
      query_minutes_back: form.query_minutes_back,
      order_timeout_seconds: form.order_timeout_seconds,
      chains
    }
    const updated = await adminAPI.usdtConfig.updateUsdtConfig(payload)
    for (const c of CHAINS) {
      hasApiKey[c.key] = !!updated.chains?.[c.key]?.has_api_key
      form.chains[c.key].api_key = ''
    }
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
