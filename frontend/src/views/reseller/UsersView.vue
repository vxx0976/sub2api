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
                  <span class="text-sm font-medium text-gray-900 dark:text-white">${{ user.balance.toFixed(2) }}</span>
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
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { resellerAPI } from '@/api'
import type { User } from '@/types'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
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
})
</script>
