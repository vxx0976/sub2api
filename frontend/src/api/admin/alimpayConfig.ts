/**
 * Admin AliMPay (Alipay 个人免签) config API
 * 独立于 SettingsView 主 save payload，单独保存
 */

import { apiClient } from '../client'

export interface AdminAliMPayConfig {
  enabled: boolean
  mode: string
  app_id: string
  has_private_key: boolean
  has_alipay_public_key: boolean
  server_url: string
  transfer_user_id: string
  business_qr_url: string
  business_qr_path: string
  amount_offset: number
  match_tolerance_seconds: number
  monitor_interval_seconds: number
  query_minutes_back: number
  order_timeout_seconds: number
}

/**
 * 敏感字段（private_key / alipay_public_key）：
 *   - undefined/null 字段不更新
 *   - 空字符串 "" 保留原值（不清空，避免误覆盖）
 *   - 非空字符串覆盖保存
 */
export interface AdminAliMPayConfigUpdate {
  enabled?: boolean
  mode?: string
  app_id?: string
  private_key?: string
  alipay_public_key?: string
  server_url?: string
  transfer_user_id?: string
  business_qr_url?: string
  business_qr_path?: string
  amount_offset?: number
  match_tolerance_seconds?: number
  monitor_interval_seconds?: number
  query_minutes_back?: number
  order_timeout_seconds?: number
}

export async function getAliMPayConfig(): Promise<AdminAliMPayConfig> {
  const { data } = await apiClient.get<AdminAliMPayConfig>('/admin/alimpay/config')
  return data
}

export async function updateAliMPayConfig(req: AdminAliMPayConfigUpdate): Promise<AdminAliMPayConfig> {
  const { data } = await apiClient.put<AdminAliMPayConfig>('/admin/alimpay/config', req)
  return data
}

export const alimpayConfigAPI = { getAliMPayConfig, updateAliMPayConfig }
export default alimpayConfigAPI
