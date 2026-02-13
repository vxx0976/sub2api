<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.users.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.users.description') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('reseller.users.search')"
            class="input w-64"
            @input="handleSearchInput"
          />
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <!-- Users Table -->
      <div v-else-if="users.length > 0" class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.users.email') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.users.username') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.users.balance') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.users.status') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.users.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.users.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="user in users"
                :key="user.id"
                class="border-b border-gray-50 dark:border-dark-800"
              >
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ user.email }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ user.username || '-' }}</span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <div class="group relative">
                      <button
                        class="font-medium text-gray-900 underline decoration-dashed decoration-gray-300 underline-offset-4 transition-colors hover:text-primary-600 dark:text-white dark:decoration-dark-500 dark:hover:text-primary-400"
                        @click="handleBalanceHistory(user)"
                      >
                        ${{ user.balance.toFixed(2) }}
                      </button>
                      <div class="pointer-events-none absolute bottom-full left-1/2 z-50 mb-1.5 -translate-x-1/2 whitespace-nowrap rounded bg-gray-900 px-2 py-1 text-xs text-white opacity-0 shadow-lg transition-opacity duration-75 group-hover:opacity-100 dark:bg-dark-600">
                        {{ t('reseller.users.balanceHistory') }}
                        <div class="absolute left-1/2 top-full -translate-x-1/2 border-4 border-transparent border-t-gray-900 dark:border-t-dark-600"></div>
                      </div>
                    </div>
                    <button
                      @click.stop="handleDeposit(user)"
                      class="rounded px-2 py-0.5 text-xs font-medium text-emerald-600 transition-colors hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-900/20"
                    >
                      {{ t('reseller.users.deposit') }}
                    </button>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="user.status === 'active'
                      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'"
                  >
                    {{ user.status === 'active' ? t('reseller.users.active') : t('reseller.users.disabled') }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(user.created_at) }}</span>
                </td>
                <td class="px-4 py-3">
                  <button
                    @click="openActionMenu(user, $event)"
                    class="action-menu-trigger flex items-center gap-1 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:hover:bg-dark-700 dark:hover:text-white"
                    :class="{ 'bg-gray-100 text-gray-900 dark:bg-dark-700 dark:text-white': activeMenuId === user.id }"
                  >
                    <Icon name="more" size="sm" />
                    <span class="text-xs">{{ t('reseller.users.more') }}</span>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <Icon name="users" size="lg" class="mb-3 text-gray-300 dark:text-gray-600" />
        <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.users.empty') }}</p>
      </div>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />
    </div>

    <!-- Action Menu (Teleported) -->
    <Teleport to="body">
      <div
        v-if="activeMenuId !== null && menuPosition"
        class="action-menu-content fixed z-[9999] w-48 overflow-hidden rounded-xl bg-white shadow-lg ring-1 ring-black/5 dark:bg-dark-800 dark:ring-white/10"
        :style="{ top: menuPosition.top + 'px', left: menuPosition.left + 'px' }"
      >
        <div class="py-1">
          <template v-for="user in users" :key="user.id">
            <template v-if="user.id === activeMenuId">
              <!-- Deposit -->
              <button
                @click="handleDeposit(user); closeActionMenu()"
                class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                <Icon name="plus" size="sm" class="text-emerald-500" :stroke-width="2" />
                {{ t('reseller.users.deposit') }}
              </button>

              <!-- Withdraw -->
              <button
                @click="handleWithdraw(user); closeActionMenu()"
                class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                <svg class="h-4 w-4 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                </svg>
                {{ t('reseller.users.withdraw') }}
              </button>

              <!-- Balance History -->
              <button
                @click="handleBalanceHistory(user); closeActionMenu()"
                class="flex w-full items-center gap-2 px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700"
              >
                <Icon name="dollar" size="sm" class="text-gray-400" :stroke-width="2" />
                {{ t('reseller.users.balanceHistory') }}
              </button>
            </template>
          </template>
        </div>
      </div>
    </Teleport>

    <!-- Balance Modal (Deposit / Withdraw) -->
    <BaseDialog
      :show="showBalanceModal"
      :title="balanceOperation === 'add' ? t('reseller.users.deposit') : t('reseller.users.withdraw')"
      width="narrow"
      @close="closeBalanceModal"
    >
      <form v-if="balanceUser" id="reseller-balance-form" @submit.prevent="handleBalanceSubmit" class="space-y-5">
        <div class="flex items-center gap-3 rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
          <div class="flex h-10 w-10 items-center justify-center rounded-full bg-primary-100">
            <span class="text-lg font-medium text-primary-700">{{ balanceUser.email.charAt(0).toUpperCase() }}</span>
          </div>
          <div class="flex-1">
            <p class="font-medium text-gray-900 dark:text-white">{{ balanceUser.email }}</p>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.users.currentBalance') }}: ${{ formatBalance(balanceUser.balance) }}</p>
          </div>
        </div>
        <div>
          <label class="input-label">{{ balanceOperation === 'add' ? t('reseller.users.depositAmount') : t('reseller.users.withdrawAmount') }}</label>
          <div class="relative flex gap-2">
            <div class="relative flex-1">
              <div class="absolute left-3 top-1/2 -translate-y-1/2 font-medium text-gray-500">$</div>
              <input v-model.number="balanceForm.amount" type="number" step="any" min="0" required class="input pl-8" />
            </div>
            <button v-if="balanceOperation === 'subtract'" type="button" @click="fillAllBalance" class="btn btn-secondary whitespace-nowrap">{{ t('reseller.users.withdrawAll') }}</button>
          </div>
        </div>
        <div>
          <label class="input-label">{{ t('reseller.users.notes') }}</label>
          <textarea
            v-model="balanceForm.notes"
            rows="3"
            class="input"
            :placeholder="balanceOperation === 'add' ? t('reseller.users.depositNotesPlaceholder') : t('reseller.users.withdrawNotesPlaceholder')"
          ></textarea>
        </div>
        <div v-if="balanceForm.amount > 0" class="rounded-xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-800 dark:bg-blue-900/20">
          <div class="flex items-center justify-between text-sm">
            <span class="text-gray-600 dark:text-gray-400">{{ t('reseller.users.newBalance') }}:</span>
            <span class="font-bold text-gray-900 dark:text-white">${{ formatBalance(calculateNewBalance()) }}</span>
          </div>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeBalanceModal" class="btn btn-secondary">{{ t('common.cancel') }}</button>
          <button
            type="submit"
            form="reseller-balance-form"
            :disabled="balanceSubmitting || !balanceForm.amount"
            class="btn"
            :class="balanceOperation === 'add' ? 'bg-emerald-600 text-white hover:bg-emerald-700' : 'btn-danger'"
          >
            {{ balanceSubmitting ? t('common.saving') : t('common.confirm') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Balance History Modal -->
    <BaseDialog
      :show="showHistoryModal"
      :title="t('reseller.users.balanceHistoryTitle')"
      width="wide"
      :close-on-click-outside="true"
      :z-index="40"
      @close="closeHistoryModal"
    >
      <div v-if="historyUser" class="space-y-4">
        <!-- User header -->
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full bg-primary-100 dark:bg-primary-900/30">
              <span class="text-lg font-medium text-primary-700 dark:text-primary-300">
                {{ historyUser.email.charAt(0).toUpperCase() }}
              </span>
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <p class="truncate font-medium text-gray-900 dark:text-white">{{ historyUser.email }}</p>
                <span
                  v-if="historyUser.username"
                  class="flex-shrink-0 rounded bg-primary-50 px-1.5 py-0.5 text-xs text-primary-600 dark:bg-primary-900/20 dark:text-primary-400"
                >
                  {{ historyUser.username }}
                </span>
              </div>
              <p class="text-xs text-gray-400 dark:text-dark-500">
                {{ t('reseller.users.createdAt') }}: {{ formatDateTime(historyUser.created_at) }}
              </p>
            </div>
            <div class="flex-shrink-0 text-right">
              <p class="text-xs text-gray-500 dark:text-dark-400">{{ t('reseller.users.currentBalance') }}</p>
              <p class="text-xl font-bold text-gray-900 dark:text-white">
                ${{ historyUser.balance?.toFixed(2) || '0.00' }}
              </p>
            </div>
          </div>
          <div class="mt-2.5 flex items-center justify-end border-t border-gray-200/60 pt-2.5 dark:border-dark-600/60">
            <p class="flex-shrink-0 text-xs text-gray-500 dark:text-dark-400">
              {{ t('reseller.users.totalRecharged') }}: <span class="font-semibold text-emerald-600 dark:text-emerald-400">${{ historyTotalRecharged.toFixed(2) }}</span>
            </p>
          </div>
        </div>

        <!-- Action buttons -->
        <div class="flex items-center gap-3">
          <button
            @click="handleDepositFromHistory"
            class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700"
          >
            <Icon name="plus" size="sm" class="text-emerald-500" :stroke-width="2" />
            {{ t('reseller.users.deposit') }}
          </button>
          <button
            @click="handleWithdrawFromHistory"
            class="flex items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-700 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-300 dark:hover:bg-dark-700"
          >
            <svg class="h-4 w-4 text-amber-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
            </svg>
            {{ t('reseller.users.withdraw') }}
          </button>
        </div>

        <!-- Loading -->
        <div v-if="historyLoading" class="flex justify-center py-8">
          <svg class="h-8 w-8 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
        </div>

        <!-- Empty state -->
        <div v-else-if="historyItems.length === 0" class="py-8 text-center">
          <p class="text-sm text-gray-500">{{ t('reseller.users.noHistory') }}</p>
        </div>

        <!-- History list -->
        <div v-else class="max-h-[28rem] space-y-3 overflow-y-auto">
          <div
            v-for="item in historyItems"
            :key="item.id"
            class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-600 dark:bg-dark-800"
          >
            <div class="flex items-start justify-between">
              <div class="flex items-start gap-3">
                <div
                  :class="[
                    'flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg',
                    getHistoryIconBg(item)
                  ]"
                >
                  <Icon :name="getHistoryIconName(item)" size="sm" :class="getHistoryIconColor(item)" />
                </div>
                <div>
                  <p class="text-sm font-medium text-gray-900 dark:text-white">
                    {{ getHistoryItemTitle(item) }}
                  </p>
                  <p
                    v-if="item.notes"
                    class="mt-0.5 text-xs text-gray-500 dark:text-dark-400"
                    :title="item.notes"
                  >
                    {{ item.notes.length > 60 ? item.notes.substring(0, 55) + '...' : item.notes }}
                  </p>
                  <p class="mt-0.5 text-xs text-gray-400 dark:text-dark-500">
                    {{ formatDateTime(item.used_at || item.created_at) }}
                  </p>
                </div>
              </div>
              <div class="text-right">
                <p :class="['text-sm font-semibold', getHistoryValueColor(item)]">
                  {{ formatHistoryValue(item) }}
                </p>
                <p
                  v-if="isAdminType(item.type)"
                  class="text-xs text-gray-400 dark:text-dark-500"
                >
                  {{ t('reseller.users.adminAdjustment') }}
                </p>
                <p
                  v-else
                  class="font-mono text-xs text-gray-400 dark:text-dark-500"
                >
                  {{ item.code.slice(0, 8) }}...
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Pagination -->
        <div v-if="historyTotalPages > 1" class="flex items-center justify-center gap-2 pt-2">
          <button
            :disabled="historyCurrentPage <= 1"
            class="btn btn-secondary px-3 py-1 text-sm"
            @click="loadHistory(historyCurrentPage - 1)"
          >
            {{ t('pagination.previous') }}
          </button>
          <span class="text-sm text-gray-500 dark:text-dark-400">
            {{ historyCurrentPage }} / {{ historyTotalPages }}
          </span>
          <button
            :disabled="historyCurrentPage >= historyTotalPages"
            class="btn btn-secondary px-3 py-1 text-sm"
            @click="loadHistory(historyCurrentPage + 1)"
          >
            {{ t('pagination.next') }}
          </button>
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { ResellerBalanceHistoryItem } from '@/api/reseller/users'
import type { User } from '@/types'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const users = ref<User[]>([])
const searchQuery = ref('')
let searchTimer: ReturnType<typeof setTimeout> | null = null

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

// ===== Action Menu =====
const activeMenuId = ref<number | null>(null)
const menuPosition = ref<{ top: number; left: number } | null>(null)

const openActionMenu = (user: User, e: MouseEvent) => {
  if (activeMenuId.value === user.id) {
    closeActionMenu()
  } else {
    const target = e.currentTarget as HTMLElement
    if (!target) {
      closeActionMenu()
      return
    }

    const rect = target.getBoundingClientRect()
    const menuWidth = 200
    const menuHeight = 140
    const padding = 8
    const viewportWidth = window.innerWidth
    const viewportHeight = window.innerHeight

    let left: number, top: number

    if (viewportWidth < 768) {
      left = Math.max(padding, Math.min(
        rect.left + rect.width / 2 - menuWidth / 2,
        viewportWidth - menuWidth - padding
      ))
      top = rect.bottom + 4
      if (top + menuHeight > viewportHeight - padding) {
        top = rect.top - menuHeight - 4
        if (top < padding) {
          top = padding
        }
      }
    } else {
      left = Math.max(padding, Math.min(
        e.clientX - menuWidth,
        viewportWidth - menuWidth - padding
      ))
      top = e.clientY
      if (top + menuHeight > viewportHeight - padding) {
        top = viewportHeight - menuHeight - padding
      }
    }

    menuPosition.value = { top, left }
    activeMenuId.value = user.id
  }
}

const closeActionMenu = () => {
  activeMenuId.value = null
  menuPosition.value = null
}

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as HTMLElement
  if (!target.closest('.action-menu-trigger') && !target.closest('.action-menu-content')) {
    closeActionMenu()
  }
}

const handleScroll = () => {
  closeActionMenu()
}

// ===== Balance Modal =====
const showBalanceModal = ref(false)
const balanceUser = ref<User | null>(null)
const balanceOperation = ref<'add' | 'subtract'>('add')
const balanceSubmitting = ref(false)
const balanceForm = reactive({ amount: 0, notes: '' })

watch(() => showBalanceModal.value, (v) => {
  if (v) {
    balanceForm.amount = 0
    balanceForm.notes = ''
  }
})

const formatBalance = (value: number) => {
  if (value === 0) return '0.00'
  const formatted = value.toFixed(8).replace(/\.?0+$/, '')
  const parts = formatted.split('.')
  if (parts.length === 1) return formatted + '.00'
  if (parts[1].length === 1) return formatted + '0'
  return formatted
}

const fillAllBalance = () => {
  if (balanceUser.value) {
    balanceForm.amount = balanceUser.value.balance
  }
}

const calculateNewBalance = () => {
  if (!balanceUser.value) return 0
  const result = balanceOperation.value === 'add'
    ? balanceUser.value.balance + balanceForm.amount
    : balanceUser.value.balance - balanceForm.amount
  return Math.abs(result) < 1e-10 ? 0 : result
}

const handleDeposit = (user: User) => {
  balanceUser.value = user
  balanceOperation.value = 'add'
  showBalanceModal.value = true
}

const handleWithdraw = (user: User) => {
  balanceUser.value = user
  balanceOperation.value = 'subtract'
  showBalanceModal.value = true
}

const closeBalanceModal = () => {
  showBalanceModal.value = false
  balanceUser.value = null
}

const handleBalanceSubmit = async () => {
  if (!balanceUser.value) return
  if (!balanceForm.amount || balanceForm.amount <= 0) {
    appStore.showError(t('reseller.users.amountRequired'))
    return
  }
  if (balanceOperation.value === 'subtract' && balanceForm.amount > balanceUser.value.balance) {
    appStore.showError(t('reseller.users.insufficientBalance'))
    return
  }
  balanceSubmitting.value = true
  try {
    await resellerAPI.users.updateBalance(
      balanceUser.value.id,
      balanceForm.amount,
      balanceOperation.value,
      balanceForm.notes
    )
    appStore.showSuccess(t('reseller.users.balanceUpdated'))
    closeBalanceModal()
    loadUsers()
  } catch (e: any) {
    console.error('Failed to update balance:', e)
    appStore.showError(e.response?.data?.detail || t('common.error'))
  } finally {
    balanceSubmitting.value = false
  }
}

// ===== Balance History Modal =====
const showHistoryModal = ref(false)
const historyUser = ref<User | null>(null)
const historyItems = ref<ResellerBalanceHistoryItem[]>([])
const historyLoading = ref(false)
const historyCurrentPage = ref(1)
const historyTotal = ref(0)
const historyTotalRecharged = ref(0)
const historyPageSize = 15

const historyTotalPages = computed(() => Math.ceil(historyTotal.value / historyPageSize) || 1)

const handleBalanceHistory = (user: User) => {
  historyUser.value = user
  showHistoryModal.value = true
}

watch(() => showHistoryModal.value, (v) => {
  if (v && historyUser.value) {
    loadHistory(1)
  }
})

const loadHistory = async (page: number) => {
  if (!historyUser.value) return
  historyLoading.value = true
  historyCurrentPage.value = page
  try {
    const res = await resellerAPI.users.getBalanceHistory(
      historyUser.value.id,
      page,
      historyPageSize
    )
    historyItems.value = res.items || []
    historyTotal.value = res.total || 0
    historyTotalRecharged.value = res.total_recharged || 0
  } catch (error) {
    console.error('Failed to load balance history:', error)
  } finally {
    historyLoading.value = false
  }
}

const closeHistoryModal = () => {
  showHistoryModal.value = false
  historyUser.value = null
}

const handleDepositFromHistory = () => {
  if (historyUser.value) {
    handleDeposit(historyUser.value)
  }
}

const handleWithdrawFromHistory = () => {
  if (historyUser.value) {
    handleWithdraw(historyUser.value)
  }
}

// History helpers
const isAdminType = (type: string) => type === 'admin_balance' || type === 'admin_concurrency'
const isBalanceType = (type: string) => type === 'balance' || type === 'admin_balance'

const getHistoryIconName = (item: ResellerBalanceHistoryItem) => {
  if (isBalanceType(item.type)) return 'dollar'
  return 'bolt'
}

const getHistoryIconBg = (item: ResellerBalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'bg-emerald-100 dark:bg-emerald-900/30'
      : 'bg-red-100 dark:bg-red-900/30'
  }
  return item.value >= 0
    ? 'bg-blue-100 dark:bg-blue-900/30'
    : 'bg-orange-100 dark:bg-orange-900/30'
}

const getHistoryIconColor = (item: ResellerBalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'text-emerald-600 dark:text-emerald-400'
      : 'text-red-600 dark:text-red-400'
  }
  return item.value >= 0
    ? 'text-blue-600 dark:text-blue-400'
    : 'text-orange-600 dark:text-orange-400'
}

const getHistoryValueColor = (item: ResellerBalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    return item.value >= 0
      ? 'text-emerald-600 dark:text-emerald-400'
      : 'text-red-600 dark:text-red-400'
  }
  return item.value >= 0
    ? 'text-blue-600 dark:text-blue-400'
    : 'text-orange-600 dark:text-orange-400'
}

const getHistoryItemTitle = (item: ResellerBalanceHistoryItem) => {
  switch (item.type) {
    case 'balance':
      return t('reseller.users.balanceAddedRedeem')
    case 'admin_balance':
      return item.value >= 0 ? t('reseller.users.balanceAddedAdmin') : t('reseller.users.balanceDeductedAdmin')
    default:
      return item.type
  }
}

const formatHistoryValue = (item: ResellerBalanceHistoryItem) => {
  if (isBalanceType(item.type)) {
    const sign = item.value >= 0 ? '+' : ''
    return `${sign}$${item.value.toFixed(2)}`
  }
  const sign = item.value >= 0 ? '+' : ''
  return `${sign}${item.value}`
}

// ===== Users List =====
async function loadUsers() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (searchQuery.value.trim()) {
      params.search = searchQuery.value.trim()
    }
    const result = await resellerAPI.users.list(params)
    users.value = result.items || []
    pagination.total = result.total
    pagination.pages = result.pages
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

function handleSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    pagination.page = 1
    loadUsers()
  }, 300)
}

function handlePageChange(page: number) {
  pagination.page = page
  loadUsers()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadUsers()
}

onMounted(() => {
  loadUsers()
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('scroll', handleScroll, true)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', handleScroll, true)
  if (searchTimer) clearTimeout(searchTimer)
})
</script>
