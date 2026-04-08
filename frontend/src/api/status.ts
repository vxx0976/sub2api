/**
 * Status API
 * Group status monitoring endpoints (public, no auth)
 */

export interface DailyStatus {
  date: string
  status: 'operational' | 'degraded' | 'down'
  rate: number
  total: number
}

export interface GroupStatusItem {
  id: number
  name: string
  platform: 'anthropic' | 'openai' | 'gemini' | 'antigravity'
  description?: string
  status: 'operational' | 'degraded' | 'down'
  success_rate: number
  total_requests: number
  uptime_30d: number
  daily_history: DailyStatus[]
}

export interface GroupStatusResponse {
  groups: GroupStatusItem[]
  updated_at: string
}

/**
 * Get group status (public endpoint, uses fetch to bypass auth interceptors)
 * GET /api/v1/group-status
 */
export async function getGroupStatus(): Promise<GroupStatusResponse> {
  const resp = await fetch('/api/v1/group-status', {
    headers: { 'Accept': 'application/json' }
  })
  if (!resp.ok) {
    throw new Error(`HTTP ${resp.status}`)
  }
  const json = await resp.json()
  if (json.code !== 0) {
    throw new Error(json.message || 'API error')
  }
  return json.data
}
