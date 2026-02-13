<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.redeem.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.redeem.description') }}</p>
        </div>
      </div>

      <!-- Generate Form -->
      <div class="card p-6">
        <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('reseller.redeem.generate') }}
        </h2>
        <div class="flex flex-wrap items-end gap-4">
          <div>
            <label class="label">{{ t('reseller.redeem.count') }}</label>
            <input
              v-model.number="generateForm.count"
              type="number"
              min="1"
              max="100"
              class="input w-32"
              placeholder="1"
            />
          </div>
          <div>
            <label class="label">{{ t('reseller.redeem.value') }}</label>
            <input
              v-model.number="generateForm.value"
              type="number"
              min="0.01"
              step="0.01"
              class="input w-40"
              placeholder="0.00"
            />
          </div>
          <div class="flex items-center gap-3">
            <span class="text-sm text-gray-500 dark:text-gray-400">
              {{ t('reseller.redeem.totalCost') }}: <strong class="text-gray-900 dark:text-white">${{ totalCost.toFixed(2) }}</strong>
            </span>
            <button
              @click="handleGenerate"
              class="btn btn-primary"
              :disabled="generating || !generateForm.count || !generateForm.value"
            >
              {{ generating ? '...' : t('reseller.redeem.generate') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <button @click="handleExport" class="btn btn-secondary" :disabled="codes.length === 0">
            {{ t('reseller.redeem.export') }}
          </button>
        </div>
        <button
          @click="loadCodes"
          :disabled="loading"
          class="btn btn-secondary"
        >
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <!-- Codes Table -->
      <div v-else-if="codes.length > 0" class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.redeem.copy') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">Code</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.redeem.value') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.announcements.status') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('common.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="code in codes"
                :key="code.id"
                class="border-b border-gray-50 dark:border-dark-800"
              >
                <td class="px-4 py-3">
                  <button
                    @click="copyCode(code.code)"
                    class="rounded p-1 text-gray-400 transition-colors hover:text-primary-600 dark:hover:text-primary-400"
                    :title="t('reseller.redeem.copy')"
                  >
                    <Icon name="copy" size="sm" />
                  </button>
                </td>
                <td class="px-4 py-3">
                  <code class="text-sm font-mono text-gray-900 dark:text-white">{{ code.code }}</code>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">${{ code.value.toFixed(2) }}</span>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="getStatusClass(code.status)"
                  >
                    {{ getStatusLabel(code.status) }}
                  </span>
                  <span v-if="code.user" class="ml-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ code.user.email }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(code.created_at) }}</span>
                </td>
                <td class="px-4 py-3">
                  <button
                    v-if="code.status !== 'used'"
                    @click="handleDelete(code)"
                    class="rounded-lg p-1.5 text-red-600 transition-colors hover:bg-red-50 dark:hover:bg-red-900/20"
                    :title="t('common.delete')"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <Icon name="gift" size="lg" class="mb-3 text-gray-300 dark:text-gray-600" />
        <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.redeem.empty') }}</p>
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

    <!-- Delete Confirmation -->
    <Teleport to="body">
      <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showDeleteModal = false">
        <div class="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-900">
          <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('common.confirmDelete') }}</h3>
          <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.redeem.deleteConfirm') }}</p>
          <div class="flex justify-end gap-3">
            <button @click="showDeleteModal = false" class="btn btn-secondary">{{ t('common.cancel') }}</button>
            <button @click="confirmDelete" class="btn btn-danger" :disabled="deleting">
              {{ deleting ? '...' : t('common.delete') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { ResellerRedeemCode } from '@/api/reseller/redeem'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const generating = ref(false)
const deleting = ref(false)
const showDeleteModal = ref(false)
const deleteTarget = ref<ResellerRedeemCode | null>(null)
const codes = ref<ResellerRedeemCode[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const generateForm = reactive({
  count: 1,
  value: 1
})

const totalCost = computed(() => (generateForm.count || 0) * (generateForm.value || 0))

async function loadCodes() {
  loading.value = true
  try {
    const result = await resellerAPI.redeem.list({
      page: pagination.page,
      page_size: pagination.page_size
    })
    codes.value = result.items || []
    pagination.total = result.total
    pagination.pages = result.pages
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleGenerate() {
  generating.value = true
  try {
    await resellerAPI.redeem.generate({
      count: generateForm.count,
      value: generateForm.value
    })
    appStore.showSuccess(t('reseller.redeem.generateSuccess'))
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    generating.value = false
  }
}

function handleDelete(code: ResellerRedeemCode) {
  deleteTarget.value = code
  showDeleteModal.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await resellerAPI.redeem.delete(deleteTarget.value.id)
    appStore.showSuccess(t('reseller.redeem.deleteSuccess'))
    showDeleteModal.value = false
    loadCodes()
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    deleting.value = false
  }
}

async function copyCode(code: string) {
  try {
    await navigator.clipboard.writeText(code)
    appStore.showSuccess(t('reseller.redeem.copied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

function handleExport() {
  const unusedCodes = codes.value.filter(c => c.status !== 'used')
  if (unusedCodes.length === 0) return
  const text = unusedCodes.map(c => c.code).join('\n')
  const blob = new Blob([text], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `redeem-codes-${new Date().toISOString().slice(0, 10)}.txt`
  a.click()
  URL.revokeObjectURL(url)
}

function getStatusClass(status: string) {
  switch (status) {
    case 'active':
    case 'unused':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'used':
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
    case 'expired':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
  }
}

function getStatusLabel(status: string) {
  switch (status) {
    case 'active':
    case 'unused':
      return t('common.active')
    case 'used':
      return t('common.used') || 'Used'
    case 'expired':
      return t('common.expired') || 'Expired'
    default:
      return status
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadCodes()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadCodes()
}

onMounted(() => {
  loadCodes()
})
</script>
