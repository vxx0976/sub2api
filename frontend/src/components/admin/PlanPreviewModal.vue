<template>
  <BaseDialog
    :show="show"
    :title="t('admin.groups.planPreview.title')"
    width="wide"
    @close="$emit('close')"
  >
    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center py-12">
      <div
        class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
      ></div>
    </div>

    <!-- Empty State -->
    <div v-else-if="plans.length === 0" class="py-12 text-center">
      <div
        class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
      >
        <Icon name="gift" size="xl" class="text-gray-400" />
      </div>
      <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.groups.planPreview.noPlans') }}
      </h3>
      <p class="text-gray-500 dark:text-dark-400">
        {{ t('admin.groups.planPreview.noPlansDesc') }}
      </p>
    </div>

    <!-- Plans Grid (Same as user PlansView) -->
    <div v-else class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="plan in plans"
        :key="plan.id"
        class="group relative flex flex-col overflow-hidden rounded-3xl border transition-all duration-300"
        :class="[
          plan.is_recommended
            ? 'border-primary-500/50 bg-gradient-to-b from-primary-50 to-white shadow-lg shadow-primary-500/10 dark:from-primary-900/20 dark:to-dark-800 dark:border-primary-500/30'
            : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow-lg dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500'
        ]"
      >
        <!-- Recommended Badge -->
        <div v-if="plan.is_recommended" class="absolute right-4 top-4">
          <span class="inline-flex items-center gap-1 rounded-full bg-primary-500 px-3 py-1 text-xs font-semibold text-white shadow-lg shadow-primary-500/30">
            <Icon name="sparkles" size="xs" />
            {{ t('plans.recommended') }}
          </span>
        </div>

        <!-- Purchasable Badge -->
        <div v-if="!plan.is_purchasable" class="absolute left-4 top-4">
          <span class="inline-flex items-center gap-1 rounded-full bg-gray-500 px-3 py-1 text-xs font-semibold text-white">
            {{ t('admin.groups.planPreview.notPurchasable') }}
          </span>
        </div>

        <!-- Card Content -->
        <div class="flex flex-1 flex-col p-6">
          <!-- Plan Name -->
          <div class="mb-4">
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
              {{ plan.name }}
            </h3>
            <p v-if="plan.description" class="mt-1 text-sm text-gray-500 dark:text-dark-400 line-clamp-2">
              {{ plan.description }}
            </p>
          </div>

          <!-- Price -->
          <div class="mb-6">
            <div class="flex items-end gap-1">
              <span class="text-lg font-medium text-gray-500 dark:text-dark-400">¥</span>
              <span class="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">
                {{ Math.floor(plan.price || 0) }}
              </span>
              <span v-if="plan.price && plan.price % 1 !== 0" class="mb-1 text-xl font-semibold text-gray-500 dark:text-dark-400">
                .{{ ((plan.price % 1) * 100).toFixed(0).padStart(2, '0') }}
              </span>
            </div>
            <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('plans.validityPeriod', { days: plan.default_validity_days || 30 }) }}
            </div>
            <!-- Promotional Badge -->
            <div v-if="plan.monthly_limit_usd && plan.price" class="mt-2 inline-flex items-center gap-1 rounded-md bg-gradient-to-r from-amber-100 to-orange-100 dark:from-amber-900/30 dark:to-orange-900/30 px-2 py-1 text-xs font-medium text-amber-700 dark:text-amber-300">
              <Icon name="sparkles" size="xs" />
              <span>{{ t('plans.pricePerDollar', { price: (plan.price / plan.monthly_limit_usd).toFixed(2) }) }}</span>
            </div>
          </div>

          <!-- Features -->
          <ul class="mb-6 flex-1 space-y-3">
            <li class="flex items-start gap-3 text-sm">
              <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <span class="text-gray-600 dark:text-gray-300">
                {{ t('plans.validFor', { days: plan.default_validity_days || 30 }) }}
              </span>
            </li>
            <li v-if="plan.daily_limit_usd" class="flex items-start gap-3 text-sm">
              <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <span class="text-gray-600 dark:text-gray-300">
                {{ t('plans.dailyLimit', { amount: plan.daily_limit_usd.toFixed(2) }) }}
              </span>
            </li>
            <li v-if="plan.weekly_limit_usd" class="flex items-start gap-3 text-sm">
              <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <span class="text-gray-600 dark:text-gray-300">
                {{ t('plans.weeklyLimit', { amount: plan.weekly_limit_usd.toFixed(2) }) }}
              </span>
            </li>
            <li v-if="plan.monthly_limit_usd" class="flex items-start gap-3 text-sm">
              <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <div class="flex-1">
                <div class="text-gray-600 dark:text-gray-300">
                  {{ t('plans.monthlyLimit', { amount: plan.monthly_limit_usd.toFixed(0) }) }}
                </div>
              </div>
            </li>
            <li v-if="!plan.daily_limit_usd && !plan.weekly_limit_usd && !plan.monthly_limit_usd" class="flex items-start gap-3 text-sm">
              <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
              </div>
              <span class="text-gray-600 dark:text-gray-300">
                {{ t('plans.unlimited') }}
              </span>
            </li>
          </ul>

          <!-- Preview-only button -->
          <button
            disabled
            class="w-full rounded-xl px-4 py-3 text-sm font-semibold transition-all duration-200 cursor-not-allowed opacity-60"
            :class="[
              plan.is_recommended
                ? 'bg-primary-500 text-white'
                : 'border border-gray-200 bg-gray-50 text-gray-700 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200'
            ]"
          >
            {{ t('admin.groups.planPreview.previewOnly') }}
          </button>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button
          type="button"
          @click="$emit('close')"
          class="btn btn-secondary"
        >
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { Group } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{
  show: boolean
}>()

defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()

const loading = ref(false)
const plans = ref<Group[]>([])

async function loadPlans() {
  loading.value = true
  try {
    // Load all subscription type groups
    const response = await adminAPI.groups.list(1, 100, {
      status: 'active'
    })
    // Filter to only show subscription type groups
    plans.value = response.items.filter(g => g.subscription_type === 'subscription')
    // Sort by sort_order
    plans.value.sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
  } catch (error) {
    console.error('Failed to load plans:', error)
    plans.value = []
  } finally {
    loading.value = false
  }
}

// Load plans when modal opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    loadPlans()
  }
}, { immediate: true })
</script>
