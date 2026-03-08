/**
 * Reseller Settings Store
 * Caches merchant KV settings for the current reseller session.
 * Loaded once on first access; cleared on logout.
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { resellerAPI } from '@/api'

export const useResellerSettingsStore = defineStore('resellerSettings', () => {
  const settings = ref<Record<string, string>>({})
  const loaded = ref(false)

  /** 商户代理功能是否启用 */
  const isAgentEnabled = computed(() => settings.value.merchant_mode === 'enabled')

  async function load() {
    if (loaded.value) return
    try {
      const data = await resellerAPI.settings.get()
      settings.value = data
      loaded.value = true
    } catch {
      // ignore – user might not be a reseller
    }
  }

  function reset() {
    settings.value = {}
    loaded.value = false
  }

  return { settings, loaded, isAgentEnabled, load, reset }
})
