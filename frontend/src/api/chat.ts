import { apiClient } from './client'
import type { ChatConversation, ChatMessage } from '@/types'

export async function createOrGetConversation(guestToken?: string): Promise<{
  conversation: ChatConversation
  messages: ChatMessage[]
}> {
  const { data } = await apiClient.post('/chat/conversations', { guest_token: guestToken })
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
  offset = 0,
  guestToken?: string
): Promise<{ messages: ChatMessage[]; total: number }> {
  const { data } = await apiClient.get(`/chat/conversations/${conversationId}/messages`, {
    params: { limit, offset, guest_token: guestToken },
  })
  return data
}

export default { createOrGetConversation, sendMessage, getMessages }
