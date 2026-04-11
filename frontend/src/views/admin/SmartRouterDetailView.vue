<template>
  <AppLayout>
    <div class="p-4 md:p-6 space-y-4">
      <!-- Header -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <router-link to="/admin/groups" class="text-gray-500 hover:text-primary-600 dark:text-gray-400">
            <Icon name="arrowLeft" size="md" />
          </router-link>
          <div>
            <h1 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
              {{ status?.name || t('admin.groups.smartRouter.detailTitle') }}
            </h1>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('admin.groups.smartRouter.detailDescription') }}
            </p>
          </div>
        </div>
        <button
          type="button"
          class="btn btn-secondary btn-sm"
          :disabled="loading"
          @click="loadStatus"
        >
          <Icon name="refresh" size="sm" :class="{ 'animate-spin': loading }" />
          <span class="ml-1">{{ t('common.refresh') }}</span>
        </button>
      </div>

      <!-- Loading -->
      <div v-if="loading && !status" class="rounded-lg border border-gray-200 dark:border-dark-600 p-8 text-center text-gray-500">
        {{ t('common.loading') }}
      </div>

      <template v-if="status">
        <!-- Pin banner -->
        <div
          v-if="status.pin_member_id"
          class="rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-900/20 dark:border-amber-700 p-3 flex items-center justify-between"
        >
          <div class="text-sm text-amber-800 dark:text-amber-200">
            <span class="font-medium">
              {{ t('admin.groups.smartRouter.pinActive') }}:
            </span>
            {{ memberName(status.pin_member_id) }}
            <span v-if="status.pin_expires_at" class="ml-2 text-xs opacity-80">
              ({{ t('admin.groups.smartRouter.pinExpiresAt') }} {{ formatDate(status.pin_expires_at) }})
            </span>
          </div>
          <button
            type="button"
            class="btn btn-secondary btn-sm"
            :disabled="actionLoading"
            @click="clearPin"
          >
            {{ t('admin.groups.smartRouter.clearPin') }}
          </button>
        </div>

        <!-- Members -->
        <div class="rounded-lg border border-gray-200 dark:border-dark-600 overflow-hidden">
          <div class="px-4 py-2.5 border-b border-gray-200 dark:border-dark-600 bg-gray-50 dark:bg-dark-700 text-sm font-medium text-gray-700 dark:text-gray-200">
            {{ t('admin.groups.smartRouter.members') }}
          </div>
          <div v-if="status.members.length === 0" class="p-6 text-center text-sm text-gray-500">
            {{ t('admin.groups.smartRouter.noMembers') }}
          </div>
          <ul v-else class="divide-y divide-gray-200 dark:divide-dark-600">
            <li
              v-for="(m, idx) in status.members"
              :key="m.group_id"
              class="p-4 flex items-center justify-between gap-3"
              :class="{ 'bg-primary-50 dark:bg-primary-900/10': status.active_member_id === m.group_id }"
            >
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <div class="w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold bg-gray-200 dark:bg-dark-500 text-gray-700 dark:text-gray-200">
                  {{ idx + 1 }}
                </div>
                <div class="min-w-0">
                  <div class="flex items-center gap-2">
                    <div class="font-medium truncate text-gray-900 dark:text-gray-100">
                      {{ m.name }}
                    </div>
                    <span
                      v-if="status.active_member_id === m.group_id"
                      class="text-[10px] rounded-full px-1.5 py-0.5 bg-primary-500 text-white"
                    >
                      {{ t('admin.groups.smartRouter.active') }}
                    </span>
                    <span
                      :class="healthPillClass(m.health_status)"
                      class="text-[10px] rounded-full px-1.5 py-0.5"
                    >
                      {{ healthLabel(m.health_status) }}
                    </span>
                  </div>
                  <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {{ t('admin.groups.smartRouter.schedulable') }}: {{ m.schedulable_accounts }}
                    · {{ t('admin.groups.smartRouter.healthy') }}: {{ m.healthy_accounts }}/{{ m.total_checked_accounts }}
                    <span v-if="m.last_health_check_at" class="ml-2">
                      · {{ formatDate(m.last_health_check_at) }}
                    </span>
                  </div>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <button
                  type="button"
                  class="btn btn-secondary btn-xs"
                  :disabled="actionLoading"
                  @click="probeMember(m.group_id)"
                >
                  {{ t('admin.groups.smartRouter.probe') }}
                </button>
                <button
                  type="button"
                  class="btn btn-secondary btn-xs"
                  :disabled="actionLoading"
                  @click="openPinDialog(m.group_id)"
                >
                  {{ t('admin.groups.smartRouter.pinHere') }}
                </button>
              </div>
            </li>
          </ul>
        </div>

        <!-- Usage -->
        <div class="rounded-lg border border-gray-200 dark:border-dark-600 overflow-hidden">
          <div class="px-4 py-2.5 border-b border-gray-200 dark:border-dark-600 bg-gray-50 dark:bg-dark-700 text-sm font-medium text-gray-700 dark:text-gray-200 flex items-center justify-between">
            <span>{{ t('admin.groups.smartRouter.usageTitle') }}</span>
            <select v-model.number="usageDays" class="input input-sm !w-auto" @change="loadUsage">
              <option :value="1">{{ t('admin.groups.smartRouter.days1') }}</option>
              <option :value="7">{{ t('admin.groups.smartRouter.days7') }}</option>
              <option :value="30">{{ t('admin.groups.smartRouter.days30') }}</option>
            </select>
          </div>
          <div v-if="!usage || usage.members.length === 0" class="p-4 text-sm text-gray-500 text-center">
            {{ t('admin.groups.smartRouter.noUsage') }}
          </div>
          <table v-else class="w-full text-sm">
            <thead class="bg-gray-50 dark:bg-dark-700 text-xs text-gray-500 dark:text-gray-400">
              <tr>
                <th class="px-4 py-2 text-left">{{ t('admin.groups.smartRouter.memberName') }}</th>
                <th class="px-4 py-2 text-right">{{ t('admin.groups.smartRouter.requests') }}</th>
                <th class="px-4 py-2 text-right">{{ t('admin.groups.smartRouter.tokens') }}</th>
                <th class="px-4 py-2 text-right">{{ t('admin.groups.smartRouter.cost') }}</th>
                <th class="px-4 py-2 text-right">{{ t('admin.groups.smartRouter.percent') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-dark-600">
              <tr v-for="m in usage.members" :key="m.group_id">
                <td class="px-4 py-2 text-gray-900 dark:text-gray-100">{{ m.name }}</td>
                <td class="px-4 py-2 text-right">{{ m.requests.toLocaleString() }}</td>
                <td class="px-4 py-2 text-right">{{ m.tokens.toLocaleString() }}</td>
                <td class="px-4 py-2 text-right">${{ m.cost.toFixed(4) }}</td>
                <td class="px-4 py-2 text-right">{{ percentOf(m.cost).toFixed(1) }}%</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Events -->
        <div class="rounded-lg border border-gray-200 dark:border-dark-600 overflow-hidden">
          <div class="px-4 py-2.5 border-b border-gray-200 dark:border-dark-600 bg-gray-50 dark:bg-dark-700 text-sm font-medium text-gray-700 dark:text-gray-200">
            {{ t('admin.groups.smartRouter.eventsTitle') }}
          </div>
          <div v-if="status.recent_events.length === 0" class="p-4 text-sm text-gray-500 text-center">
            {{ t('admin.groups.smartRouter.noEvents') }}
          </div>
          <ul v-else class="divide-y divide-gray-200 dark:divide-dark-600 text-sm">
            <li
              v-for="e in status.recent_events"
              :key="e.id"
              class="px-4 py-2 flex items-center gap-3"
            >
              <span :class="eventDotClass(e.reason)" class="w-2 h-2 rounded-full shrink-0"></span>
              <span class="font-mono text-xs text-gray-500 dark:text-gray-400 shrink-0">
                {{ formatDate(e.occurred_at) }}
              </span>
              <span class="text-xs text-gray-700 dark:text-gray-300 truncate">
                {{ eventLabel(e.reason) }}:
                {{ memberName(e.from_member_id) }}
                →
                {{ memberName(e.to_member_id) }}
                <span v-if="e.note" class="ml-2 opacity-60">{{ e.note }}</span>
              </span>
            </li>
          </ul>
        </div>
      </template>
    </div>

    <!-- Pin Dialog -->
    <BaseDialog
      :show="showPinDialog"
      :title="t('admin.groups.smartRouter.pinDialog.title')"
      @close="showPinDialog = false"
    >
      <div class="space-y-3 text-sm">
        <p class="text-gray-600 dark:text-gray-400">
          {{ t('admin.groups.smartRouter.pinDialog.description') }}
          <span class="font-medium text-gray-900 dark:text-gray-100">
            {{ memberName(pinTarget) }}
          </span>
        </p>
        <div>
          <label class="input-label">{{ t('admin.groups.smartRouter.pinDialog.ttl') }}</label>
          <select v-model.number="pinTtl" class="input">
            <option :value="1800">{{ t('admin.groups.smartRouter.pinDialog.ttl30m') }}</option>
            <option :value="7200">{{ t('admin.groups.smartRouter.pinDialog.ttl2h') }}</option>
            <option :value="28800">{{ t('admin.groups.smartRouter.pinDialog.ttl8h') }}</option>
            <option :value="86400">{{ t('admin.groups.smartRouter.pinDialog.ttl24h') }}</option>
            <option :value="0">{{ t('admin.groups.smartRouter.pinDialog.ttlPermanent') }}</option>
          </select>
        </div>
      </div>
      <template #footer>
        <button type="button" class="btn btn-secondary" @click="showPinDialog = false">
          {{ t('common.cancel') }}
        </button>
        <button type="button" class="btn btn-primary" :disabled="actionLoading" @click="confirmPin">
          {{ t('admin.groups.smartRouter.pinDialog.confirm') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type { SmartRouterStatus, SmartRouterUsageResponse } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()

const groupId = computed(() => Number(route.params.id))

const status = ref<SmartRouterStatus | null>(null)
const usage = ref<SmartRouterUsageResponse | null>(null)
const usageDays = ref(7)
const loading = ref(false)
const actionLoading = ref(false)

const showPinDialog = ref(false)
const pinTarget = ref<number | null>(null)
const pinTtl = ref(7200)

const loadStatus = async () => {
  loading.value = true
  try {
    status.value = await adminAPI.groups.getFailoverStatus(groupId.value)
  } catch (err: any) {
    appStore.showError(err?.response?.data?.detail || String(err))
  } finally {
    loading.value = false
  }
}

const loadUsage = async () => {
  try {
    usage.value = await adminAPI.groups.getFailoverUsage(groupId.value, usageDays.value)
  } catch (err: any) {
    appStore.showError(err?.response?.data?.detail || String(err))
  }
}

const memberName = (id?: number | null): string => {
  if (!id) return '—'
  const m = status.value?.members.find((x) => x.group_id === id)
  return m?.name || `#${id}`
}

const healthLabel = (s: string): string => {
  if (s === 'available') return t('admin.groups.smartRouter.healthAvailable')
  if (s === 'unavailable') return t('admin.groups.smartRouter.healthUnavailable')
  return t('admin.groups.smartRouter.healthUnknown')
}

const healthPillClass = (s: string): string => {
  if (s === 'available') return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
  if (s === 'unavailable') return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  return 'bg-gray-100 text-gray-600 dark:bg-gray-700/40 dark:text-gray-300'
}

const eventLabel = (reason: string): string => {
  return t(`admin.groups.smartRouter.eventReason.${reason}`, reason)
}

const eventDotClass = (reason: string): string => {
  switch (reason) {
    case 'soft_demote':
      return 'bg-red-500'
    case 'probe_recovered':
      return 'bg-green-500'
    case 'health_reconcile':
      return 'bg-blue-500'
    case 'config_change':
      return 'bg-sky-500'
    case 'manual_pin':
      return 'bg-amber-500'
    case 'manual_unpin':
    case 'manual_unpin_expired':
      return 'bg-gray-500'
    case 'bootstrap':
      return 'bg-primary-500'
    default:
      return 'bg-gray-400'
  }
}

const formatDate = (iso: string | null | undefined): string => {
  if (!iso) return ''
  try {
    return new Date(iso).toLocaleString()
  } catch {
    return iso
  }
}

const probeMember = async (memberId: number) => {
  actionLoading.value = true
  try {
    const result = await adminAPI.groups.triggerFailoverProbe(groupId.value, memberId)
    if (result.success) {
      appStore.showSuccess(t('admin.groups.smartRouter.probeSuccess'))
    } else {
      appStore.showError(t('admin.groups.smartRouter.probeFailed'))
    }
    await loadStatus()
  } catch (err: any) {
    appStore.showError(err?.response?.data?.detail || String(err))
  } finally {
    actionLoading.value = false
  }
}

const openPinDialog = (memberId: number) => {
  pinTarget.value = memberId
  pinTtl.value = 7200
  showPinDialog.value = true
}

const confirmPin = async () => {
  if (!pinTarget.value) return
  actionLoading.value = true
  try {
    await adminAPI.groups.setFailoverPin(groupId.value, pinTarget.value, pinTtl.value)
    appStore.showSuccess(t('admin.groups.smartRouter.pinSet'))
    showPinDialog.value = false
    await loadStatus()
  } catch (err: any) {
    appStore.showError(err?.response?.data?.detail || String(err))
  } finally {
    actionLoading.value = false
  }
}

const percentOf = (cost: number): number => {
  if (!usage.value || usage.value.members.length === 0) return 0
  const total = usage.value.members.reduce((sum, m) => sum + m.cost, 0)
  if (total <= 0) return 0
  return (cost / total) * 100
}

const clearPin = async () => {
  actionLoading.value = true
  try {
    await adminAPI.groups.clearFailoverPin(groupId.value)
    appStore.showSuccess(t('admin.groups.smartRouter.pinCleared'))
    await loadStatus()
  } catch (err: any) {
    appStore.showError(err?.response?.data?.detail || String(err))
  } finally {
    actionLoading.value = false
  }
}

onMounted(() => {
  loadStatus()
  loadUsage()
})
</script>
