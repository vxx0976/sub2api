import { apiClient } from './client'
import type { ChatConversation, ChatMessage } from '@/types'

export async function createOrGetConversation(guestToken?: string): Promise<{
  conversation: ChatConversation
  messages: ChatMessage[]
  admin_online?: boolean
}> {
  const { data } = await apiClient.post('/chat/conversations', { guest_token: guestToken })
  return data
}

// 探查访客当前的开启会话(不创建),用于页面加载时判断是否有未读管理员回复
export async function peekConversation(guestToken?: string): Promise<{
  conversation: ChatConversation | null
  messages: ChatMessage[]
  admin_online?: boolean
}> {
  const { data } = await apiClient.get('/chat/conversation', {
    params: { guest_token: guestToken },
  })
  return data
}

export async function sendMessage(
  conversationId: number,
  content: string,
  guestToken?: string
): Promise<ChatMessage> {
  const { data } = await apiClient.post(`/chat/conversations/${conversationId}/messages`, {
    content,
    guest_token: guestToken,
  })
  return data
}

export async function getMessages(
  conversationId: number,
  limit = 50,
  offset?: number,
  guestToken?: string
): Promise<{ messages: ChatMessage[]; total: number }> {
  const { data } = await apiClient.get(`/chat/conversations/${conversationId}/messages`, {
    params: { limit, guest_token: guestToken, ...(offset === undefined ? {} : { offset }) },
  })
  return data
}

export default { createOrGetConversation, peekConversation, sendMessage, getMessages }
