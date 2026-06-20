/**
 * Admin USDT (TRC20) config API
 * 独立于 SettingsView 主 save payload，单独保存。
 */

import { apiClient } from '../client'

export interface AdminUsdtConfig {
  enabled: boolean
  receiving_address: string
  tron_api_base_url: string
  has_tron_api_key: boolean
  manual_rate: number
  rate_auto_fetch: boolean
  rate_markup: number
  amount_offset: number
  confirm_seconds: number
  monitor_interval_seconds: number
  query_minutes_back: number
  order_timeout_seconds: number
}

/**
 * 敏感字段（tron_api_key）：
 *   - undefined/null 字段不更新
 *   - 空字符串 "" 保留原值（不清空）
 *   - 非空字符串覆盖保存
 */
export interface AdminUsdtConfigUpdate {
  enabled?: boolean
  receiving_address?: string
  tron_api_base_url?: string
  tron_api_key?: string
  manual_rate?: number
  rate_auto_fetch?: boolean
  rate_markup?: number
  amount_offset?: number
  confirm_seconds?: number
  monitor_interval_seconds?: number
  query_minutes_back?: number
  order_timeout_seconds?: number
}

export async function getUsdtConfig(): Promise<AdminUsdtConfig> {
  const { data } = await apiClient.get<AdminUsdtConfig>('/admin/usdt/config')
  return data
}

export async function updateUsdtConfig(req: AdminUsdtConfigUpdate): Promise<AdminUsdtConfig> {
  const { data } = await apiClient.put<AdminUsdtConfig>('/admin/usdt/config', req)
  return data
}

export const usdtConfigAPI = { getUsdtConfig, updateUsdtConfig }
export default usdtConfigAPI
