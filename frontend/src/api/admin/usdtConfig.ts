/**
 * Admin USDT (多链) config API
 * 独立于 SettingsView 主 save payload，单独保存。
 */

import { apiClient } from '../client'

export interface AdminUsdtChainConfig {
  enabled: boolean
  address: string
  has_api_key: boolean
  api_base_url: string
}

export interface AdminUsdtConfig {
  enabled: boolean // 主开关
  manual_rate: number
  rate_auto_fetch: boolean
  rate_markup: number
  amount_offset: number
  amount_tolerance: number
  confirm_seconds: number
  monitor_interval_seconds: number
  query_minutes_back: number
  order_timeout_seconds: number
  chains: Record<string, AdminUsdtChainConfig>
}

export interface AdminUsdtChainConfigUpdate {
  enabled?: boolean
  address?: string
  api_key?: string // 空字符串=保留原值
  api_base_url?: string
}

export interface AdminUsdtConfigUpdate {
  enabled?: boolean
  manual_rate?: number
  rate_auto_fetch?: boolean
  rate_markup?: number
  amount_offset?: number
  amount_tolerance?: number
  confirm_seconds?: number
  monitor_interval_seconds?: number
  query_minutes_back?: number
  order_timeout_seconds?: number
  chains?: Record<string, AdminUsdtChainConfigUpdate>
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
