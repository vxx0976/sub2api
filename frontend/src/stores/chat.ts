import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { createOrGetConversation, peekConversation, sendMessage, getMessages } from '@/api/chat'
import type { ChatConversation, ChatMessage } from '@/types'
import { useAuthStore } from './auth'

function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

const CONV_ID_KEY = 'chat_conversation_id'
// 已读位按「会话 id」持久化(chat_read_<id>):消息 id 全局自增,按会话隔离可避免
// 同一浏览器多账号互相串读;刷新/换设备后仍能正确判断管理员回复是否已读。
function readKey(convId: number): string {
  return `chat_read_${convId}`
}

export const useChatStore = defineStore('chat', () => {
  const isOpen = ref(false)
  const conversation = ref<ChatConversation | null>(null)
  const messages = ref<ChatMessage[]>([])
  const loading = ref(false)
  const sending = ref(false)
  const hasNewMessage = ref(false)
  // 客服(管理员)是否在线:开会话/探查时由后端返回,离线则在对话框内提示联系方式
  const adminOnline = ref(true)
  const welcomeShown = ref(!!localStorage.getItem('chat_welcome_shown'))

  // 轮询兜底定时器：面板关闭时 WS 已断开，靠轮询发现管理员的新回复
  let pollTimer: ReturnType<typeof setInterval> | null = null

  const guestToken = computed(() => {
    let token = localStorage.getItem('chat_guest_token')
    if (!token) {
      token = generateUUID()
      localStorage.setItem('chat_guest_token', token)
    }
    return token
  })

  function isGuest(): boolean {
    return !useAuthStore().isAuthenticated
  }

  function getLastRead(convId: number): number {
    return Number(localStorage.getItem(readKey(convId)) || 0)
  }

  function markRead(convId: number, maxId: number) {
    if (convId > 0 && maxId > getLastRead(convId)) {
      localStorage.setItem(readKey(convId), String(maxId))
    }
  }

  // 记住访客会话 id（仅访客落盘，作为轮询的兜底来源；登录用户由 restore 探查恢复）
  function rememberConversation(id: number) {
    if (isGuest()) localStorage.setItem(CONV_ID_KEY, String(id))
  }

  function maxMessageId(list: ChatMessage[]): number {
    return list.reduce((mx, m) => Math.max(mx, m.id), 0)
  }

  function hasUnreadAdmin(list: ChatMessage[], convId: number): boolean {
    const lastRead = getLastRead(convId)
    return list.some((m) => m.sender_type === 'admin' && m.id > lastRead)
  }

  // 当前要轮询的会话 id：优先内存中的会话，访客回退到持久化的会话 id
  function activeConversationId(): number {
    if (conversation.value) return conversation.value.id
    if (isGuest()) return Number(localStorage.getItem(CONV_ID_KEY) || 0)
    return 0
  }

  // 页面加载时探查既有会话(不创建空会话),恢复 convId 并据此点亮红点。
  // 解决"登录用户刷新后内存无会话→轮询拿不到 convId→红点不亮"的盲区。
  async function restoreConversation() {
    if (conversation.value) return
    let token: string | undefined
    if (isGuest()) {
      token = localStorage.getItem('chat_guest_token') || undefined
      if (!token) return // 从未产生过访客身份 → 必无会话；不为探查而生成 token
    }
    try {
      const result = await peekConversation(token)
      if (!result?.conversation) {
        if (typeof result?.admin_online === 'boolean') adminOnline.value = result.admin_online
        return
      }
      conversation.value = result.conversation
      messages.value = result.messages || []
      adminOnline.value = result.admin_online ?? true
      rememberConversation(result.conversation.id)
      if (!isOpen.value && hasUnreadAdmin(messages.value, result.conversation.id)) {
        hasNewMessage.value = true
      }
    } catch {
      // ignore
    }
  }

  async function openChat() {
    isOpen.value = true
    hasNewMessage.value = false

    // 本会话已在内存中：直接打开，并把已读位推进到最新（含轮询期间合并进来的消息）
    if (conversation.value) {
      markRead(conversation.value.id, maxMessageId(messages.value))
      return
    }
    if (loading.value) return

    loading.value = true
    try {
      const authStore = useAuthStore()
      const token = authStore.isAuthenticated ? undefined : guestToken.value
      const result = await createOrGetConversation(token)
      conversation.value = result.conversation
      messages.value = result.messages || []
      adminOnline.value = result.admin_online ?? true
      rememberConversation(result.conversation.id)
      markRead(result.conversation.id, maxMessageId(messages.value))
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
      markRead(conversation.value.id, msg.id)
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
        // 收到管理员消息说明客服在线,隐藏离线提示
        if (msg.sender_type === 'admin') adminOnline.value = true
        if (!isOpen.value) {
          hasNewMessage.value = true
        } else {
          markRead(conversation.value.id, msg.id)
        }
      }
    }
  }

  // 轮询兜底：面板关闭时检查是否有管理员新回复，有则点亮红点
  async function pollNewMessages() {
    if (isOpen.value) return // 打开时由 WS 实时接收，无需轮询
    const convId = activeConversationId()
    if (!convId) return
    try {
      const token = isGuest() ? guestToken.value : undefined
      const data = await getMessages(convId, 50, 0, token)
      const list = data?.messages || []
      if (!list.length) return

      const unread = hasUnreadAdmin(list, convId)

      // 若内存中正是该会话，顺带合并新消息，便于打开面板时立即看到最新
      if (conversation.value && conversation.value.id === convId) {
        for (const m of list) {
          if (!messages.value.some((x) => x.id === m.id)) messages.value.push(m)
        }
      }

      if (unread) {
        // 有管理员回复说明客服可达，隐藏离线提示（与 WS 路径一致）
        adminOnline.value = true
        hasNewMessage.value = true
      } else {
        // 没有未读的管理员回复（可能只是自己的消息），推进已读位
        markRead(convId, maxMessageId(list))
      }
    } catch {
      // 静默失败，下次轮询自动重试
    }
  }

  function onVisibilityPoll() {
    if (document.visibilityState === 'visible') pollNewMessages()
  }

  function startPolling(intervalMs = 15000) {
    if (pollTimer) return
    pollNewMessages()
    pollTimer = setInterval(() => {
      if (document.visibilityState === 'visible') pollNewMessages()
    }, intervalMs)
    document.addEventListener('visibilitychange', onVisibilityPoll)
  }

  function stopPolling() {
    if (pollTimer) {
      clearInterval(pollTimer)
      pollTimer = null
    }
    document.removeEventListener('visibilitychange', onVisibilityPoll)
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
    adminOnline.value = true
  }

  return {
    isOpen,
    conversation,
    messages,
    loading,
    sending,
    hasNewMessage,
    adminOnline,
    welcomeShown,
    guestToken,
    openChat,
    closeChat,
    send,
    receiveMessage,
    restoreConversation,
    startPolling,
    stopPolling,
    pollNewMessages,
    markWelcomeShown,
    reset,
  }
})
