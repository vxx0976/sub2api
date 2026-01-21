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
            <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
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
            </div>
          </div>
        </div>

        <!-- Tier Settings -->
        <div class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <div class="flex items-center justify-between">
              <div>
                <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('admin.settings.recharge.tiers') }}
                </h2>
                <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  {{ t('admin.settings.recharge.tiersHint') }}
                </p>
              </div>
              <button
                type="button"
                @click="addTier"
                class="btn btn-secondary btn-sm"
              >
                <Icon name="plus" size="sm" />
                {{ t('admin.settings.recharge.addTier') }}
              </button>
            </div>
          </div>
          <div class="p-6">
            <div v-if="config.tiers.length === 0" class="text-center py-8 text-gray-500">
              {{ t('admin.settings.recharge.noTiers') }}
            </div>
            <div v-else class="space-y-4">
              <div
                v-for="(tier, index) in config.tiers"
                :key="index"
                class="rounded-lg border border-gray-200 p-4 dark:border-dark-600"
              >
                <div class="mb-3 flex items-center justify-between">
                  <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                    {{ t('admin.settings.recharge.tier') }} {{ index + 1 }}
                  </span>
                  <button
                    type="button"
                    @click="removeTier(index)"
                    class="text-red-600 hover:text-red-700 dark:text-red-400"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </div>
                <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
                  <div>
                    <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">
                      {{ t('admin.settings.recharge.tierMin') }}
                    </label>
                    <input
                      v-model.number="tier.min"
                      type="number"
                      min="0"
                      step="0.01"
                      class="input input-sm"
                      placeholder="0"
                    />
                  </div>
                  <div>
                    <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">
                      {{ t('admin.settings.recharge.tierMax') }}
                      <span class="text-gray-400">({{ t('common.optional') }})</span>
                    </label>
                    <input
                      v-model.number="tier.max"
                      type="number"
                      :min="tier.min"
                      step="0.01"
                      class="input input-sm"
                      :placeholder="t('admin.settings.recharge.tierMaxPlaceholder')"
                    />
                  </div>
                  <div>
                    <label class="mb-1 block text-xs text-gray-600 dark:text-gray-400">
                      {{ t('admin.settings.recharge.tierMultiplier') }}
                    </label>
                    <input
                      v-model.number="tier.multiplier"
                      type="number"
                      min="1"
                      max="10"
                      step="0.01"
                      class="input input-sm"
                      placeholder="1.0"
                    />
                  </div>
                </div>
                <p v-if="tierErrors[index]" class="mt-2 text-xs text-red-600 dark:text-red-400">
                  {{ tierErrors[index] }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Preview -->
        <div v-if="config.tiers.length > 0" class="card">
          <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ t('admin.settings.recharge.preview') }}
            </h2>
          </div>
          <div class="p-6">
            <div class="space-y-2">
              <div
                v-for="(tier, index) in sortedTiers"
                :key="index"
                class="flex items-center justify-between rounded-lg bg-gray-50 px-4 py-3 dark:bg-dark-700"
              >
                <span class="text-sm text-gray-700 dark:text-gray-300">
                  ¥{{ tier.min }} {{ tier.max ? `- ¥${tier.max}` : '+' }}
                </span>
                <span class="font-semibold text-primary-600 dark:text-primary-400">
                  {{ tier.multiplier }}×
                  <span v-if="tier.multiplier > 1.0" class="ml-2 text-xs text-green-600 dark:text-green-400">
                    (+{{ ((tier.multiplier - 1) * 100).toFixed(0) }}%)
                  </span>
                </span>
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
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const saving = ref(false)

const config = ref<RechargeConfig>({
  enabled: false,
  min_amount: 10,
  max_amount: 10000,
  tiers: []
})

const originalConfig = ref<RechargeConfig | null>(null)

// Tier validation errors
const tierErrors = computed(() => {
  const errors: { [key: number]: string } = {}

  config.value.tiers.forEach((tier, index) => {
    // Check min < max
    if (tier.max !== null && tier.max !== undefined && tier.max <= tier.min) {
      errors[index] = t('admin.settings.recharge.tierValidation.maxGreaterThanMin')
    }

    // Check multiplier range
    if (tier.multiplier < 1.0 || tier.multiplier > 10.0) {
      errors[index] = t('admin.settings.recharge.tierValidation.multiplierRange')
    }

    // Check for overlaps with previous tiers
    for (let i = 0; i < index; i++) {
      const prevTier = config.value.tiers[i]
      const prevMax = prevTier.max ?? Infinity

      if (tier.min < prevMax && (tier.max === null || tier.max > prevTier.min)) {
        errors[index] = t('admin.settings.recharge.tierValidation.overlap')
        break
      }
    }
  })

  return errors
})

// Sorted tiers for preview
const sortedTiers = computed(() => {
  return [...config.value.tiers].sort((a, b) => a.min - b.min)
})

// Form validation
const isValid = computed(() => {
  if (config.value.min_amount >= config.value.max_amount) return false
  if (config.value.min_amount < 1) return false
  if (Object.keys(tierErrors.value).length > 0) return false
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

function addTier() {
  const lastTier = config.value.tiers[config.value.tiers.length - 1]
  const newMin = lastTier ? (lastTier.max ?? lastTier.min + 100) : config.value.min_amount

  config.value.tiers.push({
    min: newMin,
    max: null,
    multiplier: 1.0
  })
}

function removeTier(index: number) {
  config.value.tiers.splice(index, 1)
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
