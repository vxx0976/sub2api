<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div
          class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"
        ></div>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="card p-12 text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/30"
        >
          <Icon name="exclamationCircle" size="xl" class="text-red-500" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('plans.loadError') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ error }}
        </p>
        <button @click="loadPlans" class="btn btn-primary mt-4">
          {{ t('common.retry') }}
        </button>
      </div>

      <!-- Empty State -->
      <div v-else-if="plans.length === 0" class="card p-12 text-center">
        <div
          class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700"
        >
          <Icon name="gift" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('plans.noPlans') }}
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          {{ t('plans.noPlansDesc') }}
        </p>
      </div>

      <!-- Plans Grid -->
      <div v-else>
        <!-- Header -->
        <div class="mb-8 text-center">
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
            {{ t('plans.title') }}
          </h1>
          <p class="mt-2 text-gray-500 dark:text-dark-400">
            {{ t('plans.subtitle') }}
          </p>
        </div>

        <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <!-- PayGo Card (First Position) -->
          <PayGoCard
            :config="rechargeConfig"
            :current-balance="userBalance"
            @recharge="showRechargeDialog = true"
          />
          <div
            v-for="plan in plans"
            :key="plan.id"
            class="group relative flex flex-col overflow-hidden rounded-3xl border transition-all duration-300"
            :class="[
              plan.id === recommendedPlanId
                ? 'border-primary-500/50 bg-gradient-to-b from-primary-50 to-white shadow-lg shadow-primary-500/10 dark:from-primary-900/20 dark:to-dark-800 dark:border-primary-500/30'
                : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow-lg dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500'
            ]"
          >
            <!-- Recommended Badge -->
            <div v-if="plan.id === recommendedPlanId" class="absolute right-4 top-4">
              <span class="inline-flex items-center gap-1 rounded-full bg-primary-500 px-3 py-1 text-xs font-semibold text-white shadow-lg shadow-primary-500/30">
                <Icon name="sparkles" size="xs" />
                {{ t('plans.recommended') }}
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
                    {{ Math.floor(plan.price) }}
                  </span>
                  <span v-if="plan.price % 1 !== 0" class="mb-1 text-xl font-semibold text-gray-500 dark:text-dark-400">
                    .{{ ((plan.price % 1) * 100).toFixed(0).padStart(2, '0') }}
                  </span>
                </div>
                <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('plans.validityPeriod', { days: plan.validity_days }) }}
                </div>
              </div>

              <!-- Features -->
              <ul class="mb-6 flex-1 space-y-3">
                <li class="flex items-start gap-3 text-sm">
                  <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                    <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
                  </div>
                  <span class="text-gray-600 dark:text-gray-300">
                    {{ t('plans.validFor', { days: plan.validity_days }) }}
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
                  <span class="text-gray-600 dark:text-gray-300">
                    {{ t('plans.monthlyLimit', { amount: plan.monthly_limit_usd.toFixed(2) }) }}
                  </span>
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

              <!-- Buy Button -->
              <button
                @click="purchasePlan(plan)"
                :disabled="purchasing === plan.id"
                class="w-full rounded-xl px-4 py-3 text-sm font-semibold transition-all duration-200"
                :class="[
                  plan.id === recommendedPlanId
                    ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/30 hover:bg-primary-600 hover:shadow-xl hover:shadow-primary-500/40 disabled:opacity-50'
                    : 'border border-gray-200 bg-gray-50 text-gray-700 hover:border-gray-300 hover:bg-gray-100 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600 disabled:opacity-50'
                ]"
              >
                <span v-if="purchasing === plan.id" class="flex items-center justify-center gap-2">
                  <div class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"></div>
                  {{ t('plans.processing') }}
                </span>
                <span v-else>{{ t('plans.purchase') }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Recharge Dialog -->
    <RechargeDialog
      :show="showRechargeDialog"
      :config="rechargeConfig"
      @close="showRechargeDialog = false"
      @confirm="handleRecharge"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import plansAPI, { type PurchasablePlan } from '@/api/plans'
import { getRechargeConfig, createRechargeOrder } from '@/api/recharge'
import type { RechargeConfig } from '@/api/recharge'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import PayGoCard from '@/components/plans/PayGoCard.vue'
import RechargeDialog from '@/components/dialogs/RechargeDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const plans = ref<PurchasablePlan[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const purchasing = ref<number | null>(null)

// Recharge related state
const rechargeConfig = ref<RechargeConfig | null>(null)
const showRechargeDialog = ref(false)

// User balance from auth store
const userBalance = computed(() => authStore.user?.balance ?? null)

// 推荐套餐：is_recommended 为 true 的
const recommendedPlanId = computed(() => {
  if (plans.value.length === 0) return null
  const recommended = plans.value.find(p => p.is_recommended)
  return recommended?.id ?? null
})

async function loadPlans() {
  try {
    loading.value = true
    error.value = null
    plans.value = await plansAPI.getPlans()
  } catch (err: any) {
    console.error('Failed to load plans:', err)
    error.value = err.message || t('plans.loadError')
  } finally {
    loading.value = false
  }
}

async function purchasePlan(plan: PurchasablePlan) {
  try {
    purchasing.value = plan.id
    const result = await plansAPI.createOrder(plan.id)
    // Redirect to payment URL
    window.location.href = result.pay_url
  } catch (err: any) {
    console.error('Failed to create order:', err)
    appStore.showError(err.message || t('plans.purchaseError'))
  } finally {
    purchasing.value = null
  }
}

async function loadRechargeConfig() {
  try {
    rechargeConfig.value = await getRechargeConfig()
  } catch (err: any) {
    console.error('Failed to load recharge config:', err)
    // Silently fail - recharge will be disabled if config can't be loaded
  }
}

async function handleRecharge(amount: number) {
  try {
    const result = await createRechargeOrder(amount)
    // Redirect to payment URL
    window.location.href = result.pay_url
  } catch (err: any) {
    console.error('Failed to create recharge order:', err)
    appStore.showError(err.message || t('recharge.rechargeFailed'))
    showRechargeDialog.value = false
  }
}

onMounted(() => {
  loadPlans()
  loadRechargeConfig()
})
</script>
