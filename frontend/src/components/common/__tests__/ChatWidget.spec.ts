import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { User } from '@/types'

const connect = vi.hoisted(() => vi.fn())
const disconnect = vi.hoisted(() => vi.fn())
const getMessages = vi.hoisted(() => vi.fn())

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/composables/useChatWebSocket', async () => {
  const { ref } = await vi.importActual<typeof import('vue')>('vue')
  return {
    useChatWebSocket: () => ({
      connect,
      disconnect,
      isConnected: ref(false),
    }),
  }
})

vi.mock('@/api/chat', () => ({
  createOrGetConversation: vi.fn(),
  peekConversation: vi.fn(),
  sendMessage: vi.fn(),
  getMessages,
}))

import ChatWidget from '@/components/common/ChatWidget.vue'
import { useAuthStore } from '@/stores/auth'
import { useChatStore } from '@/stores/chat'

function setupAuthenticatedChat() {
  localStorage.setItem('auth_token', 'access-token')
  const pinia = createPinia()
  setActivePinia(pinia)

  const authStore = useAuthStore()
  authStore.$patch((state) => {
    state.token = 'access-token'
    state.user = { id: 7, role: 'user' } as unknown as User
  })

  const chatStore = useChatStore()
  chatStore.$patch({
    conversation: {
      id: 42,
      visitor_name: 'User',
      status: 'open',
      admin_unread_count: 0,
      last_message_preview: 'Existing message',
      created_at: '2026-07-13T00:00:00Z',
      updated_at: '2026-07-13T00:00:00Z',
    },
    messages: [{
      id: 9,
      conversation_id: 42,
      sender_type: 'admin',
      content: 'Existing message',
      created_at: '2026-07-13T00:01:00Z',
    }],
  })

  return { pinia, authStore, chatStore }
}

describe('ChatWidget WebSocket lifecycle', () => {
  beforeEach(() => {
    localStorage.clear()
    connect.mockReset()
    disconnect.mockReset()
    getMessages.mockReset().mockResolvedValue({ messages: [], total: 0 })
  })

  it('authenticates logged-in WebSockets and reconnects an existing conversation when reopened', async () => {
    const { pinia } = setupAuthenticatedChat()

    const wrapper = mount(ChatWidget, {
      global: {
        plugins: [pinia],
        stubs: { Teleport: true, Transition: false },
      },
    })
    await flushPromises()

    const toggle = wrapper.get('button[title="chat.title"]')
    await toggle.trigger('click')
    await flushPromises()

    expect(connect).toHaveBeenCalledTimes(1)
    expect(connect).toHaveBeenLastCalledWith(
      `ws://${window.location.host}/api/v1/chat/ws?conversation_id=42`,
      expect.objectContaining({
        protocols: expect.any(Function),
      }),
    )
    const firstOptions = connect.mock.calls[0][1] as { protocols: () => string[] | undefined }
    expect(firstOptions.protocols()).toEqual(['sub2api-chat', 'jwt.access-token'])
    localStorage.setItem('auth_token', 'refreshed-access-token')
    expect(firstOptions.protocols()).toEqual(['sub2api-chat', 'jwt.refreshed-access-token'])

    await toggle.trigger('click')
    await flushPromises()
    expect(disconnect).toHaveBeenCalledTimes(1)

    await toggle.trigger('click')
    await flushPromises()

    expect(connect).toHaveBeenCalledTimes(2)
    wrapper.unmount()
  })

  it('disconnects and clears the previous conversation when the user logs out', async () => {
    const { pinia, authStore, chatStore } = setupAuthenticatedChat()
    const wrapper = mount(ChatWidget, {
      global: {
        plugins: [pinia],
        stubs: { Teleport: true, Transition: false },
      },
    })
    await flushPromises()

    await wrapper.get('button[title="chat.title"]').trigger('click')
    await flushPromises()
    expect(connect).toHaveBeenCalledTimes(1)

    localStorage.removeItem('auth_token')
    authStore.$patch((state) => {
      state.token = null
      state.user = null
    })
    await flushPromises()

    expect(disconnect).toHaveBeenCalledTimes(1)
    expect(chatStore.isOpen).toBe(false)
    expect(chatStore.conversation).toBeNull()
    expect(chatStore.messages).toEqual([])

    wrapper.unmount()
  })
})
