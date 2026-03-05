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
import { ref, reactive, onMounted } from 'vue'
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
const availableLocaleOptions = availableLocales

const settings = ref<Record<string, string>>({
  contact_info: '',
  default_locale: '',
  contact_wechat: '',
  contact_telegram: '',
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

onMounted(() => {
  loadSettings()
})
</script>
