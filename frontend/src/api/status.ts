/**
 * Status API
 * System health check endpoints
 */

import apiClient from './client'

export interface ServiceStatus {
  status: 'healthy' | 'degraded' | 'down' | 'unknown'
  latency?: number
}

export interface SystemStatusResponse {
  api: ServiceStatus
  database: ServiceStatus
  cache: ServiceStatus
}

/**
 * Get system status from our backend
 * GET /api/v1/status
 */
export function getSystemStatus() {
  return apiClient.get<SystemStatusResponse>('/status')
}

// Claude Status Page API types
export interface ClaudeComponent {
  id: string
  name: string
  status: 'operational' | 'degraded_performance' | 'partial_outage' | 'major_outage' | 'under_maintenance'
  created_at: string
  updated_at: string
  position: number
  showcase: boolean
  start_date: string
}

export interface ClaudeStatusResponse {
  page: {
    id: string
    name: string
    url: string
    time_zone: string
    updated_at: string
  }
  components: ClaudeComponent[]
}

export interface ClaudeIncident {
  id: string
  name: string
  status: 'investigating' | 'identified' | 'monitoring' | 'resolved' | 'postmortem'
  created_at: string
  updated_at: string
  resolved_at: string | null
  impact: 'none' | 'minor' | 'major' | 'critical'
  shortlink: string
  components: Array<{
    id: string
    name: string
  }>
}

export interface ClaudeIncidentsResponse {
  page: {
    id: string
    name: string
    url: string
  }
  incidents: ClaudeIncident[]
}

/**
 * Get Claude status from official status page
 * Uses Statuspage.io API
 */
export async function getClaudeStatus(): Promise<ClaudeStatusResponse | null> {
  try {
    const response = await fetch('https://status.claude.com/api/v2/components.json', {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      }
    })
    if (!response.ok) {
      return null
    }
    return await response.json()
  } catch {
    return null
  }
}

/**
 * Get Claude incidents from official status page
 * Uses Statuspage.io API
 */
export async function getClaudeIncidents(): Promise<ClaudeIncidentsResponse | null> {
  try {
    const response = await fetch('https://status.claude.com/api/v2/incidents.json', {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      }
    })
    if (!response.ok) {
      return null
    }
    return await response.json()
  } catch {
    return null
  }
}

/**
 * Map Statuspage status to our status format
 */
export function mapClaudeStatus(status: ClaudeComponent['status']): 'operational' | 'degraded' | 'down' {
  switch (status) {
    case 'operational':
      return 'operational'
    case 'degraded_performance':
    case 'partial_outage':
      return 'degraded'
    case 'major_outage':
    case 'under_maintenance':
      return 'down'
    default:
      return 'operational'
  }
}

/**
 * Map incident impact to uptime status
 */
export function mapIncidentImpact(impact: ClaudeIncident['impact']): 'operational' | 'degraded' | 'down' {
  switch (impact) {
    case 'none':
      return 'operational'
    case 'minor':
      return 'degraded'
    case 'major':
    case 'critical':
      return 'down'
    default:
      return 'operational'
  }
}

/**
 * Generate 90-day uptime history from incidents
 * Only shows major/critical incidents to better match official status page display
 * @param incidents Array of incidents
 * @param componentName Name of component to filter (e.g., 'Claude Code', 'Claude API')
 * @returns Array of 90 status strings (oldest to newest)
 */
export function generateUptimeFromIncidents(
  incidents: ClaudeIncident[],
  componentName?: string
): string[] {
  const history: string[] = Array(90).fill('operational')
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  for (const incident of incidents) {
    // Only show significant incidents (major/critical)
    // Minor incidents don't significantly impact the day's status
    if (incident.impact === 'none' || incident.impact === 'minor') {
      continue
    }

    // Filter by component if specified - use exact match for better accuracy
    if (componentName) {
      const hasComponent = incident.components.some(c =>
        c.name === componentName ||
        c.name.toLowerCase() === componentName.toLowerCase()
      )
      if (!hasComponent) continue
    }

    const startDate = new Date(incident.created_at)
    startDate.setHours(0, 0, 0, 0)

    // Only mark the start day, not the entire duration
    // Most incidents are resolved within hours
    const endDate = new Date(incident.created_at)
    endDate.setHours(23, 59, 59, 999)

    // Check each day in the 90-day window
    for (let i = 0; i < 90; i++) {
      const dayDate = new Date(today)
      dayDate.setDate(today.getDate() - (89 - i))

      // Check if this day is the incident day
      if (dayDate >= startDate && dayDate <= endDate) {
        const status = mapIncidentImpact(incident.impact)
        // Keep the worst status for each day
        if (history[i] === 'operational' ||
            (history[i] === 'degraded' && status === 'down')) {
          history[i] = status
        }
      }
    }
  }

  return history
}
