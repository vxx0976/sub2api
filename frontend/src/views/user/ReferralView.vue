<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Page Header -->
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">
            {{ t('referral.title') }}
          </h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
            {{ t('referral.subtitle') }}
          </p>
        </div>
      </div>

    <!-- Referral Code Card -->
    <div class="overflow-hidden rounded-xl bg-gradient-to-r from-primary-500 to-primary-600 p-6 text-white shadow-lg">
      <div class="flex flex-col items-center justify-between gap-4 sm:flex-row">
        <div class="text-center sm:text-left">
          <h2 class="text-lg font-medium opacity-90">{{ t('referral.yourCode') }}</h2>
          <p class="mt-1 text-sm opacity-75">{{ t('referral.shareCodeHint') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <div
            class="rounded-lg bg-white/20 px-6 py-3 font-mono text-2xl font-bold tracking-wider backdrop-blur-sm"
          >
            {{ codeLoading ? '...' : referralCode }}
          </div>
          <button
            @click="copyCode"
            class="rounded-lg bg-white/20 p-3 transition-colors hover:bg-white/30"
            :title="t('common.copiedToClipboard')"
          >
            <ClipboardIcon class="h-5 w-5" />
          </button>
        </div>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-blue-100 p-2.5 dark:bg-blue-900/30">
            <UsersIcon class="h-5 w-5 text-blue-600 dark:text-blue-400" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('referral.totalInvited') }}</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ statsLoading ? '-' : stats?.total_invited || 0 }}
            </p>
          </div>
        </div>
      </div>

      <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-green-100 p-2.5 dark:bg-green-900/30">
            <CheckCircleIcon class="h-5 w-5 text-green-600 dark:text-green-400" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('referral.totalRewarded') }}</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ statsLoading ? '-' : stats?.total_rewarded || 0 }}
            </p>
          </div>
        </div>
      </div>

      <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-amber-100 p-2.5 dark:bg-amber-900/30">
            <ClockIcon class="h-5 w-5 text-amber-600 dark:text-amber-400" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('referral.pendingPayment') }}</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              {{ statsLoading ? '-' : stats?.pending_payment || 0 }}
            </p>
          </div>
        </div>
      </div>

      <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-primary-100 p-2.5 dark:bg-primary-900/30">
            <CurrencyDollarIcon class="h-5 w-5 text-primary-600 dark:text-primary-400" />
          </div>
          <div>
            <p class="text-sm text-gray-500 dark:text-dark-400">{{ t('referral.totalEarnings') }}</p>
            <p class="text-2xl font-bold text-gray-900 dark:text-white">
              ${{ statsLoading ? '-' : (stats?.total_earnings || 0).toFixed(2) }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Reward Rules -->
    <div class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
      <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('referral.rewardRules') }}
      </h3>
      <ul class="space-y-3 text-sm text-gray-600 dark:text-dark-300">
        <li class="flex items-start gap-2">
          <span class="mt-1 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-medium text-primary-600 dark:bg-primary-900/30 dark:text-primary-400">1</span>
          <span>{{ t('referral.rule1') }}</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-1 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-medium text-primary-600 dark:bg-primary-900/30 dark:text-primary-400">2</span>
          <span>{{ t('referral.rule2') }}</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-1 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-medium text-primary-600 dark:bg-primary-900/30 dark:text-primary-400">3</span>
          <span>{{ t('referral.rule3') }}</span>
        </li>
        <li class="flex items-start gap-2">
          <span class="mt-1 flex h-5 w-5 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-medium text-primary-600 dark:bg-primary-900/30 dark:text-primary-400">4</span>
          <span>{{ t('referral.rule4') }}</span>
        </li>
      </ul>
    </div>

    <!-- Invitees List -->
    <div class="rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
      <div class="border-b border-gray-200 px-6 py-4 dark:border-dark-700">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('referral.inviteeList') }}
        </h3>
      </div>

      <div v-if="inviteesLoading" class="flex items-center justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
      </div>

      <div v-else-if="invitees.length === 0" class="py-12 text-center text-gray-500 dark:text-dark-400">
        {{ t('referral.noInvitees') }}
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 dark:bg-dark-700/50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                {{ t('referral.inviteeEmail') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                {{ t('referral.joinedAt') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                {{ t('referral.status') }}
              </th>
              <th class="px-6 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-dark-400">
                {{ t('referral.yourEarning') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-dark-700">
            <tr v-for="invitee in invitees" :key="invitee.id" class="hover:bg-gray-50 dark:hover:bg-dark-700/30">
              <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-900 dark:text-white">
                {{ invitee.email }}
              </td>
              <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-500 dark:text-dark-400">
                {{ formatDate(invitee.joined_at) }}
              </td>
              <td class="whitespace-nowrap px-6 py-4">
                <span
                  :class="[
                    'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
                    invitee.status === 'rewarded'
                      ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                      : 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400'
                  ]"
                >
                  {{ invitee.status === 'rewarded' ? t('referral.statusRewarded') : t('referral.statusPending') }}
                </span>
              </td>
              <td class="whitespace-nowrap px-6 py-4 text-sm text-gray-900 dark:text-white">
                <span v-if="invitee.referrer_earning" class="font-medium text-green-600 dark:text-green-400">
                  +${{ invitee.referrer_earning.toFixed(2) }}
                </span>
                <span v-else class="text-gray-400 dark:text-dark-500">-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-between border-t border-gray-200 px-6 py-4 dark:border-dark-700">
        <p class="text-sm text-gray-500 dark:text-dark-400">
          {{ t('referral.showingPage', { current: currentPage, total: totalPages }) }}
        </p>
        <div class="flex gap-2">
          <button
            @click="loadInvitees(currentPage - 1)"
            :disabled="currentPage <= 1"
            class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-600 dark:text-dark-300 dark:hover:bg-dark-700"
          >
            {{ t('common.back') }}
          </button>
          <button
            @click="loadInvitees(currentPage + 1)"
            :disabled="currentPage >= totalPages"
            class="rounded-lg border border-gray-300 px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-600 dark:text-dark-300 dark:hover:bg-dark-700"
          >
            {{ t('common.next') }}
          </button>
        </div>
      </div>
    </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { useI18n } from 'vue-i18n'
import { AppLayout } from '@/components/layout'
import { useAppStore } from '@/stores'
import { referralAPI, type ReferralStats, type InviteeInfo } from '@/api'

const { t } = useI18n()
const appStore = useAppStore()

// Icon components
const ClipboardIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184'
        })
      ]
    )
}

const UsersIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z'
        })
      ]
    )
}

const CheckCircleIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
        })
      ]
    )
}

const ClockIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z'
        })
      ]
    )
}

const CurrencyDollarIcon = {
  render: () =>
    h(
      'svg',
      { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', 'stroke-width': '1.5' },
      [
        h('path', {
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          d: 'M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z'
        })
      ]
    )
}

// State
const referralCode = ref('')
const codeLoading = ref(true)
const stats = ref<ReferralStats | null>(null)
const statsLoading = ref(true)
const invitees = ref<InviteeInfo[]>([])
const inviteesLoading = ref(true)
const currentPage = ref(1)
const totalPages = ref(1)

// Load referral code
async function loadReferralCode() {
  codeLoading.value = true
  try {
    const response = await referralAPI.getReferralCode()
    referralCode.value = response.code
  } catch (error) {
    console.error('Failed to load referral code:', error)
    appStore.showError(t('referral.loadCodeFailed'))
  } finally {
    codeLoading.value = false
  }
}

// Load stats
async function loadStats() {
  statsLoading.value = true
  try {
    stats.value = await referralAPI.getReferralStats()
  } catch (error) {
    console.error('Failed to load referral stats:', error)
  } finally {
    statsLoading.value = false
  }
}

// Load invitees
async function loadInvitees(page: number = 1) {
  inviteesLoading.value = true
  try {
    const response = await referralAPI.getInvitees(page, 10)
    invitees.value = response.items
    currentPage.value = response.page
    totalPages.value = response.total_pages
  } catch (error) {
    console.error('Failed to load invitees:', error)
  } finally {
    inviteesLoading.value = false
  }
}

// Generate referral link (to home page, register button will carry the ref)
function getReferralLink(): string {
  const baseUrl = window.location.origin
  return `${baseUrl}/home?ref=${referralCode.value}`
}

// Copy referral link to clipboard
async function copyCode() {
  try {
    const link = getReferralLink()
    await navigator.clipboard.writeText(link)
    appStore.showSuccess(t('referral.linkCopied'))
  } catch (error) {
    console.error('Failed to copy link:', error)
    appStore.showError(t('common.copyFailed'))
  }
}

// Format date
function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
  loadReferralCode()
  loadStats()
  loadInvitees()
})
</script>
