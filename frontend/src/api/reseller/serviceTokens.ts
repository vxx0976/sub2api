/**
 * Reseller M2M Service Token API endpoints
 *
 * 服务令牌用于分销商后端非交互式调用 key 管理接口（重置配额/启用-禁用/创建 key）。
 * 明文仅在创建时返回一次。
 */

import { apiClient } from '../client'

export interface ResellerServiceToken {
  id: number
  name: string
  token_prefix: string
  status: string
  last_used_at: string | null
  expires_at: string | null
  created_at: string
}

export interface CreateServiceTokenRequest {
  name?: string
  expires_in_days?: number // 省略或 <=0 表示永不过期
}

export interface CreateServiceTokenResponse {
  token: string // 明文，仅此一次可见
  info: ResellerServiceToken
}

/**
 * List the reseller's service tokens (safe view, no plaintext)
 */
export async function list(): Promise<ResellerServiceToken[]> {
  const { data } = await apiClient.get<ResellerServiceToken[]>('/reseller/service-tokens')
  return data
}

/**
 * Issue a new service token. The plaintext is returned exactly once.
 */
export async function create(payload: CreateServiceTokenRequest): Promise<CreateServiceTokenResponse> {
  const { data } = await apiClient.post<CreateServiceTokenResponse>('/reseller/service-tokens', payload)
  return data
}

/**
 * Revoke a service token
 */
export async function revoke(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/reseller/service-tokens/${id}`)
  return data
}

export const serviceTokensAPI = {
  list,
  create,
  revoke
}

export default serviceTokensAPI
