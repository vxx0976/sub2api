<template>
  <!-- Reseller domain with purchase_url: embed in iframe -->
  <div v-if="isResellerDomain && resellerPurchaseUrl" class="flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <PublicHeader />
    <main class="flex-1">
      <iframe
        :src="resellerPurchaseUrl"
        class="h-[calc(100vh-57px)] w-full border-0"
        allowfullscreen
        referrerpolicy="no-referrer-when-downgrade"
      ></iframe>
    </main>
  </div>

  <!-- Normal pricing page -->
  <div v-else class="min-h-screen bg-gray-50 dark:bg-dark-950">
    <!-- Header -->
    <PublicHeader />

    <!-- Main Content -->
    <main class="relative z-10 py-16 lg:py-24">
      <div class="mx-auto max-w-7xl px-6">
        <div class="mx-auto max-w-3xl text-center">
          <h1 class="text-balance text-3xl font-semibold tracking-tight text-gray-900 dark:text-white sm:text-4xl lg:text-5xl">
            {{ t('home.pricing.title') }}
          </h1>
          <p class="mt-4 text-sm leading-relaxed text-gray-600 dark:text-gray-400 sm:text-base lg:text-lg">
            {{ t('home.pricing.subtitle') }}
          </p>
        </div>

        <!-- All Plans -->
        <div class="mt-12">
          <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            <!-- Subscription Plans -->
            <div
              v-for="plan in subscriptionPlans"
              :key="plan.key"
              class="group relative flex flex-col overflow-hidden rounded-3xl border transition-all duration-300"
              :class="[
                plan.recommended
                  ? 'border-primary-500/50 bg-gradient-to-b from-primary-50 to-white shadow-lg shadow-primary-500/10 dark:from-primary-900/20 dark:to-dark-800 dark:border-primary-500/30'
                  : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow-lg dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500'
              ]"
            >
              <!-- Recommended Badge -->
              <div v-if="plan.recommended" class="absolute right-4 top-4">
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
                </div>

                <!-- Price -->
                <div class="mb-6">
                  <div class="flex items-end gap-1">
                    <span class="text-lg font-medium text-gray-500 dark:text-dark-400">¥</span>
                    <span class="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">
                      {{ plan.price }}
                    </span>
                  </div>
                  <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ plan.credit }}
                  </div>
                  <!-- Promotional Badge -->
                  <div class="mt-2 inline-flex items-center gap-1 rounded-md bg-gradient-to-r from-amber-100 to-orange-100 dark:from-amber-900/30 dark:to-orange-900/30 px-2 py-1 text-xs font-medium text-amber-700 dark:text-amber-300">
                    <Icon name="sparkles" size="xs" />
                    <span>{{ plan.unitPrice }}</span>
                  </div>
                </div>

                <!-- Features List -->
                <ul class="mb-6 flex-1 space-y-3">
                  <li v-for="(feature, idx) in plan.features" :key="idx" class="flex items-start gap-3 text-sm">
                    <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                      <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
                    </div>
                    <span class="text-gray-600 dark:text-gray-300">{{ feature }}</span>
                  </li>
                </ul>

                <!-- Buy Button -->
                <router-link
                  :to="isAuthenticated ? '/plans' : '/login?redirect=/plans'"
                  class="w-full rounded-xl px-4 py-3 text-sm font-semibold text-center transition-all duration-200"
                  :class="[
                    plan.recommended
                      ? 'bg-primary-500 text-white shadow-lg shadow-primary-500/30 hover:bg-primary-600 hover:shadow-xl hover:shadow-primary-500/40'
                      : 'border border-gray-200 bg-gray-50 text-gray-700 hover:border-gray-300 hover:bg-gray-100 dark:border-dark-600 dark:bg-dark-700 dark:text-gray-200 dark:hover:bg-dark-600'
                  ]"
                >
                  {{ t('plans.purchase') }}
                </router-link>
              </div>
            </div>

            <!-- PayGo Plan -->
            <div class="group relative flex flex-col overflow-hidden rounded-3xl border border-emerald-300/50 bg-gradient-to-br from-emerald-50 via-white to-teal-50 shadow-lg shadow-emerald-500/10 transition-all duration-300 hover:shadow-xl dark:border-emerald-500/30 dark:from-emerald-900/20 dark:via-dark-800 dark:to-teal-900/10">
              <!-- Card Content -->
              <div class="flex flex-1 flex-col p-6">
                <!-- Plan Name -->
                <div class="mb-4">
                  <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                    💰 {{ t('plans.paygo.title') }}
                  </h3>
                  <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ t('plans.paygo.description') }}
                  </p>
                </div>

                <!-- Price -->
                <div class="mb-6">
                  <div class="text-3xl font-bold tracking-tight text-emerald-600 dark:text-emerald-400">
                    {{ t('recharge.rechargeNow') }}
                  </div>
                  <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ t('plans.paygo.description') }}
                  </div>
                </div>

                <!-- Features List -->
                <ul class="mb-6 flex-1 space-y-3">
                  <li class="flex items-start gap-3 text-sm">
                    <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                      <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
                    </div>
                    <span class="text-gray-600 dark:text-gray-300">{{ t('plans.paygo.features.anyAmount') }}</span>
                  </li>
                  <li class="flex items-start gap-3 text-sm">
                    <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                      <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
                    </div>
                    <span class="text-gray-600 dark:text-gray-300">{{ t('plans.paygo.features.payAsYouGo') }}</span>
                  </li>
                  <li class="flex items-start gap-3 text-sm">
                    <div class="mt-0.5 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                      <Icon name="check" size="xs" class="text-emerald-600 dark:text-emerald-400" />
                    </div>
                    <span class="text-gray-600 dark:text-gray-300">{{ t('plans.paygo.features.neverExpires') }}</span>
                  </li>
                </ul>

                <!-- Action Button -->
                <router-link
                  :to="isAuthenticated ? '/plans' : '/login?redirect=/plans'"
                  class="flex w-full items-center justify-center gap-2 rounded-xl bg-primary-500 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-primary-500/30 transition-all duration-200 hover:bg-primary-600 hover:shadow-xl hover:shadow-primary-500/40"
                >
                  <Icon name="plus" size="sm" />
                  {{ t('recharge.rechargeNow') }}
                </router-link>
              </div>
            </div>

            <!-- Custom Dedicated Line Plan -->
            <div class="group relative flex flex-col overflow-hidden rounded-3xl border border-purple-300/50 bg-gradient-to-br from-purple-50 via-white to-amber-50 shadow-lg shadow-purple-500/10 transition-all duration-300 hover:shadow-xl dark:border-purple-500/30 dark:from-purple-900/20 dark:via-dark-800 dark:to-amber-900/10">
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
                  <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ t('plans.enterprise.description') }}
                  </p>
                </div>

                <!-- Price - Contact Us -->
                <div class="mb-6">
                  <div class="text-3xl font-bold tracking-tight text-purple-600 dark:text-purple-400">
                    {{ t('plans.enterprise.contactUs') }}
                  </div>
                  <div class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                    {{ t('plans.enterprise.customized') }}
                  </div>
                </div>

                <!-- Features List -->
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
                <button
                  @click="showWechatModal = true"
                  class="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-purple-500 to-purple-600 px-4 py-3 text-sm font-semibold text-white shadow-lg shadow-purple-500/30 transition-all duration-200 hover:from-purple-600 hover:to-purple-700 hover:shadow-xl hover:shadow-purple-500/40"
                >
                  <Icon name="chat" size="sm" />
                  {{ t('plans.enterprise.contact') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <p class="mx-auto mt-10 max-w-4xl text-center text-xs leading-relaxed text-gray-500 dark:text-gray-400">
          {{ t('home.pricing.note') }}
        </p>
      </div>
    </main>

    <!-- WeChat QR Code Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="showWechatModal"
          class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
          @click.self="showWechatModal = false"
        >
          <div class="relative w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-800">
            <button
              @click="showWechatModal = false"
              class="absolute right-4 top-4 text-gray-400 transition-colors hover:text-gray-600 dark:hover:text-gray-300"
            >
              <Icon name="x" size="md" />
            </button>
            <div class="text-center">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ t('pricing.custom.wechatTitle') }}
              </h3>
              <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">
                {{ t('pricing.custom.wechatDesc') }}
              </p>
              <div class="mx-auto mt-4 inline-flex items-center gap-2 rounded-full bg-green-50 px-4 py-2 dark:bg-green-900/20">
                <svg class="h-4 w-4 text-green-600 dark:text-green-400" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 01.213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 00.167-.054l1.903-1.114a.864.864 0 01.717-.098 10.16 10.16 0 002.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178A1.17 1.17 0 014.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178 1.17 1.17 0 01-1.162-1.178c0-.651.52-1.18 1.162-1.18z"/>
                </svg>
                <span class="text-sm font-medium text-green-700 dark:text-green-300">mayione1</span>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import PublicHeader from '@/components/layout/PublicHeader.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const showWechatModal = ref(false)

// Reseller domain detection
const isResellerDomain = computed(() => !!appStore.cachedPublicSettings?.reseller_id)
const resellerPurchaseUrl = computed(() => {
  const s = appStore.cachedPublicSettings
  return (s?.purchase_enabled && s?.purchase_url) ? s.purchase_url : ''
})

// Subscription plans
const planConfig = [
  { key: 'starter', price: '9.9', recommended: false },
  { key: 'lite', price: '19.9', recommended: false },
  { key: 'standard', price: '49.9', recommended: true },
  { key: 'pro', price: '99.9', recommended: false }
]

const subscriptionPlans = computed(() =>
  planConfig.map(({ key, price, recommended }) => ({
    key,
    recommended,
    price,
    name: t(`pricing.plans.${key}.name`),
    credit: t(`pricing.plans.${key}.credit`),
    unitPrice: t(`pricing.plans.${key}.unitPrice`),
    features: [
      t(`pricing.plans.${key}.f1`),
      t(`pricing.plans.${key}.f2`),
      t(`pricing.plans.${key}.f3`),
      t(`pricing.plans.${key}.f4`)
    ]
  }))
)
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
