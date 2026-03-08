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

        <!-- Pricing Config (only when agent mode enabled) -->
        <div v-if="resellerSettingsStore.isAgentEnabled" class="card p-6">
          <h2 class="mb-1 text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('reseller.settings.pricingSection') }}
          </h2>
          <p class="mb-4 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.settings.pricingHint') }}</p>
          <div class="space-y-4">
            <!-- 进价 + 卖价 → 自动计算倍率 -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="label">{{ t('reseller.settings.platformCost') }}</label>
                <input
                  :value="platformCost || t('reseller.settings.platformCostNotSet')"
                  type="text"
                  class="input cursor-not-allowed bg-gray-50 dark:bg-gray-800"
                  readonly
                />
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.settings.platformCostHint') }}</p>
              </div>
              <div>
                <label class="label">{{ t('reseller.settings.sellingPrice') }}</label>
                <input
                  v-model.number="sellingPrice"
                  type="number"
                  min="0"
                  step="0.01"
                  class="input"
                  placeholder="1.00"
                />
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('reseller.settings.sellingPriceHint') }}</p>
              </div>
            </div>
            <!-- 计算结果展示 -->
            <div v-if="computedMultiplier !== null" class="rounded-lg bg-blue-50 px-4 py-3 dark:bg-blue-900/20">
              <div class="flex items-center justify-between">
                <span class="text-sm text-blue-700 dark:text-blue-300">{{ t('reseller.settings.computedMultiplier') }}</span>
                <span class="text-lg font-bold text-blue-800 dark:text-blue-200">× {{ computedMultiplier }}</span>
              </div>
              <p class="mt-0.5 text-xs text-blue-600 dark:text-blue-400">
                {{ t('reseller.settings.computedMultiplierHint', { cost: platformCost, price: sellingPrice, profit: computedProfit }) }}
              </p>
            </div>
            <div v-else-if="platformCost > 0 && sellingPrice > 0 && sellingPrice < platformCost" class="rounded-lg bg-red-50 px-4 py-3 dark:bg-red-900/20">
              <p class="text-sm text-red-600 dark:text-red-400">{{ t('reseller.settings.pricingError') }}</p>
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
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import { useAppStore, useResellerSettingsStore } from '@/stores'
import { availableLocales } from '@/i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()
const resellerSettingsStore = useResellerSettingsStore()

const loading = ref(true)
const saving = ref(false)
const availableLocaleOptions = availableLocales

const settings = ref<Record<string, string>>({
  contact_info: '',
  default_locale: '',
  contact_wechat: '',
  contact_telegram: '',
})

// 平台进价（从全局设置读取，只读）
const platformCost = ref<number>(0)
const sellingPrice = ref<number>(0)

// 自动计算倍率：卖价 / 进价，保留4位小数
const computedMultiplier = computed<string | null>(() => {
  if (platformCost.value > 0 && sellingPrice.value >= platformCost.value) {
    return (sellingPrice.value / platformCost.value).toFixed(4)
  }
  return null
})

// 每刀利润
const computedProfit = computed<string>(() => {
  if (platformCost.value > 0 && sellingPrice.value >= platformCost.value) {
    return (sellingPrice.value - platformCost.value).toFixed(4)
  }
  return '0'
})

// 倍率变化时同步写入 settings
watch(computedMultiplier, (val) => {
  if (val !== null) {
    settings.value.price_multiplier = val
  }
})

// SimpleAnnouncement list (stored as JSON in settings.announcements)
const announcementList = reactive<{ title: string; date: string }[]>([])

async function loadSettings() {
  loading.value = true
  try {
    const data = await resellerAPI.settings.get()
    // 同步到 store（确保侧边栏 isAgentEnabled 也能实时更新）
    resellerSettingsStore.settings = data
    resellerSettingsStore.loaded = true
    // 从商户自己的 KV 设置中读取管理员配置的进价（只读）
    if (data.platform_cost) {
      platformCost.value = parseFloat(data.platform_cost) || 0
    }
    // Merge loaded settings with defaults
    for (const key of Object.keys(data)) {
      settings.value[key] = data[key]
    }
    // 回显卖价
    if (data.selling_price) sellingPrice.value = parseFloat(data.selling_price) || 0
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
    // 保存卖价，供回显用（进价由管理员配置，不保存）
    if (sellingPrice.value > 0) payload.selling_price = String(sellingPrice.value)
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
