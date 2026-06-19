<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Header -->
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">{{ t('reseller.serviceTokens.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.description') }}</p>
        </div>
      </div>

      <!-- Create Form -->
      <div class="card p-6">
        <h2 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('reseller.serviceTokens.create') }}
        </h2>
        <div class="flex flex-wrap items-end gap-4">
          <div class="flex-1 min-w-[200px]">
            <label class="label">{{ t('reseller.serviceTokens.name') }}</label>
            <input
              v-model="createForm.name"
              type="text"
              maxlength="100"
              class="input w-full"
              :placeholder="t('reseller.serviceTokens.namePlaceholder')"
            />
          </div>
          <div>
            <label class="label">{{ t('reseller.serviceTokens.expiresInDays') }}</label>
            <input
              v-model.number="createForm.expires_in_days"
              type="number"
              min="0"
              class="input w-40"
              :placeholder="t('reseller.serviceTokens.neverExpires')"
            />
            <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">{{ t('reseller.serviceTokens.expiresInDaysHint') }}</p>
          </div>
          <button
            @click="handleCreate"
            class="btn btn-primary"
            :disabled="creating"
          >
            {{ creating ? '...' : t('reseller.serviceTokens.issue') }}
          </button>
        </div>
      </div>

      <!-- Usage Hint -->
      <div class="card p-6">
        <h2 class="mb-3 text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('reseller.serviceTokens.usageTitle') }}
        </h2>
        <p class="mb-3 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.usageHint') }}</p>
        <pre class="overflow-x-auto rounded-lg bg-gray-50 p-4 text-xs text-gray-700 dark:bg-dark-800 dark:text-gray-300"><code># {{ t('reseller.serviceTokens.usageResetComment') }}
curl -X POST {{ origin }}/api/v1/reseller-api/keys/&lt;KEY_ID&gt;/reset-quota \
     -H "X-Reseller-Token: rst-xxxx"

# {{ t('reseller.serviceTokens.usageEnableComment') }}
curl -X POST {{ origin }}/api/v1/reseller-api/keys/&lt;KEY_ID&gt;/enable \
     -H "X-Reseller-Token: rst-xxxx"</code></pre>
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-end">
        <button
          @click="loadTokens"
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

      <!-- Tokens Table -->
      <div v-else-if="tokens.length > 0" class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="border-b border-gray-100 dark:border-dark-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.columns.name') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.columns.prefix') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.columns.status') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.columns.lastUsed') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.columns.expiresAt') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('common.createdAt') }}</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="token in tokens"
                :key="token.id"
                class="border-b border-gray-50 dark:border-dark-800"
              >
                <td class="px-4 py-3">
                  <span class="text-sm font-medium text-gray-900 dark:text-white">{{ token.name || '—' }}</span>
                </td>
                <td class="px-4 py-3">
                  <code class="text-sm font-mono text-gray-900 dark:text-white">{{ token.token_prefix }}…</code>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium"
                    :class="getStatusClass(token.status)"
                  >
                    {{ getStatusLabel(token.status) }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">
                    {{ token.last_used_at ? formatDateTime(token.last_used_at) : t('reseller.serviceTokens.never') }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">
                    {{ token.expires_at ? formatDateTime(token.expires_at) : t('reseller.serviceTokens.neverExpires') }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(token.created_at) }}</span>
                </td>
                <td class="px-4 py-3">
                  <button
                    v-if="token.status === 'active'"
                    @click="handleRevoke(token)"
                    class="rounded-lg p-1.5 text-red-600 transition-colors hover:bg-red-50 dark:hover:bg-red-900/20"
                    :title="t('reseller.serviceTokens.revoke')"
                  >
                    <Icon name="trash" size="sm" />
                  </button>
                  <span v-else class="text-sm text-gray-300 dark:text-gray-600">—</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="card flex flex-col items-center justify-center py-12 text-center">
        <Icon name="code" size="lg" class="mb-3 text-gray-300 dark:text-gray-600" />
        <p class="text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.empty') }}</p>
      </div>
    </div>

    <!-- New Token Modal (plaintext shown once) -->
    <Teleport to="body">
      <div v-if="newToken" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="newToken = ''">
        <div class="mx-4 w-full max-w-lg rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-900">
          <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('reseller.serviceTokens.createdTitle') }}</h3>
          <p class="mb-4 text-sm text-amber-600 dark:text-amber-400">{{ t('reseller.serviceTokens.createdWarning') }}</p>
          <div class="flex items-center gap-2 rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
            <code class="flex-1 break-all text-sm font-mono text-gray-900 dark:text-white">{{ newToken }}</code>
            <button
              @click="copyToken(newToken)"
              class="shrink-0 rounded p-1.5 text-gray-400 transition-colors hover:text-primary-600 dark:hover:text-primary-400"
              :title="t('reseller.serviceTokens.copy')"
            >
              <Icon name="copy" size="sm" />
            </button>
          </div>
          <div class="mt-4 flex justify-end">
            <button @click="newToken = ''" class="btn btn-primary">{{ t('common.confirm') }}</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Revoke Confirmation -->
    <Teleport to="body">
      <div v-if="showRevokeModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click.self="showRevokeModal = false">
        <div class="mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-xl dark:bg-dark-900">
          <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">{{ t('reseller.serviceTokens.revoke') }}</h3>
          <p class="mb-4 text-sm text-gray-500 dark:text-gray-400">{{ t('reseller.serviceTokens.revokeConfirm') }}</p>
          <div class="flex justify-end gap-3">
            <button @click="showRevokeModal = false" class="btn btn-secondary">{{ t('common.cancel') }}</button>
            <button @click="confirmRevoke" class="btn btn-danger" :disabled="revoking">
              {{ revoking ? '...' : t('reseller.serviceTokens.revoke') }}
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
import type { ResellerServiceToken } from '@/api/reseller/serviceTokens'
import { useAppStore } from '@/stores'
import { formatDateTime } from '@/utils/format'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(true)
const creating = ref(false)
const revoking = ref(false)
const showRevokeModal = ref(false)
const revokeTarget = ref<ResellerServiceToken | null>(null)
const newToken = ref('')
const tokens = ref<ResellerServiceToken[]>([])

const origin = computed(() => window.location.origin)

const createForm = reactive<{ name: string; expires_in_days: number | null }>({
  name: '',
  expires_in_days: null
})

async function loadTokens() {
  loading.value = true
  try {
    tokens.value = await resellerAPI.serviceTokens.list() || []
  } catch (error: any) {
    appStore.showError(error.message || t('common.loadFailed'))
  } finally {
    loading.value = false
  }
}

async function handleCreate() {
  creating.value = true
  try {
    const payload: { name?: string; expires_in_days?: number } = {}
    if (createForm.name.trim()) payload.name = createForm.name.trim()
    if (createForm.expires_in_days && createForm.expires_in_days > 0) payload.expires_in_days = createForm.expires_in_days

    const res = await resellerAPI.serviceTokens.create(payload)
    newToken.value = res.token
    createForm.name = ''
    createForm.expires_in_days = null
    loadTokens()
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    creating.value = false
  }
}

function handleRevoke(token: ResellerServiceToken) {
  revokeTarget.value = token
  showRevokeModal.value = true
}

async function confirmRevoke() {
  if (!revokeTarget.value) return
  revoking.value = true
  try {
    await resellerAPI.serviceTokens.revoke(revokeTarget.value.id)
    appStore.showSuccess(t('reseller.serviceTokens.revokeSuccess'))
    showRevokeModal.value = false
    loadTokens()
  } catch (error: any) {
    appStore.showError(error.message || t('common.operationFailed'))
  } finally {
    revoking.value = false
  }
}

async function copyToken(value: string) {
  try {
    await navigator.clipboard.writeText(value)
    appStore.showSuccess(t('reseller.serviceTokens.copied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

function getStatusClass(status: string) {
  switch (status) {
    case 'active':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'revoked':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-600 dark:bg-dark-700 dark:text-gray-400'
  }
}

function getStatusLabel(status: string) {
  switch (status) {
    case 'active':
      return t('reseller.serviceTokens.statusActive')
    case 'revoked':
      return t('reseller.serviceTokens.statusRevoked')
    default:
      return status
  }
}

onMounted(() => {
  loadTokens()
})
</script>
