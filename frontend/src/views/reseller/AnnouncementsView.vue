<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.announcements.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.announcements.description') }}</p>
        </div>
        <button
          @click="openCreateModal"
          class="inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
        >
          <Icon name="plus" size="sm" />
          {{ t('reseller.announcements.create') }}
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <!-- Announcements List -->
      <div v-else-if="announcements.length > 0" class="space-y-4">
        <div
          v-for="item in announcements"
          :key="item.id"
          class="card p-4"
        >
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white truncate">{{ item.title }}</h3>
                <span
                  class="inline-flex flex-shrink-0 items-center rounded-full px-2 py-0.5 text-xs font-medium"
                  :class="getStatusClass(item.status)"
                >
                  {{ getStatusLabel(item.status) }}
                </span>
              </div>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{{ item.content }}</p>
              <div v-if="item.starts_at || item.ends_at" class="mt-1 flex items-center gap-2 text-xs text-gray-400 dark:text-dark-500">
                <span v-if="item.starts_at">{{ formatDateTime(item.starts_at) }}</span>
                <span v-if="item.starts_at && item.ends_at">-</span>
                <span v-if="item.ends_at">{{ formatDateTime(item.ends_at) }}</span>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <button
                @click="openEditModal(item)"
                class="rounded-lg p-1.5 text-blue-600 transition-colors hover:bg-blue-50 dark:hover:bg-blue-900/20"
                :title="t('common.edit')"
              >
                <Icon name="edit" size="sm" />
              </button>
              <button
                @click="handleDelete(item)"
                class="rounded-lg p-1.5 text-red-600 transition-colors hover:bg-red-50 dark:hover:bg-red-900/20"
                :title="t('common.delete')"
              >
                <Icon name="trash" size="sm" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <Icon name="chatBubble" size="lg" class="mb-3 text-gray-300 dark:text-gray-600" />
        <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.announcements.empty') }}</p>
        <button
          @click="openCreateModal"
          class="mt-4 inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-primary-700"
        >
          <Icon name="plus" size="sm" />
          {{ t('reseller.announcements.create') }}
        </button>
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

    <!-- Create/Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="closeModal">
        <div class="mx-4 w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-900">
          <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
            {{ editingItem ? t('reseller.announcements.editTitle') : t('reseller.announcements.createTitle') }}
          </h3>
          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div>
              <label class="label">{{ t('reseller.announcements.titleLabel') }}</label>
              <input v-model="form.title" type="text" class="input" required />
            </div>
            <div>
              <label class="label">{{ t('reseller.announcements.content') }}</label>
              <textarea v-model="form.content" class="input" rows="4" required />
            </div>
            <div>
              <label class="label">{{ t('reseller.announcements.status') }}</label>
              <select v-model="form.status" class="input">
                <option value="draft">{{ t('reseller.announcements.statusDraft') }}</option>
                <option value="active">{{ t('reseller.announcements.statusActive') }}</option>
                <option value="archived">{{ t('reseller.announcements.statusArchived') }}</option>
              </select>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="label">{{ t('reseller.announcements.startsAt') }}</label>
                <input v-model="form.starts_at" type="datetime-local" class="input" />
              </div>
              <div>
                <label class="label">{{ t('reseller.announcements.endsAt') }}</label>
                <input v-model="form.ends_at" type="datetime-local" class="input" />
              </div>
            </div>
            <div class="flex justify-end gap-3 pt-2">
              <button type="button" @click="closeModal" class="btn btn-secondary">{{ t('common.cancel') }}</button>
              <button type="submit" class="btn btn-primary" :disabled="submitting">
                {{ submitting ? '...' : t('common.save') }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Delete Confirmation -->
    <Teleport to="body">
      <div v-if="showDeleteModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showDeleteModal = false">
        <div class="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-900">
          <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('common.confirmDelete') }}</h3>
          <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.announcements.deleteConfirm') }}</p>
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
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { ResellerAnnouncement } from '@/api/reseller/announcements'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const submitting = ref(false)
const deleting = ref(false)
const showModal = ref(false)
const showDeleteModal = ref(false)
const editingItem = ref<ResellerAnnouncement | null>(null)
const deleteTarget = ref<ResellerAnnouncement | null>(null)
const announcements = ref<ResellerAnnouncement[]>([])

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  pages: 0
})

const form = reactive({
  title: '',
  content: '',
  status: 'draft',
  starts_at: '',
  ends_at: ''
})

function resetForm() {
  form.title = ''
  form.content = ''
  form.status = 'draft'
  form.starts_at = ''
  form.ends_at = ''
}

function openCreateModal() {
  editingItem.value = null
  resetForm()
  showModal.value = true
}

function openEditModal(item: ResellerAnnouncement) {
  editingItem.value = item
  form.title = item.title
  form.content = item.content
  form.status = item.status
  form.starts_at = item.starts_at ? toLocalDatetime(item.starts_at) : ''
  form.ends_at = item.ends_at ? toLocalDatetime(item.ends_at) : ''
  showModal.value = true
}

function closeModal() {
  showModal.value = false
  editingItem.value = null
}

function toLocalDatetime(isoStr: string): string {
  try {
    const d = new Date(isoStr)
    const offset = d.getTimezoneOffset()
    const local = new Date(d.getTime() - offset * 60000)
    return local.toISOString().slice(0, 16)
  } catch {
    return ''
  }
}

async function loadAnnouncements() {
  loading.value = true
  try {
    const result = await resellerAPI.announcements.list({
      page: pagination.page,
      page_size: pagination.page_size
    })
    announcements.value = result.items || []
    pagination.total = result.total
    pagination.pages = result.pages
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  submitting.value = true
  try {
    if (editingItem.value) {
      await resellerAPI.announcements.update(editingItem.value.id, {
        title: form.title,
        content: form.content,
        status: form.status,
        starts_at: form.starts_at || null,
        ends_at: form.ends_at || null
      })
      appStore.showSuccess(t('reseller.announcements.updateSuccess'))
    } else {
      await resellerAPI.announcements.create({
        title: form.title,
        content: form.content,
        status: form.status,
        starts_at: form.starts_at || undefined,
        ends_at: form.ends_at || undefined
      })
      appStore.showSuccess(t('reseller.announcements.createSuccess'))
    }
    closeModal()
    loadAnnouncements()
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    submitting.value = false
  }
}

function handleDelete(item: ResellerAnnouncement) {
  deleteTarget.value = item
  showDeleteModal.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  deleting.value = true
  try {
    await resellerAPI.announcements.delete(deleteTarget.value.id)
    appStore.showSuccess(t('reseller.announcements.deleteSuccess'))
    showDeleteModal.value = false
    loadAnnouncements()
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    deleting.value = false
  }
}

function getStatusClass(status: string) {
  switch (status) {
    case 'active':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'draft':
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
    case 'archived':
      return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
  }
}

function getStatusLabel(status: string) {
  switch (status) {
    case 'active': return t('reseller.announcements.statusActive')
    case 'draft': return t('reseller.announcements.statusDraft')
    case 'archived': return t('reseller.announcements.statusArchived')
    default: return status
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadAnnouncements()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadAnnouncements()
}

onMounted(() => {
  loadAnnouncements()
})
</script>
