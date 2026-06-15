import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { createOrGetConversation, sendMessage } from '@/api/chat'
import type { ChatConversation, ChatMessage } from '@/types'
import { useAuthStore } from './auth'

function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

export const useChatStore = defineStore('chat', () => {
  const isOpen = ref(false)
  const conversation = ref<ChatConversation | null>(null)
  const messages = ref<ChatMessage[]>([])
  const loading = ref(false)
  const sending = ref(false)
  const hasNewMessage = ref(false)
  const welcomeShown = ref(!!localStorage.getItem('chat_welcome_shown'))

  const guestToken = computed(() => {
    let token = localStorage.getItem('chat_guest_token')
    if (!token) {
      token = generateUUID()
      localStorage.setItem('chat_guest_token', token)
    }
    return token
  })

  async function openChat() {
    isOpen.value = true
    hasNewMessage.value = false

    if (conversation.value || loading.value) return

    loading.value = true
    try {
      const authStore = useAuthStore()
      const token = authStore.isAuthenticated ? undefined : guestToken.value
      const result = await createOrGetConversation(token)
      conversation.value = result.conversation
      messages.value = result.messages || []
    } catch (e) {
      console.error('Failed to open chat:', e)
    } finally {
      loading.value = false
    }
  }

  function closeChat() {
    isOpen.value = false
  }

  async function send(content: string) {
    if (!conversation.value || !content.trim()) return

    sending.value = true
    try {
      const authStore = useAuthStore()
      const token = authStore.isAuthenticated ? undefined : guestToken.value
      const msg = await sendMessage(conversation.value.id, content.trim(), token)
      if (!messages.value.some((m) => m.id === msg.id)) {
        messages.value.push(msg)
      }
    } catch (e) {
      console.error('Failed to send message:', e)
    } finally {
      sending.value = false
    }
  }

  function receiveMessage(msg: ChatMessage) {
    if (conversation.value && msg.conversation_id === conversation.value.id) {
      const exists = messages.value.some((m) => m.id === msg.id)
      if (!exists) {
        messages.value.push(msg)
        if (!isOpen.value) {
          hasNewMessage.value = true
        }
      }
    }
  }

  function markWelcomeShown() {
    welcomeShown.value = true
    localStorage.setItem('chat_welcome_shown', '1')
  }

  function reset() {
    isOpen.value = false
    conversation.value = null
    messages.value = []
    hasNewMessage.value = false
  }

  return {
    isOpen,
    conversation,
    messages,
    loading,
    sending,
    hasNewMessage,
    welcomeShown,
    guestToken,
    openChat,
    closeChat,
    send,
    receiveMessage,
    markWelcomeShown,
    reset,
  }
})
