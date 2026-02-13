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

        <!-- Contact Info -->
        <div class="card p-6">
          <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('reseller.settings.contactSection') }}
          </h2>
          <div class="space-y-4">
            <div>
              <label class="label">{{ t('reseller.settings.contactWechat') }}</label>
              <input
                v-model="settings.contact_wechat"
                type="text"
                class="input"
              />
            </div>
            <div>
              <label class="label">{{ t('reseller.settings.contactTelegram') }}</label>
              <input
                v-model="settings.contact_telegram"
                type="text"
                class="input"
              />
            </div>
          </div>
        </div>

        <!-- Announcements -->
        <div class="card p-6">
          <h2 class="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('reseller.settings.announcementsSection') }}
          </h2>
          <p class="mb-4 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.settings.announcementsHint') }}</p>
          <div class="space-y-3">
            <div
              v-for="(item, index) in announcementList"
              :key="index"
              class="flex items-center gap-2"
            >
              <input
                v-model="item.title"
                type="text"
                class="input flex-1"
                :placeholder="t('reseller.settings.announcementTitle')"
              />
              <input
                v-model="item.date"
                type="text"
                class="input w-32"
                :placeholder="t('reseller.settings.announcementDate')"
              />
              <button
                type="button"
                @click="announcementList.splice(index, 1)"
                class="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg text-gray-400 transition-colors hover:bg-red-50 hover:text-red-500 dark:hover:bg-red-900/20"
              >
                <Icon name="x" size="sm" />
              </button>
            </div>
            <button
              type="button"
              @click="announcementList.push({ title: '', date: '' })"
              class="btn btn-sm btn-secondary"
            >
              {{ t('reseller.settings.addAnnouncement') }}
            </button>
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
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import { useAppStore } from '@/stores'
import { availableLocales } from '@/i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

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
  contact_wechat: '',
  contact_telegram: '',
  tg_bot_token: '',
  tg_chat_id: ''
})

// SimpleAnnouncement list (stored as JSON in settings.announcements)
const announcementList = reactive<{ title: string; date: string }[]>([])

async function loadSettings() {
  loading.value = true
  try {
    const data = await resellerAPI.settings.get()
    // Merge loaded settings with defaults
    for (const key of Object.keys(data)) {
      settings.value[key] = data[key]
    }
    // Parse announcements JSON
    announcementList.length = 0
    if (data.announcements) {
      try {
        const arr = JSON.parse(data.announcements)
        if (Array.isArray(arr)) {
          arr.forEach((a: any) => announcementList.push({ title: a.title || '', date: a.date || '' }))
        }
      } catch { /* ignore parse errors */ }
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
    // Serialize announcements (filter out empty titles)
    const validAnnouncements = announcementList.filter(a => a.title.trim())
    payload.announcements = validAnnouncements.length > 0 ? JSON.stringify(validAnnouncements) : ''
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
