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
                <!-- Promotional Badge -->
                <div v-if="plan.monthly_limit_usd" class="mt-2 inline-flex items-center gap-1 rounded-md bg-gradient-to-r from-amber-100 to-orange-100 dark:from-amber-900/30 dark:to-orange-900/30 px-2 py-1 text-xs font-medium text-amber-700 dark:text-amber-300">
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
                    {{ t('plans.monthlyLimit', { amount: plan.monthly_limit_usd.toFixed(0) }) }}
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

              <!-- External Buy Link (Taobao) -->
              <div v-if="plan.external_buy_url" class="mb-2 text-right">
                <a
                  :href="plan.external_buy_url"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="inline-flex items-center gap-1 text-xs text-gray-500 hover:text-primary-500 dark:text-dark-400 dark:hover:text-primary-400 transition-colors"
                >
                  {{ t('plans.buyOnTaobao') }}
                  <Icon name="externalLink" size="xs" />
                </a>
              </div>

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

          <!-- PayGo Card (Before Enterprise) -->
          <PayGoCard
            :config="rechargeConfig"
            :current-balance="userBalance"
            @recharge="showRechargeDialog = true"
          />

          <!-- Crypto Payment Card -->
          <div
            v-if="cryptoAddresses.length > 0"
            class="group relative flex flex-col overflow-hidden rounded-3xl border border-gray-200 bg-white transition-all duration-300 hover:border-gray-300 hover:shadow-lg dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500"
          >
            <div class="flex flex-1 flex-col p-6">
              <div class="mb-4">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('plans.cryptoBanner.title') }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('plans.cryptoBanner.descriptionPrefix') }}
                  <a href="https://t.me/Umasou" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-0.5 font-medium text-blue-500 hover:text-blue-600 dark:text-blue-400 dark:hover:text-blue-300">
                    Telegram
                    <Icon name="externalLink" size="xs" />
                  </a>
                  {{ t('plans.cryptoBanner.descriptionSuffix') }}
                </p>
              </div>

              <div class="flex-1 space-y-2">
                <div
                  v-for="(addr, index) in cryptoAddresses"
                  :key="index"
                  class="rounded-xl bg-gray-50 px-3 py-2.5 dark:bg-dark-700"
                >
                  <div class="flex items-center justify-between gap-2">
                    <span class="inline-flex items-center rounded-md bg-gray-200 px-1.5 py-0.5 text-[11px] font-medium text-gray-600 dark:bg-dark-600 dark:text-dark-300">
                      {{ addr.chain }}
                    </span>
                    <button
                      @click="copyCryptoAddress(addr.address)"
                      class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-200 hover:text-gray-600 dark:text-dark-500 dark:hover:bg-dark-600 dark:hover:text-dark-300"
                      :title="t('plans.cryptoBanner.copySuccess')"
                    >
                      <Icon name="clipboard" size="xs" />
                    </button>
                  </div>
                  <p class="mt-1 font-mono text-xs text-gray-500 dark:text-dark-400 break-all leading-relaxed">
                    {{ addr.address }}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <!-- Custom Enterprise Card -->
          <div
            class="group relative flex flex-col overflow-hidden rounded-3xl border border-purple-300/50 bg-gradient-to-br from-purple-50 via-white to-amber-50 shadow-lg shadow-purple-500/10 transition-all duration-300 hover:shadow-xl dark:border-purple-500/30 dark:from-purple-900/20 dark:via-dark-800 dark:to-amber-900/10"
          >
            <!-- Premium Badge -->
            <div class="absolute right-4 top-4">
              <span class="inline-flex items-center gap-1 rounded-full bg-gradient-to-r from-purple-500 to-amber-500 px-3 py-1 text-xs font-semibold text-white shadow-lg">
                <Icon name="sparkles" size="xs" />
                {{ t('plans.enterprise.badge') }}
              </span>
            </div>

            <!-- Card Content -->
            <div class="flex flex-1 flex-col p-6">
              <!-- Plan Name -->
              <div class="mb-4">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ t('plans.enterprise.name') }}
                </h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400 line-clamp-2">
                  {{ t('plans.enterprise.description') }}
                </p>
              </div>

              <!-- Price - Contact Us -->
              <div class="mb-6">
                <div class="flex items-end gap-1">
                  <span class="text-3xl font-bold tracking-tight text-purple-600 dark:text-purple-400">
                    {{ t('plans.enterprise.contactUs') }}
                  </span>
                </div>
                <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                  {{ t('plans.enterprise.customized') }}
                </div>
              </div>

              <!-- Features -->
              <ul class="mb-6 flex-1 space-y-3">
                <li class="flex items-start gap-3 text-sm">
                  <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-purple-100 dark:bg-purple-900/30">
                    <Icon name="check" size="xs" class="text-purple-600 dark:text-purple-400" />
                  </div>
                  <span class="text-gray-600 dark:text-gray-300">{{ t('plans.enterprise.features.dedicated') }}</span>
                </li>
                <li class="flex items-start gap-3 text-sm">
                  <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-purple-100 dark:bg-purple-900/30">
                    <Icon name="check" size="xs" class="text-purple-600 dark:text-purple-400" />
                  </div>
                  <span class="text-gray-600 dark:text-gray-300">{{ t('plans.enterprise.features.unlimited') }}</span>
                </li>
                <li class="flex items-start gap-3 text-sm">
                  <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-purple-100 dark:bg-purple-900/30">
                    <Icon name="check" size="xs" class="text-purple-600 dark:text-purple-400" />
                  </div>
                  <span class="text-gray-600 dark:text-gray-300">{{ t('plans.enterprise.features.priority') }}</span>
                </li>
                <li class="flex items-start gap-3 text-sm">
                  <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-purple-100 dark:bg-purple-900/30">
                    <Icon name="check" size="xs" class="text-purple-600 dark:text-purple-400" />
                  </div>
                  <span class="text-gray-600 dark:text-gray-300">{{ t('plans.enterprise.features.sla') }}</span>
                </li>
              </ul>

              <!-- Contact Button -->
              <a
                href="javascript:void(0)"
                @click="contactEnterprise"
                class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-purple-500 to-purple-600 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-purple-500/30 transition-all duration-200 hover:from-purple-600 hover:to-purple-700 hover:shadow-xl hover:shadow-purple-500/40"
              >
                <Icon name="chat" size="sm" />
                {{ t('plans.enterprise.contact') }}
              </a>
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

    <!-- Contact Dialog -->
    <ContactDialog
      :show="showContactDialog"
      @close="showContactDialog = false"
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
import ContactDialog from '@/components/dialogs/ContactDialog.vue'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

const plans = ref<PurchasablePlan[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const purchasing = ref<number | null>(null)

// Crypto addresses from public settings
interface CryptoAddress {
  chain: string
  address: string
}
const cryptoAddresses = computed<CryptoAddress[]>(() => {
  const raw = appStore.cachedPublicSettings?.crypto_addresses
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed
  } catch { /* ignore */ }
  return []
})

async function copyCryptoAddress(address: string) {
  try {
    await navigator.clipboard.writeText(address)
    appStore.showSuccess(t('plans.cryptoBanner.copySuccess'))
  } catch {
    // fallback
    const ta = document.createElement('textarea')
    ta.value = address
    document.body.appendChild(ta)
    ta.select()
    document.execCommand('copy')
    document.body.removeChild(ta)
    appStore.showSuccess(t('plans.cryptoBanner.copySuccess'))
  }
}

// Recharge related state
const rechargeConfig = ref<RechargeConfig | null>(null)
const showRechargeDialog = ref(false)
const showContactDialog = ref(false)

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

function contactEnterprise() {
  showContactDialog.value = true
}

onMounted(() => {
  loadPlans()
  loadRechargeConfig()
})
</script>
