<template>
  <BaseDialog
    :show="show"
    :title="t('admin.settings.recharge.title')"
    width="wide"
    @close="$emit('close')"
  >
    <!-- Loading State -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
    </div>

    <!-- Settings Form -->
    <form v-else @submit.prevent="saveSettings" class="space-y-6">
      <!-- Basic Settings -->
      <div class="space-y-4">
        <h3 class="text-base font-medium text-gray-900 dark:text-white">
          {{ t('admin.settings.recharge.basicSettings') }}
        </h3>

        <!-- Enable Recharge -->
        <div class="flex items-center justify-between rounded-lg bg-gray-50 p-4 dark:bg-dark-700">
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
            <div class="peer h-6 w-11 rounded-full bg-gray-200 after:absolute after:left-[2px] after:top-[2px] after:h-5 after:w-5 after:rounded-full after:border after:border-gray-300 after:bg-white after:transition-all after:content-[''] peer-checked:bg-primary-600 peer-checked:after:translate-x-full peer-checked:after:border-white peer-focus:outline-none dark:border-gray-600 dark:bg-gray-700"></div>
          </label>
        </div>

        <!-- Min/Max Amount -->
        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
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
          </div>
        </div>
      </div>

      <!-- Tier Settings -->
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-base font-medium text-gray-900 dark:text-white">
              {{ t('admin.settings.recharge.tiers') }}
            </h3>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.settings.recharge.tiersHint') }}
            </p>
          </div>
          <button
            type="button"
            @click="addTier"
            class="btn btn-secondary btn-sm"
          >
            <Icon name="plus" size="sm" class="mr-1" />
            {{ t('admin.settings.recharge.addTier') }}
          </button>
        </div>

        <div v-if="config.tiers.length === 0" class="rounded-lg bg-gray-50 p-6 text-center text-gray-500 dark:bg-dark-700">
          {{ t('admin.settings.recharge.noTiers') }}
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="(tier, index) in config.tiers"
            :key="index"
            class="rounded-lg border border-gray-200 p-3 dark:border-dark-600"
          >
            <div class="mb-2 flex items-center justify-between">
              <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
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
            <div class="grid grid-cols-3 gap-3">
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

      <!-- Preview -->
      <div v-if="config.tiers.length > 0" class="space-y-3">
        <h3 class="text-base font-medium text-gray-900 dark:text-white">
          {{ t('admin.settings.recharge.preview') }}
        </h3>
        <div class="space-y-2">
          <div
            v-for="(tier, index) in sortedTiers"
            :key="index"
            class="flex items-center justify-between rounded-lg bg-gray-50 px-4 py-2 dark:bg-dark-700"
          >
            <span class="text-sm text-gray-700 dark:text-gray-300">
              ¥{{ tier.min }} {{ tier.max ? `- ¥${tier.max}` : '+' }}
            </span>
            <span class="font-semibold text-primary-600 dark:text-primary-400">
              {{ tier.multiplier }}×
              <span v-if="tier.multiplier > 1.0" class="ml-1 text-xs text-green-600 dark:text-green-400">
                (+{{ ((tier.multiplier - 1) * 100).toFixed(0) }}%)
              </span>
            </span>
          </div>
        </div>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          type="button"
          @click="$emit('close')"
          class="btn btn-secondary"
          :disabled="saving"
        >
          {{ t('common.cancel') }}
        </button>
        <button
          type="button"
          @click="saveSettings"
          class="btn btn-primary"
          :disabled="saving || !isValid"
        >
          <span v-if="saving">{{ t('common.saving') }}</span>
          <span v-else>{{ t('common.save') }}</span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { getRechargeConfigAdmin, updateRechargeConfig, type RechargeConfig } from '@/api/recharge'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  show: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'saved'): void
}>()

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const saving = ref(false)

const config = ref<RechargeConfig>({
  enabled: false,
  min_amount: 10,
  max_amount: 10000,
  tiers: []
})

// Tier validation errors
const tierErrors = computed(() => {
  const errors: { [key: number]: string } = {}

  config.value.tiers.forEach((tier, index) => {
    if (tier.max !== null && tier.max !== undefined && tier.max <= tier.min) {
      errors[index] = t('admin.settings.recharge.tierValidation.maxGreaterThanMin')
    }

    if (tier.multiplier < 1.0 || tier.multiplier > 10.0) {
      errors[index] = t('admin.settings.recharge.tierValidation.multiplierRange')
    }

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

const sortedTiers = computed(() => {
  return [...config.value.tiers].sort((a, b) => a.min - b.min)
})

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
    emit('saved')
    emit('close')
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

// Load config when modal opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadConfig()
  }
}, { immediate: true })
</script>
