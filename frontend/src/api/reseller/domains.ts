/**
 * Reseller Domain API endpoints
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface ResellerDomain {
  id: number
  reseller_id: number
  domain: string
  site_name: string
  site_logo: string
  brand_color: string
  custom_css?: string
  subtitle?: string
  home_content?: string
  doc_url?: string
  verify_token?: string
  verified: boolean
  verified_at?: string
  created_at: string
  updated_at: string
  home_template?: string
  purchase_enabled: boolean
  purchase_url?: string
  default_locale?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string
  login_redirect?: string
}

export interface CreateDomainRequest {
  domain: string
  site_name?: string
  site_logo?: string
  brand_color?: string
  custom_css?: string
  subtitle?: string
  home_content?: string
  doc_url?: string
  home_template?: string
  purchase_enabled?: boolean
  purchase_url?: string
  default_locale?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string
  login_redirect?: string
}

export interface UpdateDomainRequest {
  site_name?: string
  site_logo?: string
  brand_color?: string
  custom_css?: string
  subtitle?: string
  home_content?: string
  doc_url?: string
  home_template?: string
  purchase_enabled?: boolean
  purchase_url?: string
  default_locale?: string
  seo_title?: string
  seo_description?: string
  seo_keywords?: string
  login_redirect?: string
}

/**
 * List reseller's domains
 */
export async function list(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<ResellerDomain>> {
  const { data } = await apiClient.get<PaginatedResponse<ResellerDomain>>('/reseller/domains', {
    params: { page, page_size: pageSize }
  })
  return data
}

/**
 * Create a new domain
 */
export async function create(domainData: CreateDomainRequest): Promise<ResellerDomain> {
  const { data } = await apiClient.post<ResellerDomain>('/reseller/domains', domainData)
  return data
}

/**
 * Update a domain
 */
export async function update(id: number, updates: UpdateDomainRequest): Promise<ResellerDomain> {
  const { data } = await apiClient.put<ResellerDomain>(`/reseller/domains/${id}`, updates)
  return data
}

/**
 * Delete a domain
 */
export async function deleteDomain(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/reseller/domains/${id}`)
  return data
}

/**
 * Verify a domain via DNS TXT record
 */
export async function verify(id: number): Promise<ResellerDomain> {
  const { data } = await apiClient.post<ResellerDomain>(`/reseller/domains/${id}/verify`)
  return data
}

/**
 * Get server info (main domain + resolved IP) for DNS setup guidance
 */
export async function getServerInfo(): Promise<{ domain: string; ip: string }> {
  const { data } = await apiClient.get<{ domain: string; ip: string }>('/reseller/domains/server-info')
  return data
}

export const domainsAPI = {
  list,
  create,
  update,
  delete: deleteDomain,
  verify,
  getServerInfo
}

export default domainsAPI
