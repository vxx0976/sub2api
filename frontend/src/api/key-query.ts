import { apiClient } from './client'

export interface KeyQueryResponse {
  key: {
    name: string
    status: string
    quota: number
    quota_used: number
    quota_remaining: number
    expires_at: string | null
    days_until_expiry: number
  }
  subscription: {
    group_name: string
    status: string
    expires_at: string
    days_remaining: number
    daily_limit_usd: number | null
    daily_usage_usd: number
    weekly_limit_usd: number | null
    weekly_usage_usd: number
    monthly_limit_usd: number | null
    monthly_usage_usd: number
  } | null
  usage_summary: {
    total_cost_7d: number
    request_count_7d: number
    top_models: Array<{ model: string; count: number; cost: number }>
  }
}

export interface UsageLog {
  id: number
  created_at: string
  model: string
  input_tokens: number
  output_tokens: number
  cache_creation_tokens: number
  cache_read_tokens: number
  total_cost: number
  actual_cost: number
  duration_ms: number | null
}

export interface Pagination {
  total: number
  page: number
  page_size: number
  pages: number
}

export interface UsageStatsResponse {
  total_requests: number
  total_input_tokens: number
  total_output_tokens: number
  total_cache_tokens: number
  total_tokens: number
  total_cost: number
  total_actual_cost: number
  average_duration_ms: number
}

export interface ModelStat {
  model: string
  requests: number
  input_tokens: number
  output_tokens: number
  total_tokens: number
  cost: number
  actual_cost: number
}

export interface TrendPoint {
  date: string
  requests: number
  input_tokens: number
  output_tokens: number
  total_tokens: number
  cost: number
  actual_cost: number
}

export async function queryApiKey(apiKey: string): Promise<KeyQueryResponse> {
  const { data } = await apiClient.post<KeyQueryResponse>('/public/key-query', { api_key: apiKey })
  return data
}

export async function queryKeyUsage(params: {
  api_key: string
  page?: number
  page_size?: number
  start_date?: string
  end_date?: string
  model?: string
}): Promise<{ list: UsageLog[]; pagination: Pagination }> {
  const { data } = await apiClient.post('/public/key-usage', params)
  // Backend returns { items, total, page, page_size, pages } via response.Paginated
  const raw = data as any
  return {
    list: raw.items ?? [],
    pagination: {
      total: raw.total ?? 0,
      page: raw.page ?? 1,
      page_size: raw.page_size ?? 20,
      pages: raw.pages ?? 1
    }
  }
}

export async function queryKeyStats(params: {
  api_key: string
  start_date?: string
  end_date?: string
}): Promise<UsageStatsResponse> {
  const { data } = await apiClient.post<UsageStatsResponse>('/public/key-usage/stats', params)
  return data
}

export async function queryKeyModels(params: {
  api_key: string
  start_date?: string
  end_date?: string
}): Promise<ModelStat[]> {
  const { data } = await apiClient.post('/public/key-usage/models', params)
  // Backend returns { models: [...], start_date, end_date }
  const raw = data as any
  return raw.models ?? []
}

export async function queryKeyTrend(params: {
  api_key: string
  start_date?: string
  end_date?: string
}): Promise<TrendPoint[]> {
  const { data } = await apiClient.post('/public/key-usage/trend', params)
  // Backend returns { trend: [...], start_date, end_date }
  const raw = data as any
  return raw.trend ?? []
}
