<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('admin.merchants.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('admin.merchants.description') }}</p>
        </div>
        <div class="flex items-center gap-3">
          <button
            @click="handleBackfillSnapshots"
            :disabled="backfilling"
            class="btn btn-secondary"
          >
            {{ backfilling ? t('common.loading') : t('admin.merchants.backfillSnapshots') }}
          </button>
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="t('admin.merchants.search')"
            class="input w-64"
            @input="handleSearchInput"
          />
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <!-- Table -->
      <div v-else-if="items.length > 0" class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">ID</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchants.username') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchants.domain') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchants.email') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchants.balance') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchants.userCount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchants.keyCount') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('admin.merchants.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in items"
                :key="item.id"
                class="border-b border-gray-50 dark:border-dark-800"
              >
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.id }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.username || '-' }}</span>
                </td>
                <td class="px-4 py-3">
                  <a v-if="item.domain" :href="'https://' + item.domain" target="_blank" rel="noopener noreferrer"
                    class="text-sm font-medium text-primary-600 hover:underline dark:text-primary-400">
                    {{ item.domain }}
                  </a>
                  <span v-else class="text-sm text-gray-400 dark:text-gray-500">-</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-600 dark:text-gray-300">{{ item.email }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">${{ item.balance?.toFixed(2) }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.user_count }}</span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-900 dark:text-white">{{ item.key_count }}</span>
                </td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <button
                      @click="handleDeposit(item)"
                      class="rounded px-3 py-1 text-sm font-medium text-emerald-600 transition-colors hover:bg-emerald-50 dark:text-emerald-400 dark:hover:bg-emerald-900/20"
                    >
                      {{ t('admin.users.deposit') }}
                    </button>
                    <button
                      @click="handleWithdraw(item)"
                      class="rounded px-3 py-1 text-sm font-medium text-red-600 transition-colors hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                    >
                      {{ t('admin.users.withdraw') }}
                    </button>
                    <button
                      @click="openSettingsDialog(item)"
                      class="rounded px-3 py-1 text-sm font-medium text-primary-600 transition-colors hover:bg-primary-50 dark:text-primary-400 dark:hover:bg-primary-900/20"
                    >
                      {{ t('admin.merchants.settings') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <p class="text-gray-500 dark:text-gray-400">{{ t('admin.merchants.empty') }}</p>
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

    <!-- Balance Modal -->
    <UserBalanceModal
      :show="showBalanceModal"
      :user="balanceMerchant"
      :operation="balanceOperation"
      :update-fn="merchantUpdateBalance"
      @close="closeBalanceModal"
      @success="handleBalanceSuccess"
    />

    <!-- Settings Dialog -->
    <BaseDialog
      :show="showSettingsDialog"
      :title="t('admin.merchants.settingsTitle')"
      width="narrow"
      @close="closeSettingsDialog"
    >
      <form v-if="editingMerchant" id="merchant-settings-form" @submit.prevent="handleSettingsSubmit" class="space-y-4">
        <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
          <p class="font-medium text-gray-900 dark:text-white">{{ editingMerchant.email }}</p>
          <p v-if="editingMerchant.username" class="text-sm text-gray-500 dark:text-gray-400">{{ editingMerchant.username }}</p>
        </div>
        <div class="flex items-center justify-between rounded-xl border border-gray-200 p-4 dark:border-dark-600">
          <div>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ t('admin.merchants.agentEnabled') }}</p>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.merchants.agentEnabledHint') }}</p>
          </div>
          <button
            type="button"
            @click="settingsForm.merchant_mode = settingsForm.merchant_mode === 'enabled' ? '' : 'enabled'"
            :class="settingsForm.merchant_mode === 'enabled'
              ? 'bg-primary-600'
              : 'bg-gray-200 dark:bg-dark-600'"
            class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full transition-colors duration-200 focus:outline-none"
          >
            <span
              :class="settingsForm.merchant_mode === 'enabled' ? 'translate-x-5' : 'translate-x-1'"
              class="inline-block h-4 w-4 translate-y-1 transform rounded-full bg-white shadow transition duration-200"
            />
          </button>
        </div>
        <div>
          <label class="input-label">{{ t('admin.merchants.platformCost') }}</label>
          <input
            v-model="settingsForm.platform_cost"
            type="number"
            min="0"
            step="0.01"
            class="input"
            placeholder="0.69"
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.merchants.platformCostHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.merchants.priceMultiplier') }}</label>
          <input
            :value="settingsForm.price_multiplier || t('admin.merchants.notSet')"
            type="text"
            class="input cursor-not-allowed bg-gray-50 dark:bg-gray-800"
            readonly
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.merchants.priceMultiplierReadonlyHint') }}</p>
        </div>
        <div>
          <label class="input-label">{{ t('admin.merchants.minWithdrawal') }}</label>
          <input
            v-model="settingsForm.min_withdrawal"
            type="text"
            class="input"
            :placeholder="t('admin.merchants.minWithdrawalPlaceholder')"
          />
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button @click="closeSettingsDialog" class="btn btn-secondary">{{ t('common.cancel') }}</button>
          <button
            type="submit"
            form="merchant-settings-form"
            :disabled="settingsSubmitting"
            class="btn"
          >
            {{ settingsSubmitting ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { MerchantUser } from '@/api/admin/merchants'
import { useAppStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import UserBalanceModal from '@/components/admin/user/UserBalanceModal.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const items = ref<MerchantUser[]>([])
const searchQuery = ref('')
const backfilling = ref(false)
let searchTimer: ReturnType<typeof setTimeout> | null = null

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

// Balance Modal
const showBalanceModal = ref(false)
const balanceMerchant = ref<MerchantUser | null>(null)
const balanceOperation = ref<'add' | 'subtract'>('add')

function handleDeposit(merchant: MerchantUser) {
  balanceMerchant.value = merchant
  balanceOperation.value = 'add'
  showBalanceModal.value = true
}

function handleWithdraw(merchant: MerchantUser) {
  balanceMerchant.value = merchant
  balanceOperation.value = 'subtract'
  showBalanceModal.value = true
}

function closeBalanceModal() {
  showBalanceModal.value = false
  balanceMerchant.value = null
}

function handleBalanceSuccess() {
  loadMerchants()
}

function merchantUpdateBalance(id: number, amount: number, operation: 'add' | 'subtract', notes?: string) {
  return adminAPI.merchants.updateMerchantBalance(id, amount, operation, notes)
}

// Settings Dialog
const showSettingsDialog = ref(false)
const settingsSubmitting = ref(false)
const editingMerchant = ref<MerchantUser | null>(null)
const settingsForm = reactive({
  merchant_mode: '',
  platform_cost: '',
  price_multiplier: '',
  min_withdrawal: ''
})

async function openSettingsDialog(merchant: MerchantUser) {
  editingMerchant.value = merchant
  settingsForm.merchant_mode = ''
  settingsForm.platform_cost = ''
  settingsForm.price_multiplier = ''
  settingsForm.min_withdrawal = ''
  showSettingsDialog.value = true
  try {
    const kv = await adminAPI.merchants.getMerchantSettings(merchant.id)
    settingsForm.merchant_mode = kv.merchant_mode || ''
    settingsForm.platform_cost = kv.platform_cost || ''
    settingsForm.price_multiplier = kv.price_multiplier || ''
    settingsForm.min_withdrawal = kv.min_withdrawal || ''
  } catch {
    // ignore, form stays empty
  }
}

function closeSettingsDialog() {
  showSettingsDialog.value = false
  editingMerchant.value = null
}

async function handleSettingsSubmit() {
  if (!editingMerchant.value) return
  settingsSubmitting.value = true
  try {
    const payload: Record<string, string> = {}
    if (settingsForm.merchant_mode) payload.merchant_mode = settingsForm.merchant_mode
    if (settingsForm.platform_cost) payload.platform_cost = settingsForm.platform_cost
    if (settingsForm.min_withdrawal) payload.min_withdrawal = settingsForm.min_withdrawal

    await adminAPI.merchants.updateMerchantSettings(editingMerchant.value.id, payload)
    appStore.showSuccess(t('admin.merchants.settingsUpdated'))
    closeSettingsDialog()
    loadMerchants()
  } catch (error: any) {
    appStore.showError(error.message || t('common.error'))
  } finally {
    settingsSubmitting.value = false
  }
}

async function handleBackfillSnapshots() {
  backfilling.value = true
  try {
    const result = await adminAPI.merchants.backfillSnapshots()
    appStore.showSuccess(t('admin.merchants.backfillSuccess', { count: result.updated }))
  } catch (error: any) {
    appStore.showError(error.message || t('common.error'))
  } finally {
    backfilling.value = false
  }
}

async function loadMerchants() {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (searchQuery.value.trim()) params.search = searchQuery.value.trim()

    const result = await adminAPI.merchants.getMerchants(params)
    items.value = result.items || []
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
    loadMerchants()
  }, 300)
}

function handlePageChange(page: number) {
  pagination.page = page
  loadMerchants()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadMerchants()
}

onMounted(() => {
  loadMerchants()
})

onUnmounted(() => {
  if (searchTimer) clearTimeout(searchTimer)
})
</script>
