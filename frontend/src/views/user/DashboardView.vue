<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-12"><LoadingSpinner /></div>
      <template v-else-if="stats">
        <UserDashboardStats :stats="stats" :balance="user?.balance || 0" :is-simple="authStore.isSimpleMode" />

        <!-- Reseller Total Recharge Card (shown for resellers) -->
        <div v-if="authStore.isReseller && commissionSummary" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <div class="card p-4">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.dashboard.totalRechargeAmount') }}</p>
            <p class="mt-1 text-2xl font-bold text-purple-600 dark:text-purple-400">¥{{ Number(commissionSummary.total_recharge).toFixed(2) }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ t('reseller.dashboard.totalRechargeHint') }}</p>
          </div>
          <div class="card p-4">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalCommission') }}</p>
            <p class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">¥{{ Number(commissionSummary.total_commission).toFixed(2) }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ t('reseller.dashboard.totalCommissionHint') }}</p>
          </div>
          <div class="card p-4">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.available') }}</p>
            <p class="mt-1 text-2xl font-bold text-blue-600 dark:text-blue-400">¥{{ Number(commissionSummary.available).toFixed(2) }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ t('reseller.dashboard.availableHint') }}</p>
          </div>
          <div class="card p-4">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.totalUsers') }}</p>
            <p class="mt-1 text-2xl font-bold text-gray-900 dark:text-white">{{ commissionSummary.total_users }}</p>
            <p class="text-xs text-gray-400 dark:text-gray-500 mt-1">{{ t('reseller.dashboard.totalUsersHint') }}</p>
          </div>
        </div>

        <!-- Reseller Today Stats (only shown when merchant mode is enabled) -->
        <div v-if="authStore.isReseller && appStore.resellerAgentEnabled && commissionSummary" class="grid grid-cols-2 gap-4 sm:grid-cols-3">
          <div class="card p-4">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.todayNewUsers') }}</p>
            <p class="mt-1 text-xl font-bold text-orange-600 dark:text-orange-400">{{ commissionSummary.today_new_users }}</p>
          </div>
          <div class="card p-4">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.commissions.todayCost') }}</p>
            <p class="mt-1 text-xl font-bold text-rose-600 dark:text-rose-400">${{ Number(commissionSummary.today_cost).toFixed(2) }}</p>
          </div>
          <div class="card p-4">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.dashboard.todayCommission') }}</p>
            <p class="mt-1 text-xl font-bold text-emerald-600 dark:text-emerald-400">¥{{ Number(commissionSummary.today_cost * commissionSummary.commission_rate).toFixed(2) }}</p>
          </div>
        </div>

        <UserDashboardCharts v-model:startDate="startDate" v-model:endDate="endDate" v-model:granularity="granularity" :loading="loadingCharts" :trend="trendData" :models="modelStats" @dateRangeChange="loadCharts" @granularityChange="loadCharts" />
        <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div class="lg:col-span-2"><UserDashboardRecentUsage :data="recentUsage" :loading="loadingUsage" /></div>
          <div class="lg:col-span-1"><UserDashboardQuickActions /></div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { usageAPI, type UserDashboardStats as UserStatsType } from '@/api/usage'
import { resellerAPI } from '@/api'
import type { CommissionSummary } from '@/api/reseller/commissions'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import UserDashboardStats from '@/components/user/dashboard/UserDashboardStats.vue'
import UserDashboardCharts from '@/components/user/dashboard/UserDashboardCharts.vue'
import UserDashboardRecentUsage from '@/components/user/dashboard/UserDashboardRecentUsage.vue'
import UserDashboardQuickActions from '@/components/user/dashboard/UserDashboardQuickActions.vue'
import type { UsageLog, TrendDataPoint, ModelStat } from '@/types'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()
const user = computed(() => authStore.user)
const stats = ref<UserStatsType | null>(null)
const loading = ref(false)
const loadingUsage = ref(false)
const loadingCharts = ref(false)
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const recentUsage = ref<UsageLog[]>([])
const commissionSummary = ref<CommissionSummary | null>(null)

const formatLD = (d: Date) => d.toISOString().split('T')[0]
const startDate = ref(formatLD(new Date(Date.now() - 6 * 86400000)))
const endDate = ref(formatLD(new Date()))
const granularity = ref('day')

const loadStats = async () => { loading.value = true; try { await authStore.refreshUser(); stats.value = await usageAPI.getDashboardStats() } catch (error) { console.error('Failed to load dashboard stats:', error) } finally { loading.value = false } }
const loadCharts = async () => { loadingCharts.value = true; try { const res = await Promise.all([usageAPI.getDashboardTrend({ start_date: startDate.value, end_date: endDate.value, granularity: granularity.value as any }), usageAPI.getDashboardModels({ start_date: startDate.value, end_date: endDate.value })]); trendData.value = res[0].trend || []; modelStats.value = res[1].models || [] } catch (error) { console.error('Failed to load charts:', error) } finally { loadingCharts.value = false } }
const loadRecent = async () => { loadingUsage.value = true; try { const res = await usageAPI.getByDateRange(startDate.value, endDate.value); recentUsage.value = res.items.slice(0, 5) } catch (error) { console.error('Failed to load recent usage:', error) } finally { loadingUsage.value = false } }
const loadCommissionSummary = async () => { try { commissionSummary.value = await resellerAPI.commissions.getCommissionSummary() } catch { /* silently ignore */ } }

onMounted(() => {
  loadStats()
  loadCharts()
  loadRecent()
  if (authStore.isReseller) loadCommissionSummary()
})
</script>
