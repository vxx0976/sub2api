/**
 * Parse curl command and extract relevant fields for channel configuration
 */

export interface ParsedCurl {
  url: string
  method: 'GET' | 'POST'
  headers: Record<string, string>
  body?: string
}

/**
 * Parse a curl command string and extract URL, method, headers, and body
 */
export function parseCurl(curlCommand: string): ParsedCurl {
  const result: ParsedCurl = {
    url: '',
    method: 'GET',
    headers: {}
  }

  // Normalize the command - remove line continuations and extra whitespace
  let normalized = curlCommand
    .replace(/\\\r?\n/g, ' ') // Remove line continuations
    .replace(/\s+/g, ' ')     // Normalize whitespace
    .trim()

  // Remove 'curl' prefix if present
  if (normalized.toLowerCase().startsWith('curl ')) {
    normalized = normalized.substring(5).trim()
  }

  // Extract URL - it's usually the first non-flag argument or after specific flags
  const urlMatch = normalized.match(/['"]?(https?:\/\/[^\s'"]+)['"]?/)
  if (urlMatch) {
    result.url = urlMatch[1]
  }

  // Check for explicit method
  const methodMatch = normalized.match(/-X\s+['"]?(\w+)['"]?/i)
  if (methodMatch) {
    const method = methodMatch[1].toUpperCase()
    if (method === 'POST' || method === 'GET') {
      result.method = method
    }
  }

  // Extract headers (-H or --header)
  const headerRegex = /(?:-H|--header)\s+['"]([^'"]+)['"]/gi
  let headerMatch
  while ((headerMatch = headerRegex.exec(normalized)) !== null) {
    const headerLine = headerMatch[1]
    const colonIndex = headerLine.indexOf(':')
    if (colonIndex > 0) {
      const key = headerLine.substring(0, colonIndex).trim()
      const value = headerLine.substring(colonIndex + 1).trim()

      // Filter headers - keep only useful ones for API calls
      const lowerKey = key.toLowerCase()
      if (isRelevantHeader(lowerKey)) {
        result.headers[key] = value
      }
    }
  }

  // Extract data/body (-d or --data or --data-raw)
  const dataMatch = normalized.match(/(?:-d|--data|--data-raw)\s+['"](.+?)['"](?:\s+-|$)/i)
  if (dataMatch) {
    result.body = dataMatch[1]
    // If there's a body and no explicit method, assume POST
    if (!methodMatch) {
      result.method = 'POST'
    }
  }

  // Also try to match $'...' style data (common in bash)
  const dataDollarMatch = normalized.match(/(?:-d|--data|--data-raw)\s+\$'(.+?)'(?:\s+-|$)/i)
  if (dataDollarMatch && !result.body) {
    result.body = dataDollarMatch[1].replace(/\\'/g, "'")
    if (!methodMatch) {
      result.method = 'POST'
    }
  }

  return result
}

/**
 * Check if a header is relevant for API balance checking
 */
function isRelevantHeader(headerName: string): boolean {
  const relevantHeaders = [
    'authorization',
    'content-type',
    'accept',
    'x-api-key',
    'api-key',
    'x-auth-token',
    'x-access-token',
    'token',
    'bearer'
  ]

  // Skip browser-specific headers that aren't needed for API calls
  const skipHeaders = [
    'sec-ch-ua',
    'sec-ch-ua-mobile',
    'sec-ch-ua-platform',
    'sec-fetch-dest',
    'sec-fetch-mode',
    'sec-fetch-site',
    'sec-gpc',
    'user-agent',
    'accept-language',
    'accept-encoding',
    'cache-control',
    'pragma',
    'referer',
    'origin',
    'cookie',
    'connection',
    'upgrade-insecure-requests',
    'dnt',
    'priority'
  ]

  if (skipHeaders.includes(headerName)) {
    return false
  }

  // Include if it's in our relevant list
  if (relevantHeaders.some(h => headerName.includes(h))) {
    return true
  }

  // Include custom headers (x-* but not sec-* which we already filtered)
  if (headerName.startsWith('x-')) {
    return true
  }

  return false
}

/**
 * Validate parsed curl result
 */
export function validateParsedCurl(parsed: ParsedCurl): string | null {
  if (!parsed.url) {
    return 'URL not found in curl command'
  }

  try {
    new URL(parsed.url)
  } catch {
    return 'Invalid URL format'
  }

  return null
}
