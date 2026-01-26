<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-start justify-between gap-4">
          <!-- Left: Search + Filters -->
          <div class="flex flex-1 flex-wrap items-center gap-3">
            <div class="relative w-full sm:w-64">
              <Icon
                name="search"
                size="md"
                class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500"
              />
              <input
                v-model="searchQuery"
                type="text"
                :placeholder="t('admin.channels.searchChannels')"
                class="input pl-10"
                @input="handleSearch"
              />
            </div>
            <div class="w-full sm:w-40">
              <Select
                v-model="filters.platform"
                :options="platformOptions"
                :placeholder="t('admin.channels.allPlatforms')"
                @change="loadChannels"
              />
            </div>
            <div class="w-full sm:w-36">
              <Select
                v-model="filters.status"
                :options="statusOptions"
                :placeholder="t('admin.channels.allStatus')"
                @change="loadChannels"
              />
            </div>
          </div>

          <!-- Right: Actions -->
          <div class="ml-auto flex flex-wrap items-center justify-end gap-3">
            <button
              @click="loadChannels"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
            <button @click="openCurlModal" class="btn btn-secondary">
              <Icon name="code" size="md" class="mr-2" />
              {{ t('admin.channels.importCurl') }}
            </button>
            <button @click="openCreateModal" class="btn btn-primary">
              <Icon name="plus" size="md" class="mr-2" />
              {{ t('admin.channels.createChannel') }}
            </button>
          </div>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="channels" :loading="loading">
          <template #cell-name="{ row }">
            <div class="flex items-center gap-3">
              <div
                v-if="row.icon_url"
                class="h-8 w-8 flex-shrink-0 overflow-hidden rounded-lg"
              >
                <img :src="row.icon_url" :alt="row.name" class="h-full w-full object-cover" />
              </div>
              <div
                v-else
                class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg text-sm font-bold text-white"
                :class="{
                  'bg-orange-500': row.platform === 'anthropic',
                  'bg-emerald-500': row.platform === 'openai',
                  'bg-blue-500': row.platform === 'gemini',
                  'bg-gray-500': !row.platform || row.platform === 'other'
                }"
              >
                {{ row.name.charAt(0).toUpperCase() }}
              </div>
              <div>
                <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
                <div v-if="row.description" class="max-w-xs truncate text-xs text-gray-500 dark:text-gray-400">
                  {{ row.description }}
                </div>
              </div>
            </div>
          </template>

          <template #cell-platform="{ value }">
            <span
              v-if="value"
              class="badge"
              :class="{
                'badge-warning': value === 'anthropic',
                'badge-success': value === 'openai',
                'badge-primary': value === 'gemini',
                'badge-gray': value === 'other'
              }"
            >
              {{ value }}
            </span>
            <span v-else class="text-sm text-gray-400">-</span>
          </template>

          <template #cell-balance="{ row }">
            <div class="flex items-center gap-2">
              <span
                class="text-sm font-semibold"
                :class="(row.cached_balance || 0) < 20
                  ? 'text-red-600 dark:text-red-400'
                  : 'text-emerald-600 dark:text-emerald-400'"
              >
                {{ row.balance_unit }}{{ (row.cached_balance || 0).toFixed(2) }}
              </span>
              <button
                @click="handleRefreshBalance(row)"
                :disabled="refreshingChannelIds.has(row.id) || !row.balance_url"
                class="rounded p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-dark-700 dark:hover:text-gray-300"
                :title="t('common.refresh')"
              >
                <Icon
                  name="refresh"
                  size="sm"
                  :class="{ 'animate-spin': refreshingChannelIds.has(row.id) }"
                />
              </button>
            </div>
          </template>

          <template #cell-status="{ value }">
            <span
              :class="['badge', value === 'active' ? 'badge-success' : value === 'error' ? 'badge-danger' : 'badge-gray']"
            >
              {{ t(`admin.channels.status.${value}`) }}
            </span>
          </template>

          <template #cell-last_check_at="{ value }">
            <span class="text-sm text-gray-500 dark:text-dark-400">
              {{ value ? formatTime(value) : '-' }}
            </span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                @click="openEditModal(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-gray-100 hover:text-primary-600 dark:hover:bg-dark-700 dark:hover:text-primary-400"
              >
                <Icon name="edit" size="sm" />
                <span class="text-xs">{{ t('common.edit') }}</span>
              </button>
              <button
                @click="confirmDelete(row)"
                class="flex flex-col items-center gap-0.5 rounded-lg p-1.5 text-gray-500 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-400"
              >
                <Icon name="trash" size="sm" />
                <span class="text-xs">{{ t('common.delete') }}</span>
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState
              :title="t('admin.channels.noChannels')"
              :description="t('admin.channels.description')"
              :action-text="t('admin.channels.createChannel')"
              @action="openCreateModal"
            />
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination && pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Create/Edit Modal -->
    <BaseDialog
      :show="showModal"
      :title="editingChannel ? t('admin.channels.editChannel') : t('admin.channels.createChannel')"
      width="wide"
      @close="closeModal"
    >
      <form @submit.prevent="handleSubmit" class="space-y-5">
        <!-- Basic Info -->
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.channels.name') }} *</label>
            <input
              v-model="form.name"
              type="text"
              class="input"
              :placeholder="t('admin.channels.namePlaceholder')"
              required
            />
          </div>
          <div>
            <label class="input-label">{{ t('admin.channels.platform') }}</label>
            <Select
              v-model="form.platform"
              :options="platformSelectOptions"
              :placeholder="t('admin.channels.selectPlatform')"
            />
          </div>
        </div>

        <div>
          <label class="input-label">{{ t('admin.channels.descriptionLabel') }}</label>
          <textarea
            v-model="form.description"
            class="input"
            rows="2"
            :placeholder="t('admin.channels.descriptionPlaceholder')"
          ></textarea>
        </div>

        <div>
          <label class="input-label">{{ t('common.status') }}</label>
          <Select
            v-model="form.status"
            :options="statusSelectOptions"
          />
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label">{{ t('admin.channels.iconUrl') }}</label>
            <input
              v-model="form.icon_url"
              type="url"
              class="input"
              :placeholder="t('admin.channels.iconUrlPlaceholder')"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.channels.iconUrlHint') }}</p>
          </div>
          <div>
            <label class="input-label">{{ t('admin.channels.websiteUrl') }}</label>
            <input
              v-model="form.website_url"
              type="url"
              class="input"
              :placeholder="t('admin.channels.websiteUrlPlaceholder')"
            />
            <p class="mt-1 text-xs text-gray-500">{{ t('admin.channels.websiteUrlHint') }}</p>
          </div>
        </div>

        <!-- Balance Config Section -->
        <div class="border-t border-gray-200 pt-5 dark:border-dark-600">
          <h3 class="mb-4 text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('admin.channels.balanceConfig') }}
          </h3>

          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div class="sm:col-span-2">
              <label class="input-label">{{ t('admin.channels.balanceUrl') }}</label>
              <input
                v-model="form.balance_url"
                type="url"
                class="input"
                :placeholder="t('admin.channels.balanceUrlPlaceholder')"
              />
            </div>
            <div>
              <label class="input-label">{{ t('admin.channels.balanceMethod') }}</label>
              <Select
                v-model="form.balance_method"
                :options="methodOptions"
              />
            </div>
            <div>
              <label class="input-label">{{ t('admin.channels.balanceUnit') }}</label>
              <input
                v-model="form.balance_unit"
                type="text"
                class="input"
                placeholder="$"
              />
            </div>
            <div class="sm:col-span-2">
              <label class="input-label">{{ t('admin.channels.balancePath') }}</label>
              <input
                v-model="form.balance_path"
                type="text"
                class="input"
                :placeholder="t('admin.channels.balancePathPlaceholder')"
              />
            </div>
            <div class="sm:col-span-2">
              <label class="input-label">{{ t('admin.channels.balanceHeaders') }}</label>
              <textarea
                v-model="headersJson"
                class="input font-mono text-sm"
                rows="3"
                :placeholder="t('admin.channels.balanceHeadersPlaceholder')"
              ></textarea>
              <p class="mt-1 text-xs text-gray-500">{{ t('admin.channels.balanceHeadersHint') }}</p>
            </div>
            <div v-if="form.balance_method === 'POST'" class="sm:col-span-2">
              <label class="input-label">{{ t('admin.channels.balanceBody') }}</label>
              <textarea
                v-model="form.balance_body"
                class="input font-mono text-sm"
                rows="3"
                :placeholder="t('admin.channels.balanceBodyPlaceholder')"
              ></textarea>
            </div>
          </div>
        </div>
      </form>

      <template #footer>
        <button @click="closeModal" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button @click="handleSubmit" :disabled="submitting" class="btn btn-primary">
          <Icon v-if="submitting" name="refresh" size="sm" class="mr-2 animate-spin" />
          {{ editingChannel ? t('common.update') : t('common.create') }}
        </button>
      </template>
    </BaseDialog>

    <!-- Delete Confirmation Dialog -->
    <ConfirmDialog
      :show="showDeleteConfirm"
      :title="t('admin.channels.deleteChannel')"
      :message="t('admin.channels.deleteConfirmMessage', { name: deletingChannel?.name || '' })"
      :confirm-text="t('common.delete')"
      :danger="true"
      @confirm="handleDelete"
      @cancel="showDeleteConfirm = false"
    />

    <!-- Curl Import Modal -->
    <BaseDialog
      :show="showCurlModal"
      :title="t('admin.channels.importCurl')"
      width="wide"
      @close="closeCurlModal"
    >
      <div class="space-y-4">
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ t('admin.channels.curlImportHint') }}
        </p>
        <div>
          <label class="input-label">{{ t('admin.channels.curlCommand') }}</label>
          <textarea
            v-model="curlInput"
            class="input font-mono text-sm"
            rows="10"
            :placeholder="t('admin.channels.curlPlaceholder')"
          ></textarea>
        </div>
        <p v-if="curlError" class="text-sm text-red-600 dark:text-red-400">
          {{ curlError }}
        </p>
      </div>

      <template #footer>
        <button @click="closeCurlModal" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button @click="handleCurlImport" class="btn btn-primary">
          <Icon name="code" size="sm" class="mr-2" />
          {{ t('admin.channels.parseCurl') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import { channelsAPI } from '@/api/admin'
import type { Channel, CreateChannelRequest, UpdateChannelRequest, PaginatedResponse } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Select from '@/components/common/Select.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import { parseCurl, validateParsedCurl } from '@/utils/curlParser'

const { t } = useI18n()
const appStore = useAppStore()

// Format time
function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return t('common.time.justNow')
  if (minutes < 60) return t('common.time.minutesAgo', { n: minutes })
  if (hours < 24) return t('common.time.hoursAgo', { n: hours })
  if (days < 7) return t('common.time.daysAgo', { n: days })

  return date.toLocaleDateString()
}

// Table columns
const columns = computed(() => [
  { key: 'name', label: t('admin.channels.name'), sortable: true },
  { key: 'platform', label: t('admin.channels.platform'), sortable: true },
  { key: 'balance', label: t('admin.channels.balance'), sortable: false },
  { key: 'status', label: t('common.status'), sortable: true },
  { key: 'last_check_at', label: t('admin.channels.lastChecked'), sortable: true },
  { key: 'actions', label: t('common.actions'), align: 'right' as const }
])

// State
const loading = ref(false)
const channels = ref<Channel[]>([])
const pagination = ref<PaginatedResponse<Channel> | null>(null)
const currentPage = ref(1)
const pageSize = ref(20)

// Filters
const searchQuery = ref('')
const filters = reactive({
  platform: null as string | null,
  status: null as string | null
})

// Modal state
const showModal = ref(false)
const editingChannel = ref<Channel | null>(null)
const submitting = ref(false)

// Delete state
const showDeleteConfirm = ref(false)
const deletingChannel = ref<Channel | null>(null)
const deleting = ref(false)

// Refresh state
const refreshingChannelIds = ref<Set<number>>(new Set())

// Curl import state
const showCurlModal = ref(false)
const curlInput = ref('')
const curlError = ref('')

// Form
const form = reactive<CreateChannelRequest>({
  name: '',
  description: null,
  platform: null,
  status: 'active',
  icon_url: null,
  website_url: null,
  balance_url: null,
  balance_method: 'GET',
  balance_headers: null,
  balance_body: null,
  balance_path: null,
  balance_unit: '$'
})

const headersJson = ref('')

// Options
const platformOptions = computed(() => [
  { value: null, label: t('admin.channels.allPlatforms') },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'other', label: t('admin.channels.other') }
])

const platformSelectOptions = computed(() => [
  { value: null, label: t('admin.channels.noPlatform') },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'other', label: t('admin.channels.other') }
])

const statusOptions = computed(() => [
  { value: null, label: t('admin.channels.allStatus') },
  { value: 'active', label: t('admin.channels.status.active') },
  { value: 'inactive', label: t('admin.channels.status.inactive') },
  { value: 'error', label: t('admin.channels.status.error') }
])

const statusSelectOptions = computed(() => [
  { value: 'active', label: t('admin.channels.status.active') },
  { value: 'inactive', label: t('admin.channels.status.inactive') }
])

const methodOptions = computed(() => [
  { value: 'GET', label: 'GET' },
  { value: 'POST', label: 'POST' }
])

// Search debounce
let searchTimeout: ReturnType<typeof setTimeout> | null = null
function handleSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    loadChannels()
  }, 300)
}

// Load channels
async function loadChannels() {
  loading.value = true
  try {
    const response = await channelsAPI.list(currentPage.value, pageSize.value, {
      platform: filters.platform || undefined,
      status: (filters.status as 'active' | 'inactive' | 'error') || undefined,
      search: searchQuery.value || undefined
    })
    channels.value = response.items
    pagination.value = response
  } catch (error) {
    console.error('Failed to load channels:', error)
    appStore.showError(t('admin.channels.loadError'))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  currentPage.value = page
  loadChannels()
}

function handlePageSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  loadChannels()
}

// Create/Edit Modal
function openCreateModal() {
  editingChannel.value = null
  resetForm()
  showModal.value = true
}

function openEditModal(channel: Channel) {
  editingChannel.value = channel
  form.name = channel.name
  form.description = channel.description || null
  form.platform = channel.platform || null
  form.status = channel.status
  form.icon_url = channel.icon_url || null
  form.website_url = channel.website_url || null
  form.balance_url = channel.balance_url || null
  form.balance_method = channel.balance_method || 'GET'
  form.balance_headers = channel.balance_headers || null
  form.balance_body = channel.balance_body || null
  form.balance_path = channel.balance_path || null
  form.balance_unit = channel.balance_unit || '$'
  headersJson.value = channel.balance_headers
    ? JSON.stringify(channel.balance_headers, null, 2)
    : ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingChannel.value = null
  resetForm()
}

function resetForm() {
  form.name = ''
  form.description = null
  form.platform = null
  form.status = 'active'
  form.icon_url = null
  form.website_url = null
  form.balance_url = null
  form.balance_method = 'GET'
  form.balance_headers = null
  form.balance_body = null
  form.balance_path = null
  form.balance_unit = '$'
  headersJson.value = ''
}

async function handleSubmit() {
  // Parse headers JSON
  if (headersJson.value.trim()) {
    try {
      form.balance_headers = JSON.parse(headersJson.value)
    } catch {
      appStore.showError(t('admin.channels.invalidHeadersJson'))
      return
    }
  } else {
    form.balance_headers = null
  }

  submitting.value = true
  try {
    if (editingChannel.value) {
      // Update
      const updates: UpdateChannelRequest = {
        name: form.name,
        description: form.description,
        platform: form.platform as any,
        status: form.status as any,
        icon_url: form.icon_url,
        website_url: form.website_url,
        balance_url: form.balance_url,
        balance_method: form.balance_method as any,
        balance_headers: form.balance_headers,
        balance_body: form.balance_body,
        balance_path: form.balance_path,
        balance_unit: form.balance_unit
      }
      await channelsAPI.update(editingChannel.value.id, updates)
      appStore.showSuccess(t('admin.channels.updateSuccess'))
    } else {
      // Create
      await channelsAPI.create(form)
      appStore.showSuccess(t('admin.channels.createSuccess'))
    }
    closeModal()
    loadChannels()
  } catch (error) {
    console.error('Failed to save channel:', error)
    appStore.showError(
      editingChannel.value
        ? t('admin.channels.updateError')
        : t('admin.channels.createError')
    )
  } finally {
    submitting.value = false
  }
}

// Delete
function confirmDelete(channel: Channel) {
  deletingChannel.value = channel
  showDeleteConfirm.value = true
}

async function handleDelete() {
  if (!deletingChannel.value) return
  deleting.value = true
  try {
    await channelsAPI.delete(deletingChannel.value.id)
    appStore.showSuccess(t('admin.channels.deleteSuccess'))
    showDeleteConfirm.value = false
    deletingChannel.value = null
    loadChannels()
  } catch (error) {
    console.error('Failed to delete channel:', error)
    appStore.showError(t('admin.channels.deleteError'))
  } finally {
    deleting.value = false
  }
}

// Refresh balance
async function handleRefreshBalance(channel: Channel) {
  refreshingChannelIds.value.add(channel.id)
  try {
    const updated = await channelsAPI.checkBalance(channel.id)
    // Update the channel in the list
    const index = channels.value.findIndex(c => c.id === channel.id)
    if (index !== -1) {
      channels.value[index] = updated
    }
    // Only show error notification, no success notification to reduce noise
    if (updated.last_error) {
      appStore.showError(t('admin.channels.balanceCheckFailed'))
    }
  } catch (error) {
    console.error('Failed to refresh balance:', error)
    appStore.showError(t('admin.channels.balanceCheckError'))
  } finally {
    refreshingChannelIds.value.delete(channel.id)
  }
}

// Curl import
function openCurlModal() {
  curlInput.value = ''
  curlError.value = ''
  showCurlModal.value = true
}

function closeCurlModal() {
  showCurlModal.value = false
  curlInput.value = ''
  curlError.value = ''
}

function handleCurlImport() {
  curlError.value = ''

  if (!curlInput.value.trim()) {
    curlError.value = t('admin.channels.curlEmpty')
    return
  }

  const parsed = parseCurl(curlInput.value)
  const error = validateParsedCurl(parsed)

  if (error) {
    curlError.value = error
    return
  }

  // Extract name from URL hostname
  try {
    const url = new URL(parsed.url)
    form.name = url.hostname.replace(/^(www\.|api\.)/, '').split('.')[0]
  } catch {
    form.name = ''
  }

  // Fill form with parsed data
  form.balance_url = parsed.url
  form.balance_method = parsed.method
  form.balance_headers = Object.keys(parsed.headers).length > 0 ? parsed.headers : null
  form.balance_body = parsed.body || null
  headersJson.value = form.balance_headers
    ? JSON.stringify(form.balance_headers, null, 2)
    : ''

  // Close curl modal and open create modal
  closeCurlModal()
  editingChannel.value = null
  showModal.value = true
}

onMounted(() => {
  loadChannels()
})
</script>
