import { ref, onUnmounted } from 'vue'

export interface ChatWSOptions {
  onMessage: (data: any) => void
  onOpen?: () => void
  onClose?: () => void
  protocols?: string[] | (() => string[] | undefined)
}

export function useChatWebSocket() {
  let ws: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let reconnectAttempts = 0
  const maxReconnectAttempts = 10
  const isConnected = ref(false)
  let shouldReconnect = true
  let currentUrl = ''
  let currentOptions: ChatWSOptions | null = null

  function connect(url: string, options: ChatWSOptions) {
    currentUrl = url
    currentOptions = options
    shouldReconnect = true
    reconnectAttempts = 0
    doConnect()
  }

  function doConnect() {
    if (ws) {
      ws.onclose = null
      ws.close()
    }

    const configuredProtocols = currentOptions?.protocols
    const protocols = typeof configuredProtocols === 'function'
      ? configuredProtocols()
      : configuredProtocols
    ws = new WebSocket(currentUrl, protocols)

    ws.onopen = () => {
      isConnected.value = true
      reconnectAttempts = 0
      currentOptions?.onOpen?.()
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        currentOptions?.onMessage(data)
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      isConnected.value = false
      currentOptions?.onClose?.()
      if (shouldReconnect && reconnectAttempts < maxReconnectAttempts) {
        const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000)
        reconnectTimer = setTimeout(() => {
          reconnectAttempts++
          doConnect()
        }, delay)
      }
    }

    ws.onerror = () => {
      // onclose will fire after onerror
    }
  }

  function disconnect() {
    shouldReconnect = false
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.onclose = null
      ws.close()
      ws = null
    }
    isConnected.value = false
  }

  onUnmounted(() => {
    disconnect()
  })

  return { connect, disconnect, isConnected }
}
