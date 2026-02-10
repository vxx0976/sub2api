<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.settings.title') }}</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.settings.description') }}</p>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else>
        <!-- Default Locale -->
        <div class="card p-6">
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('reseller.settings.defaultLocale') }}
          </h2>
          <div class="space-y-4">
            <div>
              <select v-model="settings.default_locale" class="input">
                <option value="">{{ t('reseller.settings.defaultLocaleBrowser') }}</option>
                <option v-for="locale in availableLocaleOptions" :key="locale.code" :value="locale.code">
                  {{ locale.name }}
                </option>
              </select>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.settings.defaultLocaleHint') }}</p>
            </div>
          </div>
        </div>

        <!-- Telegram Bot -->
        <div class="card p-6">
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('reseller.settings.telegram.title') }}
          </h2>
          <div class="space-y-4">
            <!-- Bot Token -->
            <div>
              <label class="label">{{ t('reseller.settings.telegram.botToken') }}</label>
              <input
                v-model="settings.tg_bot_token"
                type="password"
                class="input"
                :placeholder="t('reseller.settings.telegram.botTokenPlaceholder')"
              />
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.settings.telegram.botTokenHint') }}</p>
            </div>

            <!-- Bind Status -->
            <div>
              <label class="label">{{ t('reseller.settings.telegram.bindStatus') }}</label>
              <div v-if="settings.tg_chat_id" class="flex items-center gap-2">
                <span class="text-sm text-green-600 dark:text-green-400">✅ {{ t('reseller.settings.telegram.bound') }} (Chat: {{ settings.tg_chat_id }})</span>
                <button @click="handleUnbindTelegram" class="btn btn-sm btn-danger" :disabled="saving">
                  {{ t('reseller.settings.telegram.unbind') }}
                </button>
              </div>
              <div v-else class="space-y-2">
                <span class="text-sm text-yellow-600 dark:text-yellow-400">⚠️ {{ t('reseller.settings.telegram.unbound') }}</span>
                <div class="flex items-center gap-2">
                  <button @click="handleGenerateBindCode" class="btn btn-sm btn-secondary" :disabled="generatingCode || !settings.tg_bot_token">
                    {{ t('reseller.settings.telegram.generateBindCode') }}
                  </button>
                </div>
                <p v-if="!settings.tg_bot_token" class="text-xs text-amber-600 dark:text-amber-400">{{ t('reseller.settings.telegram.saveTokenFirst') }}</p>
                <div v-if="bindCode" class="mt-2 rounded-md bg-gray-50 dark:bg-gray-800 p-3">
                  <p class="text-sm text-gray-700 dark:text-gray-300">{{ t('reseller.settings.telegram.bindInstructions') }}</p>
                  <code class="mt-1 block text-sm font-mono text-primary-600 dark:text-primary-400">/bind {{ bindCode }}</code>
                  <p v-if="polling" class="mt-2 flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
                    <svg class="h-3 w-3 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>
                    {{ t('reseller.settings.telegram.waitingForBind') }}
                  </p>
                </div>
              </div>
            </div>

          </div>
        </div>

        <!-- Save Button -->
        <div class="flex justify-end">
          <button
            @click="saveSettings"
            class="btn btn-primary"
            :disabled="saving"
          >
            {{ saving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import { useAppStore } from '@/stores'
import { availableLocales } from '@/i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const saving = ref(false)
const generatingCode = ref(false)
const bindCode = ref('')
const polling = ref(false)
let pollTimer: ReturnType<typeof setInterval> | null = null

const availableLocaleOptions = availableLocales

const settings = ref<Record<string, string>>({
  contact_info: '',
  crypto_addresses: '',
  default_locale: '',
  tg_bot_token: '',
  tg_chat_id: ''
})

async function loadSettings() {
  loading.value = true
  try {
    const data = await resellerAPI.settings.get()
    // Merge loaded settings with defaults
    for (const key of Object.keys(data)) {
      settings.value[key] = data[key]
    }
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function saveSettings() {
  saving.value = true
  try {
    const payload: Record<string, string> = {}
    for (const [key, value] of Object.entries(settings.value)) {
      // Don't send protected keys
      if (key === 'tg_chat_id' || key === 'tg_bind_code') continue
      payload[key] = value
    }
    await resellerAPI.settings.update(payload)
    appStore.showSuccess(t('common.saveSuccess'))
  } catch (error: any) {
    appStore.showError(error.message || t('common.saveFailed'))
  } finally {
    saving.value = false
  }
}

async function handleGenerateBindCode() {
  generatingCode.value = true
  try {
    // Save settings first (including bot token) to ensure backend has the token
    const payload: Record<string, string> = {}
    for (const [key, value] of Object.entries(settings.value)) {
      if (key === 'tg_chat_id' || key === 'tg_bind_code') continue
      payload[key] = value
    }
    await resellerAPI.settings.update(payload)

    const result = await resellerAPI.settings.generateBindCode()
    bindCode.value = result.bind_code
    // Start polling to detect when bind succeeds
    startBindPolling()
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    generatingCode.value = false
  }
}

function startBindPolling() {
  stopBindPolling()
  polling.value = true
  let attempts = 0
  const maxAttempts = 60 // poll for ~3 minutes
  pollTimer = setInterval(async () => {
    attempts++
    if (attempts > maxAttempts) {
      stopBindPolling()
      return
    }
    try {
      const data = await resellerAPI.settings.get()
      if (data.tg_chat_id) {
        settings.value.tg_chat_id = data.tg_chat_id
        bindCode.value = ''
        stopBindPolling()
        appStore.showSuccess(t('reseller.settings.telegram.bindSuccess'))
      }
    } catch {
      // ignore polling errors
    }
  }, 3000)
}

function stopBindPolling() {
  polling.value = false
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function handleUnbindTelegram() {
  saving.value = true
  try {
    await resellerAPI.settings.unbindTelegram()
    settings.value.tg_chat_id = ''
    appStore.showSuccess(t('common.saveSuccess'))
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadSettings()
})

onBeforeUnmount(() => {
  stopBindPolling()
})
</script>
