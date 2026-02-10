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
                <button @click="unbindTelegram" class="btn btn-sm btn-danger" :disabled="saving">
                  {{ t('reseller.settings.telegram.unbind') }}
                </button>
              </div>
              <div v-else class="space-y-2">
                <span class="text-sm text-yellow-600 dark:text-yellow-400">⚠️ {{ t('reseller.settings.telegram.unbound') }}</span>
                <div class="flex items-center gap-2">
                  <button @click="handleGenerateBindCode" class="btn btn-sm btn-secondary" :disabled="generatingCode">
                    {{ t('reseller.settings.telegram.generateBindCode') }}
                  </button>
                </div>
                <div v-if="bindCode" class="mt-2 rounded-md bg-gray-50 dark:bg-gray-800 p-3">
                  <p class="text-sm text-gray-700 dark:text-gray-300">{{ t('reseller.settings.telegram.bindInstructions') }}</p>
                  <code class="mt-1 block text-sm font-mono text-primary-600 dark:text-primary-400">/bind {{ bindCode }}</code>
                </div>
              </div>
            </div>

            <!-- Feature Toggles -->
            <div>
              <label class="label">{{ t('reseller.settings.telegram.features') }}</label>
              <div class="space-y-2">
                <label v-for="feature in tgFeatures" :key="feature.id" class="flex items-center gap-2">
                  <input
                    type="checkbox"
                    :checked="isFeatureEnabled(feature.id)"
                    @change="toggleFeature(feature.id)"
                    class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                  />
                  <span class="text-sm text-gray-700 dark:text-gray-300">{{ feature.label }}</span>
                </label>
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
import { ref, computed, onMounted } from 'vue'
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

const availableLocaleOptions = availableLocales

const settings = ref<Record<string, string>>({
  contact_info: '',
  crypto_addresses: '',
  default_locale: '',
  tg_bot_token: '',
  tg_chat_id: '',
  tg_features: ''
})

const tgFeatures = computed(() => [
  { id: 'admin_keys', label: t('reseller.settings.telegram.featureAdminKeys') },
  { id: 'admin_stats', label: t('reseller.settings.telegram.featureAdminStats') },
  { id: 'admin_notify', label: t('reseller.settings.telegram.featureAdminNotify') },
  { id: 'user_query', label: t('reseller.settings.telegram.featureUserQuery') },
  { id: 'user_notify', label: t('reseller.settings.telegram.featureUserNotify') }
])

const defaultFeatures = 'admin_keys,admin_stats,admin_notify,user_query,user_notify'

function getEnabledFeatures(): Set<string> {
  const raw = settings.value.tg_features || defaultFeatures
  return new Set(raw.split(',').map(f => f.trim()).filter(Boolean))
}

function isFeatureEnabled(id: string): boolean {
  return getEnabledFeatures().has(id)
}

function toggleFeature(id: string) {
  const enabled = getEnabledFeatures()
  if (enabled.has(id)) {
    enabled.delete(id)
  } else {
    enabled.add(id)
  }
  settings.value.tg_features = Array.from(enabled).join(',')
}

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
    const result = await resellerAPI.settings.generateBindCode()
    bindCode.value = result.bind_code
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    generatingCode.value = false
  }
}

async function unbindTelegram() {
  saving.value = true
  try {
    await resellerAPI.settings.update({ tg_chat_id: '' })
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
</script>
