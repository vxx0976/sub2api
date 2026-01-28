<template>
  <div class="min-h-screen bg-gray-50 dark:bg-dark-950">
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

        <!-- Subscription Plans -->
        <div class="mt-12">
          <h2 class="mb-6 text-center text-xl font-semibold text-gray-900 dark:text-white">
            {{ t('pricing.subscriptionPlans') }}
          </h2>
          <div class="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            <div
              v-for="plan in subscriptionPlans"
              :key="plan.key"
              class="relative flex flex-col overflow-hidden rounded-3xl border bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800"
              :class="plan.recommended ? 'ring-2 ring-primary-500' : 'border-gray-200'"
            >
              <!-- Recommended Badge -->
              <div v-if="plan.recommended" class="absolute right-4 top-4">
                <span class="inline-flex items-center gap-1 rounded-full bg-primary-500 px-3 py-1 text-xs font-semibold text-white">
                  <Icon name="sparkles" size="xs" />
                  {{ t('home.pricing.recommended') }}
                </span>
              </div>

              <!-- Plan Name -->
              <div class="text-sm font-semibold text-gray-900 dark:text-white">{{ plan.name }}</div>

              <!-- Price -->
              <div class="mt-4 flex items-end gap-2">
                <span class="text-xl font-semibold text-gray-500 dark:text-gray-400">¥</span>
                <span class="text-5xl font-semibold tracking-tight text-gray-900 dark:text-white">{{ plan.price }}</span>
              </div>
              <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
                {{ plan.credit }}
              </div>

              <!-- Promotional Badge -->
              <div class="mt-3 inline-flex items-center gap-1 self-start rounded-md bg-gradient-to-r from-amber-100 to-orange-100 px-2.5 py-1 text-xs font-medium text-amber-700 dark:from-amber-900/30 dark:to-orange-900/30 dark:text-amber-400">
                <Icon name="sparkles" size="xs" />
                <span>{{ plan.unitPrice }}</span>
              </div>

              <!-- Features List -->
              <ul class="mt-6 space-y-3 text-sm text-gray-700 dark:text-gray-300">
                <li v-for="(feature, idx) in plan.features" :key="idx" class="flex items-start gap-2">
                  <Icon
                    name="check"
                    size="sm"
                    class="mt-0.5 flex-shrink-0"
                    :class="plan.recommended ? 'text-primary-500' : 'text-emerald-500'"
                  />
                  <span class="leading-relaxed">{{ feature }}</span>
                </li>
              </ul>

              <!-- Buy Button -->
              <div class="mt-auto pt-6">
                <router-link
                  :to="isAuthenticated ? '/plans' : '/login?redirect=/plans'"
                  class="inline-flex w-full items-center justify-center rounded-2xl px-4 py-3 text-sm font-semibold transition-colors"
                  :class="plan.recommended
                    ? 'bg-primary-500 text-white hover:bg-primary-600'
                    : 'border border-gray-300 bg-white text-gray-900 hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-700 dark:text-white dark:hover:bg-dark-600'"
                >
                  {{ t('home.pricing.buy') }}
                </router-link>
              </div>
            </div>
          </div>
        </div>

        <!-- Other Plans (PayGo & Custom) -->
        <div class="mt-16">
          <h2 class="mb-6 text-center text-xl font-semibold text-gray-900 dark:text-white">
            {{ t('pricing.otherPlans') }}
          </h2>
          <div class="grid gap-6 sm:grid-cols-2">
            <!-- PayGo -->
            <div class="relative flex flex-col overflow-hidden rounded-3xl border border-emerald-200 bg-gradient-to-br from-emerald-50 to-teal-50 p-6 shadow-sm dark:border-emerald-800 dark:from-emerald-900/20 dark:to-teal-900/20">
              <div class="flex items-center gap-3">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-emerald-500">
                  <Icon name="bolt" size="lg" class="text-white" />
                </div>
                <div>
                  <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('pricing.paygo.name') }}</div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('pricing.paygo.tagline') }}</div>
                </div>
              </div>

              <div class="mt-6">
                <div class="flex items-baseline gap-1">
                  <span class="text-3xl font-bold text-gray-900 dark:text-white">¥0.50</span>
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ t('pricing.perDollar') }}</span>
                </div>
                <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  {{ t('pricing.paygo.description') }}
                </p>
              </div>

              <ul class="mt-6 space-y-3 text-sm text-gray-700 dark:text-gray-300">
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
                  <span>{{ t('pricing.paygo.feature1') }}</span>
                </li>
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
                  <span>{{ t('pricing.paygo.feature2') }}</span>
                </li>
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
                  <span>{{ t('pricing.paygo.feature3') }}</span>
                </li>
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-emerald-500" />
                  <span>{{ t('pricing.paygo.feature4') }}</span>
                </li>
              </ul>

              <div class="mt-auto pt-6">
                <router-link
                  :to="isAuthenticated ? '/plans' : '/login?redirect=/plans'"
                  class="inline-flex w-full items-center justify-center rounded-2xl border border-emerald-300 bg-white px-4 py-3 text-sm font-semibold text-emerald-700 transition-colors hover:bg-emerald-50 dark:border-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300 dark:hover:bg-emerald-900/50"
                >
                  {{ t('pricing.paygo.action') }}
                </router-link>
              </div>
            </div>

            <!-- Custom Dedicated Line -->
            <div class="relative flex flex-col overflow-hidden rounded-3xl border border-purple-200 bg-gradient-to-br from-purple-50 to-indigo-50 p-6 shadow-sm dark:border-purple-800 dark:from-purple-900/20 dark:to-indigo-900/20">
              <div class="flex items-center gap-3">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br from-purple-500 to-indigo-600">
                  <Icon name="server" size="lg" class="text-white" />
                </div>
                <div>
                  <div class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('pricing.custom.name') }}</div>
                  <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('pricing.custom.tagline') }}</div>
                </div>
              </div>

              <div class="mt-6">
                <div class="flex items-baseline gap-1">
                  <span class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('pricing.custom.price') }}</span>
                </div>
                <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
                  {{ t('pricing.custom.description') }}
                </p>
              </div>

              <ul class="mt-6 space-y-3 text-sm text-gray-700 dark:text-gray-300">
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-purple-500" />
                  <span>{{ t('pricing.custom.feature1') }}</span>
                </li>
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-purple-500" />
                  <span>{{ t('pricing.custom.feature2') }}</span>
                </li>
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-purple-500" />
                  <span>{{ t('pricing.custom.feature3') }}</span>
                </li>
                <li class="flex items-start gap-2">
                  <Icon name="check" size="sm" class="mt-0.5 flex-shrink-0 text-purple-500" />
                  <span>{{ t('pricing.custom.feature4') }}</span>
                </li>
              </ul>

              <div class="mt-auto pt-6">
                <button
                  @click="showWechatModal = true"
                  class="inline-flex w-full items-center justify-center rounded-2xl border border-purple-300 bg-white px-4 py-3 text-sm font-semibold text-purple-700 transition-colors hover:bg-purple-50 dark:border-purple-700 dark:bg-purple-900/30 dark:text-purple-300 dark:hover:bg-purple-900/50"
                >
                  <Icon name="chatBubble" size="sm" class="mr-2" />
                  {{ t('pricing.custom.action') }}
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
              <img
                src="/wechat-qrcode.png"
                alt="WeChat QR Code"
                class="mx-auto mt-4 h-48 w-48 rounded-lg"
              />
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
import { useAuthStore } from '@/stores'
import PublicHeader from '@/components/layout/PublicHeader.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const authStore = useAuthStore()

const isAuthenticated = computed(() => authStore.isAuthenticated)
const showWechatModal = ref(false)

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
