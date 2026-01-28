<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Welcome Banner -->
      <div class="rounded-2xl bg-gradient-to-r from-primary-500 to-primary-600 p-6 text-white shadow-lg">
        <div class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
          <div>
            <h1 class="text-2xl font-bold">
              {{ t('consoleHome.welcome') }}{{ userName ? `，${userName}` : '' }}
            </h1>
            <p class="mt-1 text-primary-100">{{ currentDate }}</p>
          </div>
          <div class="flex items-center gap-2 rounded-xl bg-white/20 px-4 py-2 backdrop-blur">
            <Icon name="dollar" size="md" />
            <span class="text-sm font-medium">{{ t('consoleHome.balance') }}: ${{ balance.toFixed(2) }}</span>
          </div>
        </div>
      </div>

      <!-- Announcements -->
      <div v-if="announcements.length > 0" class="card p-6">
        <div class="mb-4 flex items-center gap-2">
          <Icon name="chatBubble" size="md" class="text-amber-500" />
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('consoleHome.announcements') }}
          </h2>
        </div>
        <div class="space-y-3">
          <div
            v-for="(announcement, index) in announcements"
            :key="index"
            class="flex items-start gap-3 rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
          >
            <div class="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-full bg-amber-100 dark:bg-amber-900/30">
              <Icon name="infoCircle" size="sm" class="text-amber-600 dark:text-amber-400" />
            </div>
            <div class="flex-1">
              <p class="text-sm font-medium text-gray-900 dark:text-white">{{ announcement.title }}</p>
              <p v-if="announcement.date" class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ announcement.date }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Promotions -->
      <div>
        <div class="mb-4 flex items-center gap-2">
          <Icon name="gift" size="md" class="text-purple-500" />
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('consoleHome.promotions') }}
          </h2>
        </div>
        <div class="grid gap-6 sm:grid-cols-2">
          <!-- Recharge Bonus Card -->
          <div class="card overflow-hidden flex flex-col">
            <div class="bg-gradient-to-br from-emerald-500 to-teal-600 p-5">
              <div class="flex items-center gap-3">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur">
                  <Icon name="dollar" size="lg" class="text-white" />
                </div>
                <h3 class="text-lg font-bold text-white">{{ t('consoleHome.rechargeBonus.title') }}</h3>
              </div>
            </div>
            <div class="flex flex-1 flex-col p-5">
              <p class="text-sm text-gray-600 dark:text-dark-400">
                {{ t('consoleHome.rechargeBonus.description') }}
              </p>
              <ul class="mt-4 space-y-2">
                <li v-for="tier in displayTiers" :key="tier.min" class="flex items-center justify-between text-sm text-gray-700 dark:text-dark-300">
                  <span class="flex items-center gap-2">
                    <Icon name="check" size="sm" class="text-emerald-500" />
                    <span>¥{{ tier.min }}{{ tier.max ? `-${tier.max}` : '+' }}</span>
                  </span>
                  <span class="flex items-center gap-2">
                    <span class="text-xs text-gray-500">¥{{ (1 / tier.multiplier).toFixed(2) }}{{ t('consoleHome.rechargeBonus.perDollar') }}</span>
                    <span class="font-medium text-emerald-600 dark:text-emerald-400">{{ tier.multiplier }}×</span>
                    <span v-if="tier.multiplier > 1" class="text-xs font-medium text-orange-500">+{{ ((tier.multiplier - 1) * 100).toFixed(0) }}%</span>
                  </span>
                </li>
              </ul>
              <router-link
                to="/plans"
                class="mt-5 inline-flex w-full items-center justify-center gap-2 rounded-xl bg-emerald-500 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-emerald-600"
              >
                {{ t('consoleHome.rechargeBonus.action') }}
                <Icon name="arrowRight" size="sm" />
              </router-link>
            </div>
          </div>

          <!-- Referral Reward Card -->
          <div class="card overflow-hidden flex flex-col">
            <div class="bg-gradient-to-br from-purple-500 to-indigo-600 p-5">
              <div class="flex items-center gap-3">
                <div class="flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur">
                  <Icon name="users" size="lg" class="text-white" />
                </div>
                <h3 class="text-lg font-bold text-white">{{ t('consoleHome.referralReward.title') }}</h3>
              </div>
            </div>
            <div class="flex flex-1 flex-col p-5">
              <p class="text-sm text-gray-600 dark:text-dark-400">
                {{ t('consoleHome.referralReward.description') }}
              </p>
              <div class="mt-4 grid grid-cols-2 gap-2">
                <div v-for="plan in referralPlans" :key="plan.price" class="flex items-center gap-2 text-sm text-gray-700 dark:text-dark-300">
                  <Icon name="check" size="sm" class="text-purple-500 flex-shrink-0" />
                  <span>{{ t('consoleHome.referralReward.planItem', { price: plan.price, reward: plan.reward }) }}</span>
                </div>
                <div class="col-span-2 flex items-center gap-2 text-sm text-gray-500 dark:text-dark-400">
                  <Icon name="chatBubble" size="sm" class="text-purple-400 flex-shrink-0" />
                  <span>{{ t('consoleHome.referralReward.moreInfo') }}</span>
                </div>
              </div>
              <router-link
                to="/referral"
                class="mt-5 inline-flex w-full items-center justify-center gap-2 rounded-xl bg-purple-500 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-purple-600"
              >
                {{ t('consoleHome.referralReward.action') }}
                <Icon name="arrowRight" size="sm" />
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <!-- Quick Links -->
      <div>
        <div class="mb-4 flex items-center gap-2">
          <Icon name="bolt" size="md" class="text-blue-500" />
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t('consoleHome.quickLinks') }}
          </h2>
        </div>
        <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
          <router-link
            v-for="link in quickLinks"
            :key="link.path"
            :to="link.path"
            class="card card-hover flex flex-col items-center gap-3 p-5 text-center"
          >
            <div
              class="flex h-12 w-12 items-center justify-center rounded-xl"
              :class="link.bgClass"
            >
              <Icon :name="link.icon" size="lg" class="text-white" />
            </div>
            <span class="text-sm font-medium text-gray-900 dark:text-white">{{ link.label }}</span>
          </router-link>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { settingsAPI } from '@/api/settings'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Announcement } from '@/types'

const { t, locale } = useI18n()
const authStore = useAuthStore()

const userName = computed(() => authStore.user?.username || authStore.user?.email?.split('@')[0] || '')
const balance = computed(() => authStore.user?.balance || 0)

const currentDate = computed(() => {
  const now = new Date()
  return now.toLocaleDateString(locale.value === 'zh' ? 'zh-CN' : 'en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    weekday: 'long'
  })
})

// Hardcoded recharge tiers (no API loading needed)
const rechargeTiers = [
  { min: 10, max: 199.99, multiplier: 2.0 },
  { min: 200, max: 499.99, multiplier: 2.1 },
  { min: 500, max: null, multiplier: 2.2 }
]

// Display top 3 tiers for the card
const displayTiers = rechargeTiers.slice(0, 3)

// Referral plans with 10% reward (minimum $1)
const referralPlans = [
  { price: 9.9, reward: 1 },
  { price: 19.9, reward: 2 },
  { price: 49.9, reward: 5 },
  { price: 99.9, reward: 10 }
]

// Default announcements (hardcoded)
const announcements = ref<Announcement[]>([
  { title: t('consoleHome.announcement1'), date: '2026-01-28' },
  { title: t('consoleHome.announcement2'), date: '2026-01-28' }
])

// Optionally load announcements from API to override defaults
onMounted(async () => {
  try {
    const settings = await settingsAPI.getPublicSettings()
    if (settings.announcements && Array.isArray(settings.announcements) && settings.announcements.length > 0) {
      announcements.value = settings.announcements
    }
  } catch (error) {
    // Keep default announcements on error
    console.error('Failed to load announcements:', error)
  }
})

const quickLinks = computed(() => [
  {
    path: '/subscriptions',
    icon: 'creditCard' as const,
    label: t('consoleHome.links.subscriptions'),
    bgClass: 'bg-gradient-to-br from-blue-500 to-blue-600'
  },
  {
    path: '/plans',
    icon: 'gift' as const,
    label: t('consoleHome.links.plans'),
    bgClass: 'bg-gradient-to-br from-emerald-500 to-emerald-600'
  },
  {
    path: '/keys',
    icon: 'key' as const,
    label: t('consoleHome.links.apiKeys'),
    bgClass: 'bg-gradient-to-br from-amber-500 to-amber-600'
  },
  {
    path: '/referral',
    icon: 'userPlus' as const,
    label: t('consoleHome.links.referral'),
    bgClass: 'bg-gradient-to-br from-purple-500 to-purple-600'
  }
])
</script>
