<template>
  <AppLayout>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">
          {{ t('admin.referrals.title') }}
        </h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          {{ t('admin.referrals.description') }}
        </p>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-6">
      <div class="card p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.referrals.stats.totalRecords') }}</div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ stats.total_records }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.referrals.stats.totalReferrers') }}</div>
        <div class="mt-1 text-2xl font-semibold text-gray-900 dark:text-white">{{ stats.total_referrers }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.referrals.stats.pending') }}</div>
        <div class="mt-1 text-2xl font-semibold text-amber-600 dark:text-amber-400">{{ stats.total_pending }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.referrals.stats.rewarded') }}</div>
        <div class="mt-1 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">{{ stats.total_rewarded }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.referrals.stats.referrerPaid') }}</div>
        <div class="mt-1 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">${{ stats.total_referrer_paid.toFixed(2) }}</div>
      </div>
      <div class="card p-4">
        <div class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.referrals.stats.inviteePaid') }}</div>
        <div class="mt-1 text-2xl font-semibold text-emerald-600 dark:text-emerald-400">${{ stats.total_invitee_paid.toFixed(2) }}</div>
      </div>
    </div>

    <!-- Search -->
    <div class="card p-4">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div class="flex-1">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('admin.referrals.searchPlaceholder')"
            class="input w-full"
            @keyup.enter="handleSearch"
          />
        </div>
        <button class="btn btn-primary" @click="handleSearch">
          <Icon name="search" size="sm" class="mr-1" />
          {{ t('common.search') }}
        </button>
      </div>
    </div>

    <!-- Table -->
    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 dark:bg-dark-800">
            <tr>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.referrals.table.referrer') }}
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.referrals.table.invitee') }}
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.referrals.table.status') }}
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.referrals.table.referrerReward') }}
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.referrals.table.inviteeReward') }}
              </th>
              <th class="px-4 py-3 text-left text-xs font-medium uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.referrals.table.createdAt') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-dark-700">
            <tr v-if="loading">
              <td colspan="6" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                <Icon name="refresh" size="lg" class="mx-auto animate-spin" />
              </td>
            </tr>
            <tr v-else-if="!records || records.length === 0">
              <td colspan="6" class="px-4 py-8 text-center text-gray-500 dark:text-gray-400">
                {{ t('admin.referrals.noRecords') }}
              </td>
            </tr>
            <tr v-for="record in records" :key="record.invitee_id" class="hover:bg-gray-50 dark:hover:bg-dark-800/50">
              <td class="whitespace-nowrap px-4 py-3">
                <div class="text-sm text-gray-900 dark:text-white">{{ record.referrer_email }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">ID: {{ record.referrer_id }}</div>
              </td>
              <td class="whitespace-nowrap px-4 py-3">
                <div class="text-sm text-gray-900 dark:text-white">{{ record.invitee_email }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">ID: {{ record.invitee_id }}</div>
              </td>
              <td class="whitespace-nowrap px-4 py-3">
                <span v-if="record.status === 'rewarded'" class="inline-flex items-center rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-medium text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-400">
                  {{ t('admin.referrals.status.rewarded') }}
                </span>
                <span v-else class="inline-flex items-center rounded-full bg-amber-100 px-2.5 py-0.5 text-xs font-medium text-amber-800 dark:bg-amber-900/30 dark:text-amber-400">
                  {{ t('admin.referrals.status.pending') }}
                </span>
              </td>
              <td class="whitespace-nowrap px-4 py-3">
                <template v-if="record.status === 'rewarded'">
                  <span v-if="record.referrer_reward > 0" class="font-medium text-emerald-600 dark:text-emerald-400">
                    +${{ record.referrer_reward.toFixed(2) }}
                  </span>
                  <span v-else class="text-gray-400">
                    {{ record.skip_referrer_reason || '-' }}
                  </span>
                </template>
                <span v-else class="text-gray-400">-</span>
              </td>
              <td class="whitespace-nowrap px-4 py-3">
                <template v-if="record.status === 'rewarded'">
                  <span class="font-medium text-emerald-600 dark:text-emerald-400">
                    +${{ record.invitee_reward.toFixed(2) }}
                  </span>
                </template>
                <span v-else class="text-gray-400">-</span>
              </td>
              <td class="whitespace-nowrap px-4 py-3 text-sm text-gray-500 dark:text-gray-400">
                {{ record.created_at }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="flex items-center justify-between border-t border-gray-200 px-4 py-3 dark:border-dark-700">
        <div class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('common.pagination.showing', { from: (currentPage - 1) * pageSize + 1, to: Math.min(currentPage * pageSize, total), total }) }}
        </div>
        <div class="flex gap-2">
          <button
            class="btn btn-secondary btn-sm"
            :disabled="currentPage <= 1"
            @click="goToPage(currentPage - 1)"
          >
            {{ t('common.pagination.previous') }}
          </button>
          <button
            class="btn btn-secondary btn-sm"
            :disabled="currentPage >= totalPages"
            @click="goToPage(currentPage + 1)"
          >
            {{ t('common.pagination.next') }}
          </button>
        </div>
      </div>
    </div>
  </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { getReferrals, getReferralStats, type AdminReferralRecord, type AdminReferralStats } from '@/api/admin/referrals'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const loading = ref(false)
const records = ref<AdminReferralRecord[]>([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const totalPages = ref(0)
const searchQuery = ref('')

const stats = reactive<AdminReferralStats>({
  total_records: 0,
  total_referrers: 0,
  total_invitees: 0,
  total_pending: 0,
  total_rewarded: 0,
  total_referrer_paid: 0,
  total_invitee_paid: 0
})

async function loadStats() {
  try {
    const data = await getReferralStats()
    Object.assign(stats, data)
  } catch (error) {
    console.error('Failed to load stats:', error)
  }
}

async function loadRecords() {
  loading.value = true
  try {
    const response = await getReferrals(currentPage.value, pageSize.value, searchQuery.value || undefined)
    records.value = response.items || []
    total.value = response.total || 0
    totalPages.value = response.total_pages || 0
  } catch (error) {
    console.error('Failed to load records:', error)
    records.value = []
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  currentPage.value = 1
  loadRecords()
}

function goToPage(page: number) {
  currentPage.value = page
  loadRecords()
}

onMounted(() => {
  loadStats()
  loadRecords()
})
</script>
