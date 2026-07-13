import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { useChatWebSocket } from '@/composables/useChatWebSocket'

class MockWebSocket {
  static instances: MockWebSocket[] = []

  onopen: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null
  close = vi.fn()

  constructor(
    readonly url: string,
    readonly protocols?: string | string[],
  ) {
    MockWebSocket.instances.push(this)
  }
}

describe('useChatWebSocket', () => {
  const originalWebSocket = globalThis.WebSocket

  beforeEach(() => {
    vi.useFakeTimers()
    MockWebSocket.instances = []
    Object.defineProperty(globalThis, 'WebSocket', {
      configurable: true,
      value: MockWebSocket,
    })
  })

  afterEach(() => {
    vi.useRealTimers()
    Object.defineProperty(globalThis, 'WebSocket', {
      configurable: true,
      value: originalWebSocket,
    })
  })

  it('resolves protocols again for every automatic reconnect', async () => {
    let accessToken = 'token-v1'
    let socketApi: ReturnType<typeof useChatWebSocket> | undefined
    const wrapper = mount(defineComponent({
      setup() {
        socketApi = useChatWebSocket()
        return () => h('div')
      },
    }))

    socketApi!.connect('ws://localhost/chat', {
      protocols: () => ['sub2api-chat', `jwt.${accessToken}`],
      onMessage: vi.fn(),
    })

    expect(MockWebSocket.instances).toHaveLength(1)
    expect(MockWebSocket.instances[0].protocols).toEqual(['sub2api-chat', 'jwt.token-v1'])

    accessToken = 'token-v2'
    MockWebSocket.instances[0].onclose?.(new CloseEvent('close'))
    await vi.advanceTimersByTimeAsync(1000)

    expect(MockWebSocket.instances).toHaveLength(2)
    expect(MockWebSocket.instances[1].protocols).toEqual(['sub2api-chat', 'jwt.token-v2'])

    wrapper.unmount()
  })
})
